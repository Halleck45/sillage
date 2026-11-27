package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

// acpAgentSpec décrit un agent piloté par ACP : comment le lancer, et les
// quelques points où il s'écarte du protocole commun. Brancher un nouvel agent
// ACP revient à écrire cette structure.
type acpAgentSpec struct {
	Binary string
	Args   []string
	Env    []string

	// Meta part dans clientCapabilities._meta (extensions de l'agent).
	Meta map[string]any
	// Permission répond aux demandes d'autorisation ; acpRejectAll par défaut.
	Permission acpPermissionPolicy
	// TurnStart reconnaît, parmi les notifications propres à l'agent, celle qui
	// annonce le début réel de son tour. Le texte reçu avant est alors jeté :
	// voir acpTurn.discardPrelude. Nil pour un agent qui n'écrit rien avant.
	TurnStart func(method string, params json.RawMessage) bool
	// FreshPrompt habille le premier prompt d'une session neuve (contexte du
	// projet, consignes). Une session reprise les connaît déjà.
	FreshPrompt func(text string) string
	// SignInHint remplace l'erreur de l'agent quand il n'est pas connecté.
	SignInHint string
}

const (
	// acpStartTimeout borne l'ouverture (initialize, puis session neuve ou
	// reprise, qui rejoue l'historique) : un agent bloqué à ce stade ne fera
	// jamais rien, autant le dire plutôt que de laisser la tâche tourner.
	acpStartTimeout = 3 * time.Minute
	// acpExitGrace : délai laissé au process pour sortir une fois son entrée
	// fermée, avant de tuer son groupe.
	acpExitGrace = 3 * time.Second
)

// runACP exécute un tour d'un agent ACP : un process par tour, comme les
// autres adaptateurs, la session faisant le lien d'un tour à l'autre quand
// l'agent sait la reprendre (Task.SessionID).
func (r *Runner) runACP(task *Task, agent Agent, handle *procHandle, spec acpAgentSpec, cliInput string) error {
	cmd := exec.Command(spec.Binary, spec.Args...)
	cmd.Dir = task.WorktreeDir
	cmd.Env = spec.Env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", spec.Binary, err)
	}
	handle.cmd = cmd

	messages := 0
	turn := newACPTurn(task.WorktreeDir)
	turn.onMessage = func(text string) {
		msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, text)
		if err == nil {
			messages++
			r.publishMessage(msg)
			*task = t
			r.publishTask(*task)
		}
	}
	turn.onActivity = func(line string) {
		if t, err := r.store.UpdateTask(task.ID, func(t *Task) {
			t.LiveActivity = &line
			t.CommandLog = appendCommandLog(t.CommandLog, line)
		}); err == nil {
			*task = t
			r.publishActivity(task.ID, &line)
			r.publishTask(*task)
		}
	}
	turn.onDenied = func(line string) {
		msg, t, err := r.store.AddMessage(task.ID, "agent", "", "[tool-denied:"+line+"]")
		if err == nil {
			r.publishMessage(msg)
			*task = t
			r.publishTask(*task)
		}
	}

	policy := spec.Permission
	if policy == nil {
		policy = acpRejectAll
	}
	client := newACPClient(stdout, stdin, acpHandlers{
		Update:     func(_ string, update acpUpdate) { turn.handle(update) },
		Permission: func(req acpPermissionRequest) string { return turn.permission(policy, req) },
		Notification: func(method string, params json.RawMessage) {
			if spec.TurnStart != nil && spec.TurnStart(method, params) {
				turn.discardPrelude()
			}
		},
	})

	result, runErr := r.acpConversation(client, turn, task, handle, spec, cliInput)

	// Fermer l'entrée dit à l'agent que c'est fini. On attend la fin de son
	// flux de sortie avant Wait (qui ferme le tube) ; un process, ou un enfant
	// qu'il a laissé derrière lui, qui garde la sortie ouverte est tué.
	_ = stdin.Close()
	select {
	case <-client.done:
	case <-time.After(acpExitGrace):
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		select {
		case <-client.done:
		case <-time.After(acpExitGrace):
		}
	}
	_ = cmd.Wait()

	if handle.interrupted.Load() {
		return nil
	}
	if runErr != nil {
		var rpcErr *acpError
		switch {
		case errors.As(runErr, &rpcErr) && rpcErr.Code == acpAuthRequired && spec.SignInHint != "":
			return errors.New(spec.SignInHint)
		case errors.As(runErr, &rpcErr):
			return fmt.Errorf("the agent reported an error: %s", truncate(rpcErr.Error(), 300))
		case errors.Is(runErr, errACPClosed):
			if line := lastLine(stderrBuf.String()); line != "" {
				return errors.New(line)
			}
			return fmt.Errorf("%s exited before answering", spec.Binary)
		case errors.Is(runErr, context.DeadlineExceeded):
			return fmt.Errorf("%s did not open a session in time", spec.Binary)
		}
		return runErr
	}
	switch result.StopReason {
	case "end_turn", "cancelled", "":
	case "refusal":
		return errors.New("the agent refused to continue")
	default:
		return fmt.Errorf("the agent stopped early: %s", result.StopReason)
	}
	// Un tour réussi qui n'a rien dit laisserait une conversation vide, sans
	// le moindre indice (même règle que runTextCLI).
	if messages == 0 {
		if line := lastLine(stderrBuf.String()); line != "" {
			return errors.New(line)
		}
		return fmt.Errorf("%s produced no output", spec.Binary)
	}
	return nil
}

// acpConversation mène l'échange : ouverture, session (reprise si possible,
// neuve sinon), un prompt, puis les compteurs de tokens.
func (r *Runner) acpConversation(client *acpClient, turn *acpTurn, task *Task, handle *procHandle, spec acpAgentSpec, cliInput string) (acpPromptResult, error) {
	startCtx, cancel := context.WithTimeout(context.Background(), acpStartTimeout)
	defer cancel()

	init, err := client.initialize(startCtx, spec.Meta)
	if err != nil {
		return acpPromptResult{}, err
	}

	sessionID := ""
	if task.SessionID != "" && init.AgentCapabilities.LoadSession {
		err := client.loadSession(startCtx, task.SessionID, task.WorktreeDir)
		switch {
		case err == nil:
			sessionID = task.SessionID
		case errors.Is(err, errACPClosed), errors.Is(err, context.DeadlineExceeded):
			return acpPromptResult{}, err
		}
	}
	if sessionID == "" {
		// Session introuvable (effacée, ou créée par un autre agent) : on
		// l'oublie et le fil est rejoué dans le prompt, comme pour un agent
		// sans reprise.
		if task.SessionID != "" {
			if t, err := r.store.UpdateTask(task.ID, func(t *Task) { t.SessionID = "" }); err == nil {
				*task = t
				cliInput = r.prepareCliInput(t, cliInput)
			}
		}
		sessionID, err = client.newSession(startCtx, task.WorktreeDir)
		if err != nil {
			return acpPromptResult{}, err
		}
		// Sans reprise possible, garder l'identifiant ferait croire au tour
		// suivant qu'il n'a pas d'historique à rejouer.
		if init.AgentCapabilities.LoadSession {
			if t, err := r.store.UpdateTask(task.ID, func(t *Task) { t.SessionID = sessionID }); err == nil {
				*task = t
			}
		}
		if spec.FreshPrompt != nil {
			cliInput = spec.FreshPrompt(cliInput)
		}
	}

	stop := func() { client.cancel(sessionID) }
	handle.softStop.Store(&stop)
	// Une interruption arrivée pendant l'ouverture a déjà signalé le process.
	if handle.interrupted.Load() {
		return acpPromptResult{}, nil
	}

	turn.begin()
	result, err := client.prompt(context.Background(), sessionID, cliInput)
	if result.StopReason == "cancelled" || handle.interrupted.Load() {
		turn.abort()
	} else {
		turn.end()
	}
	if err != nil {
		return result, err
	}

	if usage := result.Usage; usage != nil {
		// inputTokens compte aussi ce qui a été relu du cache ; on le retire
		// pour compter comme l'adaptateur claude (entrée neuve et écriture de
		// cache, pas la relecture).
		input := usage.InputTokens - usage.CachedReadTokens
		if input < 0 {
			input = 0
		}
		if t, err := r.store.UpdateTask(task.ID, func(t *Task) {
			t.Tokens.Input += input
			t.Tokens.Output += usage.OutputTokens
		}); err == nil {
			*task = t
		}
		r.publishTokens()
	}
	return result, nil
}
