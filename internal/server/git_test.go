package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// diffFixture est un diff unifié inline couvrant : un fichier modifié
// (contexte, ajout, suppression) et un nouveau fichier.
const diffFixture = `diff --git a/main.go b/main.go
index 1111111..2222222 100644
--- a/main.go
+++ b/main.go
@@ -1,4 +1,5 @@
 package main
 
-func old() {}
+func new() {}
+func extra() {}
diff --git a/newfile.txt b/newfile.txt
new file mode 100644
index 0000000..3333333
--- /dev/null
+++ b/newfile.txt
@@ -0,0 +1,2 @@
+hello
+world
`

func TestParseDiffFixture(t *testing.T) {
	files := parseDiff(diffFixture)
	if len(files) != 2 {
		t.Fatalf("attendu 2 fichiers, reçu %d (%+v)", len(files), files)
	}

	main := files[0]
	if main.Path != "main.go" {
		t.Fatalf("chemin attendu 'main.go', reçu %q", main.Path)
	}
	if main.Additions != 2 || main.Deletions != 1 {
		t.Fatalf("compteurs inattendus pour main.go : +%d -%d", main.Additions, main.Deletions)
	}
	if len(main.Hunks) != 1 {
		t.Fatalf("attendu 1 hunk pour main.go, reçu %d", len(main.Hunks))
	}
	hunk := main.Hunks[0]
	if hunk.Header != "@@ -1,4 +1,5 @@" {
		t.Fatalf("en-tête de hunk inattendu : %q", hunk.Header)
	}
	wantTypes := []string{"ctx", "ctx", "del", "add", "add"}
	if len(hunk.Lines) != len(wantTypes) {
		t.Fatalf("attendu %d lignes, reçu %d (%+v)", len(wantTypes), len(hunk.Lines), hunk.Lines)
	}
	for i, want := range wantTypes {
		if hunk.Lines[i].Type != want {
			t.Errorf("ligne %d : type attendu %q, reçu %q", i, want, hunk.Lines[i].Type)
		}
	}

	newFile := files[1]
	if newFile.Path != "newfile.txt" {
		t.Fatalf("chemin attendu 'newfile.txt', reçu %q", newFile.Path)
	}
	if newFile.Additions != 2 || newFile.Deletions != 0 {
		t.Fatalf("compteurs inattendus pour newfile.txt : +%d -%d", newFile.Additions, newFile.Deletions)
	}
}

// runTestGit exécute une commande git pour la mise en place du test (pas de
// dépendance au code testé).
func runTestGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s a échoué : %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func TestWorktreeAndDiff(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}

	repo := t.TempDir()
	dataDir := t.TempDir()

	runTestGit(t, repo, "init")
	runTestGit(t, repo, "config", "user.email", "test@example.com")
	runTestGit(t, repo, "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# projet\n"), 0o644); err != nil {
		t.Fatalf("écriture README impossible : %v", err)
	}
	runTestGit(t, repo, "add", "-A")
	runTestGit(t, repo, "commit", "-m", "initial")

	wantBase := strings.TrimSpace(runTestGit(t, repo, "rev-parse", "--abbrev-ref", "HEAD"))

	dir, base, err := CreateWorktree(repo, dataDir, "t1", "sillage/100-test", "")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	if base != wantBase {
		t.Fatalf("base attendue %q, reçue %q", wantBase, base)
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		t.Fatalf("le worktree %s devrait exister : %v", dir, err)
	}

	// Modifie un fichier suivi et ajoute un fichier non suivi.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# projet\n\nModifié.\n"), 0o644); err != nil {
		t.Fatalf("modification README impossible : %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nouveau.txt"), []byte("contenu\n"), 0o644); err != nil {
		t.Fatalf("création nouveau.txt impossible : %v", err)
	}

	files, err := Diff(dir, base)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("attendu 2 fichiers modifiés, reçu %d (%+v)", len(files), files)
	}

	var paths []string
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	if !contains(paths, "README.md") || !contains(paths, "nouveau.txt") {
		t.Fatalf("chemins inattendus : %v", paths)
	}
}

// TestGitCommonDirFromWorktree : le worktree d'une tâche n'a qu'un fichier
// `.git` pointeur ; c'est le dossier git du dépôt d'origine qu'un agent en bac
// à sable doit voir pour que ses commandes git fonctionnent.
func TestGitCommonDirFromWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}

	repo := t.TempDir()
	dataDir := t.TempDir()
	runTestGit(t, repo, "init")
	runTestGit(t, repo, "config", "user.email", "test@example.com")
	runTestGit(t, repo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# projet\n"), 0o644); err != nil {
		t.Fatalf("écriture README impossible : %v", err)
	}
	runTestGit(t, repo, "add", "-A")
	runTestGit(t, repo, "commit", "-m", "initial")

	dir, _, err := CreateWorktree(repo, dataDir, "t1", "sillage/100-test", "")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	common, err := GitCommonDir(dir)
	if err != nil {
		t.Fatalf("GitCommonDir: %v", err)
	}
	if !filepath.IsAbs(common) {
		t.Fatalf("le dossier git commun doit être absolu, reçu %q", common)
	}
	// Le chemin doit désigner le .git du dépôt d'origine, pas celui du worktree.
	want, err := filepath.EvalSymlinks(filepath.Join(repo, ".git"))
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	got, err := filepath.EvalSymlinks(common)
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	if got != want {
		t.Fatalf("dossier git commun attendu %q, reçu %q", want, got)
	}
	if _, err := GitCommonDir(""); err == nil {
		t.Fatal("un répertoire vide doit être une erreur")
	}
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

// TestParseGithubRemote couvre le fallback PR : parse d'URL de remote github.com
// aux formats ssh (scp-like) et https, avec ou sans suffixe ".git".
func TestParseGithubRemote(t *testing.T) {
	cases := []struct {
		in                  string
		wantOwner, wantRepo string
		wantOk              bool
	}{
		{"git@github.com:Halleck45/sillage.git", "Halleck45", "sillage", true},
		{"git@github.com:Halleck45/sillage", "Halleck45", "sillage", true},
		{"https://github.com/Halleck45/sillage.git", "Halleck45", "sillage", true},
		{"https://github.com/Halleck45/sillage", "Halleck45", "sillage", true},
		{"https://github.com/Halleck45/sillage/", "Halleck45", "sillage", true},
		{"  https://github.com/Halleck45/sillage.git  ", "Halleck45", "sillage", true},
		{"https://gitlab.com/foo/bar.git", "", "", false},
		{"git@gitlab.com:foo/bar.git", "", "", false},
		{"", "", "", false},
	}
	for _, c := range cases {
		owner, repo, ok := ParseGithubRemote(c.in)
		if ok != c.wantOk || owner != c.wantOwner || repo != c.wantRepo {
			t.Errorf("ParseGithubRemote(%q) = (%q, %q, %v), attendu (%q, %q, %v)",
				c.in, owner, repo, ok, c.wantOwner, c.wantRepo, c.wantOk)
		}
	}
}

// TestGithubBranchURL couvre le calcul de branchUrl fourni par la réponse de
// ship : lecture seule (git remote get-url), jamais de commande réseau.
func TestGithubBranchURL(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	cases := []struct {
		remote string
		want   string
	}{
		{"git@github.com:acme/demo.git", "https://github.com/acme/demo/tree/sillage/100-demo"},
		{"https://github.com/acme/demo.git", "https://github.com/acme/demo/tree/sillage/100-demo"},
		{"git@gitlab.com:acme/demo.git", ""},
		{"", ""},
	}
	for _, c := range cases {
		dir := t.TempDir()
		runTestGit(t, dir, "init")
		if c.remote != "" {
			runTestGit(t, dir, "remote", "add", "origin", c.remote)
		}
		if got := githubBranchURL(dir, "sillage/100-demo"); got != c.want {
			t.Errorf("githubBranchURL(remote=%q) = %q, attendu %q", c.remote, got, c.want)
		}
	}
}

// --- Livraison par chantier (voir docs/SPEC-LIVRAISON.md) ---

// initTestRepo initialise un dépôt git avec un commit initial sur main.
func initTestRepo(t *testing.T, repo string) {
	t.Helper()
	runTestGit(t, repo, "init", "-b", "main")
	runTestGit(t, repo, "config", "user.email", "test@example.com")
	runTestGit(t, repo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# projet\n"), 0o644); err != nil {
		t.Fatalf("écriture README impossible : %v", err)
	}
	runTestGit(t, repo, "add", "-A")
	runTestGit(t, repo, "commit", "-m", "initial")
}

// deliveryFixture monte un dépôt git réel (avec remote bare), un serveur, un
// projet et un chantier, prêts pour les tests de livraison.
type deliveryFixture struct {
	srv     *Server
	repo    string
	bare    string
	dataDir string
	project Project
	card    Card
}

func newDeliveryFixture(t *testing.T, delivery Delivery) *deliveryFixture {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}

	bare := t.TempDir()
	runTestGit(t, bare, "init", "--bare", "-b", "main")
	repo := t.TempDir()
	initTestRepo(t, repo)
	runTestGit(t, repo, "remote", "add", "origin", bare)

	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	srv := NewServer(store, "", dataDir, fstest.MapFS{})
	project, err := store.AddProject("demo", "", "", []Repo{{Name: "demo", Path: repo}}, nil, &delivery)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := store.AddCard(project.ID, "Refonte auth", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	// Les répertoires temporaires sont supprimés à la fin du test, mais un
	// rebase de fond ou un agent interrompu peut encore y écrire : leurs
	// goroutines survivent au retour de la fonction qui les a lancées. Ce
	// nettoyage est enregistré après les t.TempDir() ci-dessus, donc il s'exécute
	// avant eux (LIFO) : on attend, puis on supprime.
	t.Cleanup(func() {
		srv.waitRebases()
		srv.runner.waitTasks()
	})
	return &deliveryFixture{srv: srv, repo: repo, bare: bare, dataDir: dataDir, project: project, card: card}
}

// addTask crée une tâche sur la branche du chantier (même chemin que
// handleCreateTask, sans lancer d'agent), y écrit un fichier, et la met en
// revue. Retourne l'identifiant de la tâche et la branche du chantier.
func (f *deliveryFixture) addTask(t *testing.T, title, file, content string) (string, CardBranch) {
	t.Helper()
	cb, err := f.srv.ensureCardBranch(f.card, f.project, f.project.Repos[0])
	if err != nil {
		t.Fatalf("ensureCardBranch: %v", err)
	}
	id, ref := f.srv.store.ReserveTaskID()
	branch := fmt.Sprintf("sillage/%d-%s", ref, Slugify(title))
	dir, base, err := CreateWorktree(f.repo, f.dataDir, id, branch, cb.Branch)
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	if base != cb.Branch {
		t.Fatalf("la tâche devrait partir de la branche du chantier %q, reçu %q", cb.Branch, base)
	}
	if _, err := f.srv.store.CreateTask(id, ref, f.card.ID, f.project.ID, title, "echo", branch, base, dir, "demo"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0o644); err != nil {
		t.Fatalf("écriture %s impossible : %v", file, err)
	}
	// filesCount : ce que le runner calcule à la fin d'un agent. Une tâche à
	// zéro fichier n'est jamais acceptée automatiquement (voir
	// autoAcceptMergedTasks), il faut donc le renseigner ici comme en vrai.
	if _, err := f.srv.store.UpdateTask(id, func(tk *Task) { tk.Status = "review"; tk.FilesCount = 1 }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	return id, cb
}

// TestRestartRefreshesCountersOfInterruptedTask : une tâche interrompue par un
// arrêt de Sillage doit revenir avec ses compteurs, relus depuis git au
// chargement. Le runner ne les écrit qu'à la fin d'une exécution : sans ce
// rattrapage, un travail commité s'affiche « 0 fichier, 0 commit », et une
// branche déjà fusionnée à la main ne serait jamais constatée acceptée
// (autoAcceptMergedTasks exige filesCount > 0).
func TestRestartRefreshesCountersOfInterruptedTask(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge"})
	id, _ := f.addTask(t, "Refonte visuelle", "style.css", ":root{--glass:1}\n")
	task, ok := f.srv.store.GetTask(id)
	if !ok {
		t.Fatalf("tâche %s introuvable", id)
	}
	if _, err := CommitAll(task.WorktreeDir, "Redesign the UI"); err != nil {
		t.Fatalf("CommitAll: %v", err)
	}
	// L'état que laisse un agent tué en pleine exécution : running, et des
	// compteurs jamais écrits.
	activity := "Bash · git status --short"
	if _, err := f.srv.store.UpdateTask(id, func(tk *Task) {
		tk.Status = "running"
		tk.LiveActivity = &activity
		tk.FilesCount, tk.DocsCount, tk.CommitsCount = 0, 0, 0
	}); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	reloaded, err := NewStore(f.dataDir)
	if err != nil {
		t.Fatalf("NewStore après redémarrage : %v", err)
	}
	got, ok := reloaded.GetTask(id)
	if !ok {
		t.Fatalf("tâche %s introuvable après redémarrage", id)
	}
	if got.Status != "review" {
		t.Fatalf("statut après redémarrage = %q, want review", got.Status)
	}
	if got.FilesCount != 1 || got.CommitsCount != 1 {
		t.Fatalf("compteurs relus attendus (1 fichier, 1 commit), reçu files=%d commits=%d", got.FilesCount, got.CommitsCount)
	}
}

// accept accepte une tâche et attend les rebases automatiques que l'acceptation
// déclenche en tâche de fond (voir rebaseSiblingTasks) : sans cette attente, un
// test verrait un état intermédiaire, ou courrait contre le nettoyage de
// t.TempDir pendant qu'un git tourne encore.
func (f *deliveryFixture) accept(t *testing.T, taskID string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetPathValue("id", taskID)
	w := httptest.NewRecorder()
	f.srv.handleAccept(w, req)
	f.srv.waitRebases()
	return w
}

func (f *deliveryFixture) ship(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.SetPathValue("id", f.card.ID)
	w := httptest.NewRecorder()
	f.srv.handleCardShip(w, req)
	return w
}

// catchUp rattrape la branche de destination dans celle du chantier et retourne
// la réponse décodée.
func (f *deliveryFixture) catchUp(t *testing.T) CatchUpResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
	req.SetPathValue("id", f.card.ID)
	w := httptest.NewRecorder()
	f.srv.handleCardCatchUp(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("rattrapage : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var resp CatchUpResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("réponse de rattrapage illisible : %v", err)
	}
	return resp
}

// delivery appelle l'aperçu de livraison (qui constate au passage les branches
// fusionnées à la main) et retourne la réponse décodée.
func (f *deliveryFixture) delivery(t *testing.T) DeliveryPreview {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetPathValue("id", f.card.ID)
	w := httptest.NewRecorder()
	f.srv.handleCardDelivery(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("aperçu de livraison : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var prev DeliveryPreview
	if err := json.Unmarshal(w.Body.Bytes(), &prev); err != nil {
		t.Fatalf("aperçu illisible : %v", err)
	}
	return prev
}

// TestAcceptThenShipWorkstream couvre le flux complet du nouveau modèle :
// la tâche part de la branche du chantier, l'acceptation y fusionne son
// travail, la livraison pousse cette branche (une seule action sortante), et le
// chantier ne devient livrable qu'à partir de la première acceptation (il n'y a
// rien à livrer avant).
func TestAcceptThenShipWorkstream(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	taskID, cb := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if !strings.HasPrefix(cb.Branch, "sillage/ws-") {
		t.Fatalf("nom de branche de chantier inattendu : %q", cb.Branch)
	}

	// Rien d'accepté : la branche du chantier est vide, la livraison est bloquée
	// des deux côtés.
	card, _ := f.srv.store.GetCard(f.card.ID)
	if card.ShipReady || card.ShipBlocker != "nothing-accepted" {
		t.Fatalf("chantier attendu non livrable (nothing-accepted), reçu ready=%v blocker=%q", card.ShipReady, card.ShipBlocker)
	}
	if w := f.ship(t, `{"confirm":true}`); w.Code != http.StatusConflict {
		t.Fatalf("livraison sans rien d'accepté : attendu 409, reçu %d (%s)", w.Code, w.Body.String())
	}
	if w := f.ship(t, `{"confirm":false}`); w.Code != http.StatusBadRequest {
		t.Fatalf("livraison sans confirmation : attendu 400, reçu %d", w.Code)
	}

	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	if task, _ := f.srv.store.GetTask(taskID); task.Status != "accepted" {
		t.Fatalf("statut attendu 'accepted', reçu %q", task.Status)
	}
	// Le travail de la tâche est dans la branche du chantier.
	if _, err := os.Stat(filepath.Join(cb.WorktreeDir, "feature.txt")); err != nil {
		t.Fatalf("feature.txt devrait être fusionné dans la branche du chantier : %v", err)
	}
	msgs := f.srv.store.GetMessages(taskID)
	last := msgs[len(msgs)-1]
	if last.Text != "[accepted:"+cb.Branch+"]" || last.Author != "agent" || last.AuthorName != "" {
		t.Fatalf("message marqueur d'acceptation inattendu : %+v", last)
	}

	card, _ = f.srv.store.GetCard(f.card.ID)
	if !card.ShipReady {
		t.Fatalf("chantier attendu livrable, blocage %q", card.ShipBlocker)
	}
	if card.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' avant livraison, reçue %q", card.Column)
	}

	w := f.ship(t, `{"confirm":true}`)
	if w.Code != http.StatusOK {
		t.Fatalf("ship: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var resp ShipResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("réponse de livraison illisible : %v", err)
	}
	if len(resp.Repos) != 1 || !resp.Repos[0].Pushed || resp.Repos[0].Error != "" {
		t.Fatalf("résultat de livraison inattendu : %+v", resp.Repos)
	}
	// La branche du chantier est réellement sur le remote.
	if out := runTestGit(t, f.bare, "branch", "--list", cb.Branch); !strings.Contains(out, "ws-") {
		t.Fatalf("la branche du chantier devrait être poussée, reçu : %q", out)
	}
	card, _ = f.srv.store.GetCard(f.card.ID)
	if card.Column != "done" {
		t.Fatalf("colonne attendue 'done' après livraison, reçue %q", card.Column)
	}
	if len(card.Branches) != 1 || card.Branches[0].ShippedAt == nil {
		t.Fatalf("shippedAt devrait être renseigné : %+v", card.Branches)
	}
}

// TestShipPartialWorkstream : livrer un chantier dont tout n'est pas fini. Une
// tâche est acceptée, une autre est encore en revue : la livraison part quand
// même, avec le travail accepté et lui seul. La carte reste en « En cours »
// (« Terminé » veut toujours dire livré ET fini), et l'acceptation suivante
// rouvre une livraison qui ne pousse que le neuf.
func TestShipPartialWorkstream(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	first, cb := f.addTask(t, "Première", "un.txt", "un\n")
	second, _ := f.addTask(t, "Seconde", "deux.txt", "deux\n")
	if w := f.accept(t, first); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}

	card, _ := f.srv.store.GetCard(f.card.ID)
	if !card.ShipReady {
		t.Fatalf("chantier attendu livrable avec une tâche encore en revue, blocage %q", card.ShipBlocker)
	}
	prev := f.delivery(t)
	if !prev.Ready || prev.Counts.Accepted != 1 || prev.Counts.Pending != 1 {
		t.Fatalf("aperçu inattendu : ready=%v accepted=%d pending=%d", prev.Ready, prev.Counts.Accepted, prev.Counts.Pending)
	}

	if w := f.ship(t, `{"confirm":true}`); w.Code != http.StatusOK {
		t.Fatalf("livraison partielle : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	// Seul le travail accepté est parti : la tâche en revue n'est pas fusionnée.
	out := runTestGit(t, f.bare, "log", "--oneline", cb.Branch)
	if !strings.Contains(out, "Première") {
		t.Fatalf("le travail accepté devrait être poussé, reçu : %q", out)
	}
	if strings.Contains(out, "Seconde") {
		t.Fatalf("le travail non accepté ne doit jamais être poussé, reçu : %q", out)
	}
	card, _ = f.srv.store.GetCard(f.card.ID)
	if card.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' (une tâche reste à relire), reçue %q", card.Column)
	}
	if card.Branches[0].ShippedAt == nil {
		t.Fatalf("shippedAt devrait être renseigné après la livraison partielle")
	}

	// La suite du chantier se livre ensuite, et là seulement la carte est terminée.
	if w := f.accept(t, second); w.Code != http.StatusOK {
		t.Fatalf("accept (2e): attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	if w := f.ship(t, `{"confirm":true}`); w.Code != http.StatusOK {
		t.Fatalf("seconde livraison : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	if out := runTestGit(t, f.bare, "log", "--oneline", cb.Branch); !strings.Contains(out, "Seconde") {
		t.Fatalf("le travail de la seconde tâche devrait être poussé, reçu : %q", out)
	}
	card, _ = f.srv.store.GetCard(f.card.ID)
	if card.Column != "done" {
		t.Fatalf("colonne attendue 'done' après la livraison complète, reçue %q", card.Column)
	}
}

// TestAcceptConflictKeepsTaskInReview : deux tâches du même chantier qui
// touchent la même ligne. "Seconde" est occupée (agent en cours, simulé sans
// process réel) au moment de la première acceptation, donc le rebase
// automatique des tâches sœurs (rebaseSiblingTasks) la laisse de côté : c'est
// seulement à sa propre acceptation, une fois libérée, que la fusion directe
// (MergeBranch dans acceptTaskInto) découvre le conflit. Elle échoue en 409,
// la tâche RESTE en revue le temps de répondre, un message marqueur de
// conflit est ajouté au fil, ET l'agent reçoit aussitôt l'instruction de
// reprendre la base (conflictRebasePrompt) : le conflit se règle donc tout
// seul, sans attendre que l'humain remarque le marqueur.
func TestAcceptConflictKeepsTaskInReview(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	first, _ := f.addTask(t, "Première", "README.md", "# projet\n\nversion A\n")
	second, cb := f.addTask(t, "Seconde", "README.md", "# projet\n\nversion B\n")

	f.srv.runner.mu.Lock()
	f.srv.runner.procs[second] = &procHandle{done: make(chan struct{})}
	f.srv.runner.mu.Unlock()

	if w := f.accept(t, first); w.Code != http.StatusOK {
		t.Fatalf("première acceptation : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}

	f.srv.runner.mu.Lock()
	delete(f.srv.runner.procs, second)
	f.srv.runner.mu.Unlock()

	w := f.accept(t, second)
	if w.Code != http.StatusConflict {
		t.Fatalf("seconde acceptation : attendu 409, reçu %d (%s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "merge conflict") {
		t.Fatalf("message de conflit attendu, reçu %q", w.Body.String())
	}
	task, _ := f.srv.store.GetTask(second)
	if task.Status != "running" {
		t.Fatalf("le conflit doit relancer l'agent (running), reçu %q", task.Status)
	}
	msgs := f.srv.store.GetMessages(second)
	if len(msgs) < 2 || !strings.HasPrefix(msgs[len(msgs)-2].Text, "[merge-conflict:") || !strings.Contains(msgs[len(msgs)-2].Text, "README.md") {
		t.Fatalf("message marqueur de conflit inattendu : %+v", msgs)
	}
	if last := msgs[len(msgs)-1]; !strings.Contains(last.Text, cb.Branch) {
		t.Fatalf("l'agent devrait recevoir la branche du chantier comme instruction, reçu %q", last.Text)
	}
	// L'arbre du chantier est intact (merge --abort) : le travail de la première
	// tâche reste livrable, malgré la seconde encore en cours.
	card, _ := f.srv.store.GetCard(f.card.ID)
	if !card.ShipReady {
		t.Fatalf("chantier attendu livrable malgré la tâche en conflit, blocage %q", card.ShipBlocker)
	}

	if _, err := f.srv.runner.Interrupt(second); err != nil {
		t.Fatalf("interrupt: %v", err)
	}
}

// TestDeliveryPreviewReportsBehind couvre le retard annoncé par l'aperçu de
// livraison, celui qui provoque les conflits : une tâche dont la branche de
// chantier a avancé sous elle (parce qu'une autre tâche a été acceptée), et le
// chantier lui-même quand sa base a avancé. Un rebase remet les compteurs à
// zéro, et une tâche acceptée n'y figure jamais (rien à rebaser).
func TestDeliveryPreviewReportsBehind(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	first, cb := f.addTask(t, "Première", "a.txt", "contenu a\n")
	second, _ := f.addTask(t, "Seconde", "b.txt", "contenu b\n")

	if prev := f.delivery(t); len(prev.Behind) != 0 {
		t.Fatalf("deux tâches parties de la même branche ne sont pas en retard, reçu %v", prev.Behind)
	}

	// La seconde tâche a son agent au travail : le rebase automatique la laisse
	// tranquille (voir rebaseSiblingTasks), et c'est justement le cas où le
	// badge de retard doit s'afficher.
	if _, err := f.srv.store.UpdateTask(second, func(tk *Task) { tk.Status = "running" }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	// Accepter la première fait avancer la branche du chantier d'un commit :
	// la seconde tâche est désormais en retard de ce commit.
	if w := f.accept(t, first); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	// Deux commits, pas un : celui de la tâche acceptée, puis le commit de
	// fusion (MergeBranch fusionne en --no-ff). On compte des commits absents,
	// comme le fait une forge, pas des acceptations.
	prev := f.delivery(t)
	if prev.Behind[second] != 2 {
		t.Fatalf("la seconde tâche devrait être en retard de deux commits, reçu %d (%v)", prev.Behind[second], prev.Behind)
	}
	if _, ok := prev.Behind[first]; ok {
		t.Fatalf("une tâche acceptée n'a rien à rebaser, elle ne doit pas figurer dans behind : %v", prev.Behind)
	}

	// Le rebase (celui que l'agent finit par faire, ou celui du bouton) : le
	// retard disparaît. Le travail en attente est commité d'abord, comme le fait
	// le rebase automatique.
	task, _ := f.srv.store.GetTask(second)
	if _, err := CommitAll(task.WorktreeDir, "Sillage: Seconde"); err != nil {
		t.Fatalf("CommitAll: %v", err)
	}
	runTestGit(t, task.WorktreeDir, "-c", "user.email=test@example.com", "-c", "user.name=Test", "rebase", cb.Branch)
	if prev := f.delivery(t); len(prev.Behind) != 0 {
		t.Fatalf("après rebase, plus aucun retard attendu, reçu %v", prev.Behind)
	}

	// Retard du chantier sur sa base : un commit arrive sur main.
	if err := os.WriteFile(filepath.Join(f.repo, "release.txt"), []byte("livré\n"), 0o644); err != nil {
		t.Fatalf("écriture impossible : %v", err)
	}
	runTestGit(t, f.repo, "add", "-A")
	runTestGit(t, f.repo, "commit", "-m", "avance sur main")

	prev = f.delivery(t)
	if len(prev.Repos) != 1 {
		t.Fatalf("un seul dépôt attendu, reçu %d", len(prev.Repos))
	}
	if prev.Repos[0].Behind != 1 {
		t.Fatalf("le chantier devrait être en retard d'un commit sur %q, reçu %d", prev.Repos[0].Base, prev.Repos[0].Behind)
	}
}

// TestShipMergeModeNeverPushes couvre le mode "merge" : la branche du chantier
// est fusionnée en fast-forward dans la branche de destination du dépôt de
// travail, et RIEN n'est poussé (le remote reste vide).
func TestShipMergeModeNeverPushes(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})

	taskID, _ := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}

	w := f.ship(t, `{"confirm":true}`)
	if w.Code != http.StatusOK {
		t.Fatalf("ship: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var resp ShipResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("réponse de livraison illisible : %v", err)
	}
	if len(resp.Repos) != 1 || !resp.Repos[0].Merged || resp.Repos[0].Pushed || resp.Repos[0].Error != "" {
		t.Fatalf("résultat de livraison inattendu : %+v", resp.Repos)
	}
	// La fusion a bien eu lieu dans le dépôt de travail...
	if _, err := os.Stat(filepath.Join(f.repo, "feature.txt")); err != nil {
		t.Fatalf("feature.txt devrait être fusionné dans main : %v", err)
	}
	// ...et absolument rien n'a été poussé.
	if out := strings.TrimSpace(runTestGit(t, f.bare, "branch", "--list")); out != "" {
		t.Fatalf("le mode merge ne doit jamais pousser, branches distantes : %q", out)
	}
}

// TestMergeLocalRefusesWhenTargetDiverged : la fusion locale n'accepte que le
// fast-forward, et ne tente jamais de résoudre quoi que ce soit.
func TestMergeLocalRefusesWhenTargetDiverged(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})

	taskID, cb := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	// main avance de son côté : plus de fast-forward possible.
	if err := os.WriteFile(filepath.Join(f.repo, "autre.txt"), []byte("divergence\n"), 0o644); err != nil {
		t.Fatalf("écriture autre.txt impossible : %v", err)
	}
	runTestGit(t, f.repo, "add", "-A")
	runTestGit(t, f.repo, "commit", "-m", "divergence")

	_, err := MergeLocal(f.repo, f.dataDir, f.card.ID, "demo", "main", cb.Branch)
	if !errors.Is(err, ErrTargetDiverged) {
		t.Fatalf("erreur attendue ErrTargetDiverged, reçue %v", err)
	}
}

// TestDeliveryPreviewTargetPosition couvre les deux positions du chantier par
// rapport à sa destination, celles dont l'UI se sert pour ne pas proposer une
// livraison qui échouerait (ou qui n'apporterait rien) :
//
//   - fusionnable en fast-forward tant que la destination n'a pas bougé ;
//   - plus rien à livrer quand la branche est arrivée dans la destination ;
//   - ni l'un ni l'autre quand les deux ont divergé.
func TestDeliveryPreviewTargetPosition(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})

	taskID, cb := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}

	prev := f.delivery(t)
	if len(prev.Repos) != 1 {
		t.Fatalf("un seul dépôt attendu, reçu %d", len(prev.Repos))
	}
	if !prev.Repos[0].FastForwardable || prev.Repos[0].MergedIntoTarget {
		t.Fatalf("chantier attendu fusionnable et pas encore arrivé, reçu %+v", prev.Repos[0])
	}

	// main avance de son côté : plus de fast-forward possible, et le chantier
	// n'est toujours pas arrivé. C'est le cas où livrer échouerait.
	if err := os.WriteFile(filepath.Join(f.repo, "ailleurs.txt"), []byte("autre\n"), 0o644); err != nil {
		t.Fatalf("écriture impossible : %v", err)
	}
	runTestGit(t, f.repo, "add", "-A")
	runTestGit(t, f.repo, "commit", "-m", "avance sur main")

	prev = f.delivery(t)
	if prev.Repos[0].FastForwardable || prev.Repos[0].MergedIntoTarget {
		t.Fatalf("après divergence : ni fusionnable ni arrivé, reçu %+v", prev.Repos[0])
	}

	// La branche du chantier est fusionnée à la main dans main : tout est
	// arrivé, il n'y a plus rien à livrer.
	runTestGit(t, f.repo, "-c", "user.email=test@example.com", "-c", "user.name=Test", "merge", "--no-edit", cb.Branch)
	prev = f.delivery(t)
	if !prev.Repos[0].MergedIntoTarget {
		t.Fatalf("branche fusionnée à la main : arrivée attendue, reçu %+v", prev.Repos[0])
	}
	if prev.Repos[0].Pending != 0 {
		t.Fatalf("plus rien à livrer attendu, reçu pending=%d", prev.Repos[0].Pending)
	}
}

// TestCatchUpUnblocksShip : quand la destination a avancé, le rattrapage la
// fusionne dans la branche du chantier, ce qui rend la fusion fast-forward (donc
// la livraison) possible à nouveau. L'historique du chantier n'est pas réécrit :
// le commit de la tâche acceptée est toujours là.
func TestCatchUpUnblocksShip(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})

	taskID, cb := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	shaBefore := HeadSha(cb.WorktreeDir)

	// main avance de son côté, sur un autre fichier : plus de fast-forward.
	if err := os.WriteFile(filepath.Join(f.repo, "release.txt"), []byte("livré\n"), 0o644); err != nil {
		t.Fatalf("écriture impossible : %v", err)
	}
	runTestGit(t, f.repo, "add", "-A")
	runTestGit(t, f.repo, "commit", "-m", "avance sur main")
	if prev := f.delivery(t); prev.Repos[0].FastForwardable {
		t.Fatal("la destination a avancé : le fast-forward ne devrait plus être possible")
	}

	resp := f.catchUp(t)
	if len(resp.Repos) != 1 || !resp.Repos[0].Merged || resp.Repos[0].Error != "" {
		t.Fatalf("rattrapage inattendu : %+v", resp.Repos)
	}
	// Le travail du chantier est intact, et il a en plus le contenu de main.
	if _, err := os.Stat(filepath.Join(cb.WorktreeDir, "feature.txt")); err != nil {
		t.Fatalf("le travail du chantier doit rester : %v", err)
	}
	if _, err := os.Stat(filepath.Join(cb.WorktreeDir, "release.txt")); err != nil {
		t.Fatalf("le chantier devrait avoir rattrapé main : %v", err)
	}
	if HeadSha(cb.WorktreeDir) == shaBefore {
		t.Fatal("le rattrapage devrait avoir ajouté un commit de fusion")
	}

	prev := f.delivery(t)
	if !prev.Repos[0].FastForwardable || prev.Repos[0].Behind != 0 {
		t.Fatalf("après rattrapage : fusion attendue possible et aucun retard, reçu %+v", prev.Repos[0])
	}
	if w := f.ship(t, `{"confirm":true}`); w.Code != http.StatusOK {
		t.Fatalf("ship après rattrapage : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	// Un second rattrapage n'a plus rien à faire.
	if resp2 := f.catchUp(t); !resp2.Repos[0].UpToDate {
		t.Fatalf("second rattrapage : « déjà à jour » attendu, reçu %+v", resp2.Repos[0])
	}
}

// TestCatchUpConflictLeavesWorkstreamIntact : un rattrapage qui conflicte est
// annulé. Le chantier reste exactement dans l'état où il était : c'est la
// condition pour pouvoir le proposer d'un simple clic.
func TestCatchUpConflictLeavesWorkstreamIntact(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})

	taskID, cb := f.addTask(t, "Ajoute une feature", "partage.txt", "version chantier\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	shaBefore := HeadSha(cb.WorktreeDir)

	// main touche le même fichier, autrement : le rattrapage ne peut pas trancher.
	if err := os.WriteFile(filepath.Join(f.repo, "partage.txt"), []byte("version main\n"), 0o644); err != nil {
		t.Fatalf("écriture impossible : %v", err)
	}
	runTestGit(t, f.repo, "add", "-A")
	runTestGit(t, f.repo, "commit", "-m", "main touche le même fichier")

	resp := f.catchUp(t)
	if resp.Repos[0].ConflictFilePaths != "partage.txt" {
		t.Fatalf("conflit attendu sur partage.txt, reçu %+v", resp.Repos[0])
	}
	if resp.Repos[0].Merged {
		t.Fatal("aucune fusion ne doit être enregistrée en cas de conflit")
	}
	if HeadSha(cb.WorktreeDir) != shaBefore {
		t.Fatal("le chantier doit rester intact après un conflit de rattrapage")
	}
	if !IsWorktreeClean(cb.WorktreeDir) {
		t.Fatal("le worktree du chantier doit être propre (fusion annulée)")
	}
	if content, err := os.ReadFile(filepath.Join(cb.WorktreeDir, "partage.txt")); err != nil || string(content) != "version chantier\n" {
		t.Fatalf("le contenu du chantier doit être intact, reçu %q (err %v)", content, err)
	}
}

// TestAcceptRebasesSiblingTasks : accepter une tâche rejoue automatiquement les
// autres tâches en revue du chantier sur la branche du chantier, avec un
// marqueur dans leur fil. Le travail non commité de ces tâches est commité
// avant (les agents ne commitent pas toujours) et n'est donc jamais perdu.
func TestAcceptRebasesSiblingTasks(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	firstID, cb := f.addTask(t, "Première tâche", "premier.txt", "un\n")
	secondID, _ := f.addTask(t, "Deuxième tâche", "second.txt", "deux\n")
	second, _ := f.srv.store.GetTask(secondID)

	if w := f.accept(t, firstID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	f.srv.waitRebases()

	// La branche du chantier est désormais contenue dans celle de la deuxième
	// tâche : elle repart du travail accepté.
	if !IsBranchMergedInto(second.WorktreeDir, cb.Branch, second.Branch) {
		t.Fatal("la deuxième tâche devrait avoir été rebasée sur la branche du chantier")
	}
	// Son propre travail est toujours là, et le fil l'annonce.
	if _, err := os.Stat(filepath.Join(second.WorktreeDir, "second.txt")); err != nil {
		t.Fatalf("le travail de la deuxième tâche ne doit pas être perdu : %v", err)
	}
	if _, err := os.Stat(filepath.Join(second.WorktreeDir, "premier.txt")); err != nil {
		t.Fatalf("la deuxième tâche devrait voir le travail accepté : %v", err)
	}
	msgs := f.srv.store.GetMessages(secondID)
	last := msgs[len(msgs)-1]
	if last.Text != "[rebased:"+cb.Branch+"]" {
		t.Fatalf("marqueur de rebase attendu, reçu %q", last.Text)
	}
	if task, _ := f.srv.store.GetTask(secondID); task.Rebasing || task.Status != "review" {
		t.Fatalf("tâche attendue en revue sans rebase en cours, reçu status=%q rebasing=%v", task.Status, task.Rebasing)
	}
}

// TestAcceptRebaseConflictAsksAgentToRebase : un rebase automatique qui
// conflicte est annulé (worktree intact, branche inchangée), posé au fil comme
// tel, ET envoyé aussitôt à l'agent comme instruction (voir
// conflictRebasePrompt) : lui seul sait reprendre la base, la tâche repart
// donc "running" au lieu de rester plantée en revue avec un marqueur que
// personne ne lit.
func TestAcceptRebaseConflictAsksAgentToRebase(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	// Les deux tâches touchent le même fichier, avec un contenu différent.
	firstID, cb := f.addTask(t, "Première tâche", "partage.txt", "version un\n")
	secondID, _ := f.addTask(t, "Deuxième tâche", "partage.txt", "version deux\n")
	second, _ := f.srv.store.GetTask(secondID)

	if w := f.accept(t, firstID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	f.srv.waitRebases()

	if IsBranchMergedInto(second.WorktreeDir, cb.Branch, second.Branch) {
		t.Fatal("un rebase en conflit ne doit rien changer à la branche de la tâche")
	}
	// Le rebase est annulé, pas laissé en plan : le worktree est utilisable.
	if !IsWorktreeClean(second.WorktreeDir) {
		t.Fatal("le worktree devrait être propre après l'annulation du rebase")
	}
	if content, err := os.ReadFile(filepath.Join(second.WorktreeDir, "partage.txt")); err != nil || string(content) != "version deux\n" {
		t.Fatalf("le travail de la tâche doit être intact, reçu %q (err %v)", content, err)
	}
	msgs := f.srv.store.GetMessages(secondID)
	if len(msgs) < 2 || msgs[len(msgs)-2].Text != "[rebase-conflict:partage.txt]" {
		t.Fatalf("marqueur de conflit de rebase attendu avant l'instruction, reçu %+v", msgs)
	}
	last := msgs[len(msgs)-1]
	if !strings.Contains(last.Text, cb.Branch) {
		t.Fatalf("l'agent devrait recevoir la branche du chantier comme instruction, reçu %q", last.Text)
	}
	if task, _ := f.srv.store.GetTask(secondID); task.Rebasing || task.Status != "running" {
		t.Fatalf("tâche attendue relancée (running) après le conflit, reçu status=%q rebasing=%v", task.Status, task.Rebasing)
	}

	if _, err := f.srv.runner.Interrupt(secondID); err != nil {
		t.Fatalf("interrupt: %v", err)
	}
}

// TestShipMergePushModeUpdatesRemoteTarget couvre le mode "merge-push" : la
// branche du chantier est fusionnée dans la branche de destination, ET cette
// branche de destination est poussée (c'est toute la différence avec "merge").
func TestShipMergePushModeUpdatesRemoteTarget(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge-push", Target: "main"})

	taskID, _ := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}

	w := f.ship(t, `{"confirm":true}`)
	if w.Code != http.StatusOK {
		t.Fatalf("ship: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var resp ShipResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("réponse de livraison illisible : %v", err)
	}
	if len(resp.Repos) != 1 || !resp.Repos[0].Merged || !resp.Repos[0].Pushed || resp.Repos[0].Error != "" {
		t.Fatalf("résultat de livraison inattendu : %+v", resp.Repos)
	}
	// La fusion a eu lieu localement...
	if _, err := os.Stat(filepath.Join(f.repo, "feature.txt")); err != nil {
		t.Fatalf("feature.txt devrait être fusionné dans main : %v", err)
	}
	// ...et main est poussée, avec le travail dedans. La branche du chantier,
	// elle, n'a aucune raison d'être poussée dans ce mode.
	branches := runTestGit(t, f.bare, "branch", "--list")
	if !strings.Contains(branches, "main") {
		t.Fatalf("main devrait être poussée, branches distantes : %q", branches)
	}
	if strings.Contains(branches, "sillage/ws-") {
		t.Fatalf("le mode merge-push ne pousse que la branche de destination, reçu : %q", branches)
	}
	if files := runTestGit(t, f.bare, "ls-tree", "--name-only", "main"); !strings.Contains(files, "feature.txt") {
		t.Fatalf("feature.txt devrait être dans main côté remote, reçu : %q", files)
	}
}

// TestMergeAndPushRefusesWhenRemoteDiverged : le mode "merge-push" rattrape le
// remote avant de fusionner, et refuse net si la branche de destination a
// vraiment divergé. Rien n'est poussé, rien n'est fusionné : un refus avant
// écriture vaut mieux qu'une fusion locale suivie d'un push rejeté.
func TestMergeAndPushRefusesWhenRemoteDiverged(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge-push", Target: "main"})

	taskID, cb := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	// Un autre clone pousse un commit sur main : le local a divergé du remote
	// (il a le travail du chantier, le remote a ce commit-ci).
	other := t.TempDir()
	runTestGit(t, other, "clone", f.repo, other)
	runTestGit(t, other, "remote", "set-url", "origin", f.bare)
	// Un clone n'hérite pas de l'identité git : sans elle, commit échoue sur une
	// machine sans configuration globale (la CI, typiquement).
	runTestGit(t, other, "config", "user.email", "test@example.com")
	runTestGit(t, other, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(other, "ailleurs.txt"), []byte("remote\n"), 0o644); err != nil {
		t.Fatalf("écriture ailleurs.txt impossible : %v", err)
	}
	runTestGit(t, other, "add", "-A")
	runTestGit(t, other, "commit", "-m", "travail venu d'ailleurs")
	runTestGit(t, other, "push", "origin", "main")
	// main local avance aussi, sur un autre commit : divergence réelle.
	if err := os.WriteFile(filepath.Join(f.repo, "ici.txt"), []byte("local\n"), 0o644); err != nil {
		t.Fatalf("écriture ici.txt impossible : %v", err)
	}
	runTestGit(t, f.repo, "add", "-A")
	runTestGit(t, f.repo, "commit", "-m", "travail local")

	_, err := MergeAndPush(f.repo, f.dataDir, f.card.ID, "demo", "main", cb.Branch)
	if !errors.Is(err, ErrTargetDiverged) {
		t.Fatalf("erreur attendue ErrTargetDiverged, reçue %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(f.repo, "feature.txt")); statErr == nil {
		t.Fatal("aucune fusion ne doit avoir lieu quand la destination a divergé")
	}
}

// TestShipPushModeDoesNotOpenPR couvre le mode "push" : la branche du chantier
// est poussée, et aucune pull request n'est ouverte (prUrl vide).
func TestShipPushModeDoesNotOpenPR(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "push"})

	taskID, cb := f.addTask(t, "Ajoute une feature", "feature.txt", "contenu\n")
	if w := f.accept(t, taskID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}

	w := f.ship(t, `{"confirm":true}`)
	if w.Code != http.StatusOK {
		t.Fatalf("ship: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var resp ShipResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("réponse de livraison illisible : %v", err)
	}
	if len(resp.Repos) != 1 || !resp.Repos[0].Pushed || resp.Repos[0].Merged || resp.Repos[0].Error != "" {
		t.Fatalf("résultat de livraison inattendu : %+v", resp.Repos)
	}
	if resp.Repos[0].PrURL != "" {
		t.Fatalf("le mode push n'ouvre aucune pull request, reçu %q", resp.Repos[0].PrURL)
	}
	// La branche du chantier est poussée, main n'est pas touchée.
	branches := runTestGit(t, f.bare, "branch", "--list")
	if !strings.Contains(branches, cb.Branch) {
		t.Fatalf("la branche du chantier devrait être poussée, reçu : %q", branches)
	}
	if strings.Contains(branches, "main") {
		t.Fatalf("le mode push ne touche pas la branche de destination, reçu : %q", branches)
	}
}

// TestAutoAcceptBranchMergedByHand : une branche de tâche fusionnée à la main
// dans la branche du chantier (hors de Sillage) est constatée à la lecture de
// l'aperçu de livraison, et la tâche passe « acceptée » avec son marqueur.
func TestAutoAcceptBranchMergedByHand(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	taskID, cb := f.addTask(t, "Fusionnée à la main", "manuel.txt", "contenu\n")

	// Du travail non commité : rien ne doit être conclu automatiquement.
	if prev := f.delivery(t); prev.Counts.Pending != 1 {
		t.Fatalf("une tâche avec un worktree sale ne doit pas être acceptée : %+v", prev.Counts)
	}
	if task, _ := f.srv.store.GetTask(taskID); task.Status != "review" {
		t.Fatalf("statut attendu 'review', reçu %q", task.Status)
	}

	// L'humain commite puis fusionne lui-même, hors de Sillage.
	task, _ := f.srv.store.GetTask(taskID)
	runTestGit(t, task.WorktreeDir, "add", "-A")
	runTestGit(t, task.WorktreeDir, "commit", "-m", "à la main")
	runTestGit(t, cb.WorktreeDir, "merge", "--no-ff", "-m", "merge manuel", task.Branch)

	// Une tâche qui n'a rien produit n'est jamais conclue automatiquement.
	if _, err := f.srv.store.UpdateTask(taskID, func(tk *Task) { tk.FilesCount = 0 }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if prev := f.delivery(t); prev.Counts.Accepted != 0 {
		t.Fatalf("une tâche sans fichier ne doit pas être acceptée automatiquement : %+v", prev.Counts)
	}
	if _, err := f.srv.store.UpdateTask(taskID, func(tk *Task) { tk.FilesCount = 1 }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	prev := f.delivery(t)
	if prev.Counts.Accepted != 1 || prev.Counts.Pending != 0 {
		t.Fatalf("la tâche fusionnée à la main devrait être acceptée : %+v", prev.Counts)
	}
	if task, _ := f.srv.store.GetTask(taskID); task.Status != "accepted" {
		t.Fatalf("statut attendu 'accepted', reçu %q", task.Status)
	}
	msgs := f.srv.store.GetMessages(taskID)
	if last := msgs[len(msgs)-1]; last.Text != "[auto-accepted:"+cb.Branch+"]" {
		t.Fatalf("message marqueur attendu [auto-accepted:...], reçu %+v", last)
	}
	// Le chantier est livrable, et la livraison n'a pas été déclenchée.
	if !prev.Ready {
		t.Fatalf("chantier attendu livrable, blocage %q", prev.Blocker)
	}
	if out := strings.TrimSpace(runTestGit(t, f.bare, "branch", "--list")); out != "" {
		t.Fatalf("aucune livraison ne devait avoir lieu, branches distantes : %q", out)
	}
}

// TestAcceptLegacyTaskCreatesWorkstreamBranch : une tâche créée avant les
// branches de chantier (aucune `branches` sur la carte, branche partie de
// `main`) reste acceptable : la branche du chantier est créée à la volée et le
// travail en cours n'est pas perdu.
func TestAcceptLegacyTaskCreatesWorkstreamBranch(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	// Reproduit l'ancien modèle : branche de tâche partie de main, carte sans
	// branche de chantier.
	id, ref := f.srv.store.ReserveTaskID()
	branch := fmt.Sprintf("sillage/%d-ancienne", ref)
	dir, base, err := CreateWorktree(f.repo, f.dataDir, id, branch, "main")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	if _, err := f.srv.store.CreateTask(id, ref, f.card.ID, f.project.ID, "Ancienne tâche", "echo", branch, base, dir, "demo"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ancien.txt"), []byte("travail en cours\n"), 0o644); err != nil {
		t.Fatalf("écriture impossible : %v", err)
	}
	if _, err := f.srv.store.UpdateTask(id, func(tk *Task) { tk.Status = "review"; tk.FilesCount = 1 }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	if card, _ := f.srv.store.GetCard(f.card.ID); len(card.Branches) != 0 {
		t.Fatalf("précondition : la carte ne devrait avoir aucune branche de chantier")
	}

	if w := f.accept(t, id); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	card, _ := f.srv.store.GetCard(f.card.ID)
	if len(card.Branches) != 1 {
		t.Fatalf("la branche du chantier devrait avoir été créée : %+v", card.Branches)
	}
	if _, err := os.Stat(filepath.Join(card.Branches[0].WorktreeDir, "ancien.txt")); err != nil {
		t.Fatalf("le travail de la tâche devrait être fusionné dans la branche du chantier : %v", err)
	}
	if task, _ := f.srv.store.GetTask(id); task.Status != "accepted" {
		t.Fatalf("statut attendu 'accepted', reçu %q", task.Status)
	}
}

// TestShipAgainAfterNewWork : un chantier livré redevient livrable dès qu'une
// tâche nouvelle y est acceptée, et la seconde livraison ne pousse que le neuf
// en réutilisant la pull request existante.
func TestShipAgainAfterNewWork(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	first, cb := f.addTask(t, "Première", "un.txt", "un\n")
	if w := f.accept(t, first); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	if w := f.ship(t, `{"confirm":true}`); w.Code != http.StatusOK {
		t.Fatalf("ship: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	card, _ := f.srv.store.GetCard(f.card.ID)
	if card.Column != "done" || card.Branches[0].ShippedAt == nil {
		t.Fatalf("chantier attendu livré : colonne %q, shippedAt %v", card.Column, card.Branches[0].ShippedAt)
	}
	// Plus rien à livrer : une seconde livraison ne fait rien.
	if prev := f.delivery(t); prev.Repos[0].Pending != 0 {
		t.Fatalf("aucun commit ne devrait rester à livrer, reçu %d", prev.Repos[0].Pending)
	}

	// Nouvelle tâche sur un chantier livré : il repart en travail.
	second, _ := f.addTask(t, "Seconde", "deux.txt", "deux\n")
	card, _ = f.srv.store.GetCard(f.card.ID)
	if card.Branches[0].ShippedAt != nil {
		t.Fatalf("le chantier ne devrait plus être marqué livré après une nouvelle tâche")
	}
	if card.Column != "doing" {
		t.Fatalf("colonne attendue 'doing' après une nouvelle tâche, reçue %q", card.Column)
	}
	// La première tâche reste acceptée : le chantier reste livrable en principe,
	// mais l'aperçu n'annonce plus rien à livrer (tout est déjà poussé).
	if !card.ShipReady {
		t.Fatalf("chantier attendu livrable, blocage %q", card.ShipBlocker)
	}
	if prev := f.delivery(t); prev.Repos[0].Pending != 0 {
		t.Fatalf("aucun commit ne devrait rester à livrer, reçu %d", prev.Repos[0].Pending)
	}

	if w := f.accept(t, second); w.Code != http.StatusOK {
		t.Fatalf("accept (2e): attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	prev := f.delivery(t)
	if !prev.Ready || prev.Repos[0].Pending == 0 {
		t.Fatalf("chantier attendu livrable à nouveau : ready=%v pending=%d", prev.Ready, prev.Repos[0].Pending)
	}

	w := f.ship(t, `{"confirm":true}`)
	if w.Code != http.StatusOK {
		t.Fatalf("ship (2e): attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var resp ShipResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("réponse de livraison illisible : %v", err)
	}
	if !resp.Repos[0].Pushed || resp.Repos[0].Error != "" {
		t.Fatalf("seconde livraison inattendue : %+v", resp.Repos[0])
	}
	// Le remote a bien reçu le nouveau commit.
	out := runTestGit(t, f.bare, "log", "--oneline", cb.Branch)
	if !strings.Contains(out, "Seconde") {
		t.Fatalf("le travail de la seconde tâche devrait être poussé, reçu : %q", out)
	}
}

// TestParseRemoteAndForgeURLs couvre la détection de forge et les URL de repli
// (pré-remplies, en lecture seule) pour GitHub et GitLab, sous-groupes compris.
func TestParseRemoteAndForgeURLs(t *testing.T) {
	cases := []struct {
		in                 string
		wantHost, wantPath string
		wantOk             bool
	}{
		{"git@github.com:acme/demo.git", "github.com", "acme/demo", true},
		{"https://github.com/acme/demo", "github.com", "acme/demo", true},
		{"git@gitlab.com:acme/groupe/demo.git", "gitlab.com", "acme/groupe/demo", true},
		{"https://gitlab.example.org:8443/acme/demo.git", "gitlab.example.org", "acme/demo", true},
		{"ssh://git@gitlab.com/acme/demo.git", "gitlab.com", "acme/demo", true},
		{"/tmp/local-bare", "", "", false},
		{"", "", "", false},
	}
	for _, c := range cases {
		host, path, ok := ParseRemote(c.in)
		if ok != c.wantOk || host != c.wantHost || path != c.wantPath {
			t.Errorf("ParseRemote(%q) = (%q, %q, %v), attendu (%q, %q, %v)",
				c.in, host, path, ok, c.wantHost, c.wantPath, c.wantOk)
		}
	}

	if got, want := GithubCompareURL("github.com", "acme/demo", "main", "sillage/ws-1-x"),
		"https://github.com/acme/demo/compare/main...sillage/ws-1-x?expand=1"; got != want {
		t.Errorf("GithubCompareURL = %q, attendu %q", got, want)
	}
	if got, want := GitlabNewMRURL("gitlab.com", "acme/groupe/demo", "main", "sillage/ws-1-x"),
		"https://gitlab.com/acme/groupe/demo/-/merge_requests/new?merge_request[source_branch]=sillage%2Fws-1-x&merge_request[target_branch]=main"; got != want {
		t.Errorf("GitlabNewMRURL = %q, attendu %q", got, want)
	}
}

// TestDetectForgeFromRemote vérifie la détection du fournisseur depuis le
// remote origin d'un dépôt réel (lecture seule, aucune commande réseau).
func TestDetectForgeFromRemote(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	cases := []struct{ remote, want string }{
		{"git@github.com:acme/demo.git", "github"},
		{"https://gitlab.com/acme/demo.git", "gitlab"},
		// GitLab auto-hébergé : reconnu au nom d'hôte, comme prévu.
		{"git@gitlab.self-hosted.example:acme/demo.git", "gitlab"},
		{"git@bitbucket.org:acme/demo.git", ""},
		{"", ""},
	}
	for _, c := range cases {
		dir := t.TempDir()
		runTestGit(t, dir, "init")
		if c.remote != "" {
			runTestGit(t, dir, "remote", "add", "origin", c.remote)
		}
		if got := DetectForge(dir).Provider; got != c.want {
			t.Errorf("DetectForge(remote=%q) = %q, attendu %q", c.remote, got, c.want)
		}
	}
}

// TestDeleteTaskInterruptsRunningAgentAndRemovesWorktree reproduit le flux de
// handleDeleteTask sur un dépôt git réel : une tâche "running" avec un agent
// simulé en cours (procHandle sans process réel, comme l'adaptateur fake) est
// interrompue puis supprimée ; le worktree est effectivement retiré du dépôt
// d'origine, mais la branche doit toujours exister (jamais supprimée).
func TestDeleteTaskInterruptsRunningAgentAndRemovesWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}

	repo := t.TempDir()
	dataDir := t.TempDir()
	runTestGit(t, repo, "init", "-b", "main")
	runTestGit(t, repo, "config", "user.email", "test@example.com")
	runTestGit(t, repo, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(repo, "README.md"), []byte("# projet\n"), 0o644); err != nil {
		t.Fatalf("écriture README impossible : %v", err)
	}
	runTestGit(t, repo, "add", "-A")
	runTestGit(t, repo, "commit", "-m", "initial")

	const branch = "sillage/100-demo"
	dir, base, err := CreateWorktree(repo, dataDir, "t1", branch, "")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}

	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("demo", "", "", []Repo{{Name: "demo", Path: repo}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	taskID, ref := s.ReserveTaskID()
	if _, err := s.CreateTask(taskID, ref, card.ID, project.ID, "Demo", "echo", branch, base, dir, "demo"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, err := s.UpdateTask(taskID, func(tk *Task) { tk.Status = "running" }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	runner := NewRunner(s, NewHub())
	// Simule un agent en cours d'exécution (sans process réel, comme l'adaptateur fake).
	handle := &procHandle{done: make(chan struct{})}
	runner.mu.Lock()
	runner.procs[taskID] = handle
	runner.mu.Unlock()

	if err := runner.DeleteTask(taskID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	if !handle.interrupted.Load() {
		t.Fatalf("l'agent en cours aurait dû être marqué interrompu")
	}
	if _, ok := s.GetTask(taskID); ok {
		t.Fatalf("la tâche ne devrait plus exister après suppression")
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("le worktree devrait avoir été retiré, reçu err=%v", err)
	}
	// La branche ne doit JAMAIS être supprimée (elle peut avoir été poussée).
	out := runTestGit(t, repo, "branch", "--list", branch)
	if !strings.Contains(out, "100-demo") {
		t.Fatalf("la branche %q devrait toujours exister, reçu : %q", branch, out)
	}
}

// TestCreateTaskWaitsForDependency : une tâche créée avec waitsForTaskId reste
// "waiting" (agent non lancé) jusqu'à l'acceptation de la tâche référencée, qui
// la démarre alors automatiquement.
func TestCreateTaskWaitsForDependency(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})

	depID, _ := f.addTask(t, "Backend", "backend.txt", "backend\n")

	body := fmt.Sprintf(`{"cardId":%q,"title":"Frontend","agentId":"echo","waitsForTaskId":%q}`, f.card.ID, depID)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	f.srv.handleCreateTask(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var waiting Task
	if err := json.Unmarshal(w.Body.Bytes(), &waiting); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if waiting.Status != "waiting" || waiting.WaitsForTaskID != depID {
		t.Fatalf("tâche attendue 'waiting' en attente de %q, reçu statut %q waitsFor %q", depID, waiting.Status, waiting.WaitsForTaskID)
	}
	if f.srv.runner.IsRunning(waiting.ID) {
		t.Fatalf("l'agent ne devrait pas encore tourner")
	}

	if w := f.accept(t, depID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}

	started, ok := f.srv.store.GetTask(waiting.ID)
	if !ok || started.Status != "running" {
		t.Fatalf("la tâche en attente aurait dû démarrer après l'acceptation, statut %q", started.Status)
	}
	if started.WaitsForTaskID != "" || started.PendingPrompt != "" {
		t.Fatalf("waitsForTaskId/pendingPrompt auraient dû être vidés, reçu %+v", started)
	}
	if !f.srv.runner.IsRunning(waiting.ID) {
		t.Fatalf("l'agent devrait maintenant tourner")
	}

	// Libère vite l'agent simulé plutôt que d'attendre ses ~3 s.
	if _, err := f.srv.runner.Interrupt(waiting.ID); err != nil {
		t.Fatalf("Interrupt: %v", err)
	}
}

// TestCreateTaskWaitsForInvalidDependency : waitsForTaskId doit référencer une
// tâche du même chantier, pas déjà terminale.
func TestCreateTaskWaitsForInvalidDependency(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	depID, _ := f.addTask(t, "Backend", "backend.txt", "backend\n")

	otherCard, err := f.srv.store.AddCard(f.project.ID, "Autre chantier", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	body := fmt.Sprintf(`{"cardId":%q,"title":"Frontend","agentId":"echo","waitsForTaskId":%q}`, otherCard.ID, depID)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	f.srv.handleCreateTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("dépendance d'un autre chantier : attendu 400, reçu %d (%s)", w.Code, w.Body.String())
	}

	if w := f.accept(t, depID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	body = fmt.Sprintf(`{"cardId":%q,"title":"Frontend","agentId":"echo","waitsForTaskId":%q}`, f.card.ID, depID)
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w = httptest.NewRecorder()
	f.srv.handleCreateTask(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("dépendance déjà acceptée : attendu 400, reçu %d (%s)", w.Code, w.Body.String())
	}
}

// TestStartWaitingTaskManualOverride : POST /api/tasks/{id}/start démarre une
// tâche "waiting" sans attendre l'acceptation de sa dépendance (dépendance
// refusée/supprimée, ou changement d'avis) ; refusé une fois la tâche démarrée.
func TestStartWaitingTaskManualOverride(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	depID, _ := f.addTask(t, "Backend", "backend.txt", "backend\n")

	body := fmt.Sprintf(`{"cardId":%q,"title":"Frontend","agentId":"echo","waitsForTaskId":%q}`, f.card.ID, depID)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	f.srv.handleCreateTask(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var waiting Task
	if err := json.Unmarshal(w.Body.Bytes(), &waiting); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}

	startReq := httptest.NewRequest(http.MethodPost, "/", nil)
	startReq.SetPathValue("id", waiting.ID)
	startW := httptest.NewRecorder()
	f.srv.handleStartWaitingTask(startW, startReq)
	if startW.Code != http.StatusOK {
		t.Fatalf("start: attendu 200, reçu %d (%s)", startW.Code, startW.Body.String())
	}
	started, _ := f.srv.store.GetTask(waiting.ID)
	if started.Status != "running" {
		t.Fatalf("statut attendu 'running', reçu %q", started.Status)
	}

	// La dépendance reste en revue, jamais acceptée : le démarrage manuel a
	// bien contourné l'attente.
	startW2 := httptest.NewRecorder()
	f.srv.handleStartWaitingTask(startW2, startReq)
	if startW2.Code != http.StatusBadRequest {
		t.Fatalf("second start: attendu 400 (n'est plus 'waiting'), reçu %d", startW2.Code)
	}

	if _, err := f.srv.runner.Interrupt(waiting.ID); err != nil {
		t.Fatalf("Interrupt: %v", err)
	}
}

// TestCancelWaitingTask : une tâche "waiting" peut être refusée avant même
// d'avoir démarré, et l'acceptation ultérieure de sa dépendance ne la
// ressuscite pas (startDependentTasks ne cible que les tâches encore "waiting").
func TestCancelWaitingTask(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	depID, _ := f.addTask(t, "Backend", "backend.txt", "backend\n")

	body := fmt.Sprintf(`{"cardId":%q,"title":"Frontend","agentId":"echo","waitsForTaskId":%q}`, f.card.ID, depID)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	f.srv.handleCreateTask(w, req)
	var waiting Task
	if err := json.Unmarshal(w.Body.Bytes(), &waiting); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}

	cancelled, err := f.srv.runner.Cancel(waiting.ID)
	if err != nil {
		t.Fatalf("Cancel: %v", err)
	}
	if cancelled.Status != "cancelled" {
		t.Fatalf("statut attendu 'cancelled', reçu %q", cancelled.Status)
	}

	if w := f.accept(t, depID); w.Code != http.StatusOK {
		t.Fatalf("accept: attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	after, _ := f.srv.store.GetTask(waiting.ID)
	if after.Status != "cancelled" {
		t.Fatalf("la tâche annulée ne devrait pas redémarrer, statut %q", after.Status)
	}
}
