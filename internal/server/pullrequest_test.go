package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeForge remplace gh le temps d'un test : la forge est GitHub, les
// recherches et créations de PR sont scriptées.
type fakeForge struct {
	found    *PullRequest
	viewed   map[int]PullRequest
	openErr  error
	openURL  string
	openCall int
	finds    int
}

func (ff *fakeForge) install(t *testing.T) {
	t.Helper()
	orig := forgePR
	t.Cleanup(func() { forgePR = orig })
	forgePR = prForge{
		provider: func(string) string { return "github" },
		find: func(_ context.Context, _, branch string) (PullRequest, bool, error) {
			ff.finds++
			if ff.found != nil && ff.found.HeadRefName == branch {
				return *ff.found, true, nil
			}
			return PullRequest{}, false, nil
		},
		view: func(_ context.Context, _ string, n int) (PullRequest, error) {
			pr, ok := ff.viewed[n]
			if !ok {
				return PullRequest{}, errors.New("no pull request found")
			}
			return pr, nil
		},
		open: func(_, branch, base, _, _ string) (string, string, error) {
			ff.openCall++
			fallback := "https://github.com/o/r/compare/" + base + "..." + branch + "?expand=1"
			if ff.openErr != nil {
				return "", fallback, ff.openErr
			}
			return ff.openURL, "", nil
		},
	}
}

func decodeShip(t *testing.T, w interface{ Bytes() []byte }) ShipResponse {
	t.Helper()
	var resp ShipResponse
	if err := json.Unmarshal(w.Bytes(), &resp); err != nil {
		t.Fatalf("réponse de livraison illisible : %v", err)
	}
	return resp
}

// TestShipRetriesPROpeningAfterFailure : gh échoue à ouvrir la PR. La branche
// est poussée, mais la livraison le dit en erreur, joint la page de compare
// sans l'enregistrer comme PR, et laisse le chantier à livrer. Relivrer ouvre
// la PR.
func TestShipRetriesPROpeningAfterFailure(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	ff := &fakeForge{openErr: errors.New("gh pr create failed: Base ref must be a branch")}
	ff.install(t)

	taskID, cb := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	w := f.ship(t, `{"confirm":true}`)
	if w.Code != http.StatusOK {
		t.Fatalf("ship: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	r := decodeShip(t, w.Body).Repos[0]
	if !r.Pushed || r.PrURL != "" || r.CompareURL == "" || !strings.Contains(r.Error, "Base ref must be a branch") {
		t.Fatalf("échec d'ouverture attendu, poussé, compare joint : %+v", r)
	}
	card, _ := f.srv.store.GetCard(f.card.ID)
	if card.Branches[0].PrURL != "" || card.Branches[0].ShippedAt != nil {
		t.Fatalf("ni PR ni livraison ne doivent être enregistrées : %+v", card.Branches[0])
	}
	if prev := f.delivery(t); prev.Repos[0].Pending == 0 || prev.Repos[0].PrTitle == "" {
		t.Fatalf("la PR reste à ouvrir, l'aperçu doit l'annoncer : %+v", prev.Repos[0])
	}

	ff.openErr, ff.openURL = nil, "https://github.com/o/r/pull/12"
	r = decodeShip(t, f.ship(t, `{"confirm":true}`).Body).Repos[0]
	if r.Error != "" || r.PrURL != "https://github.com/o/r/pull/12" || r.Skipped {
		t.Fatalf("la relivraison doit ouvrir la PR : %+v", r)
	}
	card, _ = f.srv.store.GetCard(f.card.ID)
	if b := card.Branches[0]; b.PrURL != r.PrURL || b.PrState != "open" || b.ShippedAt == nil {
		t.Fatalf("PR et livraison attendues : %+v", b)
	}
	if ff.openCall != 2 {
		t.Fatalf("deux tentatives d'ouverture attendues, reçu %d", ff.openCall)
	}
	_ = cb
}

// TestManualPRIsLinkedAndReused : une PR ouverte à la main pour la branche du
// chantier est retrouvée par l'aperçu, et la livraison la met à jour par le
// push au lieu d'en ouvrir une seconde.
func TestManualPRIsLinkedAndReused(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	ff := &fakeForge{}
	ff.install(t)

	taskID, cb := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	ff.found = &PullRequest{Number: 9, URL: "https://github.com/o/r/pull/9", State: "OPEN", HeadRefName: cb.Branch, BaseRefName: "main"}

	prev := f.delivery(t)
	if prev.Repos[0].PrURL != ff.found.URL || prev.Repos[0].PrState != "open" || prev.Repos[0].PrTitle != "" {
		t.Fatalf("la PR ouverte à la main doit être rattachée : %+v", prev.Repos[0])
	}
	r := decodeShip(t, f.ship(t, `{"confirm":true}`).Body).Repos[0]
	if r.Error != "" || !r.Pushed || r.PrURL != ff.found.URL {
		t.Fatalf("livraison sur la PR existante attendue : %+v", r)
	}
	if ff.openCall != 0 {
		t.Fatalf("aucune PR ne doit être ouverte, reçu %d ouvertures", ff.openCall)
	}
}

// TestClosedPRIsNotReused : une PR fermée sans fusion n'est pas réutilisée,
// la livraison suivante en ouvre une nouvelle.
func TestClosedPRIsNotReused(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	ff := &fakeForge{openURL: "https://github.com/o/r/pull/3"}
	ff.install(t)

	taskID, _ := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	if _, err := f.srv.store.SetCardBranchPR(f.card.ID, "demo", "https://github.com/o/r/pull/2", "open"); err != nil {
		t.Fatal(err)
	}
	// gh ne trouve plus de PR vivante pour la branche : elle a été fermée.
	r := decodeShip(t, f.ship(t, `{"confirm":true}`).Body).Repos[0]
	if r.Error != "" || r.PrURL != "https://github.com/o/r/pull/3" || ff.openCall != 1 {
		t.Fatalf("une nouvelle PR doit être ouverte : %+v (ouvertures %d)", r, ff.openCall)
	}
}

func TestMigrateComparePrURLs(t *testing.T) {
	s := &Store{Cards: map[string]Card{"c": {ID: "c", Branches: []CardBranch{
		{RepoName: "a", PrURL: "https://github.com/o/r/compare/main...x?expand=1"},
		{RepoName: "b", PrURL: "https://github.com/o/r/pull/4", PrState: "merged"},
		{RepoName: "c", PrURL: "https://gitlab.com/g/p/-/merge_requests/7"},
	}}}}
	migrateComparePrURLs(s)
	got := s.Cards["c"].Branches
	if got[0].PrURL != "" || got[1].PrURL == "" || got[2].PrURL == "" {
		t.Fatalf("seule l'URL de compare doit disparaître : %+v", got)
	}
}

func TestPickPR(t *testing.T) {
	prs := []PullRequest{
		{Number: 5, State: "CLOSED", HeadRefName: "x"},
		{Number: 4, State: "OPEN", HeadRefName: "x", IsCrossRepository: true},
		{Number: 3, State: "MERGED", HeadRefName: "x"},
		{Number: 2, State: "MERGED", HeadRefName: "x"},
	}
	pr, ok, _ := pickPR(prs, "x")
	if !ok || pr.Number != 3 {
		t.Fatalf("la fusionnée la plus récente du dépôt même est attendue, reçu %+v", pr)
	}
	pr, ok, _ = pickPR(append([]PullRequest{{Number: 6, State: "OPEN", HeadRefName: "x"}}, prs...), "x")
	if !ok || pr.Number != 6 {
		t.Fatalf("une PR ouverte l'emporte, reçu %+v", pr)
	}
	if _, ok, _ := pickPR(prs[:2], "x"); ok {
		t.Fatal("ni une PR fermée ni une PR de fork ne comptent")
	}
}

// TestCardBaseUsesOriginDefaultBranch : sans destination réglée, un chantier en
// mode "pr" part de la branche par défaut d'origin, pas de la branche qui se
// trouve extraite (souvent une branche locale, base impossible pour une PR).
// Les modes de fusion gardent la branche courante.
func TestCardBaseUsesOriginDefaultBranch(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	runTestGit(t, f.repo, "push", "-q", "origin", "main")
	runTestGit(t, f.repo, "fetch", "-q", "origin")
	runTestGit(t, f.repo, "remote", "set-head", "origin", "main")
	runTestGit(t, f.repo, "checkout", "-q", "-b", "docs/local-only")

	repo := f.project.Repos[0]
	if base, err := cardBase(f.project, repo); err != nil || base != "main" {
		t.Fatalf("mode pr : base main attendue, reçu %q (%v)", base, err)
	}
	merge := f.project
	merge.Delivery = Delivery{Mode: "merge"}
	if base, err := cardBase(merge, repo); err != nil || base != "docs/local-only" {
		t.Fatalf("mode merge : branche courante attendue, reçu %q (%v)", base, err)
	}
	cb, err := f.srv.ensureCardBranch(f.card, f.project, repo)
	if err != nil || cb.Base != "main" {
		t.Fatalf("branche de chantier sur main attendue, reçu %+v (%v)", cb, err)
	}
}

// TestAdoptPRBranch : un chantier attaché à une PR ouverte travaille sur sa
// branche de tête, et sa livraison pousse sur elle sans ouvrir de PR.
func TestAdoptPRBranch(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	runTestGit(t, f.repo, "push", "-q", "origin", "main")
	// La branche de la PR existe sur origin, poussée par quelqu'un d'autre.
	other := t.TempDir()
	runTestGit(t, other, "clone", "-q", f.bare, other)
	runTestGit(t, other, "config", "user.email", "test@example.com")
	runTestGit(t, other, "config", "user.name", "Test")
	runTestGit(t, other, "checkout", "-q", "-b", "feat/from-someone")
	if err := os.WriteFile(filepath.Join(other, "pr.txt"), []byte("pr\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, other, "add", "-A")
	runTestGit(t, other, "commit", "-q", "-m", "work in the PR")
	runTestGit(t, other, "push", "-q", "origin", "feat/from-someone")

	pr := PullRequest{Number: 42, URL: "https://github.com/o/r/pull/42", State: "OPEN", HeadRefName: "feat/from-someone", BaseRefName: "main"}
	ff := &fakeForge{viewed: map[int]PullRequest{
		42: pr,
		43: {Number: 43, State: "OPEN", HeadRefName: "main", BaseRefName: "main", IsCrossRepository: true},
		44: {Number: 44, State: "OPEN", HeadRefName: "--upload-pack=x", BaseRefName: "main"},
	}}
	ff.install(t)
	repo := f.project.Repos[0]

	if _, err := f.srv.ensureCardBranch(f.card, f.project, repo, 43); err == nil || !strings.Contains(err.Error(), "fork") {
		t.Fatalf("une PR de fork doit être refusée, reçu %v", err)
	}
	if _, err := f.srv.ensureCardBranch(f.card, f.project, repo, 44); err == nil {
		t.Fatal("un nom de branche qui ressemble à une option doit être refusé")
	}
	cb, err := f.srv.ensureCardBranch(f.card, f.project, repo, 42)
	if err != nil {
		t.Fatalf("adoption : %v", err)
	}
	if cb.Branch != "feat/from-someone" || cb.Base != "main" || !cb.Adopted || cb.PrURL != pr.URL {
		t.Fatalf("branche adoptée inattendue : %+v", cb)
	}
	if _, err := os.Stat(filepath.Join(cb.WorktreeDir, "pr.txt")); err != nil {
		t.Fatalf("le worktree doit porter le travail de la PR : %v", err)
	}
	// Redemander la même PR est sans effet ; une autre est refusée.
	if again, err := f.srv.ensureCardBranch(f.card, f.project, repo, 42); err != nil || again.Branch != cb.Branch {
		t.Fatalf("même PR : branche existante attendue, reçu %+v (%v)", again, err)
	}

	// Une tâche du chantier, acceptée, puis livrée : le push met la PR à jour.
	ff.found = &pr
	taskID, _ := f.addTask(t, "Corrige la PR", "fix.txt", "fix\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	r := decodeShip(t, f.ship(t, `{"confirm":true}`).Body).Repos[0]
	if r.Error != "" || !r.Pushed || r.PrURL != pr.URL || ff.openCall != 0 {
		t.Fatalf("livraison sur la PR adoptée attendue : %+v (ouvertures %d)", r, ff.openCall)
	}
	if tree := runTestGit(t, f.bare, "ls-tree", "--name-only", "feat/from-someone"); !strings.Contains(tree, "fix.txt") {
		t.Fatalf("le correctif doit être poussé sur la branche de la PR, reçu %q", tree)
	}
}

// TestMergedPRIsNotShippedAgain : une PR fusionnée sur la forge (même dans une
// autre branche que la base enregistrée) clôt la livraison : rien n'est
// annoncé à livrer tant qu'aucun commit nouveau n'arrive.
func TestMergedPRIsNotShippedAgain(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	ff := &fakeForge{openErr: errors.New("gh pr create failed")}
	ff.install(t)
	taskID, cb := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	_ = f.ship(t, `{"confirm":true}`) // poussée, PR non ouverte
	ff.found = &PullRequest{Number: 16, URL: "https://github.com/o/r/pull/16", State: "MERGED", HeadRefName: cb.Branch, BaseRefName: "main"}
	f.srv.prLookups = prLookups{}
	prev := f.delivery(t)
	if r := prev.Repos[0]; r.PrState != "merged" || r.Pending != 0 || r.PrTitle != "" {
		t.Fatalf("PR fusionnée : rien à livrer attendu, reçu %+v", r)
	}
	r := decodeShip(t, f.ship(t, `{"confirm":true}`).Body).Repos[0]
	if !r.Skipped || ff.openCall != 1 {
		t.Fatalf("aucune nouvelle tentative attendue : %+v (ouvertures %d)", r, ff.openCall)
	}
}
