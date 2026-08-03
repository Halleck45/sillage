package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
	"time"
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

// --- Migration des statuts disparus ("ready" -> review, "shipped"/"done" -> accepted) ---

func TestMigrateTaskStatuses(t *testing.T) {
	dir := t.TempDir()
	legacy := `{
  "Projects": {"p1": {"id":"p1","name":"p","repos":[{"name":"p","path":"/tmp/p"}],"unread":0,"tokens":{"input":0,"output":0,"costUsd":0},"checkCmd":""}},
  "Cards": {"c1": {"id":"c1","projectId":"p1","column":"doing","title":"Carte"}},
  "Tasks": {
    "t1": {"id":"t1","cardId":"c1","projectId":"p1","ref":100,"title":"T","agentId":"echo","branch":"sillage/100-t","status":"ready","base":"main","worktreeDir":"/tmp/wt"},
    "t2": {"id":"t2","cardId":"c1","projectId":"p1","ref":101,"title":"T2","agentId":"echo","branch":"sillage/101-t2","status":"shipped","base":"main","worktreeDir":"/tmp/wt2"}
  },
  "Messages": {}, "Agents": {},
  "NextProjectN": 1, "NextCardN": 1, "NextTaskN": 2, "NextMessageN": 0, "NextRef": 102
}`
	if err := os.WriteFile(filepath.Join(dir, "state.json"), []byte(legacy), 0o644); err != nil {
		t.Fatalf("écriture state.json legacy impossible : %v", err)
	}

	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t1, ok := s.GetTask("t1")
	if !ok {
		t.Fatalf("tâche t1 introuvable après migration")
	}
	if t1.Status != "review" {
		t.Fatalf("statut migré attendu 'review', reçu %q", t1.Status)
	}
	// Une tâche livrée seule (ancien modèle) devient une tâche acceptée : la
	// livraison est désormais une action de chantier.
	t2, ok := s.GetTask("t2")
	if !ok || t2.Status != "accepted" {
		t.Fatalf("statut migré attendu 'accepted' pour une tâche 'shipped', reçu %+v (ok=%v)", t2, ok)
	}
	// Le chantier reçoit une référence : sans elle, sa branche s'appellerait
	// `sillage/ws-0-<slug>`.
	c1, ok := s.GetCard("c1")
	if !ok || c1.Ref == 0 {
		t.Fatalf("le chantier devrait avoir reçu une référence, reçu %+v (ok=%v)", c1, ok)
	}
	// Le mode de livraison par défaut est celui d'avant : ouvrir une PR.
	p1, ok := s.GetProject("p1")
	if !ok || p1.Delivery.Mode != "pr" {
		t.Fatalf("mode de livraison migré attendu 'pr', reçu %+v (ok=%v)", p1.Delivery, ok)
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

	single, err := s.AddProject("mono", "", "", []Repo{{Path: "/tmp/mono"}}, nil, nil)
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

	multi, err := s.AddProject("multi", "", "", []Repo{{Name: "api", Path: "/tmp/api"}, {Name: "web", Path: "/tmp/web"}}, nil, nil)
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
	// gc.auto bas : chaque commit écrit un blob complet de state.json, il faut
	// que git compacte souvent au lieu d'attendre son seuil de 6700 objets.
	gcAuto, err := runGit(dir, gitDefaultTimeout, "config", "--get", "gc.auto")
	if err != nil {
		t.Fatalf("config --get gc.auto: %v", err)
	}
	if strings.TrimSpace(gcAuto) != "256" {
		t.Fatalf("gc.auto attendu 256, reçu %q", gcAuto)
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

// --- Fréquence du commit automatique ---

// currentCommitTimer lit le minuteur de commit sous verrou.
func currentCommitTimer(s *Store) *time.Timer {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.commitTimer
}

// TestWorkspaceCommitThrottledNotDebounced vérifie que des sauvegardes
// rapprochées ne repoussent pas le commit automatique : le minuteur armé par
// la première doit survivre aux suivantes (throttle). Avec un debounce, une
// activité continue d'agent (plusieurs sauvegardes par seconde) réarmerait le
// minuteur sans fin et ne commiterait jamais ; à l'inverse, commiter à chaque
// sauvegarde gonflerait le dépôt en objets libres.
func TestWorkspaceCommitThrottledNotDebounced(t *testing.T) {
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	// NewStore sauvegarde déjà une fois : le minuteur doit être armé.
	first := currentCommitTimer(s)
	if first == nil {
		t.Fatalf("le minuteur de commit devrait être armé après la première sauvegarde")
	}

	// Plusieurs mutations rapprochées, comme un agent qui travaille.
	project, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if _, err := s.AddCard(project.ID, "Ma carte", "", ""); err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	if _, err := s.UpdateSettings(nil, nil); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	if got := currentCommitTimer(s); got != first {
		t.Fatalf("le minuteur a été réarmé par les sauvegardes suivantes (debounce) : attendu %p, reçu %p", first, got)
	}

	// L'intervalle doit rester assez large pour que le nombre de commits sur
	// une journée de travail reste faible.
	if workspaceCommitInterval < time.Minute {
		t.Fatalf("workspaceCommitInterval trop court : %v", workspaceCommitInterval)
	}
}

// waitForCommitCount attend que dataDir atteigne le nombre de commits attendu.
func waitForCommitCount(t *testing.T, dataDir string, want int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var got string
	for time.Now().Before(deadline) {
		got = strings.TrimSpace(runTestGit(t, dataDir, "rev-list", "--count", "HEAD"))
		if got == strconv.Itoa(want) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("%d commits attendus dans l'espace de travail, reçu %q", want, got)
}

// TestWorkspaceCommitFiresOncePerInterval vérifie le mécanisme complet sur un
// espace de travail réellement versionné : plusieurs mutations dans le même
// créneau ne produisent qu'UN commit, et un créneau suivant en produit un
// nouveau (le minuteur se réarme après avoir tiré).
func TestWorkspaceCommitFiresOncePerInterval(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	dir := t.TempDir()

	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := InitWorkspaceGit(dir, ""); err != nil {
		t.Fatalf("InitWorkspaceGit: %v", err)
	}
	baseCount := strings.TrimSpace(runTestGit(t, dir, "rev-list", "--count", "HEAD"))
	if baseCount != "1" {
		t.Fatalf("un seul commit initial attendu, reçu %q", baseCount)
	}

	// Créneau court pour le test, et on repart d'un minuteur non armé.
	s.mu.Lock()
	if s.commitTimer != nil {
		s.commitTimer.Stop()
		s.commitTimer = nil
	}
	s.commitInterval = 40 * time.Millisecond
	s.mu.Unlock()

	// Trois mutations dans le même créneau : un seul commit attendu.
	project, err := s.AddProject("sillage", "", "", []Repo{{Path: "/tmp/sillage"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if _, err := s.AddCard(project.ID, "Ma carte", "", ""); err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	if _, err := s.UpdateSettings(nil, nil); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	waitForCommitCount(t, dir, 2)

	// Le créneau est libéré après le tir : une nouvelle mutation en réarme un.
	if got := currentCommitTimer(s); got != nil {
		t.Fatalf("le minuteur devrait être libéré après le commit, reçu %p", got)
	}
	if _, err := s.AddCard(project.ID, "Autre carte", "", ""); err != nil {
		t.Fatalf("AddCard (second créneau): %v", err)
	}
	waitForCommitCount(t, dir, 3)
}

// --- Synchronisation automatique de l'espace de travail (autoSync) ---

// newAutoSyncTestServer construit un Server dont le dataDir du Store et celui
// du Server sont le MÊME répertoire (contrairement à newTestServer, qui les
// isole volontairement pour les tests de confirmation) : nécessaire ici, les
// vérifications git (WorkspaceGitEnabled, InitWorkspaceGit...) et le Store
// doivent porter sur le même dossier, comme en production (main.go).
func newAutoSyncTestServer(t *testing.T) (*Server, string) {
	t.Helper()
	dataDir := t.TempDir()
	s, err := NewStore(dataDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	srv := NewServer(s, "", dataDir, fstest.MapFS{})
	t.Cleanup(srv.stopAutoSync)
	return srv, dataDir
}

func TestUpdateWorkspaceAutoSyncRequiresGitAndRemote(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	srv, dataDir := newAutoSyncTestServer(t)

	// Ni git ni remote : refusé.
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"autoSync":true}`))
	w := httptest.NewRecorder()
	srv.handleUpdateWorkspace(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("attendu 400 (ni git ni remote), reçu %d (body=%s)", w.Code, w.Body.String())
	}

	// git initialisé mais sans remote : toujours refusé.
	if err := InitWorkspaceGit(dataDir, ""); err != nil {
		t.Fatalf("InitWorkspaceGit: %v", err)
	}
	req2 := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"autoSync":true}`))
	w2 := httptest.NewRecorder()
	srv.handleUpdateWorkspace(w2, req2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("attendu 400 (pas de remote), reçu %d (body=%s)", w2.Code, w2.Body.String())
	}
	if ws := srv.store.GetWorkspace(); ws.AutoSync {
		t.Fatalf("autoSync ne devrait pas avoir été activé")
	}

	// git + remote présents : accepté.
	req3 := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"remote":"`+dataDir+`","autoSync":true}`))
	w3 := httptest.NewRecorder()
	srv.handleUpdateWorkspace(w3, req3)
	if w3.Code != http.StatusOK {
		t.Fatalf("attendu 200 (git+remote fournis dans le même appel), reçu %d (body=%s)", w3.Code, w3.Body.String())
	}
	if ws := srv.store.GetWorkspace(); !ws.AutoSync {
		t.Fatalf("autoSync devrait être activé")
	}
}

func TestAutoSyncTickSuccessUpdatesLastSyncAt(t *testing.T) {
	srv, _ := newAutoSyncTestServer(t)
	if _, err := srv.store.UpdateWorkspace(func(w *Workspace) {
		w.SetupDone, w.SyncRemote, w.AutoSync = true, "origin", true
	}); err != nil {
		t.Fatalf("UpdateWorkspace: %v", err)
	}

	orig := syncPushFn
	defer func() { syncPushFn = orig }()
	gitEnabled := workspaceGitEnabledFn
	defer func() { workspaceGitEnabledFn = gitEnabled }()
	workspaceGitEnabledFn = func(string) bool { return true }
	syncPushFn = func(string) (string, error) { return "ok", nil }

	srv.autoSyncTick()

	ws := srv.store.GetWorkspace()
	if ws.LastSyncAt == nil {
		t.Fatalf("lastSyncAt devrait être renseigné après un tick réussi")
	}
	if got := srv.getLastSyncError(); got != "" {
		t.Fatalf("lastSyncError devrait être vide après un tick réussi, reçu %q", got)
	}
}

func TestAutoSyncTickFailureSetsLastSyncError(t *testing.T) {
	srv, _ := newAutoSyncTestServer(t)
	if _, err := srv.store.UpdateWorkspace(func(w *Workspace) {
		w.SetupDone, w.SyncRemote, w.AutoSync = true, "origin", true
	}); err != nil {
		t.Fatalf("UpdateWorkspace: %v", err)
	}

	orig := syncPushFn
	defer func() { syncPushFn = orig }()
	gitEnabled := workspaceGitEnabledFn
	defer func() { workspaceGitEnabledFn = gitEnabled }()
	workspaceGitEnabledFn = func(string) bool { return true }
	syncPushFn = func(string) (string, error) { return "", fmt.Errorf("network unreachable") }

	srv.autoSyncTick()

	got := srv.getLastSyncError()
	if got == "" {
		t.Fatalf("lastSyncError devrait être renseigné après un tick en échec")
	}
	if strings.Contains(got, ErrSyncConflict.Error()) {
		t.Fatalf("cet échec ne devrait pas être signalé comme un conflit, reçu %q", got)
	}
	if ws := srv.store.GetWorkspace(); !ws.AutoSync || ws.LastSyncAt != nil {
		t.Fatalf("autoSync doit rester actif et lastSyncAt rester vide après un échec non-conflit : %+v", ws)
	}
}

func TestAutoSyncTickPausesOnConflictUntilManualSyncSucceeds(t *testing.T) {
	srv, _ := newAutoSyncTestServer(t)
	if _, err := srv.store.UpdateWorkspace(func(w *Workspace) {
		w.SetupDone, w.SyncRemote, w.AutoSync = true, "origin", true
	}); err != nil {
		t.Fatalf("UpdateWorkspace: %v", err)
	}

	origSync := syncPushFn
	defer func() { syncPushFn = origSync }()
	gitEnabled := workspaceGitEnabledFn
	defer func() { workspaceGitEnabledFn = gitEnabled }()
	workspaceGitEnabledFn = func(string) bool { return true }

	calls := 0
	syncPushFn = func(dir string) (string, error) {
		calls++
		return "", fmt.Errorf("%w: the remote workspace diverged, resolve manually in %s", ErrSyncConflict, dir)
	}

	srv.autoSyncTick()
	if calls != 1 {
		t.Fatalf("le premier tick devrait appeler la synchronisation une fois, reçu %d appel(s)", calls)
	}
	if got := srv.getLastSyncError(); !strings.Contains(got, ErrSyncConflict.Error()) {
		t.Fatalf("lastSyncError devrait signaler un conflit, reçu %q", got)
	}

	// Tant que le conflit n'est pas résolu, le tick suivant ne doit rien tenter.
	srv.autoSyncTick()
	if calls != 1 {
		t.Fatalf("le tick devrait être en pause pendant un conflit non résolu, appels=%d", calls)
	}
	if ws := srv.store.GetWorkspace(); !ws.AutoSync {
		t.Fatalf("autoSync ne doit pas être désactivé par un conflit : la pause n'est que temporaire")
	}

	// Une synchronisation manuelle réussie efface l'erreur et relance (ici
	// simulée directement : c'est ce que fait handleWorkspaceSync en cas de
	// succès).
	srv.setLastSyncError("")
	syncPushFn = func(string) (string, error) { calls++; return "output", nil }

	srv.autoSyncTick()
	if calls != 2 {
		t.Fatalf("le tick devrait reprendre après la résolution du conflit, appels=%d", calls)
	}
	if ws := srv.store.GetWorkspace(); ws.LastSyncAt == nil {
		t.Fatalf("lastSyncAt devrait être renseigné après la reprise réussie")
	}
}
