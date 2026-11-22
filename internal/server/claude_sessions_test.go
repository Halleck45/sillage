package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeSession(t *testing.T, cfg, cwd, id string, lines ...string) {
	t.Helper()
	dir := filepath.Join(cfg, "projects", claudeProjectDirName(cwd))
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, id+".jsonl"), []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestListClaudeSessions(t *testing.T) {
	cfg := t.TempDir()
	repo := "/work/api"
	user := func(cwd, ep, text string) string {
		return `{"type":"user","cwd":"` + cwd + `","gitBranch":"main","entrypoint":"` + ep + `","timestamp":"2026-01-02T10:00:00Z","message":{"role":"user","content":"` + text + `"}}`
	}
	// Une session interactive, avec un retour d'outil et une commande interne
	// qui ne comptent pas comme demandes.
	writeSession(t, cfg, repo, "a1",
		user(repo, "cli", "Fix the login bug"),
		`{"type":"user","cwd":"`+repo+`","entrypoint":"cli","toolUseResult":{"x":1},"message":{"role":"user","content":[{"type":"tool_result","content":"ok"}]}}`,
		user(repo, "cli", "<command-name>/clear</command-name>"),
		`{"type":"ai-title","aiTitle":"Login bug"}`,
		user(repo, "cli", "Now add a test"),
		`not json`)
	// Une exécution de Sillage (claude -p) : écartée.
	writeSession(t, cfg, repo, "b2", user(repo, "sdk-cli", "You are triaging a GitHub pull request"))
	// Un sous-dossier du dépôt : gardé.
	writeSession(t, cfg, repo+"/web", "c3", user(repo+"/web", "cli", "Style the header"))
	// Un dépôt voisin au nom encodé identique (« /work/api-v2 ») : écarté par le cwd.
	writeSession(t, cfg, "/work/api-v2", "d4", user("/work/api-v2", "cli", "Other project"))

	got := listClaudeSessions(Project{Repos: []Repo{{Name: "api", Path: repo}}}, []string{cfg}, time.Now())
	ids := map[string]ClaudeSession{}
	for _, s := range got {
		ids[s.ID] = s
	}
	if len(got) != 2 || ids["a1"].ID == "" || ids["c3"].ID == "" {
		t.Fatalf("attendu a1 et c3, reçu %+v", got)
	}
	a := ids["a1"]
	if a.Title != "Login bug" || a.PromptsN != 2 || a.LastPrompt != "Now add a test" || a.Branch != "main" || a.RepoName != "api" {
		t.Fatalf("session mal lue : %+v", a)
	}
	if !a.Active {
		t.Fatal("un fichier écrit à l'instant appartient à une session active")
	}
	if !strings.Contains(a.ResumeCmd, "claude --resume a1") || !strings.Contains(a.ResumeCmd, "CLAUDE_CONFIG_DIR=") {
		t.Fatalf("commande de reprise inattendue : %s", a.ResumeCmd)
	}
	if ids["c3"].Title != "Style the header" {
		t.Fatalf("titre de repli attendu sur la première demande, reçu %q", ids["c3"].Title)
	}
}
