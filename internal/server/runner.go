package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// claudeAllowedTools est la liste figée d'outils autorisés à l'agent claude.
// Jamais de Bash(git push:*) ici : le push n'est déclenché que par Ship (git.go).
const claudeAllowedTools = "Read,Edit,Write,Glob,Grep,WebFetch,Bash(go build:*),Bash(go test:*),Bash(go vet:*),Bash(ls:*),Bash(cat:*),Bash(mkdir:*)"

// procHandle représente un processus (ou une simulation) en cours pour une tâche.
type procHandle struct {
	cmd         *exec.Cmd
	cancel      context.CancelFunc // utilisé par l'adaptateur fake (pas de process réel)
	interrupted atomic.Bool
	done        chan struct{}
}

// Runner exécute au plus un agent par tâche à la fois.
type Runner struct {
	mu    sync.Mutex
	procs map[string]*procHandle
	store *Store
	hub   *Hub
}

// NewRunner crée un runner lié à un store et un hub SSE.
func NewRunner(store *Store, hub *Hub) *Runner {
	return &Runner{procs: map[string]*procHandle{}, store: store, hub: hub}
}

func (r *Runner) publishTask(t Task)       { r.hub.Publish(Event{Name: "task", Data: t}) }
func (r *Runner) publishMessage(m Message) { r.hub.Publish(Event{Name: "message", Data: m}) }
func (r *Runner) publishAgents()           { r.hub.Publish(Event{Name: "agents", Data: r.store.ListAgents()}) }
func (r *Runner) publishTokens() {
	r.hub.Publish(Event{Name: "tokens", Data: r.store.TokensSnapshot()})
}
func (r *Runner) publishCards(projectID string) {
	r.hub.Publish(Event{Name: "cards", Data: r.store.CardsByProject(projectID)})
}
func (r *Runner) publishProject(p Project) { r.hub.Publish(Event{Name: "project", Data: p}) }
func (r *Runner) publishWorkspace(w WorkspaceStatus) {
	r.hub.Publish(Event{Name: "workspace", Data: w})
}
func (r *Runner) publishSettings(s Settings) { r.hub.Publish(Event{Name: "settings", Data: s}) }
func (r *Runner) publishActivity(taskID string, line *string) {
	r.hub.Publish(Event{Name: "activity", Data: map[string]any{"taskId": taskID, "line": line}})
}

// Start lance (ou relance) l'agent d'une tâche. initial=true correspond au
// lancement initial (text = titre + description, aucun Message ajouté) ;
// initial=false correspond à un nouveau message utilisateur (text = son
// contenu, ajouté comme Message puis transmis à l'agent). L'AuthorName du
// message utilisateur est le displayName des Settings (vide si non renseigné,
// le frontend affiche alors "Vous"/"You").
func (r *Runner) Start(taskID string, initial bool, text string) error {
	r.mu.Lock()
	if _, exists := r.procs[taskID]; exists {
		r.mu.Unlock()
		return fmt.Errorf("an agent is already running for this task")
	}
	r.mu.Unlock()

	task, ok := r.store.GetTask(taskID)
	if !ok {
		return fmt.Errorf("task not found")
	}
	agent, ok := r.store.GetAgent(task.AgentID)
	if !ok {
		return fmt.Errorf("agent not found")
	}

	task, err := r.store.UpdateTask(taskID, func(t *Task) { t.Status = "running"; t.LiveActivity = nil })
	if err != nil {
		return err
	}
	r.publishTask(task)
	r.publishAgents()

	if !initial {
		authorName := r.store.GetSettings().DisplayName
		msg, updated, err := r.store.AddMessage(taskID, "user", authorName, text)
		if err != nil {
			return err
		}
		r.publishMessage(msg)
		task = updated
		r.publishTask(task)
	}

	handle := &procHandle{done: make(chan struct{})}
	r.mu.Lock()
	r.procs[taskID] = handle
	r.mu.Unlock()

	go r.run(task, agent, handle, text)
	return nil
}

// Interrupt arrête l'agent en cours pour une tâche : SIGINT au groupe de
// process, puis SIGKILL après 5 s si nécessaire. Le statut passe
// immédiatement à "review".
func (r *Runner) Interrupt(taskID string) (Task, error) {
	r.mu.Lock()
	handle, ok := r.procs[taskID]
	r.mu.Unlock()
	if !ok {
		return Task{}, fmt.Errorf("no agent is running for this task")
	}
	handle.interrupted.Store(true)

	task, err := r.store.UpdateTask(taskID, func(t *Task) { t.Status = "review" })
	if err != nil {
		return Task{}, err
	}
	r.publishTask(task)
	r.publishCards(task.ProjectID)
	r.publishAgents()

	go killProcessGroup(handle)
	return task, nil
}

// Cancel annule une tâche (autorisé depuis running/review/ready, voir
// Store.CancelTask). Si elle est en cours d'exécution, l'agent est interrompu
// au préalable (même mécanique qu'Interrupt) avant de passer au statut final
// "cancelled" (et non "review").
func (r *Runner) Cancel(taskID string) (Task, error) {
	task, ok := r.store.GetTask(taskID)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}

	var handle *procHandle
	if task.Status == "running" {
		r.mu.Lock()
		h, hasProc := r.procs[taskID]
		r.mu.Unlock()
		if hasProc {
			h.interrupted.Store(true)
			handle = h
		}
	}

	updated, err := r.store.CancelTask(taskID)
	if err != nil {
		return Task{}, err
	}
	r.publishTask(updated)
	r.publishCards(updated.ProjectID)
	r.publishAgents()

	if handle != nil {
		go killProcessGroup(handle)
	}
	return updated, nil
}

func killProcessGroup(handle *procHandle) {
	if handle.cmd != nil && handle.cmd.Process != nil {
		pid := handle.cmd.Process.Pid
		_ = syscall.Kill(-pid, syscall.SIGINT)
		select {
		case <-handle.done:
			return
		case <-time.After(5 * time.Second):
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		}
		return
	}
	if handle.cancel != nil {
		handle.cancel()
	}
}

// run exécute l'adaptateur adéquat puis finalise la tâche (checks, compteurs).
func (r *Runner) run(task Task, agent Agent, handle *procHandle, cliInput string) {
	defer func() {
		r.mu.Lock()
		delete(r.procs, task.ID)
		r.mu.Unlock()
		close(handle.done)
	}()

	project, _ := r.store.GetProject(task.ProjectID)

	var runErr error
	switch agent.Cli {
	case "claude":
		runErr = r.runClaude(&task, agent, project, handle, cliInput)
	case "codex":
		runErr = r.runCodex(&task, agent, project, handle, cliInput)
	case "fake":
		runErr = r.runFake(&task, agent, handle)
	default:
		runErr = fmt.Errorf("unknown cli: %s", agent.Cli)
	}

	if runErr != nil {
		msg, updated, err := r.store.AddMessage(task.ID, "agent", agent.Name, "⚠️ "+runErr.Error())
		if err == nil {
			r.publishMessage(msg)
			task = updated
		}
	}

	r.finalize(task.ID)
}

// finalize exécute le check éventuel du projet et clôt l'exécution.
func (r *Runner) finalize(taskID string) {
	task, ok := r.store.GetTask(taskID)
	if !ok {
		return
	}
	project, _ := r.store.GetProject(task.ProjectID)

	checks := []Check{}
	if project.CheckCmd != "" {
		ok, _ := runCheck(task.WorktreeDir, project.CheckCmd, 120*time.Second)
		checks = []Check{{Label: project.CheckCmd, Ok: ok}}
	}

	filesCount, docsCount := 0, 0
	if files, err := Diff(task.WorktreeDir, task.Base); err == nil {
		filesCount = len(files)
		for _, f := range files {
			if isDocFile(f.Path) {
				docsCount++
			}
		}
	}

	updated, err := r.store.UpdateTask(taskID, func(t *Task) {
		if t.Status == "running" {
			t.Status = "review"
		}
		t.Unread = true
		t.LiveActivity = nil
		t.Checks = checks
		t.FilesCount = filesCount
		t.DocsCount = docsCount
	})
	if err != nil {
		return
	}
	r.publishTask(updated)
	r.publishCards(updated.ProjectID)
	r.publishAgents()
	r.publishActivity(updated.ID, nil)
}

func runCheck(dir, command string, timeout time.Duration) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return err == nil, string(out)
}

func isDocFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".txt", ".rst":
		return true
	default:
		return false
	}
}

func isImageFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

// buildSystemPrompt combine le contexte de l'agent et celui du projet pour
// --append-system-prompt (adaptateur claude) : contexte agent, puis ligne
// vide, puis "Project context:\n<projectContext>" si projectContext est non
// vide. Retourne une chaîne vide si les deux le sont (pas de flag ajouté).
func buildSystemPrompt(agentContext, projectContext string) string {
	if projectContext == "" {
		return agentContext
	}
	if agentContext == "" {
		return "Project context:\n" + projectContext
	}
	return agentContext + "\n\nProject context:\n" + projectContext
}

// summarizeToolUse construit la ligne d'activité affichée pour un tool_use claude.
func summarizeToolUse(name string, input map[string]any) string {
	for _, key := range []string{"file_path", "path", "command", "pattern", "query", "url"} {
		if v, ok := input[key]; ok {
			return name + " · " + truncate(fmt.Sprintf("%v", v), 80)
		}
	}
	b, _ := json.Marshal(input)
	return name + " · " + truncate(string(b), 80)
}

// --- Adaptateur claude ---

type claudeContentBlock struct {
	Type  string         `json:"type"`
	Text  string         `json:"text"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

type claudeEnvelope struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype"`
	SessionID string `json:"session_id"`
	Message   *struct {
		Content []claudeContentBlock `json:"content"`
	} `json:"message"`
	Result       string  `json:"result"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	Usage        *struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	} `json:"usage"`
	IsError bool `json:"is_error"`
}

func (r *Runner) runClaude(task *Task, agent Agent, project Project, handle *procHandle, cliInput string) error {
	args := []string{
		"-p", "--output-format", "stream-json", "--verbose",
		"--permission-mode", "acceptEdits",
		"--allowedTools", claudeAllowedTools,
	}
	if agent.Model != "" {
		args = append(args, "--model", agent.Model)
	}
	if systemPrompt := buildSystemPrompt(agent.ContextPrompt, project.ContextPrompt); systemPrompt != "" {
		args = append(args, "--append-system-prompt", systemPrompt)
	}
	if task.SessionID != "" {
		args = append(args, "--resume", task.SessionID)
	}
	args = append(args, cliInput)

	cmd := exec.Command("claude", args...)
	cmd.Dir = task.WorktreeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude: %w", err)
	}
	handle.cmd = cmd

	lastAgentText := ""
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var env claudeEnvelope
		if err := json.Unmarshal([]byte(line), &env); err != nil {
			continue
		}
		switch env.Type {
		case "system":
			if env.Subtype == "init" && env.SessionID != "" {
				if t, err := r.store.UpdateTask(task.ID, func(t *Task) { t.SessionID = env.SessionID }); err == nil {
					*task = t
				}
			}
		case "assistant":
			if env.Message == nil {
				continue
			}
			for _, block := range env.Message.Content {
				switch block.Type {
				case "text":
					if strings.TrimSpace(block.Text) == "" {
						continue
					}
					lastAgentText = block.Text
					msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, block.Text)
					if err == nil {
						r.publishMessage(msg)
						*task = t
						r.publishTask(*task)
					}
				case "tool_use":
					line := summarizeToolUse(block.Name, block.Input)
					if t, err := r.store.UpdateTask(task.ID, func(t *Task) { t.LiveActivity = &line }); err == nil {
						*task = t
						r.publishActivity(task.ID, &line)
						r.publishTask(*task)
					}
				}
			}
		case "result":
			inputTok, outputTok := 0, 0
			if env.Usage != nil {
				inputTok = env.Usage.InputTokens + env.Usage.CacheCreationInputTokens
				outputTok = env.Usage.OutputTokens
			}
			cost := env.TotalCostUSD
			if t, err := r.store.UpdateTask(task.ID, func(t *Task) {
				t.Tokens.Input += inputTok
				t.Tokens.Output += outputTok
				t.Tokens.CostUsd += cost
			}); err == nil {
				*task = t
			}
			r.publishTokens()
			if env.Result != "" && env.Result != lastAgentText {
				msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, env.Result)
				if err == nil {
					r.publishMessage(msg)
					*task = t
					r.publishTask(*task)
				}
			}
			if env.IsError {
				return fmt.Errorf("the agent reported an error")
			}
		}
	}

	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if waitErr == nil && scanErr != nil {
		waitErr = scanErr
	}
	if waitErr != nil && !handle.interrupted.Load() {
		msg := strings.TrimSpace(stderrBuf.String())
		if msg == "" {
			msg = waitErr.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// --- Adaptateur codex (best-effort) ---

func (r *Runner) runCodex(task *Task, agent Agent, project Project, handle *procHandle, cliInput string) error {
	// workspace-write : écriture limitée au worktree, réseau coupé ; le push
	// reste impossible et ne passe que par Ship (git.go) après validation humaine.
	// SILLAGE_CODEX_SANDBOX permet de choisir un autre mode (ex : danger-full-access
	// sur les machines où bwrap est bloqué par AppArmor) ; le confinement d'Sillage
	// (worktree dédié, pas de push agent) reste alors la seule barrière.
	sandbox := os.Getenv("SILLAGE_CODEX_SANDBOX")
	if sandbox == "" {
		sandbox = "workspace-write"
	}
	if project.ContextPrompt != "" {
		cliInput = "Project context:\n" + project.ContextPrompt + "\n\n---\n\n" + cliInput
	}
	args := []string{"exec", "--json", "--sandbox", sandbox, "-C", task.WorktreeDir}
	if agent.Model != "" {
		args = append(args, "--model", agent.Model)
	}
	args = append(args, cliInput)

	cmd := exec.Command("codex", args...)
	cmd.Dir = task.WorktreeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderrBuf, rawBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start codex: %w", err)
	}
	handle.cmd = cmd

	parsedAny := false
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		rawBuf.WriteString(line)
		rawBuf.WriteByte('\n')
		if strings.TrimSpace(line) == "" {
			continue
		}
		var generic map[string]any
		if err := json.Unmarshal([]byte(line), &generic); err != nil {
			continue
		}
		if text, ok := extractCodexMessage(generic); ok {
			parsedAny = true
			msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, text)
			if err == nil {
				r.publishMessage(msg)
				*task = t
				r.publishTask(*task)
			}
		}
		if tok, ok := extractCodexTokens(generic); ok {
			parsedAny = true
			if t, err := r.store.UpdateTask(task.ID, func(t *Task) {
				t.Tokens.Input += tok.Input
				t.Tokens.Output += tok.Output
				t.Tokens.CostUsd += tok.CostUsd
			}); err == nil {
				*task = t
			}
			r.publishTokens()
		}
	}

	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if waitErr == nil && scanErr != nil {
		waitErr = scanErr
	}
	if !parsedAny && rawBuf.Len() > 0 {
		msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, strings.TrimSpace(rawBuf.String()))
		if err == nil {
			r.publishMessage(msg)
			*task = t
			r.publishTask(*task)
		}
	}
	if waitErr != nil && !handle.interrupted.Load() {
		msg := strings.TrimSpace(stderrBuf.String())
		if msg == "" {
			msg = waitErr.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func extractCodexMessage(m map[string]any) (string, bool) {
	if msg, ok := m["msg"].(map[string]any); ok {
		if msg["type"] == "agent_message" {
			if text, ok := msg["message"].(string); ok && text != "" {
				return text, true
			}
		}
	}
	if m["type"] == "item.completed" {
		if item, ok := m["item"].(map[string]any); ok && item["type"] == "agent_message" {
			if text, ok := item["text"].(string); ok && text != "" {
				return text, true
			}
		}
	}
	return "", false
}

func extractCodexTokens(m map[string]any) (Tokens, bool) {
	raw, ok := m["token_count"]
	if !ok {
		// Forme récente : {"type":"turn.completed","usage":{"input_tokens":N,...}}.
		if m["type"] == "turn.completed" {
			raw, ok = m["usage"]
		}
		if !ok {
			return Tokens{}, false
		}
	}
	tc, ok := raw.(map[string]any)
	if !ok {
		return Tokens{}, false
	}
	tok := Tokens{}
	found := false
	if v, ok := tc["input_tokens"].(float64); ok {
		tok.Input = int(v)
		found = true
	}
	if v, ok := tc["output_tokens"].(float64); ok {
		tok.Output = int(v)
		found = true
	}
	if v, ok := tc["cost_usd"].(float64); ok {
		tok.CostUsd = v
		found = true
	}
	return tok, found
}

// --- Adaptateur fake (test/démo, sans exec) ---

func (r *Runner) runFake(task *Task, agent Agent, handle *procHandle) error {
	ctx, cancel := context.WithCancel(context.Background())
	handle.cancel = cancel
	defer cancel()

	steps := []string{"Analyse du projet", "Rédaction des modifications", "Finalisation"}
	for _, step := range steps {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
		line := step
		if t, err := r.store.UpdateTask(task.ID, func(t *Task) { t.LiveActivity = &line }); err == nil {
			*task = t
			r.publishActivity(task.ID, &line)
			r.publishTask(*task)
		}
	}

	content := fmt.Sprintf("# Sillage : tâche de test\n\nGénéré le %s par l'agent %s.\n", time.Now().UTC().Format(time.RFC3339), agent.Name)
	_ = os.WriteFile(filepath.Join(task.WorktreeDir, "SILLAGE-TEST.md"), []byte(content), 0o644)

	msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, "Tâche simulée terminée : SILLAGE-TEST.md mis à jour.")
	if err == nil {
		r.publishMessage(msg)
		*task = t
		r.publishTask(*task)
	}

	if t, err := r.store.UpdateTask(task.ID, func(t *Task) {
		t.Tokens.Input += 1200
		t.Tokens.Output += 340
		t.Tokens.CostUsd += 0.004
	}); err == nil {
		*task = t
	}
	r.publishTokens()
	return nil
}
