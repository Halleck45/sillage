package server

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// workspaceGitignore est le contenu figé du .gitignore de l'espace de
// travail : seuls state.json, config.json et .gitignore sont versionnés.
const workspaceGitignore = "worktrees/\n*.tmp\n"

// WorkspaceGitEnabled indique si dataDir est initialisé comme dépôt git.
func WorkspaceGitEnabled(dataDir string) bool {
	info, err := os.Stat(filepath.Join(dataDir, ".git"))
	return err == nil && info.IsDir()
}

// WorkspaceDirty indique s'il existe des changements non commités dans
// dataDir (commit auto en attente ou modifications non commitées).
func WorkspaceDirty(dataDir string) bool {
	out, err := runGit(dataDir, 10*time.Second, "status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

// WorkspaceLastCommitAt retourne la date du dernier commit de dataDir, ou nil
// s'il n'y en a aucun (ou si dataDir n'est pas un dépôt git).
func WorkspaceLastCommitAt(dataDir string) *time.Time {
	out, err := runGit(dataDir, 10*time.Second, "log", "-1", "--format=%cI")
	if err != nil {
		return nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, out)
	if err != nil {
		return nil
	}
	return &t
}

// workspaceTrackedFiles retourne, parmi state.json/config.json/.gitignore,
// ceux qui existent réellement sur disque dans dataDir : `git add` échoue si
// on lui passe un chemin inexistant (ex : config.json absent quand
// SILLAGE_PASSWORD est utilisée, qui ne l'écrit jamais sur disque).
func workspaceTrackedFiles(dataDir string) []string {
	candidates := []string{"state.json", "config.json", ".gitignore"}
	var out []string
	for _, name := range candidates {
		if _, err := os.Stat(filepath.Join(dataDir, name)); err == nil {
			out = append(out, name)
		}
	}
	return out
}

// gitAddWorkspaceFiles ajoute à l'index les fichiers versionnés de l'espace
// de travail qui existent réellement sur disque (jamais -A, jamais de chemin
// inexistant).
func gitAddWorkspaceFiles(dataDir string) (string, error) {
	files := workspaceTrackedFiles(dataDir)
	if len(files) == 0 {
		return "", nil
	}
	args := append([]string{"add"}, files...)
	return runGit(dataDir, gitDefaultTimeout, args...)
}

// SetWorkspaceRemote définit (ou remplace) le remote "origin" de dataDir.
func SetWorkspaceRemote(dataDir, remote string) error {
	if _, err := runGit(dataDir, gitDefaultTimeout, "remote", "get-url", "origin"); err == nil {
		if _, err := runGit(dataDir, gitDefaultTimeout, "remote", "set-url", "origin", remote); err != nil {
			return fmt.Errorf("git remote set-url failed: %w", err)
		}
		return nil
	}
	if _, err := runGit(dataDir, gitDefaultTimeout, "remote", "add", "origin", remote); err != nil {
		return fmt.Errorf("git remote add failed: %w", err)
	}
	return nil
}

// InitWorkspaceGit initialise dataDir comme dépôt git (branche main), écrit
// .gitignore, crée un premier commit si nécessaire, et définit origin si
// remote est fourni. N'exécute jamais de push.
func InitWorkspaceGit(dataDir, remote string) error {
	if !WorkspaceGitEnabled(dataDir) {
		if _, err := runGit(dataDir, gitDefaultTimeout, "init", "-b", "main"); err != nil {
			return fmt.Errorf("git init failed: %w", err)
		}
		// Identité locale par défaut : ne jamais dépendre d'une config git
		// globale absente (poste de dev minimal, CI...).
		_, _ = runGit(dataDir, gitDefaultTimeout, "config", "user.email", "sillage@localhost")
		_, _ = runGit(dataDir, gitDefaultTimeout, "config", "user.name", "Sillage")
	}

	gitignorePath := filepath.Join(dataDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte(workspaceGitignore), 0o644); err != nil {
			return fmt.Errorf("failed to write .gitignore: %w", err)
		}
	}

	if _, err := gitAddWorkspaceFiles(dataDir); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}
	status, err := runGit(dataDir, gitDefaultTimeout, "status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) != "" {
		if _, err := runGit(dataDir, gitDefaultTimeout, "commit", "-m", "sillage: init workspace"); err != nil {
			return fmt.Errorf("git commit failed: %w", err)
		}
	}

	if remote != "" {
		if err := SetWorkspaceRemote(dataDir, remote); err != nil {
			return err
		}
	}
	return nil
}

// commitWorkspace committe silencieusement state.json/config.json/.gitignore
// si dataDir est un dépôt git et qu'il y a des changements en attente.
// Jamais de push : appelée par le debounce automatique après chaque
// sauvegarde du state (voir Store.scheduleWorkspaceCommit).
func commitWorkspace(dataDir string) {
	if !WorkspaceGitEnabled(dataDir) {
		return
	}
	if _, err := gitAddWorkspaceFiles(dataDir); err != nil {
		return
	}
	status, err := runGit(dataDir, gitDefaultTimeout, "status", "--porcelain")
	if err != nil || strings.TrimSpace(status) == "" {
		return
	}
	_, _ = runGit(dataDir, gitDefaultTimeout, "commit", "-m", "sillage: update")
}

// CloneWorkspace clone remote dans un répertoire temporaire voisin de
// dataDir (même système de fichiers, pour un déplacement atomique ensuite)
// et vérifie qu'il s'agit bien d'un espace de travail Sillage (state.json à
// la racine). Ne modifie pas encore dataDir : voir ReplaceWorkspaceFiles.
func CloneWorkspace(dataDir, remote string) (cloneDir string, err error) {
	parent := filepath.Dir(dataDir)
	cloneDir, err = os.MkdirTemp(parent, "sillage-clone-*")
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitPushTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "clone", remote, cloneDir)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, cloneErr := cmd.CombinedOutput()
	if cloneErr != nil {
		os.RemoveAll(cloneDir)
		return "", fmt.Errorf("git clone failed: %w: %s", cloneErr, strings.TrimSpace(string(out)))
	}

	if _, statErr := os.Stat(filepath.Join(cloneDir, "state.json")); statErr != nil {
		os.RemoveAll(cloneDir)
		return "", fmt.Errorf("remote does not look like a Sillage workspace")
	}
	return cloneDir, nil
}

// ReplaceWorkspaceFiles remplace state.json, config.json et .git de dataDir
// par ceux du clone, puis supprime le répertoire temporaire du clone.
func ReplaceWorkspaceFiles(dataDir, cloneDir string) error {
	for _, name := range []string{"state.json", "config.json", ".git"} {
		_ = os.RemoveAll(filepath.Join(dataDir, name))
	}
	for _, name := range []string{"state.json", "config.json", ".git"} {
		src := filepath.Join(cloneDir, name)
		if _, err := os.Stat(src); err != nil {
			continue // config.json (ou autre) peut être absent d'un vieux clone : pas bloquant
		}
		if err := os.Rename(src, filepath.Join(dataDir, name)); err != nil {
			return fmt.Errorf("failed to move %s from clone: %w", name, err)
		}
	}
	return os.RemoveAll(cloneDir)
}

// workspaceStatus assemble l'objet exposé par GET /api/workspace, l'événement
// SSE "workspace" et le champ State.Workspace : état persisté (Store) combiné
// à des faits git calculés à la volée sur s.dataDir.
func (s *Server) workspaceStatus() WorkspaceStatus {
	ws := s.store.GetWorkspace()
	gitEnabled := WorkspaceGitEnabled(s.dataDir)
	status := WorkspaceStatus{
		SetupDone:  ws.SetupDone,
		GitEnabled: gitEnabled,
		Remote:     ws.SyncRemote,
		LastSyncAt: ws.LastSyncAt,
	}
	if gitEnabled {
		status.Dirty = WorkspaceDirty(s.dataDir)
		status.LastCommitAt = WorkspaceLastCommitAt(s.dataDir)
	}
	return status
}
