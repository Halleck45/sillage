package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// La vérification projet (Project.CheckCmd) tourne en tâche de fond dans le
// worktree d'une tâche : à la fin de chaque exécution d'agent, après un rebase
// automatique, et à la demande. Elle ne tourne jamais en même temps que l'agent
// de la même tâche : tout départ d'agent, rebase, acceptation, refus ou
// suppression l'annule d'abord (CancelCheck), parce qu'elle lit (et souvent
// écrit, artefacts de build) le même arbre de travail.

const (
	// checkTimeout borne une vérification. Large : c'est la suite de tests du
	// projet, et un délai dépassé est rendu comme tel (Reason "timeout"), pas
	// comme un échec anonyme.
	checkTimeout = 10 * time.Minute
	// checkOutputTail : taille de la fin de sortie gardée. Assez pour la trace
	// d'un test cassé, pas assez pour gonfler state.json.
	checkOutputTail = 12 << 10
)

type checkHandle struct {
	cancel    context.CancelFunc
	done      chan struct{}
	cancelled bool // protégé par Runner.mu
}

// tailBuffer garde les derniers max octets écrits.
type tailBuffer struct {
	mu  sync.Mutex
	buf []byte
	max int
}

func (b *tailBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	if len(b.buf) > b.max {
		b.buf = append([]byte(nil), b.buf[len(b.buf)-b.max:]...)
	}
	return len(p), nil
}

func (b *tailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.buf)
}

// WorkFingerprint résume le travail d'une tâche : le diff du worktree (non
// commité compris) depuis son point de départ sur la branche base. Les lignes
// qui bougent sans que le code de la tâche change (index git, en-têtes de
// hunk) sont ignorées : un rebase sans conflit garde la même empreinte, une
// relecture faite avant lui reste donc valable. Vide si git ne répond pas.
func WorkFingerprint(dir, base string) string {
	if dir == "" || base == "" {
		return ""
	}
	from := base
	if mb, err := runGit(dir, gitDefaultTimeout, "merge-base", base, "HEAD"); err == nil && strings.TrimSpace(mb) != "" {
		from = strings.TrimSpace(mb)
	}
	if _, err := runGit(dir, gitDefaultTimeout, "add", "-A", "-N"); err != nil {
		return ""
	}
	out, err := runGit(dir, gitDefaultTimeout, "diff", "--no-color", "--no-ext-diff", from)
	if err != nil {
		return ""
	}
	h := sha256.New()
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "@@") {
			continue
		}
		h.Write([]byte(line))
		h.Write([]byte{'\n'})
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// StartCheck lance la vérification projet d'une tâche en tâche de fond, après
// avoir annulé celle qui tournait. Rend false sans rien lancer si le projet
// n'a pas de commande ou si un agent travaille dans le worktree.
func (r *Runner) StartCheck(taskID string) bool {
	if r.IsRunning(taskID) {
		return false
	}
	return r.startCheck(taskID)
}

// startCheck est StartCheck sans le contrôle de l'agent : finalize l'appelle
// alors que le process de l'agent est encore compté (il a fini, mais le defer
// de run ne l'a pas encore retiré).
func (r *Runner) startCheck(taskID string) bool {
	r.CancelCheck(taskID)
	task, ok := r.store.GetTask(taskID)
	if !ok || task.WorktreeDir == "" {
		return false
	}
	project, _ := r.store.GetProject(task.ProjectID)
	command := strings.TrimSpace(project.CheckCmd)
	if command == "" {
		if len(task.Checks) > 0 || task.Checking {
			if updated, err := r.store.PatchTask(taskID, func(t *Task) {
				t.Checks = []Check{}
				t.Checking = false
				t.CheckStartedAt = nil
			}); err == nil {
				r.publishTask(updated)
			}
		}
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	handle := &checkHandle{cancel: cancel, done: make(chan struct{})}
	r.mu.Lock()
	r.checks[taskID] = handle
	r.mu.Unlock()

	now := time.Now().UTC()
	if updated, err := r.store.PatchTask(taskID, func(t *Task) {
		t.Checking = true
		t.CheckStartedAt = &now
	}); err == nil {
		r.publishTask(updated)
	}

	r.checkWG.Add(1)
	go r.runCheckJob(ctx, handle, task, project, command)
	return true
}

func (r *Runner) runCheckJob(ctx context.Context, handle *checkHandle, task Task, project Project, command string) {
	defer r.checkWG.Done()
	defer handle.cancel()

	sha := WorkFingerprint(task.WorktreeDir, task.Base)
	env := []string(nil)
	if runtime, err := preparePermissionRuntime(r.store.DataDir(), task, project); err == nil {
		env = runtime.Env
	}
	started := time.Now()
	ok, reason, output := runCheckCommand(ctx, task.WorktreeDir, command, env)
	at := time.Now().UTC()

	r.mu.Lock()
	cancelled := handle.cancelled
	if r.checks[task.ID] == handle {
		delete(r.checks, task.ID)
	}
	r.mu.Unlock()

	updated, err := r.store.PatchTask(task.ID, func(t *Task) {
		t.Checking = false
		t.CheckStartedAt = nil
		if cancelled {
			return // annulée pour laisser la place : le résultat ne dit rien
		}
		c := Check{Label: command, Ok: ok, Sha: sha, At: &at, DurationMs: time.Since(started).Milliseconds(), Output: output}
		if !ok {
			c.Reason = reason
		}
		t.Checks = []Check{c}
		if sha != "" {
			t.WorkSha = sha
		}
	})
	close(handle.done)
	if err == nil {
		r.publishTask(updated)
	}
}

// CancelCheck arrête la vérification en cours d'une tâche et attend qu'elle ait
// rendu la main (processus tués, état écrit). Sans effet s'il n'y en a pas.
func (r *Runner) CancelCheck(taskID string) {
	r.mu.Lock()
	handle := r.checks[taskID]
	if handle != nil {
		handle.cancelled = true
		delete(r.checks, taskID)
	}
	r.mu.Unlock()
	if handle == nil {
		return
	}
	handle.cancel()
	<-handle.done
}

// waitChecks attend la fin des vérifications en cours. Réservé aux tests.
func (r *Runner) waitChecks() { r.checkWG.Wait() }

// runCheckCommand exécute command avec sh dans dir, dans son propre groupe de
// process : une suite de tests lance des enfants (go test → binaires de test,
// npm → node), et l'annulation doit les viser tous.
func runCheckCommand(ctx context.Context, dir, command string, env []string) (ok bool, reason, output string) {
	out := &tailBuffer{max: checkOutputTail}
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Un petit-enfant qui garde la sortie ouverte ne doit pas bloquer Wait.
	cmd.WaitDelay = 5 * time.Second
	if err := cmd.Start(); err != nil {
		return false, "error: " + err.Error(), ""
	}
	done := make(chan struct{})
	waitErr := make(chan error, 1)
	go func() { waitErr <- cmd.Wait(); close(done) }()

	var err error
	select {
	case err = <-waitErr:
	case <-ctx.Done():
		killGroup(cmd, done)
		err = <-waitErr
	}
	text := strings.TrimRight(out.String(), "\n")
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		return false, "timeout", text
	case ctx.Err() != nil:
		return false, "cancelled", text
	case err == nil:
		return true, "", text
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, fmt.Sprintf("exit %d", exitErr.ExitCode()), text
	}
	return false, "error: " + err.Error(), text
}
