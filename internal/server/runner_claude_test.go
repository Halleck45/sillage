package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeClaudeScript imite le CLI claude : une reprise (--resume) échoue comme
// une session créée sous un autre compte, un départ frais répond. Le dernier
// argument (le prompt) est gardé dans prompt.txt.
const fakeClaudeScript = `#!/bin/sh
for last; do :; done
printf '%s' "$last" > "$(dirname "$0")/prompt.txt"
for a in "$@"; do
  if [ "$a" = "--resume" ]; then
    echo "No conversation found with session ID: old-session" >&2
    echo '{"type":"result","subtype":"error_during_execution","is_error":true,"num_turns":0}'
    exit 1
  fi
done
echo '{"type":"system","subtype":"init","session_id":"new-session"}'
echo '{"type":"assistant","message":{"content":[{"type":"text","text":"repris"}]}}'
echo '{"type":"result","subtype":"success","is_error":false,"result":"repris"}'
`

func TestClaudeLostSessionRestartsWithTranscript(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(fakeClaudeScript), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	id, _ := f.addTask(t, "Billets", "a.go", "package a\n")
	store := f.srv.store
	if _, _, err := store.AddMessage(id, "user", "", "écris le premier billet"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.AddMessage(id, "agent", "Muse", "premier billet écrit"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := store.AddMessage(id, "agent", "Muse", "⚠️ the agent reported an error"); err != nil {
		t.Fatal(err)
	}
	task, err := store.UpdateTask(id, func(tk *Task) { tk.SessionID = "old-session" })
	if err != nil {
		t.Fatal(err)
	}
	agent := Agent{ID: "muse", Name: "Muse", Cli: "claude"}
	handle := &procHandle{done: make(chan struct{})}
	if err := f.srv.runner.runClaude(&task, agent, f.project, f.card, handle, "réessaie", permissionRuntime{}); err != nil {
		t.Fatalf("la reprise perdue aurait dû repartir d'une session neuve : %v", err)
	}
	if task.SessionID != "new-session" {
		t.Fatalf("session attendue new-session, reçue %q", task.SessionID)
	}
	prompt, _ := os.ReadFile(filepath.Join(bin, "prompt.txt"))
	p := string(prompt)
	if !strings.Contains(p, "premier billet écrit") || !strings.HasSuffix(p, "réessaie") || strings.Contains(p, "⚠️") {
		t.Fatalf("le départ frais doit rejouer le fil sans les erreurs, prompt :\n%s", p)
	}
}

func TestClaudeErrorCarriesReason(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	bin := t.TempDir()
	script := "#!/bin/sh\necho 'Not logged in' >&2\necho '{\"type\":\"result\",\"is_error\":true}'\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "claude"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	id, _ := f.addTask(t, "X", "x.go", "package x\n")
	task, _ := f.srv.store.GetTask(id)
	err := f.srv.runner.runClaude(&task, Agent{Name: "Muse", Cli: "claude"}, f.project, f.card, &procHandle{done: make(chan struct{})}, "go", permissionRuntime{})
	if err == nil || !strings.Contains(err.Error(), "Not logged in") {
		t.Fatalf("l'erreur doit porter la raison du CLI, reçu %v", err)
	}
}

func TestRestartTaskDropsWorkAndConversation(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	id, _ := f.addTask(t, "Billets", "billet.md", "brouillon\n")
	task, _ := f.srv.store.GetTask(id)
	runTestGit(t, task.WorktreeDir, "add", "-A")
	runTestGit(t, task.WorktreeDir, "commit", "-q", "-m", "wip")
	if err := os.WriteFile(filepath.Join(task.WorktreeDir, "notes.txt"), []byte("non suivi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	store := f.srv.store
	store.AddMessage(id, "user", "", "ancienne consigne")
	store.AddMessage(id, "agent", "Écho", "ancienne réponse")
	store.UpdateTask(id, func(tk *Task) { tk.SessionID = "old"; tk.Review = &TaskReview{Verdict: "approved"} })

	got, msg, err := f.srv.runner.RestartTask(id)
	if err != nil {
		t.Fatalf("RestartTask: %v", err)
	}
	for _, name := range []string{"billet.md", "notes.txt"} {
		if _, err := os.Stat(filepath.Join(task.WorktreeDir, name)); !os.IsNotExist(err) {
			t.Fatalf("%s aurait dû disparaître du worktree", name)
		}
	}
	if got.SessionID != "" || got.Review != nil || got.CommitsCount != 0 || got.FilesCount != 0 || got.Status != "review" {
		t.Fatalf("tâche mal remise à zéro : %+v", got)
	}
	if msg.Text != "[restarted:echo]" {
		t.Fatalf("marqueur inattendu : %q", msg.Text)
	}
	store.AddMessage(id, "user", "", "nouvelle consigne")
	if in := f.srv.runner.prepareCliInput(got, "nouvelle consigne"); strings.Contains(in, "ancienne") {
		t.Fatalf("rien d'avant le redémarrage ne doit être rejoué :\n%s", in)
	}
	store.UpdateTask(id, func(tk *Task) { tk.Status = "accepted" })
	if _, _, err := f.srv.runner.RestartTask(id); err == nil {
		t.Fatal("une tâche acceptée ne doit pas pouvoir repartir de zéro")
	}
}
