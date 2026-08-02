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
		return string(out), fmt.Errorf("commande git expirée (%s) : git %s", timeout, strings.Join(args, " "))
	}
	if err != nil {
		return string(out), fmt.Errorf("git %s : %w : %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
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
		return out.String(), fmt.Errorf("git add a échoué : %w", err)
	}

	statusOut, err := runGit(dir, gitDefaultTimeout, "status", "--porcelain")
	if err != nil {
		return out.String(), err
	}
	if strings.TrimSpace(statusOut) != "" {
		commitOut, err := runGit(dir, gitDefaultTimeout, "commit", "-m", "Atelier: "+title)
		out.WriteString(commitOut)
		if err != nil {
			return out.String(), fmt.Errorf("git commit a échoué : %w", err)
		}
	}

	pushOut, err := runGit(dir, gitPushTimeout, "push", "-u", "origin", branch)
	out.WriteString(pushOut)
	if err != nil {
		return out.String(), fmt.Errorf("échec du push (remote 'origin' absent ou inaccessible) : %w", err)
	}
	return out.String(), nil
}
