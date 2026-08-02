package server

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// --- Migration des dépôts (path legacy -> repos) ---

func TestMigrateProjectPathToRepos(t *testing.T) {
	dir := t.TempDir()
	legacy := `{
  "Projects": {"p1": {"id":"p1","name":"sillage","path":"/tmp/sillage","unread":0,"tokens":{"input":0,"output":0,"costUsd":0},"checkCmd":""}},
  "Cards": {}, "Tasks": {}, "Messages": {}, "Agents": {},
  "NextProjectN": 1, "NextCardN": 0, "NextTaskN": 0, "NextMessageN": 0, "NextRef": 100
}`
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(legacy), 0o644); err != nil {
		t.Fatalf("écriture state.json legacy impossible : %v", err)
	}

	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, ok := s.GetProject("p1")
	if !ok {
		t.Fatalf("projet p1 introuvable après migration")
	}
	if len(p.Repos) != 1 || p.Repos[0].Name != "sillage" || p.Repos[0].Path != "/tmp/sillage" {
		t.Fatalf("migration path->repos inattendue : %+v", p.Repos)
	}
}

// --- Migration de l'espace de travail (workspace absent -> setupDone=true local) ---

func TestMigrateLegacyWorkspaceSetupDoneLocal(t *testing.T) {
	dir := t.TempDir()
	legacy := `{
  "Projects": {}, "Cards": {}, "Tasks": {}, "Messages": {}, "Agents": {},
  "NextProjectN": 0, "NextCardN": 0, "NextTaskN": 0, "NextMessageN": 0, "NextRef": 100
}`
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(legacy), 0o644); err != nil {
		t.Fatalf("écriture state.json legacy impossible : %v", err)
	}

	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if !s.GetWorkspace().SetupDone {
		t.Fatalf("un espace déjà utilisé (sans champ Workspace) devrait migrer vers setupDone=true")
	}
}

// Contrôle : une installation neuve (aucun state.json) ne doit pas marquer
// setupDone : l'onboarding doit s'afficher.
func TestNewInstallHasOnboardingPending(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if s.GetWorkspace().SetupDone {
		t.Fatalf("une installation neuve ne devrait pas avoir setupDone=true (onboarding attendu)")
	}
}

// --- Validation des dépôts ---

func TestNormalizeReposValidation(t *testing.T) {
	if _, err := NormalizeRepos(nil); err == nil {
		t.Fatalf("une liste vide devrait être refusée")
	}
	if _, err := NormalizeRepos([]Repo{{Path: ""}}); err == nil {
		t.Fatalf("un chemin vide devrait être refusé")
	}
	if _, err := NormalizeRepos([]Repo{{Name: "x", Path: "/a"}, {Name: "x", Path: "/b"}}); err == nil {
		t.Fatalf("des noms dupliqués devraient être refusés")
	}
	out, err := NormalizeRepos([]Repo{{Path: "/tmp/mon-projet"}})
	if err != nil {
		t.Fatalf("NormalizeRepos: %v", err)
	}
	if out[0].Name != "mon-projet" {
		t.Fatalf("nom par défaut attendu 'mon-projet', reçu %q", out[0].Name)
	}
}

func TestValidateRepoPath(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}

	notADir := filepath.Join(t.TempDir(), "absent")
	if err := ValidateRepoPath(notADir); err == nil {
		t.Fatalf("un chemin inexistant devrait être refusé")
	}

	notAGitRepo := t.TempDir()
	if err := ValidateRepoPath(notAGitRepo); err == nil {
		t.Fatalf("un répertoire qui n'est pas un dépôt git devrait être refusé")
	}

	gitRepo := t.TempDir()
	runTestGit(t, gitRepo, "init")
	if err := ValidateRepoPath(gitRepo); err != nil {
		t.Fatalf("un dépôt git valide devrait être accepté : %v", err)
	}
}

func TestResolveTaskRepo(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	single, err := s.AddProject("mono", []Repo{{Path: "/tmp/mono"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	repo, err := s.ResolveTaskRepo(single.ID, "")
	if err != nil {
		t.Fatalf("ResolveTaskRepo (repo unique) : %v", err)
	}
	if repo.Path != "/tmp/mono" {
		t.Fatalf("repo attendu /tmp/mono, reçu %+v", repo)
	}

	multi, err := s.AddProject("multi", []Repo{{Name: "api", Path: "/tmp/api"}, {Name: "web", Path: "/tmp/web"}})
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if _, err := s.ResolveTaskRepo(multi.ID, ""); err == nil {
		t.Fatalf("repoName manquant avec plusieurs repos devrait être refusé")
	} else if err.Error() != "repoName required (project has several repositories)" {
		t.Fatalf("message d'erreur inattendu : %q", err.Error())
	}
	if _, err := s.ResolveTaskRepo(multi.ID, "inconnu"); err == nil {
		t.Fatalf("un repoName inconnu devrait être refusé")
	}
	repo, err = s.ResolveTaskRepo(multi.ID, "web")
	if err != nil {
		t.Fatalf("ResolveTaskRepo (web) : %v", err)
	}
	if repo.Path != "/tmp/web" {
		t.Fatalf("repo attendu /tmp/web, reçu %+v", repo)
	}
}

// --- Setup de l'espace de travail (local / init) ---

func TestInitWorkspaceGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("écriture state.json impossible : %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("écriture config.json impossible : %v", err)
	}

	if err := InitWorkspaceGit(dir, ""); err != nil {
		t.Fatalf("InitWorkspaceGit: %v", err)
	}
	if !WorkspaceGitEnabled(dir) {
		t.Fatalf("dataDir devrait être un dépôt git après InitWorkspaceGit")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gitignore")); err != nil {
		t.Fatalf(".gitignore devrait avoir été créé : %v", err)
	}
	if WorkspaceDirty(dir) {
		t.Fatalf("le dépôt devrait être propre après le premier commit")
	}
	if WorkspaceLastCommitAt(dir) == nil {
		t.Fatalf("un commit initial devrait exister")
	}

	// Ré-exécuter init avec un remote doit rester idempotent et le définir.
	if err := InitWorkspaceGit(dir, "https://example.invalid/workspace.git"); err != nil {
		t.Fatalf("InitWorkspaceGit (avec remote): %v", err)
	}
	out, err := runGit(dir, gitDefaultTimeout, "remote", "get-url", "origin")
	if err != nil {
		t.Fatalf("remote get-url: %v", err)
	}
	if strings.TrimSpace(out) != "https://example.invalid/workspace.git" {
		t.Fatalf("remote attendu https://example.invalid/workspace.git, reçu %q", out)
	}
}

// --- Rapatriement (clone) d'un espace de travail existant ---

// setupBareWorkspaceRemote crée un dépôt bare (le "remote") contenant un
// commit initial avec state.json/config.json/.gitignore, pour simuler un
// espace de travail Sillage déjà existant et synchronisé.
func setupBareWorkspaceRemote(t *testing.T, stateContent string) (bareDir string) {
	t.Helper()
	bareDir = t.TempDir()
	runTestGit(t, bareDir, "init", "--bare", "-b", "main")

	seed := t.TempDir()
	runTestGit(t, seed, "init", "-b", "main")
	runTestGit(t, seed, "config", "user.email", "seed@example.com")
	runTestGit(t, seed, "config", "user.name", "Seed")
	if err := os.WriteFile(filepath.Join(seed, "state.json"), []byte(stateContent), 0o644); err != nil {
		t.Fatalf("écriture state.json impossible : %v", err)
	}
	if err := os.WriteFile(filepath.Join(seed, "config.json"), []byte(`{"passwordHash":""}`), 0o644); err != nil {
		t.Fatalf("écriture config.json impossible : %v", err)
	}
	if err := os.WriteFile(filepath.Join(seed, ".gitignore"), []byte(workspaceGitignore), 0o644); err != nil {
		t.Fatalf("écriture .gitignore impossible : %v", err)
	}
	runTestGit(t, seed, "add", "-A")
	runTestGit(t, seed, "commit", "-m", "seed")
	runTestGit(t, seed, "remote", "add", "origin", bareDir)
	runTestGit(t, seed, "push", "-u", "origin", "main")
	return bareDir
}

func TestWorkspaceCloneRejectsNonSillageRemote(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	bareDir := t.TempDir()
	runTestGit(t, bareDir, "init", "--bare", "-b", "main")

	seed := t.TempDir()
	runTestGit(t, seed, "init", "-b", "main")
	runTestGit(t, seed, "config", "user.email", "seed@example.com")
	runTestGit(t, seed, "config", "user.name", "Seed")
	if err := os.WriteFile(filepath.Join(seed, "readme.txt"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("écriture readme.txt impossible : %v", err)
	}
	runTestGit(t, seed, "add", "-A")
	runTestGit(t, seed, "commit", "-m", "not a sillage workspace")
	runTestGit(t, seed, "remote", "add", "origin", bareDir)
	runTestGit(t, seed, "push", "-u", "origin", "main")

	parent := t.TempDir()
	dataDir := filepath.Join(parent, "data")
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		t.Fatalf("mkdir dataDir: %v", err)
	}

	_, err := CloneWorkspace(dataDir, "file://"+bareDir)
	if err == nil {
		t.Fatalf("le clone d'un remote sans state.json devrait échouer")
	}
	if err.Error() != "remote does not look like a Sillage workspace" {
		t.Fatalf("message d'erreur inattendu : %q", err.Error())
	}
}

func TestWorkspaceCloneAndReload(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	stateJSON := `{"Projects":{"p1":{"id":"p1","name":"projet-distant","repos":[],"unread":0,"tokens":{"input":0,"output":0,"costUsd":0},"checkCmd":""}},"Cards":{},"Tasks":{},"Messages":{},"Agents":{},"NextProjectN":1,"NextCardN":0,"NextTaskN":0,"NextMessageN":0,"NextRef":100}`
	bareDir := setupBareWorkspaceRemote(t, stateJSON)

	parent := t.TempDir()
	dataDir := filepath.Join(parent, "data")
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		t.Fatalf("mkdir dataDir: %v", err)
	}

	// Store existant AVANT le rapatriement : espace local vide.
	s, err := NewStore(dataDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if _, ok := s.GetProject("p1"); ok {
		t.Fatalf("le projet distant ne devrait pas encore exister avant le clone")
	}

	cloneDir, err := CloneWorkspace(dataDir, "file://"+bareDir)
	if err != nil {
		t.Fatalf("CloneWorkspace: %v", err)
	}
	if err := ReplaceWorkspaceFiles(dataDir, cloneDir); err != nil {
		t.Fatalf("ReplaceWorkspaceFiles: %v", err)
	}
	if _, err := os.Stat(cloneDir); !os.IsNotExist(err) {
		t.Fatalf("le répertoire temporaire du clone devrait avoir été supprimé")
	}

	// Rechargement en mémoire du même Store (même pointeur), sans redémarrage.
	if err := s.ReloadFromDisk(); err != nil {
		t.Fatalf("ReloadFromDisk: %v", err)
	}
	p, ok := s.GetProject("p1")
	if !ok || p.Name != "projet-distant" {
		t.Fatalf("le state rapatrié devrait contenir le projet distant, reçu %+v (ok=%v)", p, ok)
	}
	if !WorkspaceGitEnabled(dataDir) {
		t.Fatalf("dataDir devrait être un dépôt git après le clone")
	}
}

// --- Synchronisation (SyncPush) ---

func TestSyncPushCommitsAndPushes(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	bareDir := setupBareWorkspaceRemote(t, `{"v":0}`)

	parent := t.TempDir()
	dataDir := filepath.Join(parent, "data")
	runTestGit(t, parent, "clone", "file://"+bareDir, dataDir)
	runTestGit(t, dataDir, "config", "user.email", "a@example.com")
	runTestGit(t, dataDir, "config", "user.name", "A")

	// Rendre l'espace de travail "sale" : modification non commitée.
	if err := os.WriteFile(filepath.Join(dataDir, "state.json"), []byte(`{"v":1}`), 0o644); err != nil {
		t.Fatalf("écriture state.json impossible : %v", err)
	}
	if !WorkspaceDirty(dataDir) {
		t.Fatalf("le dépôt devrait être sale avant SyncPush")
	}

	if _, err := SyncPush(dataDir); err != nil {
		t.Fatalf("SyncPush: %v", err)
	}
	if WorkspaceDirty(dataDir) {
		t.Fatalf("le dépôt devrait être propre après SyncPush")
	}

	out := runTestGit(t, dataDir, "log", "origin/main", "-1", "--format=%s")
	if strings.TrimSpace(out) != "sillage: update" {
		t.Fatalf("le remote devrait avoir reçu le commit auto, log : %q", out)
	}
}

func TestSyncPushConflict(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	bareDir := setupBareWorkspaceRemote(t, `{"v":0}`)

	parent := t.TempDir()
	dataDirA := filepath.Join(parent, "a")
	dataDirB := filepath.Join(parent, "b")
	runTestGit(t, parent, "clone", "file://"+bareDir, dataDirA)
	runTestGit(t, parent, "clone", "file://"+bareDir, dataDirB)
	for _, d := range []string{dataDirA, dataDirB} {
		runTestGit(t, d, "config", "user.email", "test@example.com")
		runTestGit(t, d, "config", "user.name", "Test")
	}

	// A modifie et pousse en premier (simule une autre machine déjà synchronisée).
	if err := os.WriteFile(filepath.Join(dataDirA, "state.json"), []byte(`{"v":"A"}`), 0o644); err != nil {
		t.Fatalf("écriture A: %v", err)
	}
	runTestGit(t, dataDirA, "add", "-A")
	runTestGit(t, dataDirA, "commit", "-m", "A")
	runTestGit(t, dataDirA, "push", "origin", "main")

	// B modifie la même ligne du même fichier, en local, sans pousser.
	if err := os.WriteFile(filepath.Join(dataDirB, "state.json"), []byte(`{"v":"B"}`), 0o644); err != nil {
		t.Fatalf("écriture B: %v", err)
	}
	runTestGit(t, dataDirB, "add", "-A")
	runTestGit(t, dataDirB, "commit", "-m", "B")

	_, err := SyncPush(dataDirB)
	if err == nil {
		t.Fatalf("SyncPush devrait échouer avec un conflit")
	}
	if !errors.Is(err, ErrSyncConflict) {
		t.Fatalf("erreur attendue ErrSyncConflict, reçu : %v", err)
	}
	wantMsg := "sync conflict: the remote workspace diverged, resolve manually in " + dataDirB
	if err.Error() != wantMsg {
		t.Fatalf("message d'erreur attendu %q, reçu %q", wantMsg, err.Error())
	}
	if isRebaseInProgress(dataDirB) {
		t.Fatalf("le rebase devrait avoir été annulé (git rebase --abort)")
	}
	out := runTestGit(t, dataDirB, "log", "-1", "--format=%s")
	if strings.TrimSpace(out) != "B" {
		t.Fatalf("le commit local de B devrait être préservé après l'abandon, log : %q", out)
	}
}

func TestSyncPushNeverTouchesProjectRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}

	// Dépôt de PROJET, avec son propre remote : ne doit jamais être touché
	// par SyncPush, qui n'opère que sur l'espace de travail (dataDir).
	projectBare := t.TempDir()
	runTestGit(t, projectBare, "init", "--bare", "-b", "main")
	projectRepo := t.TempDir()
	runTestGit(t, projectRepo, "init", "-b", "main")
	runTestGit(t, projectRepo, "config", "user.email", "p@example.com")
	runTestGit(t, projectRepo, "config", "user.name", "P")
	if err := os.WriteFile(filepath.Join(projectRepo, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("écriture main.go impossible : %v", err)
	}
	runTestGit(t, projectRepo, "add", "-A")
	runTestGit(t, projectRepo, "commit", "-m", "project commit")
	runTestGit(t, projectRepo, "remote", "add", "origin", projectBare)
	runTestGit(t, projectRepo, "push", "-u", "origin", "main")
	beforeLog := runTestGit(t, projectBare, "log", "-1", "--format=%H", "main")

	// Espace de travail séparé, avec son propre remote.
	workspaceBare := setupBareWorkspaceRemote(t, `{"v":0}`)
	parent := t.TempDir()
	dataDir := filepath.Join(parent, "data")
	runTestGit(t, parent, "clone", "file://"+workspaceBare, dataDir)
	runTestGit(t, dataDir, "config", "user.email", "w@example.com")
	runTestGit(t, dataDir, "config", "user.name", "W")
	if err := os.WriteFile(filepath.Join(dataDir, "state.json"), []byte(`{"v":1}`), 0o644); err != nil {
		t.Fatalf("écriture state.json impossible : %v", err)
	}

	if _, err := SyncPush(dataDir); err != nil {
		t.Fatalf("SyncPush: %v", err)
	}

	afterLog := runTestGit(t, projectBare, "log", "-1", "--format=%H", "main")
	if beforeLog != afterLog {
		t.Fatalf("le dépôt de projet ne devrait jamais être modifié par SyncPush : avant=%q après=%q", beforeLog, afterLog)
	}
	wsLog := runTestGit(t, dataDir, "log", "origin/main", "-1", "--format=%s")
	if strings.TrimSpace(wsLog) != "sillage: update" {
		t.Fatalf("le remote de l'espace de travail devrait avoir reçu le commit, log : %q", wsLog)
	}
}
