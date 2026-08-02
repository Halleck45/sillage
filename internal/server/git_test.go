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

	dir, base, err := CreateWorktree(repo, dataDir, "t1", "atelier/100-test")
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
