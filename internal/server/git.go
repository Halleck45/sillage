package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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

// GitCommonDir retourne le dossier git partagé du dépôt auquel dir appartient
// (le `.git` du dépôt d'origine quand dir est un worktree lié, où `.git` n'est
// qu'un fichier pointeur). Sert à donner ce dossier à un agent qui travaille
// dans un bac à sable ne montrant que les répertoires qu'on lui ajoute : sans
// lui, toute commande git échoue par « ce n'est pas un dépôt git ».
func GitCommonDir(dir string) (string, error) {
	if dir == "" {
		return "", errors.New("empty directory")
	}
	out, err := runGit(dir, gitDefaultTimeout, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
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
// branche partant de base (la branche du chantier, voir CreateCardWorktree).
// Si base est vide, la branche courante du dépôt est utilisée. Retourne le
// répertoire du worktree et la branche de base effective.
func CreateWorktree(repoPath, dataDir, taskID, branch, base string) (dir, resolvedBase string, err error) {
	return createWorktree(repoPath, filepath.Join(dataDir, "worktrees", taskID), branch, base)
}

// CreateCardWorktree crée la branche de feature d'un chantier sur un dépôt,
// avec son worktree dédié : c'est là que les tâches acceptées sont fusionnées,
// et c'est cette branche que la livraison pousse. base vide = branche courante
// du dépôt.
func CreateCardWorktree(repoPath, dataDir, cardID, repoName, branch, base string) (dir, resolvedBase string, err error) {
	name := "ws-" + cardID + "-" + Slugify(repoName)
	return createWorktree(repoPath, filepath.Join(dataDir, "worktrees", name), branch, base)
}

// createWorktree est l'implémentation commune : crée dir comme worktree du
// dépôt repoPath, sur la branche branch partant de base.
func createWorktree(repoPath, dir, branch, base string) (string, string, error) {
	if base == "" {
		var err error
		base, err = currentBranch(repoPath)
		if err != nil {
			return "", "", err
		}
	}
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

// OpenDifftool lance l'outil de diff configuré par l'utilisateur (`git config
// diff.tool`, réglage standard de git : VS Code, Meld, Kaleidoscope...) sur un
// fichier du worktree, entre base et l'état courant. Ne bloque pas : l'outil
// est une appli graphique qui peut rester ouverte longtemps, l'appelant ne
// doit pas attendre sa fermeture.
func OpenDifftool(dir, base, path string) error {
	if out, err := runGit(dir, 5*time.Second, "config", "diff.tool"); err != nil || strings.TrimSpace(out) == "" {
		return errors.New("no diff tool configured: run `git config --global diff.tool <tool>` (e.g. vscode)")
	}
	cmd := exec.Command("git", "difftool", "--no-prompt", base, "--", path)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch diff tool: %w", err)
	}
	go cmd.Wait() // évite un zombie ; l'appelant n'a plus de raison d'attendre l'outil
	return nil
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

// CountCommits compte les commits de from..to (lecture seule, aucun réseau).
// Une révision inconnue (par exemple origin/<branche> avant tout push) rend
// une erreur : l'appelant décide quoi en faire.
func CountCommits(dir, from, to string) (int, error) {
	out, err := runGit(dir, gitDefaultTimeout, "rev-list", "--count", from+".."+to)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.TrimSpace(out))
	if err != nil {
		return 0, err
	}
	return n, nil
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

// TaskWorkCounts compte ce que la branche d'une tâche a produit par rapport à
// sa base : fichiers touchés, documents parmi eux, et commits. Best-effort, car
// ces trois nombres ne sont qu'un affichage : un worktree disparu ou une base
// inconnue rendent des zéros plutôt qu'une erreur.
func TaskWorkCounts(worktreeDir, base string) (files, docs, commits int) {
	if worktreeDir == "" || base == "" {
		return 0, 0, 0
	}
	if changed, err := Diff(worktreeDir, base); err == nil {
		files = len(changed)
		for _, f := range changed {
			if isDocFile(f.Path) {
				docs++
			}
		}
	}
	if log, err := Commits(worktreeDir, base); err == nil {
		commits = len(log)
	}
	return files, docs, commits
}

// CommitAll indexe tout l'arbre de travail de dir et commite s'il y a quelque
// chose à commiter (sinon ne fait rien, sans erreur). Aucun réseau.
func CommitAll(dir, message string) (string, error) {
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
	if strings.TrimSpace(statusOut) == "" {
		return out.String(), nil
	}
	commitOut, err := runGit(dir, gitDefaultTimeout, "commit", "-m", message)
	out.WriteString(commitOut)
	if err != nil {
		return out.String(), fmt.Errorf("git commit failed: %w", err)
	}
	return out.String(), nil
}

// ErrMergeConflict signale un conflit de fusion (branche de tâche dans la
// branche du chantier, ou branche de chantier dans la branche de destination).
var ErrMergeConflict = errors.New("merge conflict")

// HeadSha retourne le SHA de HEAD d'un worktree (chaîne vide en cas d'échec).
// Sert à savoir si une fusion a réellement ajouté des commits, sans dépendre
// de la sortie de git, qui est localisée selon la machine.
func HeadSha(dir string) string {
	out, err := runGit(dir, gitDefaultTimeout, "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// IsWorktreeClean indique si un worktree n'a aucune modification en attente
// (fichiers non suivis compris). Une erreur git vaut « pas propre » : on ne
// prend jamais de décision automatique sur une information manquante.
func IsWorktreeClean(dir string) bool {
	out, err := runGit(dir, gitDefaultTimeout, "status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == ""
}

// IsBranchMergedInto indique si branch est entièrement contenue dans target
// (tous ses commits en sont des ancêtres). Lecture seule : sert à détecter une
// fusion faite à la main, hors de Sillage.
func IsBranchMergedInto(dir, branch, target string) bool {
	_, err := runGit(dir, gitDefaultTimeout, "merge-base", "--is-ancestor", branch, target)
	return err == nil
}

// MergeBranch fusionne branch dans la branche courante du worktree dir, avec
// un merge commit dédié (--no-ff) : l'historique du chantier garde une trace
// lisible de chaque tâche acceptée. Aucun réseau.
//
// En cas de conflit, la fusion est annulée (git merge --abort, l'arbre revient
// intact) et l'erreur enveloppe ErrMergeConflict avec la liste des fichiers en
// conflit : rien n'est résolu automatiquement.
func MergeBranch(dir, branch, message string) (string, []string, error) {
	out, err := runGit(dir, gitDefaultTimeout, "merge", "--no-ff", "-m", message, branch)
	if err == nil {
		return out, nil, nil
	}
	conflicts := conflictedFiles(dir)
	_, _ = runGit(dir, gitDefaultTimeout, "merge", "--abort")
	if len(conflicts) > 0 {
		return out, conflicts, fmt.Errorf("%w: %s", ErrMergeConflict, strings.Join(conflicts, ", "))
	}
	return out, nil, fmt.Errorf("git merge failed: %w", err)
}

// ErrRebaseConflict signale un conflit de rebase (branche de tâche rejouée sur
// la branche du chantier). Le rebase est alors annulé : l'arbre revient intact.
var ErrRebaseConflict = errors.New("rebase conflict")

// RebaseOnto rejoue la branche courante du worktree dir au-dessus de onto (la
// branche du chantier), pour qu'une tâche encore en revue reparte du travail
// déjà accepté. Aucun réseau.
//
// En cas de conflit, `git rebase --abort` remet le worktree exactement dans son
// état d'avant, et l'erreur enveloppe ErrRebaseConflict avec la liste des
// fichiers en conflit : rien n'est résolu automatiquement, seul l'agent (ou
// l'humain) sait le faire.
//
// L'appelant doit s'assurer que le worktree est propre et qu'aucun agent n'y
// travaille : un rebase réécrit l'historique de la branche de la tâche.
func RebaseOnto(dir, onto string) (string, []string, error) {
	out, err := runGit(dir, gitDefaultTimeout, "rebase", onto)
	if err == nil {
		return out, nil, nil
	}
	conflicts := conflictedFiles(dir)
	_, _ = runGit(dir, gitDefaultTimeout, "rebase", "--abort")
	if len(conflicts) > 0 {
		return out, conflicts, fmt.Errorf("%w: %s", ErrRebaseConflict, strings.Join(conflicts, ", "))
	}
	return out, nil, fmt.Errorf("git rebase failed: %w", err)
}

// conflictedFiles liste les fichiers en conflit d'une fusion en cours.
func conflictedFiles(dir string) []string {
	out, err := runGit(dir, gitDefaultTimeout, "diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files
}

// pushBranch pousse une branche vers origin. C'est, avec SyncPush (espace de
// travail, jamais un dépôt de projet), le SEUL endroit du code qui exécute
// `git push` sur un dépôt de projet : deux appelants seulement, Ship (branche
// de chantier) et mergeThenPush (branche de destination du mode "merge-push").
// Jamais de --force, jamais de refspec construite depuis une entrée utilisateur.
func pushBranch(dir, branch string, setUpstream bool) (string, error) {
	args := []string{"push"}
	if setUpstream {
		args = append(args, "-u")
	}
	args = append(args, "origin", branch)

	out, err := runGit(dir, gitPushTimeout, args...)
	if err != nil {
		return out, fmt.Errorf("push failed (remote 'origin' missing or unreachable): %w", err)
	}
	return out, nil
}

// Ship pousse la branche d'un chantier vers origin (les commits ont déjà été
// faits à l'acceptation des tâches, voir MergeBranch ; un arbre sale est tout
// de même commité pour ne rien perdre). Utilisé par les modes de livraison
// "pr" et "push" (voir Delivery.Mode) : la différence entre les deux est
// l'ouverture de la pull request, pas le push.
func Ship(dir, branch, title string) (string, error) {
	var out strings.Builder

	commitOut, err := CommitAll(dir, "Sillage: "+title)
	out.WriteString(commitOut)
	if err != nil {
		return out.String(), err
	}

	pushOut, err := pushBranch(dir, branch, true)
	out.WriteString(pushOut)
	if err != nil {
		return out.String(), err
	}
	return out.String(), nil
}

// --- Livraison en fusion dans la branche de destination ("merge", "merge-push") ---

// ErrTargetDiverged signale que la branche de destination a divergé : la
// fusion locale n'accepte que le fast-forward, aucune résolution automatique.
var ErrTargetDiverged = errors.New("target branch has diverged")

// ErrTargetBusy signale que la branche de destination est empruntée par un
// worktree dans un état qui interdit toute fusion sûre (autre branche
// courante, ou modifications non commitées dans le dépôt de travail).
var ErrTargetBusy = errors.New("target branch is busy")

// MergeLocal fusionne la branche d'un chantier dans la branche de destination,
// en local et en fast-forward uniquement. Ne pousse JAMAIS rien : pousser une
// branche partagée reste une décision humaine, prise dans un terminal.
//
// Deux chemins, dans cet ordre :
//  1. worktree transitoire dédié (retiré à la fin), quand target n'est
//     empruntée par aucun worktree : le dépôt de travail n'est pas touché ;
//  2. repli dans le dépôt lui-même quand target y est déjà empruntée (cas
//     courant : main est la branche courante du dépôt de travail). Autorisé
//     uniquement si target EST la branche courante et que l'arbre est propre :
//     jamais de changement de branche, jamais de stash, jamais de force. Sinon
//     ErrTargetBusy, avec la commande à jouer à la main.
func MergeLocal(repoPath, dataDir, cardID, repoName, target, source string) (string, error) {
	return mergeIntoTarget(repoPath, dataDir, cardID, repoName, target, source, false)
}

// MergeAndPush est MergeLocal suivie du push de la branche de destination
// (mode "merge-push") : le projet dont la livraison consiste à faire avancer
// la branche principale, remote comprise, sans passer par une pull request.
//
// Le remote est rattrapé AVANT la fusion (voir fastForwardFromRemote) : sans
// ça, le push serait rejeté dès que la branche de destination a avancé côté
// remote, et l'utilisateur se retrouverait avec une fusion locale à moitié
// livrée. Toujours fast-forward, jamais de --force.
func MergeAndPush(repoPath, dataDir, cardID, repoName, target, source string) (string, error) {
	return mergeIntoTarget(repoPath, dataDir, cardID, repoName, target, source, true)
}

// mergeIntoTarget porte les deux modes de fusion : le choix du répertoire où
// fusionner (worktree transitoire, ou dépôt de travail en repli) est commun,
// seul le push final diffère.
func mergeIntoTarget(repoPath, dataDir, cardID, repoName, target, source string, push bool) (string, error) {
	dir := filepath.Join(dataDir, "worktrees", ".merge-"+cardID+"-"+Slugify(repoName))
	_, _ = runGit(repoPath, gitDefaultTimeout, "worktree", "prune")
	_ = os.RemoveAll(dir)

	if _, err := runGit(repoPath, gitDefaultTimeout, "worktree", "add", dir, target); err == nil {
		defer RemoveWorktree(repoPath, dir)
		return mergeThenPush(dir, target, source, push)
	}

	cur, err := currentBranch(repoPath)
	if err != nil {
		return "", err
	}
	if cur != target {
		return "", fmt.Errorf("%w: %s is checked out in another worktree; run `git merge --ff-only %s` there", ErrTargetBusy, target, source)
	}
	statusOut, err := runGit(repoPath, gitDefaultTimeout, "status", "--porcelain")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(statusOut) != "" {
		return "", fmt.Errorf("%w: %s has uncommitted changes in %s", ErrTargetBusy, target, repoPath)
	}
	return mergeThenPush(repoPath, target, source, push)
}

// mergeThenPush enchaîne, dans le répertoire où target est empruntée :
// rattrapage du remote (si push), fusion fast-forward, push de target (si
// push). Le rattrapage est fait d'abord : une divergence est un refus avant
// toute écriture, pas une fusion locale suivie d'un push rejeté.
func mergeThenPush(dir, target, source string, push bool) (string, error) {
	var out strings.Builder

	if push {
		fetchOut, err := fastForwardFromRemote(dir, target)
		out.WriteString(fetchOut)
		if err != nil {
			return out.String(), err
		}
	}

	mergeOut, err := mergeFastForward(dir, target, source)
	out.WriteString(mergeOut)
	if err != nil {
		return out.String(), err
	}
	if !push {
		return out.String(), nil
	}

	pushOut, err := pushBranch(dir, target, false)
	out.WriteString(pushOut)
	if err != nil {
		return out.String(), err
	}
	return out.String(), nil
}

// mergeFastForward exécute la fusion fast-forward proprement dite et traduit
// un refus de git en ErrTargetDiverged (message actionnable).
func mergeFastForward(dir, target, source string) (string, error) {
	out, err := runGit(dir, gitDefaultTimeout, "merge", "--ff-only", source)
	if err != nil {
		return out, fmt.Errorf("%w: %s cannot fast-forward to %s, merge manually", ErrTargetDiverged, target, source)
	}
	return out, nil
}

// fastForwardFromRemote met la branche de destination locale au niveau de son
// homologue distante, en fast-forward uniquement. Trois cas :
//
//   - la branche n'existe pas encore sur le remote : rien à rattraper, le push
//     final la créera ;
//   - le local contient déjà le distant (à jour, ou en avance) : rien à faire ;
//   - le distant a des commits que le local n'a pas : fast-forward, ou refus
//     ErrTargetDiverged si les deux ont divergé (aucune résolution automatique).
func fastForwardFromRemote(dir, target string) (string, error) {
	var out strings.Builder

	fetchOut, err := runGit(dir, gitPushTimeout, "fetch", "origin", target)
	out.WriteString(fetchOut)
	if err != nil {
		if isMissingRemoteRef(err) {
			return out.String(), nil
		}
		return out.String(), fmt.Errorf("fetch failed (remote 'origin' missing or unreachable): %w", err)
	}
	if _, err := runGit(dir, gitDefaultTimeout, "merge-base", "--is-ancestor", "FETCH_HEAD", "HEAD"); err == nil {
		return out.String(), nil
	}
	mergeOut, err := runGit(dir, gitDefaultTimeout, "merge", "--ff-only", "FETCH_HEAD")
	out.WriteString(mergeOut)
	if err != nil {
		return out.String(), fmt.Errorf("%w: %s and origin/%s have diverged, reconcile manually", ErrTargetDiverged, target, target)
	}
	return out.String(), nil
}

// isMissingRemoteRef reconnaît l'échec d'un fetch dont la référence distante
// n'existe pas (branche jamais poussée, remote vierge). Le message de git est
// localisé selon la machine, d'où le double test.
func isMissingRemoteRef(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "couldn't find remote ref") ||
		strings.Contains(msg, "impossible de trouver la référence distante")
}

// --- Ouverture de pull/merge request (lecture seule, jamais de push) ---

// remoteSSHRe et remoteHTTPSRe reconnaissent les deux formats d'URL de remote
// utilisés par git : ssh de type scp (git@host:owner/repo.git) ou https
// (https://host/owner/repo[.git], sous-groupes compris pour GitLab).
var remoteSSHRe = regexp.MustCompile(`^(?:[^@]+@)?([^:/]+):(.+?)(\.git)?/?$`)
var remoteHTTPSRe = regexp.MustCompile(`^[a-z]+://(?:[^@/]+@)?([^:/]+)(?::\d+)?/(.+?)(\.git)?/?$`)

// ParseRemote extrait l'hôte et le chemin (owner/repo, sous-groupes GitLab
// compris) d'une URL de remote git.
func ParseRemote(remoteURL string) (host, path string, ok bool) {
	remoteURL = strings.TrimSpace(remoteURL)
	if m := remoteHTTPSRe.FindStringSubmatch(remoteURL); m != nil {
		return m[1], m[2], true
	}
	if m := remoteSSHRe.FindStringSubmatch(remoteURL); m != nil {
		return m[1], m[2], true
	}
	return "", "", false
}

// ParseGithubRemote extrait le owner et le repo d'une URL de remote
// github.com, au format ssh (git@github.com:owner/repo.git) ou https
// (https://github.com/owner/repo[.git]).
func ParseGithubRemote(remoteURL string) (owner, repo string, ok bool) {
	host, path, ok := ParseRemote(remoteURL)
	if !ok || host != "github.com" {
		return "", "", false
	}
	owner, repo, found := strings.Cut(path, "/")
	if !found || owner == "" || repo == "" {
		return "", "", false
	}
	return owner, repo, true
}

// ForgeInfo décrit le remote origin d'un dépôt. RemoteURL vide signifie
// « aucun remote origin » ; Provider vide avec un RemoteURL renseigné signifie
// « forge inconnue » (les deux cas n'ont pas la même conséquence produit).
type ForgeInfo struct {
	Provider  string // github|gitlab|""
	Host      string
	Path      string // owner/repo, sous-groupes GitLab compris
	RemoteURL string
}

// DetectForge déduit le fournisseur de forge du remote origin d'un dépôt.
// Jamais persisté : redéduit à chaque opération, un remote peut changer.
func DetectForge(dir string) ForgeInfo {
	out, err := runGit(dir, gitDefaultTimeout, "remote", "get-url", "origin")
	if err != nil {
		return ForgeInfo{}
	}
	info := ForgeInfo{RemoteURL: strings.TrimSpace(out)}
	if info.RemoteURL == "" {
		return ForgeInfo{}
	}
	host, path, ok := ParseRemote(info.RemoteURL)
	if !ok {
		return info
	}
	info.Host, info.Path = host, path
	switch {
	case host == "github.com":
		info.Provider = "github"
	case strings.Contains(host, "gitlab"):
		info.Provider = "gitlab"
	}
	return info
}

// OpenPR ouvre une pull request (GitHub) ou une merge request (GitLab) pour une
// branche déjà poussée sur origin. Aucune commande d'écriture git : la branche
// a été poussée par Ship. En l'absence du CLI de la forge, ou en cas d'échec,
// replie sur une URL de création pré-remplie, en lecture seule.
func OpenPR(dir, branch, base, title string) (string, error) {
	info := DetectForge(dir)
	switch info.Provider {
	case "github":
		if _, lookErr := exec.LookPath("gh"); lookErr == nil {
			if u, err := runForgeCreate(dir, "gh", "pr", "create", "--head", branch, "--base", base, "--title", title, "--body", prBody(title)); err == nil && u != "" {
				return u, nil
			}
		}
		return GithubCompareURL(info.Host, info.Path, base, branch), nil
	case "gitlab":
		if _, lookErr := exec.LookPath("glab"); lookErr == nil {
			if u, err := runForgeCreate(dir, "glab", "mr", "create", "--source-branch", branch, "--target-branch", base, "--title", title, "--description", prBody(title), "--yes"); err == nil && u != "" {
				return u, nil
			}
		}
		return GitlabNewMRURL(info.Host, info.Path, base, branch), nil
	}
	return "", fmt.Errorf("origin remote is not a github or gitlab repository")
}

// GithubCompareURL construit l'URL de création de pull request GitHub
// pré-remplie (lecture seule, aucun effet de bord).
func GithubCompareURL(host, path, base, branch string) string {
	return fmt.Sprintf("https://%s/%s/compare/%s...%s?expand=1", host, path, base, branch)
}

// GitlabNewMRURL construit l'URL de création de merge request GitLab
// pré-remplie (lecture seule, aucun effet de bord).
func GitlabNewMRURL(host, path, base, branch string) string {
	return fmt.Sprintf("https://%s/%s/-/merge_requests/new?merge_request[source_branch]=%s&merge_request[target_branch]=%s",
		host, path, url.QueryEscape(branch), url.QueryEscape(base))
}

// prBody est le corps de la pull/merge request créée par Sillage.
func prBody(title string) string {
	return "Created with Sillage\n\n" + title
}

// runForgeCreate exécute le CLI de la forge (gh ou glab, timeout 60 s,
// environnement hérité) et extrait l'URL de la requête depuis sa sortie.
func runForgeCreate(dir, bin string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s failed: %w: %s", bin, args[0], err, strings.TrimSpace(string(out)))
	}
	if u := extractURL(string(out)); u != "" {
		return u, nil
	}
	return "", fmt.Errorf("%s succeeded but no URL was found in its output", bin)
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
		// rien à rebaser, on pousse directement.
		if !isMissingRemoteRef(pullErr) {
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
