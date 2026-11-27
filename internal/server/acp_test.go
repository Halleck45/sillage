package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestACPHelperAgent n'est pas un test : c'est un faux agent ACP, lancé par
// les tests ci-dessous en relançant le binaire de test (SILLAGE_ACP_HELPER
// choisit le scénario). Il parle le protocole sur stdio et note dans
// SILLAGE_ACP_LOG ce qu'il a reçu, pour que le test puisse le vérifier.
func TestACPHelperAgent(t *testing.T) {
	scenario := os.Getenv("SILLAGE_ACP_HELPER")
	if scenario == "" {
		return
	}
	defer os.Exit(0)

	logFile, _ := os.OpenFile(os.Getenv("SILLAGE_ACP_LOG"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	note := func(format string, args ...any) { fmt.Fprintf(logFile, format+"\n", args...) }
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	out := json.NewEncoder(os.Stdout)

	type message struct {
		ID     json.RawMessage `json:"id"`
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
		Result json.RawMessage `json:"result"`
	}
	read := func() (message, bool) {
		if !in.Scan() {
			return message{}, false
		}
		var m message
		_ = json.Unmarshal(in.Bytes(), &m)
		return m, true
	}
	result := func(id json.RawMessage, v any) {
		_ = out.Encode(map[string]any{"jsonrpc": "2.0", "id": id, "result": v})
	}
	fail := func(id json.RawMessage, code int, text string) {
		_ = out.Encode(map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": text}})
	}
	update := func(v map[string]any) {
		_ = out.Encode(map[string]any{"jsonrpc": "2.0", "method": "session/update", "params": map[string]any{"sessionId": "s", "update": v}})
	}
	say := func(text string) {
		update(map[string]any{"sessionUpdate": "agent_message_chunk", "content": map[string]any{"type": "text", "text": text}})
	}
	turnStart := func() {
		_ = out.Encode(map[string]any{"jsonrpc": "2.0", "method": "github.com/copilot/sessionEvent",
			"params": map[string]any{"sessionId": "s", "type": "assistant.turn_start", "data": map[string]any{}}})
	}

	cwd := ""
	for {
		m, ok := read()
		if !ok {
			return
		}
		switch m.Method {
		case "initialize":
			note("initialize %s", m.Params)
			result(m.ID, map[string]any{"protocolVersion": 1, "agentCapabilities": map[string]any{"loadSession": scenario != "noload"}})
		case "session/new":
			var p struct {
				Cwd string `json:"cwd"`
			}
			_ = json.Unmarshal(m.Params, &p)
			cwd = p.Cwd
			note("session/new")
			if scenario == "auth" {
				fail(m.ID, -32000, "Authentication required")
				continue
			}
			result(m.ID, map[string]any{"sessionId": "sess-new"})
		case "session/load":
			var p struct {
				SessionID string `json:"sessionId"`
				Cwd       string `json:"cwd"`
			}
			_ = json.Unmarshal(m.Params, &p)
			cwd = p.Cwd
			note("session/load %s", p.SessionID)
			if scenario == "lost" {
				fail(m.ID, -32002, "Session not found")
				continue
			}
			// Rejeu de l'historique, avant la réponse.
			update(map[string]any{"sessionUpdate": "user_message_chunk", "content": map[string]any{"type": "text", "text": "old question"}})
			say("OLD ANSWER")
			update(map[string]any{"sessionUpdate": "tool_call", "toolCallId": "old", "kind": "execute", "title": "old command"})
			result(m.ID, map[string]any{})
		case "session/prompt":
			var p struct {
				Prompt []struct {
					Text string `json:"text"`
				} `json:"prompt"`
			}
			_ = json.Unmarshal(m.Params, &p)
			if len(p.Prompt) > 0 {
				note("prompt %q", p.Prompt[0].Text)
			}
			switch scenario {
			case "silent":
				fmt.Fprintln(os.Stderr, "boom: the model is unavailable")
				result(m.ID, map[string]any{"stopReason": "end_turn"})
			case "cancel":
				turnStart()
				say("Working")
				update(map[string]any{"sessionUpdate": "tool_call", "toolCallId": "c1", "kind": "execute", "rawInput": map[string]any{"command": "sleep 60"}})
				next, ok := read()
				if !ok {
					return
				}
				note("after prompt: %s", next.Method)
				say("Info: Operation cancelled by user")
				result(m.ID, map[string]any{"stopReason": "cancelled"})
			default:
				say("Warning: startup noise that is not an answer.")
				turnStart()
				say("Looking")
				say(" at it.")
				update(map[string]any{"sessionUpdate": "tool_call", "toolCallId": "t1", "kind": "read", "title": "Viewing a.txt",
					"status": "pending", "rawInput": map[string]any{"path": filepath.Join(cwd, "sub", "a.txt")}})
				update(map[string]any{"sessionUpdate": "tool_call_update", "toolCallId": "t1", "status": "completed"})
				turnStart()
				say("Now writing.")
				_ = out.Encode(map[string]any{"jsonrpc": "2.0", "id": 100, "method": "session/request_permission", "params": map[string]any{
					"sessionId": "s",
					"toolCall":  map[string]any{"toolCallId": "t2", "kind": "fetch", "title": "Fetch", "rawInput": map[string]any{"url": "https://example.com"}},
					"options": []map[string]any{
						{"optionId": "yes", "kind": "allow_once", "name": "Allow once"},
						{"optionId": "always", "kind": "allow_always", "name": "Always allow"},
						{"optionId": "no", "kind": "reject_once", "name": "Deny"},
					},
				}})
				reply, ok := read()
				if !ok {
					return
				}
				note("permission %s", reply.Result)
				update(map[string]any{"sessionUpdate": "tool_call", "toolCallId": "t3", "kind": "execute", "status": "pending",
					"rawInput": map[string]any{"command": "git push origin HEAD"}})
				for i := 0; i < 2; i++ {
					update(map[string]any{"sessionUpdate": "tool_call_update", "toolCallId": "t3", "status": "failed",
						"rawOutput": map[string]any{"code": "denied", "message": "denied by rules"}})
				}
				_ = out.Encode(map[string]any{"jsonrpc": "2.0", "id": 101, "method": "fs/read_text_file", "params": map[string]any{"path": "/etc/passwd"}})
				reply, ok = read()
				if !ok {
					return
				}
				note("fs %s", string(reply.Result))
				say("Done.")
				result(m.ID, map[string]any{"stopReason": "end_turn",
					"usage": map[string]any{"inputTokens": 100, "outputTokens": 7, "cachedReadTokens": 60, "cachedWriteTokens": 30}})
			}
		default:
			note("other %s", m.Method)
		}
	}
}

// acpTestRun prépare une tâche et lance runACP contre le faux agent.
type acpTestRun struct {
	store   *Store
	runner  *Runner
	task    Task
	handle  *procHandle
	logPath string
	spec    acpAgentSpec
}

func newACPTestRun(t *testing.T, scenario string) *acpTestRun {
	t.Helper()
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	p, _ := s.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil, nil)
	card, _ := s.AddCard(p.ID, "Carte", "", "")
	taskID := mkTaskWithStatus(t, s, card.ID, p.ID, "running")
	worktree := t.TempDir()
	task, err := s.UpdateTask(taskID, func(tk *Task) { tk.WorktreeDir = worktree })
	if err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	logPath := filepath.Join(t.TempDir(), "agent.log")
	return &acpTestRun{
		store:   s,
		runner:  NewRunner(s, NewHub()),
		task:    task,
		handle:  &procHandle{done: make(chan struct{})},
		logPath: logPath,
		spec: acpAgentSpec{
			Binary:      os.Args[0],
			Args:        []string{"-test.run=^TestACPHelperAgent$"},
			Env:         append(os.Environ(), "SILLAGE_ACP_HELPER="+scenario, "SILLAGE_ACP_LOG="+logPath),
			Meta:        copilotMeta,
			TurnStart:   copilotTurnStart,
			FreshPrompt: func(text string) string { return "CONTEXT\n\n" + text },
			SignInHint:  copilotSignInHint,
		},
	}
}

func (a *acpTestRun) run(input string) error {
	return a.runner.runACP(&a.task, Agent{Name: "Octo", Cli: "copilot"}, a.handle, a.spec, input)
}

func (a *acpTestRun) agentLog(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(a.logPath)
	if err != nil {
		t.Fatalf("agent log: %v", err)
	}
	return string(data)
}

func (a *acpTestRun) texts() []string {
	var out []string
	for _, m := range a.store.GetMessages(a.task.ID) {
		out = append(out, m.Text)
	}
	return out
}

// TestRunACPTurnsTheStreamIntoAConversation suit un tour complet : le texte
// d'avant le début du tour est jeté, la réponse est coupée à chaque outil,
// chaque outil laisse une ligne d'activité, et rien n'est jamais accordé.
func TestRunACPTurnsTheStreamIntoAConversation(t *testing.T) {
	a := newACPTestRun(t, "basic")
	if err := a.run("Fix the bug"); err != nil {
		t.Fatalf("runACP: %v", err)
	}

	want := []string{
		"Looking at it.",
		"Now writing.",
		"[tool-denied:fetch · https://example.com]",
		"[tool-denied:execute · git push origin HEAD]",
		"Done.",
	}
	if got := a.texts(); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("conversation = %q, want %q", got, want)
	}

	task, _ := a.store.GetTask(a.task.ID)
	if task.SessionID != "sess-new" {
		t.Fatalf("the session should be kept for the next turn, got %q", task.SessionID)
	}
	// 100 tokens d'entrée dont 60 relus du cache : 40 comptés, comme claude.
	if task.Tokens.Input != 40 || task.Tokens.Output != 7 {
		t.Fatalf("tokens = %+v, want 40 in / 7 out", task.Tokens)
	}
	var lines []string
	for _, entry := range task.CommandLog {
		lines = append(lines, entry.Text)
	}
	wantLines := "read · sub/a.txt|fetch · https://example.com|execute · git push origin HEAD"
	if got := strings.Join(lines, "|"); got != wantLines {
		t.Fatalf("command log = %q, want %q", got, wantLines)
	}

	log := a.agentLog(t)
	if !strings.Contains(log, `permission {"outcome":{"optionId":"no","outcome":"selected"}}`) {
		t.Fatalf("a permission request must be rejected, agent saw:\n%s", log)
	}
	// Une méthode non servie reçoit une erreur, donc aucun résultat.
	if !strings.Contains(log, "fs \n") {
		t.Fatalf("file access must not be served, agent saw:\n%s", log)
	}
	if !strings.Contains(log, `prompt "CONTEXT\n\nFix the bug"`) {
		t.Fatalf("a fresh session should get the context, agent saw:\n%s", log)
	}
	if !strings.Contains(log, `"interactive":false`) || !strings.Contains(log, `"readTextFile":false`) {
		t.Fatalf("initialize should announce an unattended client without file access:\n%s", log)
	}
}

// TestRunACPResumesASession : la session connue est reprise, l'historique que
// l'agent rejoue n'est pas pris pour du travail neuf, et le contexte n'est pas
// renvoyé.
func TestRunACPResumesASession(t *testing.T) {
	a := newACPTestRun(t, "basic")
	a.task, _ = a.store.UpdateTask(a.task.ID, func(tk *Task) { tk.SessionID = "sess-old" })
	if err := a.run("And now?"); err != nil {
		t.Fatalf("runACP: %v", err)
	}
	for _, text := range a.texts() {
		if strings.Contains(text, "OLD ANSWER") {
			t.Fatalf("replayed history must not become a message: %q", a.texts())
		}
	}
	task, _ := a.store.GetTask(a.task.ID)
	for _, entry := range task.CommandLog {
		if strings.Contains(entry.Text, "old command") {
			t.Fatalf("replayed tools must not be logged: %+v", task.CommandLog)
		}
	}
	log := a.agentLog(t)
	if !strings.Contains(log, "session/load sess-old") || strings.Contains(log, "session/new") {
		t.Fatalf("the known session should be loaded, agent saw:\n%s", log)
	}
	if !strings.Contains(log, `prompt "And now?"`) {
		t.Fatalf("a resumed session should get the bare message, agent saw:\n%s", log)
	}
	if task.SessionID != "sess-old" {
		t.Fatalf("session = %q, want it unchanged", task.SessionID)
	}
}

// TestRunACPForgetsALostSession : une session introuvable est oubliée et le
// tour repart sur une session neuve, avec le contexte.
func TestRunACPForgetsALostSession(t *testing.T) {
	a := newACPTestRun(t, "lost")
	a.task, _ = a.store.UpdateTask(a.task.ID, func(tk *Task) { tk.SessionID = "sess-gone" })
	if err := a.run("And now?"); err != nil {
		t.Fatalf("runACP: %v", err)
	}
	task, _ := a.store.GetTask(a.task.ID)
	if task.SessionID != "sess-new" {
		t.Fatalf("session = %q, want the new one", task.SessionID)
	}
	log := a.agentLog(t)
	if !strings.Contains(log, "session/load sess-gone") || !strings.Contains(log, "session/new") {
		t.Fatalf("expected a load then a new session, agent saw:\n%s", log)
	}
	if !strings.Contains(log, `prompt "CONTEXT\n\nTask: `) {
		t.Fatalf("the new session should get the context and the task, agent saw:\n%s", log)
	}
}

// TestRunACPWithoutLoadSessionKeepsNoSession : garder l'identifiant d'un agent
// qui ne sait pas reprendre priverait le tour suivant de son historique.
func TestRunACPWithoutLoadSessionKeepsNoSession(t *testing.T) {
	a := newACPTestRun(t, "noload")
	if err := a.run("Fix the bug"); err != nil {
		t.Fatalf("runACP: %v", err)
	}
	if task, _ := a.store.GetTask(a.task.ID); task.SessionID != "" {
		t.Fatalf("session = %q, want none", task.SessionID)
	}
}

func TestRunACPNeverEndsSilently(t *testing.T) {
	a := newACPTestRun(t, "silent")
	err := a.run("Fix the bug")
	if err == nil || !strings.Contains(err.Error(), "the model is unavailable") {
		t.Fatalf("stderr should surface on a silent turn, got %v", err)
	}
}

func TestRunACPExplainsAMissingSignIn(t *testing.T) {
	a := newACPTestRun(t, "auth")
	err := a.run("Fix the bug")
	if err == nil || err.Error() != copilotSignInHint {
		t.Fatalf("error = %v, want the sign-in hint", err)
	}
}

// TestRunACPInterruptCancelsTheSession : interrompre passe d'abord par
// session/cancel, pour que l'agent enregistre sa session avant de sortir.
func TestRunACPInterruptCancelsTheSession(t *testing.T) {
	a := newACPTestRun(t, "cancel")
	errc := make(chan error, 1)
	go func() {
		err := a.run("Fix the bug")
		close(a.handle.done)
		errc <- err
	}()

	deadline := time.Now().Add(20 * time.Second)
	for {
		task, _ := a.store.GetTask(a.task.ID)
		if task.LiveActivity != nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the agent never started working")
		}
		time.Sleep(10 * time.Millisecond)
	}
	a.handle.interrupted.Store(true)
	killProcessGroup(a.handle)

	select {
	case err := <-errc:
		if err != nil {
			t.Fatalf("an interrupted run is not an error, got %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("the run did not stop")
	}
	if log := a.agentLog(t); !strings.Contains(log, "after prompt: session/cancel") {
		t.Fatalf("the agent should have been asked to cancel, it saw:\n%s", log)
	}
	if got := a.texts(); len(got) != 1 || got[0] != "Working" {
		t.Fatalf("what the agent said before the interruption should be kept, got %q", got)
	}
}

func TestACPRejectAllNeverGrants(t *testing.T) {
	cases := []struct {
		options []acpPermissionOption
		want    string
	}{
		{[]acpPermissionOption{{ID: "a", Kind: "allow_once"}, {ID: "b", Kind: "allow_always"}, {ID: "c", Kind: "reject_once"}}, "c"},
		{[]acpPermissionOption{{ID: "a", Kind: "allow_always"}, {ID: "d", Kind: "reject_always"}}, "d"},
		// Aucun refus proposé : on annule plutôt que d'accorder.
		{[]acpPermissionOption{{ID: "a", Kind: "allow_once"}}, ""},
		{nil, ""},
	}
	for _, c := range cases {
		if got := acpRejectAll(acpPermissionRequest{Options: c.options}); got != c.want {
			t.Fatalf("acpRejectAll(%+v) = %q, want %q", c.options, got, c.want)
		}
	}
}

func TestACPToolLine(t *testing.T) {
	cwd := "/data/worktrees/t-1"
	cases := []struct {
		call acpToolCall
		want string
	}{
		{acpToolCall{Kind: "execute", Title: "Run tests", RawInput: json.RawMessage(`{"command":"go test\n ./..."}`)}, "execute · go test ./..."},
		{acpToolCall{Kind: "read", RawInput: json.RawMessage(`{"path":"/data/worktrees/t-1/web/app.js"}`)}, "read · web/app.js"},
		{acpToolCall{Kind: "read", RawInput: json.RawMessage(`{"path":"/etc/hosts"}`)}, "read · /etc/hosts"},
		{acpToolCall{Kind: "edit", Title: "apply_patch", RawInput: json.RawMessage(`"*** Begin Patch"`), Path: "/data/worktrees/t-1/a.go"}, "edit · a.go"},
		{acpToolCall{Kind: "other", Title: "Thinking"}, "tool · Thinking"},
		{acpToolCall{}, "tool"},
	}
	for _, c := range cases {
		if got := acpToolLine(c.call, cwd); got != c.want {
			t.Fatalf("acpToolLine(%+v) = %q, want %q", c.call, got, c.want)
		}
	}
}

// TestCopilotArgsServeACPButKeepOutboundCommandsDenied : les autorisations
// sont fixées par les règles du CLI, jamais par un mode qui autorise tout.
func TestCopilotArgsServeACPButKeepOutboundCommandsDenied(t *testing.T) {
	args := copilotArgs(Agent{Model: "gpt-test"})
	joined := strings.Join(args, " ")
	for _, required := range []string{
		"--acp",
		"--no-ask-user",
		"--allow-tool=read,write,shell",
		"--deny-tool=" + copilotDeniedTools,
		"--disable-builtin-mcps",
		"--sync-session=off",
		"--model=gpt-test",
	} {
		if !strings.Contains(joined, required) {
			t.Fatalf("Copilot arguments should contain %q: %v", required, args)
		}
	}
	for _, required := range []string{"git push", "gh:*", "glab:*", ".github/copilot"} {
		if !strings.Contains(copilotDeniedTools, required) {
			t.Fatalf("Copilot deny rules should cover %q", required)
		}
	}
	// --no-remote a été retiré du CLI, qui refuse désormais de démarrer avec.
	for _, forbidden := range []string{"--allow-all", "--yolo", "--autopilot", "dangerously-skip-permissions", "--no-remote"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("Copilot arguments should not contain %q: %v", forbidden, args)
		}
	}
}

// TestPrepareCopilotHomeCopiesTheAccountAndNothingElse : le dossier isolé
// reprend le nom du compte, jamais le reste (hooks, jeton en clair, réglages).
func TestPrepareCopilotHomeCopiesTheAccountAndNothingElse(t *testing.T) {
	userHome := t.TempDir()
	userConfig := "// User settings belong in settings.json.\n// This file is managed automatically.\n" +
		`{"lastLoggedInUser":{"host":"https://github.com","login":"octocat"},` +
		`"loggedInUsers":[{"host":"https://github.com","login":"octocat"}],` +
		`"copilot_tokens":{"https://github.com:octocat":"secret"},"trustedFolders":["/somewhere"]}`
	if err := os.WriteFile(filepath.Join(userHome, "config.json"), []byte(userConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userHome, "settings.json"), []byte(`{"hooks":{"PermissionRequest":[{"command":"false"}]}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	dataDir := t.TempDir()
	home, err := prepareCopilotHome(dataDir, "t-1", userHome)
	if err != nil {
		t.Fatalf("prepareCopilotHome: %v", err)
	}
	if want := filepath.Join(dataDir, agentHomesDir, "t-1", "copilot"); home != want {
		t.Fatalf("home = %q, want %q", home, want)
	}
	config, _ := os.ReadFile(filepath.Join(home, "config.json"))
	if !strings.Contains(string(config), "octocat") {
		t.Fatalf("the account should be carried over: %s", config)
	}
	for _, leaked := range []string{"secret", "copilot_tokens", "trustedFolders"} {
		if strings.Contains(string(config), leaked) {
			t.Fatalf("%q must not be copied: %s", leaked, config)
		}
	}
	settings, _ := os.ReadFile(filepath.Join(home, "settings.json"))
	if strings.Contains(string(settings), "hooks") || !strings.Contains(string(settings), `"remoteExport": false`) {
		t.Fatalf("settings should be Sillage's own, got %s", settings)
	}
	if info, _ := os.Stat(home); info.Mode().Perm() != 0o700 {
		t.Fatalf("home mode = %v, want 0700", info.Mode().Perm())
	}

	// Tour suivant : l'état que le CLI a écrit reste, le compte suit celui de
	// l'utilisateur (ici déconnecté), les réglages sont remis d'aplomb.
	if err := os.WriteFile(filepath.Join(home, "config.json"), []byte("// managed\n"+`{"firstLaunchAt":"2026-01-01","loggedInUsers":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "settings.json"), []byte(`{"remoteExport": true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userHome, "config.json"), []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareCopilotHome(dataDir, "t-1", userHome); err != nil {
		t.Fatalf("prepareCopilotHome (second turn): %v", err)
	}
	config, _ = os.ReadFile(filepath.Join(home, "config.json"))
	if !strings.Contains(string(config), "firstLaunchAt") || strings.Contains(string(config), "loggedInUsers") {
		t.Fatalf("CLI state should stay and the account follow the user, got %s", config)
	}
	settings, _ = os.ReadFile(filepath.Join(home, "settings.json"))
	if string(settings) != copilotSettings {
		t.Fatalf("settings should be rewritten every turn, got %s", settings)
	}
}

func TestCopilotEnvIsolatesHomeAndDropsClassicTokens(t *testing.T) {
	env := copilotEnv([]string{"PATH=/bin", "GITHUB_TOKEN=ghp_classic", "COPILOT_HOME=/home/me/.copilot", "GH_TOKEN=github_pat_x"}, "/data/agent-homes/t-1/copilot")
	joined := strings.Join(env, " ")
	if strings.Contains(joined, "ghp_classic") || strings.Contains(joined, "/home/me/.copilot") {
		t.Fatalf("classic token and user home must not reach copilot: %v", env)
	}
	for _, kept := range []string{"PATH=/bin", "GH_TOKEN=github_pat_x", "COPILOT_HOME=/data/agent-homes/t-1/copilot"} {
		if !slicesContain(env, kept) {
			t.Fatalf("env should contain %q: %v", kept, env)
		}
	}
}

func TestCopilotTurnStartIgnoresSubagents(t *testing.T) {
	const method = "github.com/copilot/sessionEvent"
	if !copilotTurnStart(method, json.RawMessage(`{"sessionId":"s","type":"assistant.turn_start","data":{}}`)) {
		t.Fatal("the foreground turn start should be recognised")
	}
	for _, params := range []string{
		`{"type":"assistant.turn_start","agentId":"sub-1"}`,
		`{"type":"assistant.idle"}`,
	} {
		if copilotTurnStart(method, json.RawMessage(params)) {
			t.Fatalf("%s is not a foreground turn start", params)
		}
	}
	if copilotTurnStart("session/other", json.RawMessage(`{"type":"assistant.turn_start"}`)) {
		t.Fatal("another notification is not a turn start")
	}
}
