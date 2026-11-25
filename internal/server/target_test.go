package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// advanceOrigin pousse un commit sur le main du remote depuis un autre clone :
// le dépôt du projet se retrouve en retard, comme un clone qu'on n'a pas tiré.
func advanceOrigin(t *testing.T, f *deliveryFixture) string {
	t.Helper()
	runTestGit(t, f.repo, "push", "-q", "origin", "main")
	other := t.TempDir()
	runTestGit(t, other, "clone", "-q", f.bare, ".")
	runTestGit(t, other, "config", "user.email", "test@example.com")
	runTestGit(t, other, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(other, "upstream.txt"), []byte("upstream\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, other, "add", "-A")
	runTestGit(t, other, "commit", "-q", "-m", "feat: upstream work")
	runTestGit(t, other, "push", "-q", "origin", "main")
	return strings.TrimSpace(runTestGit(t, other, "rev-parse", "HEAD"))
}

func targetRepos(t *testing.T, w *httptest.ResponseRecorder) []TargetStatus {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var res struct{ Repos []TargetStatus }
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	return res.Repos
}

// Le cas qui a produit 135 fichiers dans une tâche de trois lignes : un main
// local en retard. L'accueil le voit après fetch, le chantier part déjà
// d'origin/main, et la mise à jour avance main en fast-forward.
func TestStaleTargetIsDetectedAvoidedAndFastForwarded(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	upstream := advanceOrigin(t, f)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetPathValue("id", f.project.ID)
	w := httptest.NewRecorder()
	f.srv.handleTargetStatus(w, req)
	repos := targetRepos(t, w)
	if len(repos) != 1 || repos[0].Behind != 1 || repos[0].Ahead != 0 || repos[0].Error != "" {
		t.Fatalf("retard attendu d'un commit après fetch, reçu %+v", repos)
	}

	// Le chantier part d'origin/main, pas du main local en retard, et sa
	// destination reste « main ».
	cb, err := f.srv.ensureCardBranch(f.card, f.project, f.project.Repos[0])
	if err != nil {
		t.Fatal(err)
	}
	if cb.Base != "main" {
		t.Fatalf("la destination du chantier doit rester main, reçu %q", cb.Base)
	}
	if got := HeadSha(cb.WorktreeDir); got != upstream {
		t.Fatalf("le chantier devrait partir d'origin/main (%s), reçu %s", upstream, got)
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetPathValue("id", f.project.ID)
	w = httptest.NewRecorder()
	f.srv.handleTargetUpdate(w, req)
	if repos := targetRepos(t, w); repos[0].Behind != 0 || repos[0].Error != "" {
		t.Fatalf("main devrait être à jour, reçu %+v", repos)
	}
	if got := HeadSha(f.repo); got != upstream {
		t.Fatalf("main extrait devrait avancer jusqu'à %s, reçu %s", upstream, got)
	}
}

// Un main local avec ses propres commits n'est jamais réécrit : ni base de
// chantier prise sur origin, ni fast-forward.
func TestDivergedTargetIsNeverRewritten(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	advanceOrigin(t, f)
	if err := os.WriteFile(filepath.Join(f.repo, "local.txt"), []byte("local\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTestGit(t, f.repo, "add", "-A")
	runTestGit(t, f.repo, "commit", "-q", "-m", "feat: local work")
	local := HeadSha(f.repo)
	if err := FetchTarget(f.repo, "main"); err != nil {
		t.Fatal(err)
	}
	if behind, ahead, ok := TargetDrift(f.repo, "main"); !ok || behind != 1 || ahead != 1 {
		t.Fatalf("divergence attendue (1/1), reçu behind=%d ahead=%d ok=%v", behind, ahead, ok)
	}
	if got := freshestRef(f.repo, "main"); got != "main" {
		t.Fatalf("un main divergent doit rester la base, reçu %q", got)
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetPathValue("id", f.project.ID)
	w := httptest.NewRecorder()
	f.srv.handleTargetUpdate(w, req)
	if repos := targetRepos(t, w); repos[0].Error == "" {
		t.Fatalf("la mise à jour d'un main divergent doit échouer, reçu %+v", repos)
	}
	if HeadSha(f.repo) != local {
		t.Fatalf("main local réécrit")
	}
}

// Branche de destination non extraite : la ref avance sans toucher au
// répertoire de travail.
func TestFastForwardTargetNotCheckedOut(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	upstream := advanceOrigin(t, f)
	runTestGit(t, f.repo, "checkout", "-q", "-b", "elsewhere")
	if err := FetchTarget(f.repo, "main"); err != nil {
		t.Fatal(err)
	}
	if err := FastForwardTarget(f.repo, "main"); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(runTestGit(t, f.repo, "rev-parse", "main")); got != upstream {
		t.Fatalf("main devrait pointer sur %s, reçu %s", upstream, got)
	}
}

// Un agent qui change de branche dans son worktree laisse la branche de la
// tâche vide : l'acceptation est refusée au lieu de « réussir » sans rien
// fusionner.
func TestAcceptRefusesWorktreeOffTaskBranch(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	id, _ := f.addTask(t, "Ailleurs", "a.txt", "a\n")
	task, _ := f.srv.store.GetTask(id)
	runTestGit(t, task.WorktreeDir, "checkout", "-q", "-b", "side-work")

	if w := f.accept(t, id); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("attendu 422, reçu %d (%s)", w.Code, w.Body.String())
	}
	if task, _ := f.srv.store.GetTask(id); task.Status != "review" {
		t.Fatalf("la tâche doit rester en revue, reçu %q", task.Status)
	}
}

// Les compteurs stockés suivent le diff affiché, même quand la branche a
// bougé hors d'une exécution d'agent.
func TestDiffRefreshesStaleWorkCounts(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	id, _ := f.addTask(t, "Compteurs", "a.txt", "a\n")
	if _, err := f.srv.store.UpdateTask(id, func(tk *Task) { tk.FilesCount, tk.CommitsCount = 135, 121 }); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetPathValue("id", id)
	w := httptest.NewRecorder()
	f.srv.handleDiff(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("diff : attendu 200, reçu %d", w.Code)
	}
	task, _ := f.srv.store.GetTask(id)
	if want, _, commits := TaskWorkCounts(task); task.FilesCount != want || task.CommitsCount != commits || task.FilesCount == 135 {
		t.Fatalf("compteurs non rafraîchis : files=%d commits=%d", task.FilesCount, task.CommitsCount)
	}
}

// Une tâche acceptée garde son diff : sa base devient le chantier d'avant son
// merge commit, pas la pointe du chantier qui contient déjà son travail.
func TestAcceptedTaskKeepsItsDiff(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	id, _ := f.addTask(t, "Acceptée", "a.txt", "a\n")
	if w := f.accept(t, id); w.Code != http.StatusOK {
		t.Fatalf("accept : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	task, _ := f.srv.store.GetTask(id)
	files, err := Diff(task.WorktreeDir, TaskDiffBase(task))
	if err != nil || len(files) != 1 || files[0].Path != "a.txt" {
		t.Fatalf("diff de la tâche acceptée attendu sur a.txt, reçu %+v (%v)", files, err)
	}
	if n, _, commits := TaskWorkCounts(task); n != 1 || commits != 1 {
		t.Fatalf("compteurs attendus 1 fichier, 1 commit, reçu %d, %d", n, commits)
	}
}

// Un chantier parti d'origin/main ne compte pas l'écart avec le main local en
// retard comme son propre travail.
func TestDeliveryPreviewIgnoresStaleLocalTarget(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	advanceOrigin(t, f)
	if err := FetchTarget(f.repo, "main"); err != nil {
		t.Fatal(err)
	}
	id, _ := f.addTask(t, "Propre", "b.txt", "b\n")
	if w := f.accept(t, id); w.Code != http.StatusOK {
		t.Fatalf("accept : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	prev := f.delivery(t)
	if row := prev.Repos[0]; row.Files != 1 {
		t.Fatalf("un seul fichier attendu dans l'aperçu, reçu %+v", row)
	}
}

// Le corps d'une PR liste les commits du chantier, sans merge commit ni
// signature de l'outil.
func TestPRBodyListsWorkstreamCommits(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	id, cb := f.addTask(t, "Corps", "c.txt", "c\n")
	task, _ := f.srv.store.GetTask(id)
	runTestGit(t, task.WorktreeDir, "add", "-A")
	runTestGit(t, task.WorktreeDir, "commit", "-q", "-m", "feat: add c")
	if w := f.accept(t, id); w.Code != http.StatusOK {
		t.Fatalf("accept : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	if got := prBody(cb.WorktreeDir, "main", cb.Branch); got != "- feat: add c" {
		t.Fatalf("corps de PR inattendu : %q", got)
	}
}

// Accepter ne crée aucun merge commit : la branche du chantier, donc la pull
// request, ne porte que les commits des agents, même quand une seconde tâche
// partie avant la première doit être rejouée.
func TestAcceptKeepsWorkstreamHistoryLinear(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	first, cb := f.addTask(t, "Un", "a.txt", "a\n")
	second, _ := f.addTask(t, "Deux", "b.txt", "b\n")
	for _, id := range []string{first, second} {
		task, _ := f.srv.store.GetTask(id)
		runTestGit(t, task.WorktreeDir, "add", "-A")
		runTestGit(t, task.WorktreeDir, "commit", "-q", "-m", "feat: "+id)
	}
	task, _ := f.srv.store.GetTask(second)
	before := HeadSha(task.WorktreeDir)
	if w := f.accept(t, first); w.Code != http.StatusOK {
		t.Fatalf("accept : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	// Le rebase de fond a rejoué la seconde : on la ramène sur sa base
	// d'origine pour que l'acceptation ait à la rejouer elle-même.
	f.srv.rebaseWG.Wait()
	runTestGit(t, task.WorktreeDir, "reset", "-q", "--hard", before)
	if w := f.accept(t, second); w.Code != http.StatusOK {
		t.Fatalf("accept : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	if merges := strings.TrimSpace(runTestGit(t, cb.WorktreeDir, "rev-list", "--merges", "main.."+cb.Branch)); merges != "" {
		t.Fatalf("aucun merge commit attendu, reçu %s", merges)
	}
	if got := prBody(cb.WorktreeDir, "main", cb.Branch); !strings.Contains(got, "feat: "+first) || !strings.Contains(got, "feat: "+second) {
		t.Fatalf("corps de PR inattendu : %q", got)
	}
	for id, path := range map[string]string{first: "a.txt", second: "b.txt"} {
		task, _ := f.srv.store.GetTask(id)
		files, err := Diff(task.WorktreeDir, TaskDiffBase(task))
		if err != nil || len(files) != 1 || files[0].Path != path {
			t.Fatalf("diff de %s attendu sur %s seul, reçu %+v (%v)", id, path, files, err)
		}
	}
}

// Le cas d'un chantier sans rien d'accepté, 127 commits derrière origin/main,
// dont une tâche avait été rebasée à la main sur origin/main : elle montrait
// 137 fichiers pour un commit. Le rattrapage avance le chantier en
// fast-forward jusqu'à origin/main (pas le main local en retard), puis rejoue
// les tâches en revue : chacune ne montre plus que son propre travail.
func TestCatchUpFastForwardsEmptyWorkstreamToOrigin(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	first, cb := f.addTask(t, "Un", "a.txt", "a\n")
	second, _ := f.addTask(t, "Deux", "b.txt", "b\n")
	for _, id := range []string{first, second} {
		task, _ := f.srv.store.GetTask(id)
		runTestGit(t, task.WorktreeDir, "add", "-A")
		runTestGit(t, task.WorktreeDir, "commit", "-q", "-m", "feat: "+id)
	}
	upstream := advanceOrigin(t, f)
	if err := FetchTarget(f.repo, "main"); err != nil {
		t.Fatal(err)
	}
	secondTask, _ := f.srv.store.GetTask(second)
	runTestGit(t, secondTask.WorktreeDir, "rebase", "-q", "origin/main")
	if files, _ := Diff(secondTask.WorktreeDir, TaskDiffBase(secondTask)); len(files) != 2 {
		t.Fatalf("avant rattrapage, la tâche rebasée à la main compte la destination : 2 fichiers attendus, reçu %+v", files)
	}
	if prev := f.delivery(t); prev.Repos[0].Behind != 1 {
		t.Fatalf("un commit de retard attendu, reçu %+v", prev.Repos[0])
	}

	resp := f.catchUp(t)
	if len(resp.Repos) != 1 || !resp.Repos[0].Merged || resp.Repos[0].Error != "" {
		t.Fatalf("rattrapage inattendu : %+v", resp.Repos)
	}
	if got := HeadSha(cb.WorktreeDir); got != upstream {
		t.Fatalf("le chantier devrait avancer en fast-forward jusqu'à origin/main (%s), reçu %s", upstream, got)
	}
	if prev := f.delivery(t); prev.Repos[0].Behind != 0 {
		t.Fatalf("plus aucun retard attendu, reçu %+v", prev.Repos[0])
	}
	f.srv.waitRebases()
	for id, path := range map[string]string{first: "a.txt", second: "b.txt"} {
		task, _ := f.srv.store.GetTask(id)
		if !IsBranchMergedInto(task.WorktreeDir, cb.Branch, task.Branch) {
			t.Fatalf("%s devrait repartir du chantier rattrapé", id)
		}
		files, err := Diff(task.WorktreeDir, TaskDiffBase(task))
		if err != nil || len(files) != 1 || files[0].Path != path {
			t.Fatalf("diff de %s attendu sur %s seul, reçu %+v (%v)", id, path, files, err)
		}
	}
	// La tâche rebasée à la main n'a pas bougé, mais sa base si : ses compteurs
	// (l'onglet « Diff 137 ») doivent suivre sans attendre l'ouverture du diff.
	if task, _ := f.srv.store.GetTask(second); task.FilesCount != 1 || task.CommitsCount != 1 {
		t.Fatalf("compteurs de la tâche rebasée à la main : 1 fichier et 1 commit attendus, reçu %d et %d", task.FilesCount, task.CommitsCount)
	}
}
