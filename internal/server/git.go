package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	gitDefaultTimeout = 60 * time.Second
	gitPushTimeout    = 120 * time.Second
)

// runGit exécute une commande git dans dir, avec un timeout et l'environnement
// hérité complété de GIT_TERMINAL_PROMPT=0 (jamais de prompt interactif).
func runGit(dir string, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("git command timed out (%s): git %s", timeout, strings.Join(args, " "))
	}
	if err != nil {
		return string(out), fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// IsGitRepo indique si path est la racine (ou un sous-répertoire) d'un dépôt git.
func IsGitRepo(path string) bool {
	_, err := runGit(path, 10*time.Second, "rev-parse", "--git-dir")
	return err == nil
}

// currentBranch retourne la branche courante d'un dépôt.
func currentBranch(repoPath string) (string, error) {
	out, err := runGit(repoPath, gitDefaultTimeout, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// CreateWorktree crée un worktree git dédié à une tâche, sur une nouvelle
// branche partant de la branche par défaut du dépôt. Retourne le répertoire
// du worktree et le nom de la branche de base.
func CreateWorktree(repoPath, dataDir, taskID, branch string) (dir, base string, err error) {
	base, err = currentBranch(repoPath)
	if err != nil {
		return "", "", err
	}
	dir = filepath.Join(dataDir, "worktrees", taskID)
	if err := os.MkdirAll(filepath.Dir(dir), 0o700); err != nil {
		return "", "", err
	}
	// Nettoie les worktrees fantômes (répertoire de données supprimé ou déplacé)
	// qui bloqueraient la création au même chemin.
	_, _ = runGit(repoPath, gitDefaultTimeout, "worktree", "prune")
	// -B : réutilise la branche si elle existe déjà (le nom est dérivé d'un
	// compteur de référence, une collision signifie un état recréé de zéro).
	if _, err := runGit(repoPath, gitDefaultTimeout, "worktree", "add", "-B", branch, dir, base); err != nil {
		return "", "", err
	}
	return dir, base, nil
}

// RemoveWorktree retire le worktree d'une tâche du dépôt d'origine :
// git worktree remove --force puis worktree prune. Best-effort (appelé lors
// d'une suppression de tâche) : un échec ne doit jamais empêcher la
// suppression, et la branche n'est JAMAIS supprimée (elle peut avoir été
// poussée).
func RemoveWorktree(repoPath, worktreeDir string) {
	if repoPath == "" || worktreeDir == "" {
		return
	}
	_, _ = runGit(repoPath, gitDefaultTimeout, "worktree", "remove", "--force", worktreeDir)
	_, _ = runGit(repoPath, gitDefaultTimeout, "worktree", "prune")
}

// Diff calcule le diff unifié entre base et l'état courant du worktree,
// y compris les fichiers non suivis (git add -A -N les fait apparaître).
func Diff(dir, base string) ([]DiffFile, error) {
	if _, err := runGit(dir, gitDefaultTimeout, "add", "-A", "-N"); err != nil {
		return nil, err
	}
	out, err := runGit(dir, gitDefaultTimeout, "diff", base)
	if err != nil {
		return nil, err
	}
	return parseDiff(out), nil
}

// parseDiff transforme la sortie de `git diff` en une liste de fichiers avec
// leurs hunks. Parseur robuste : ignore les sections binaires, gère les renommages.
func parseDiff(output string) []DiffFile {
	var files []DiffFile
	var current *DiffFile
	var currentHunk *Hunk
	binary := false

	flushHunk := func() {
		if current != nil && currentHunk != nil {
			current.Hunks = append(current.Hunks, *currentHunk)
			currentHunk = nil
		}
	}
	flushFile := func() {
		flushHunk()
		if current != nil {
			files = append(files, *current)
			current = nil
		}
	}

	for _, line := range strings.Split(output, "\n") {
		switch {
		case strings.HasPrefix(line, "diff --git "):
			flushFile()
			current = &DiffFile{}
			binary = false
			parts := strings.SplitN(line, " ", 4)
			if len(parts) == 4 {
				current.Path = strings.TrimPrefix(parts[3], "b/")
			}
		case strings.HasPrefix(line, "rename to "):
			if current != nil {
				current.Path = strings.TrimPrefix(line, "rename to ")
			}
		case strings.HasPrefix(line, "--- "):
			p := strings.TrimPrefix(line, "--- ")
			if p != "/dev/null" && current != nil {
				current.Path = strings.TrimPrefix(p, "a/")
			}
		case strings.HasPrefix(line, "+++ "):
			p := strings.TrimPrefix(line, "+++ ")
			if p != "/dev/null" && current != nil {
				current.Path = strings.TrimPrefix(p, "b/")
			}
		case strings.HasPrefix(line, "Binary files "):
			binary = true
		case strings.HasPrefix(line, "@@"):
			if binary || current == nil {
				continue
			}
			flushHunk()
			currentHunk = &Hunk{Header: line}
		case strings.HasPrefix(line, "rename from "),
			strings.HasPrefix(line, "similarity index "),
			strings.HasPrefix(line, "index "),
			strings.HasPrefix(line, "new file mode "),
			strings.HasPrefix(line, "deleted file mode "):
			// métadonnées ignorées pour le comptage
		default:
			if binary || current == nil || currentHunk == nil || line == "" {
				continue
			}
			switch line[0] {
			case '+':
				current.Additions++
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{Type: "add", Text: line[1:]})
			case '-':
				current.Deletions++
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{Type: "del", Text: line[1:]})
			case ' ':
				currentHunk.Lines = append(currentHunk.Lines, DiffLine{Type: "ctx", Text: line[1:]})
			case '\\':
				// "\ No newline at end of file" : ignoré
			}
		}
	}
	flushFile()
	if files == nil {
		files = []DiffFile{}
	}
	return files
}

// Commits liste les commits de base..HEAD (le plus récent en premier, comme git log).
func Commits(dir, base string) ([]CommitInfo, error) {
	out, err := runGit(dir, gitDefaultTimeout, "log", base+"..HEAD", "--pretty=format:%h|%s|%cr")
	if err != nil {
		return nil, err
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return []CommitInfo{}, nil
	}
	var commits []CommitInfo
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}
		commits = append(commits, CommitInfo{Hash: parts[0], Subject: parts[1], RelTime: parts[2]})
	}
	return commits, nil
}

// Ship commite les changements en attente puis pousse la branche vers origin.
// C'est le SEUL endroit du code qui exécute `git push`.
func Ship(dir, branch, title string) (string, error) {
	var out strings.Builder

	addOut, err := runGit(dir, gitDefaultTimeout, "add", "-A")
	out.WriteString(addOut)
	if err != nil {
		return out.String(), fmt.Errorf("git add failed: %w", err)
	}

	statusOut, err := runGit(dir, gitDefaultTimeout, "status", "--porcelain")
	if err != nil {
		return out.String(), err
	}
	if strings.TrimSpace(statusOut) != "" {
		commitOut, err := runGit(dir, gitDefaultTimeout, "commit", "-m", "Sillage: "+title)
		out.WriteString(commitOut)
		if err != nil {
			return out.String(), fmt.Errorf("git commit failed: %w", err)
		}
	}

	pushOut, err := runGit(dir, gitPushTimeout, "push", "-u", "origin", branch)
	out.WriteString(pushOut)
	if err != nil {
		return out.String(), fmt.Errorf("push failed (remote 'origin' missing or unreachable): %w", err)
	}
	return out.String(), nil
}

// --- Ouverture de pull request (lecture seule, jamais de push) ---

// githubSSHRemoteRe et githubHTTPSRemoteRe reconnaissent les deux formats
// d'URL de remote github.com utilisés par git (ssh de type scp, ou https).
var githubSSHRemoteRe = regexp.MustCompile(`^git@github\.com:([^/]+)/(.+?)(\.git)?/?$`)
var githubHTTPSRemoteRe = regexp.MustCompile(`^https?://github\.com/([^/]+)/(.+?)(\.git)?/?$`)

// ParseGithubRemote extrait le owner et le repo d'une URL de remote
// github.com, au format ssh (git@github.com:owner/repo.git) ou https
// (https://github.com/owner/repo[.git]).
func ParseGithubRemote(remoteURL string) (owner, repo string, ok bool) {
	remoteURL = strings.TrimSpace(remoteURL)
	if m := githubSSHRemoteRe.FindStringSubmatch(remoteURL); m != nil {
		return m[1], m[2], true
	}
	if m := githubHTTPSRemoteRe.FindStringSubmatch(remoteURL); m != nil {
		return m[1], m[2], true
	}
	return "", "", false
}

// OpenPR ouvre une pull request pour une branche déjà poussée sur origin.
// Tente d'abord `gh pr create` (aucune commande d'écriture : la branche est
// supposée déjà poussée par Ship). En l'absence de gh, ou en cas d'échec,
// replie sur une URL de comparaison GitHub en lecture seule.
func OpenPR(dir, branch, base, title string) (string, error) {
	if _, lookErr := exec.LookPath("gh"); lookErr == nil {
		if url, err := runGhPRCreate(dir, branch, title); err == nil && url != "" {
			return url, nil
		}
	}
	return fallbackCompareURL(dir, base, branch)
}

// runGhPRCreate exécute `gh pr create` (timeout 60 s, environnement hérité)
// et extrait l'URL de la pull request depuis sa sortie standard.
func runGhPRCreate(dir, branch, title string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	body := "Created with Sillage\n\n" + title
	cmd := exec.CommandContext(ctx, "gh", "pr", "create", "--head", branch, "--title", title, "--body", body)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh pr create failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	if url := extractURL(string(out)); url != "" {
		return url, nil
	}
	return "", fmt.Errorf("gh pr create succeeded but no URL was found in its output")
}

// extractURL retourne la première ligne ressemblant à une URL http(s) dans s.
func extractURL(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "https://") || strings.HasPrefix(line, "http://") {
			return line
		}
	}
	return ""
}

// fallbackCompareURL construit une URL de comparaison GitHub en lecture
// seule, sans exécuter aucune commande d'écriture.
func fallbackCompareURL(dir, base, branch string) (string, error) {
	out, err := runGit(dir, gitDefaultTimeout, "remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("no github remote configured")
	}
	owner, repo, ok := ParseGithubRemote(out)
	if !ok {
		return "", fmt.Errorf("origin remote is not a github.com repository")
	}
	return fmt.Sprintf("https://github.com/%s/%s/compare/%s...%s?expand=1", owner, repo, base, branch), nil
}

// githubBranchURL construit l'URL de la branche sur GitHub
// (https://github.com/<owner>/<repo>/tree/<branch>) si le remote "origin" est
// un dépôt github.com, chaîne vide sinon. Jamais d'erreur : c'est une
// information optionnelle jointe à la réponse de Ship.
func githubBranchURL(dir, branch string) string {
	out, err := runGit(dir, gitDefaultTimeout, "remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	owner, repo, ok := ParseGithubRemote(out)
	if !ok {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/%s/tree/%s", owner, repo, branch)
}

// --- Synchronisation de l'espace de travail (dataDir uniquement) ---

// ErrSyncConflict signale un conflit de rebase lors de la synchronisation de
// l'espace de travail (le remote a divergé).
var ErrSyncConflict = errors.New("sync conflict")

// syncPushFn indirige l'appel à SyncPush pour Server.autoSyncTick, afin de
// pouvoir simuler succès/échec/conflit en test sans dépôt git réel.
var syncPushFn = SyncPush

// SyncPush synchronise l'espace de travail (dataDir) avec son remote : commit
// des changements en attente, puis `git pull --rebase origin main`, puis
// `git push -u origin main`. C'est, avec Ship (dépôts de projet), le SEUL
// autre endroit du code qui exécute `git push` : il ne peut opérer que sur
// dataDir, jamais sur un dépôt de projet. L'appelant (handlers.go) ne doit
// lui passer que le dataDir configuré côté serveur, jamais une valeur issue
// d'une entrée utilisateur.
func SyncPush(dataDir string) (string, error) {
	var out strings.Builder

	addOut, err := gitAddWorkspaceFiles(dataDir)
	out.WriteString(addOut)
	if err != nil {
		return out.String(), fmt.Errorf("git add failed: %w", err)
	}
	statusOut, err := runGit(dataDir, gitDefaultTimeout, "status", "--porcelain")
	if err != nil {
		return out.String(), err
	}
	if strings.TrimSpace(statusOut) != "" {
		commitOut, err := runGit(dataDir, gitDefaultTimeout, "commit", "-m", "sillage: update")
		out.WriteString(commitOut)
		if err != nil {
			return out.String(), fmt.Errorf("git commit failed: %w", err)
		}
	}

	pullOut, pullErr := runGit(dataDir, gitPushTimeout, "pull", "--rebase", "origin", "main")
	out.WriteString(pullOut)
	if pullErr != nil {
		// Première sync vers un remote vierge : pas de branche main distante,
		// rien à rebaser, on pousse directement. (Message git localisé selon
		// la machine, d'où le double test.)
		errLower := strings.ToLower(pullErr.Error())
		emptyRemote := strings.Contains(errLower, "couldn't find remote ref") ||
			strings.Contains(errLower, "impossible de trouver la référence distante")
		if !emptyRemote {
			conflict := isRebaseInProgress(dataDir)
			_, _ = runGit(dataDir, gitDefaultTimeout, "rebase", "--abort")
			if conflict {
				return out.String(), fmt.Errorf("%w: the remote workspace diverged, resolve manually in %s", ErrSyncConflict, dataDir)
			}
			return out.String(), fmt.Errorf("git pull --rebase failed: %w", pullErr)
		}
	}

	pushOut, err := runGit(dataDir, gitPushTimeout, "push", "-u", "origin", "main")
	out.WriteString(pushOut)
	if err != nil {
		return out.String(), fmt.Errorf("push failed: %w", err)
	}
	return out.String(), nil
}

// isRebaseInProgress indique si un rebase git est en cours dans dataDir :
// distingue un vrai conflit d'un simple échec (réseau, remote absent...).
func isRebaseInProgress(dataDir string) bool {
	gitDir := filepath.Join(dataDir, ".git")
	for _, name := range []string{"rebase-apply", "rebase-merge"} {
		if info, err := os.Stat(filepath.Join(gitDir, name)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}
