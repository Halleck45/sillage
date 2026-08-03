package server

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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

	dir, base, err := CreateWorktree(repo, dataDir, "t1", "sillage/100-test")
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

// TestShipFromReviewAddsMarkerMessage reproduit le flux de handleShip :
// ship accepté directement depuis "review" (l'étape "ready" a disparu),
// puis message marqueur "[shipped:<branch>]" (author=agent, authorName vide).
func TestShipFromReviewAddsMarkerMessage(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}

	bare := t.TempDir()
	runTestGit(t, bare, "init", "--bare", "-b", "main")

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
	runTestGit(t, repo, "remote", "add", "origin", bare)

	dir, base, err := CreateWorktree(repo, dataDir, "t1", "sillage/100-demo")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "feature.txt"), []byte("contenu\n"), 0o644); err != nil {
		t.Fatalf("écriture feature.txt impossible : %v", err)
	}

	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("demo", "", "", []Repo{{Path: repo}}, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := s.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	taskID, ref := s.ReserveTaskID()
	task, err := s.CreateTask(taskID, ref, card.ID, project.ID, "Demo", "echo", "sillage/100-demo", base, dir, "demo")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, err := s.UpdateTask(taskID, func(tk *Task) { tk.Status = "review" }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	task, _ = s.GetTask(taskID)
	if task.Status != "review" {
		t.Fatalf("précondition : la tâche devrait être en review")
	}

	// Reproduit le flux de handleShip : ship accepté depuis "review".
	output, err := Ship(task.WorktreeDir, task.Branch, task.Title)
	if err != nil {
		t.Fatalf("Ship: %v (output=%s)", err, output)
	}
	task, err = s.UpdateTask(taskID, func(tk *Task) { tk.Status = "shipped" })
	if err != nil {
		t.Fatalf("UpdateTask (shipped): %v", err)
	}
	if task.Status != "shipped" {
		t.Fatalf("statut attendu 'shipped', reçu %q", task.Status)
	}

	msg, task, err := s.AddMessage(task.ID, "agent", "", "[shipped:"+task.Branch+"]")
	if err != nil {
		t.Fatalf("AddMessage: %v", err)
	}
	wantText := "[shipped:sillage/100-demo]"
	if msg.Author != "agent" || msg.AuthorName != "" || msg.Text != wantText {
		t.Fatalf("message marqueur inattendu : %+v", msg)
	}

	if branchURL := githubBranchURL(task.WorktreeDir, task.Branch); branchURL != "" {
		t.Fatalf("branchUrl attendu vide (remote non-GitHub), reçu %q", branchURL)
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
	dir, base, err := CreateWorktree(repo, dataDir, "t1", branch)
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}

	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	project, err := s.AddProject("demo", "", "", []Repo{{Name: "demo", Path: repo}}, nil)
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
