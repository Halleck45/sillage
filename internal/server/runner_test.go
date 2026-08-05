package server

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestParseCodexTokenStreamKeepsLastCumulativeTotal reproduit le bug réel :
// les événements token_count de codex portent des totaux CUMULÉS par
// exécution, pas des deltas. Un flux de 3 événements croissants ne doit
// jamais être sommé (100+250+400) : seul le dernier événement doit compter.
func TestParseCodexTokenStreamKeepsLastCumulativeTotal(t *testing.T) {
	lines := []string{
		`{"token_count":{"input_tokens":100,"output_tokens":10}}`,
		`{"token_count":{"input_tokens":250,"output_tokens":25}}`,
		`{"token_count":{"input_tokens":400,"output_tokens":40}}`,
	}
	tok, found := parseCodexTokenStream(lines)
	if !found {
		t.Fatalf("un événement de tokens aurait dû être détecté")
	}
	if tok.Input != 400 || tok.Output != 40 {
		t.Fatalf("tokens attendus 400 input / 40 output (dernier cumulé, pas la somme), reçu %+v", tok)
	}
}

// TestParseCodexTokenStreamTurnCompletedForm couvre la forme récente
// {"type":"turn.completed","usage":{...}}, avec la même règle de non-cumul.
func TestParseCodexTokenStreamTurnCompletedForm(t *testing.T) {
	lines := []string{
		`{"type":"turn.completed","usage":{"input_tokens":50,"output_tokens":5}}`,
		`{"type":"turn.completed","usage":{"input_tokens":120,"output_tokens":12}}`,
	}
	tok, found := parseCodexTokenStream(lines)
	if !found {
		t.Fatalf("un événement de tokens aurait dû être détecté")
	}
	if tok.Input != 120 || tok.Output != 12 {
		t.Fatalf("tokens attendus 120 input / 12 output, reçu %+v", tok)
	}
}

// TestParseCodexTokenStreamNoEvent vérifie qu'un flux sans événement de
// tokens (messages seuls, lignes vides, JSON invalide) ne trouve rien.
func TestParseCodexTokenStreamNoEvent(t *testing.T) {
	lines := []string{
		``,
		`{"msg":{"type":"agent_message","message":"bonjour"}}`,
		`not even json`,
	}
	tok, found := parseCodexTokenStream(lines)
	if found {
		t.Fatalf("aucun événement de tokens ne devrait être trouvé, reçu %+v", tok)
	}
}

// TestParseCodexRateLimits couvre le seul endroit où codex publie le quota de
// compte (fichier de session, jamais le flux --json) : deux fenêtres,
// primaire (5h) et secondaire (hebdomadaire).
func TestParseCodexRateLimits(t *testing.T) {
	var generic map[string]any
	raw := `{"type":"event_msg","payload":{"type":"token_count","info":{},"rate_limits":{
		"primary":{"used_percent":44.0,"window_minutes":300,"resets_at":1771005432},
		"secondary":{"used_percent":49.0,"window_minutes":10080,"resets_at":1771409457}}}}`
	if err := json.Unmarshal([]byte(raw), &generic); err != nil {
		t.Fatalf("fixture JSON invalide : %v", err)
	}
	windows, ok := parseCodexRateLimits(generic)
	if !ok {
		t.Fatalf("rate_limits aurait dû être détecté")
	}
	if len(windows) != 2 {
		t.Fatalf("attendu 2 fenêtres, reçu %d : %+v", len(windows), windows)
	}
	if windows[0].Label != "5h" || windows[0].UsedPercent != 44.0 {
		t.Fatalf("fenêtre primaire inattendue : %+v", windows[0])
	}
	if windows[1].Label != "week" || windows[1].UsedPercent != 49.0 {
		t.Fatalf("fenêtre secondaire inattendue : %+v", windows[1])
	}
}

// TestParseCodexRateLimitsIgnoresOtherEvents vérifie qu'un événement sans
// rate_limits (message, token_count sans compte, autre type) ne renvoie rien.
func TestParseCodexRateLimitsIgnoresOtherEvents(t *testing.T) {
	cases := []string{
		`{"type":"event_msg","payload":{"type":"agent_message","message":"bonjour"}}`,
		`{"type":"event_msg","payload":{"type":"token_count","info":{}}}`,
		`{"msg":{"type":"agent_message"}}`,
	}
	for _, raw := range cases {
		var generic map[string]any
		if err := json.Unmarshal([]byte(raw), &generic); err != nil {
			t.Fatalf("fixture JSON invalide : %v", err)
		}
		if windows, ok := parseCodexRateLimits(generic); ok {
			t.Fatalf("aucune fenêtre attendue pour %q, reçu %+v", raw, windows)
		}
	}
}

// TestQuotaWindowLabel couvre les deux fenêtres connues et le repli générique.
func TestQuotaWindowLabel(t *testing.T) {
	if got := quotaWindowLabel(300); got != "5h" {
		t.Fatalf("attendu \"5h\", reçu %q", got)
	}
	if got := quotaWindowLabel(10080); got != "week" {
		t.Fatalf("attendu \"week\", reçu %q", got)
	}
	if got := quotaWindowLabel(60); got != "60m" {
		t.Fatalf("attendu un repli générique \"60m\", reçu %q", got)
	}
}

// TestReadCodexRateLimits couvre la localisation du fichier de session par
// thread_id (structure sessions/AAAA/MM/JJ/rollout-...-<threadID>.jsonl) et
// la lecture du dernier instantané de quota qu'il contient.
func TestReadCodexRateLimits(t *testing.T) {
	root := t.TempDir()
	oldDir := codexSessionsDir
	codexSessionsDir = func() string { return root }
	defer func() { codexSessionsDir = oldDir }()

	dayDir := filepath.Join(root, "2026", "08", "04")
	if err := os.MkdirAll(dayDir, 0o755); err != nil {
		t.Fatalf("MkdirAll : %v", err)
	}
	threadID := "019fcb94-5cf5-72c1-a18f-05fc099ddeae"
	sessionPath := filepath.Join(dayDir, "rollout-2026-08-04T09-02-05-"+threadID+".jsonl")
	content := strings.Join([]string{
		`{"type":"event_msg","payload":{"type":"token_count","info":{},"rate_limits":{"primary":{"used_percent":22.0,"window_minutes":10080,"resets_at":1786296634}}}}`,
		`{"type":"event_msg","payload":{"type":"token_count","info":{},"rate_limits":{"primary":{"used_percent":23.0,"window_minutes":10080,"resets_at":1786296634}}}}`,
	}, "\n")
	if err := os.WriteFile(sessionPath, []byte(content+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile : %v", err)
	}

	windows, ok := readCodexRateLimits(threadID)
	if !ok {
		t.Fatalf("un instantané aurait dû être trouvé")
	}
	if len(windows) != 1 || windows[0].UsedPercent != 23.0 {
		t.Fatalf("attendu le DERNIER instantané (23%%), reçu %+v", windows)
	}
	if !windows[0].ResetsAt.Equal(time.Unix(1786296634, 0)) {
		t.Fatalf("resetsAt inattendu : %v", windows[0].ResetsAt)
	}
}

// TestReadCodexRateLimitsMissingFile vérifie qu'un thread_id sans fichier de
// session correspondant ne fait pas planter le best-effort.
func TestReadCodexRateLimitsMissingFile(t *testing.T) {
	root := t.TempDir()
	oldDir := codexSessionsDir
	codexSessionsDir = func() string { return root }
	defer func() { codexSessionsDir = oldDir }()

	if windows, ok := readCodexRateLimits("unknown-thread"); ok {
		t.Fatalf("aucun instantané attendu, reçu %+v", windows)
	}
	if windows, ok := readCodexRateLimits(""); ok {
		t.Fatalf("thread_id vide : aucun instantané attendu, reçu %+v", windows)
	}
}

// TestBuildSystemPrompt couvre l'adaptateur claude : chaque bloc (agent,
// projet, chantier) n'apparaît que s'il est non vide, séparés par des lignes
// vides, dans cet ordre.
func TestBuildSystemPrompt(t *testing.T) {
	if got := buildSystemPrompt("", "", ""); got != "" {
		t.Fatalf("attendu une chaîne vide quand tout est vide, reçu %q", got)
	}
	if got := buildSystemPrompt("Agent context", "", ""); got != "Agent context" {
		t.Fatalf("résultat inattendu (agent seul) : %q", got)
	}
	if got := buildSystemPrompt("", "Project ctx", ""); got != "Project context:\nProject ctx" {
		t.Fatalf("résultat inattendu (projet seul) : %q", got)
	}
	if got := buildSystemPrompt("", "", "Workstream ctx"); got != "Workstream context:\nWorkstream ctx" {
		t.Fatalf("résultat inattendu (chantier seul) : %q", got)
	}
	want := "Agent context\n\nProject context:\nProject ctx\n\nWorkstream context:\nWorkstream ctx"
	if got := buildSystemPrompt("Agent context", "Project ctx", "Workstream ctx"); got != want {
		t.Fatalf("résultat combiné attendu %q, reçu %q", want, got)
	}
	// Projet seul manquant, agent + chantier présents : pas de bloc vide entre les deux.
	want2 := "Agent context\n\nWorkstream context:\nWorkstream ctx"
	if got := buildSystemPrompt("Agent context", "", "Workstream ctx"); got != want2 {
		t.Fatalf("résultat attendu %q, reçu %q", want2, got)
	}
}

// TestContextPartsCodexPrefix couvre l'adaptateur codex : le préfixe de
// prompt combine "Project context:"/"Workstream context:" (blocs omis s'ils
// sont vides) avant le séparateur "---".
func TestContextPartsCodexPrefix(t *testing.T) {
	prefixFor := func(project, workstream string) string {
		blocks := strings.Join(contextParts(project, workstream), "\n\n")
		if blocks == "" {
			return ""
		}
		return blocks + "\n\n---\n\n"
	}

	if got := prefixFor("", ""); got != "" {
		t.Fatalf("aucun préfixe attendu quand tout est vide, reçu %q", got)
	}
	want := "Project context:\nP\n\nWorkstream context:\nW\n\n---\n\n"
	if got := prefixFor("P", "W"); got != want {
		t.Fatalf("préfixe attendu %q, reçu %q", want, got)
	}
	want2 := "Workstream context:\nW\n\n---\n\n"
	if got := prefixFor("", "W"); got != want2 {
		t.Fatalf("préfixe attendu %q, reçu %q", want2, got)
	}
}

func TestCodexArgsAllowLinkedWorktreeGitWritesWithoutExposingHooks(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	root := t.TempDir()
	repo := filepath.Join(root, "repo with spaces")
	worktrees := filepath.Join(root, "task worktrees")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("create repository directory: %v", err)
	}
	initTestRepo(t, repo)
	worktree, _, err := CreateWorktree(repo, worktrees, "task one", "sillage/test-codex-git", "main")
	if err != nil {
		t.Fatalf("CreateWorktree: %v", err)
	}

	gitDir, err := resolveCodexGitDir(worktree, "--absolute-git-dir")
	if err != nil {
		t.Fatalf("resolve worktree git-dir: %v", err)
	}
	commonDir, err := resolveCodexGitDir(worktree, "--git-common-dir")
	if err != nil {
		t.Fatalf("resolve common git-dir: %v", err)
	}
	dirs := codexGitWritableDirs(worktree)

	for _, required := range []string{
		gitDir,
		filepath.Join(commonDir, "objects"),
		filepath.Join(commonDir, "refs"),
		filepath.Join(commonDir, "logs"),
	} {
		required, err = canonicalExistingDir(required)
		if err != nil {
			t.Fatalf("resolve required directory: %v", err)
		}
		if !containsString(dirs, required) {
			t.Errorf("Codex writable directories should contain %q: %v", required, dirs)
		}
	}
	for _, protected := range []string{commonDir, filepath.Join(commonDir, "hooks")} {
		protected, err = canonicalExistingDir(protected)
		if err != nil {
			t.Fatalf("resolve protected directory: %v", err)
		}
		if containsString(dirs, protected) {
			t.Errorf("Codex writable directories must not expose %q: %v", protected, dirs)
		}
	}

	args := codexArgs(worktree, "workspace-write", Agent{Model: "gpt-test"}, "Commit the change")
	for _, dir := range dirs {
		if !containsAdjacentArgs(args, "--add-dir", dir) {
			t.Errorf("Codex arguments should grant write access to %q: %v", dir, args)
		}
	}
	if !containsAdjacentArgs(args, "--model", "gpt-test") {
		t.Errorf("Codex arguments should preserve the selected model: %v", args)
	}
	if args[len(args)-1] != "Commit the change" {
		t.Errorf("Codex prompt should be the final argument: %v", args)
	}

	// Reproduce the sandbox boundary without launching Codex: the common
	// git-dir is read-only, and only the directories selected above are made
	// writable. A real add and commit must still succeed.
	t.Cleanup(func() {
		_ = filepath.Walk(commonDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil
			}
			if info.IsDir() {
				return os.Chmod(path, 0o755)
			}
			return os.Chmod(path, 0o644)
		})
	})
	chmodTree(t, commonDir, 0o555, 0o444)
	for _, dir := range dirs {
		chmodTree(t, dir, 0o755, 0o644)
	}
	if err := os.WriteFile(filepath.Join(worktree, "codex.txt"), []byte("committed from the task worktree\n"), 0o644); err != nil {
		t.Fatalf("write task file: %v", err)
	}
	runTestGit(t, worktree, "add", "codex.txt")
	runTestGit(t, worktree, "commit", "-m", "Test Codex worktree commit")
	if status := strings.TrimSpace(runTestGit(t, worktree, "status", "--porcelain")); status != "" {
		t.Fatalf("worktree should be clean after commit, got %q", status)
	}
}

func TestCodexGitWritableDirsSkipMetadataInsideWorkspace(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	repo := t.TempDir()
	initTestRepo(t, repo)
	if dirs := codexGitWritableDirs(repo); len(dirs) != 0 {
		t.Fatalf("a regular repository should already be writable through the workspace: %v", dirs)
	}
	if dirs := codexGitWritableDirs(t.TempDir()); len(dirs) != 0 {
		t.Fatalf("a non-Git directory should not add writable roots: %v", dirs)
	}
}

func TestCodexGitWritableDirsSupportExternalSeparateGitDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}

	root := t.TempDir()
	worktree := filepath.Join(root, "workspace")
	gitDir := filepath.Join(root, "external metadata", "repository.git")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(gitDir), 0o755); err != nil {
		t.Fatalf("create metadata parent: %v", err)
	}
	runTestGit(t, root, "init", "-b", "main", "--separate-git-dir", gitDir, worktree)

	want, err := canonicalExistingDir(gitDir)
	if err != nil {
		t.Fatalf("resolve separate git-dir: %v", err)
	}
	if got := codexGitWritableDirs(worktree); len(got) != 1 || got[0] != want {
		t.Fatalf("external separate git-dir should be added as one writable root: got %v, want %q", got, want)
	}
}

func TestCodexArgsDoNotExtendNonWorkspaceSandboxModes(t *testing.T) {
	args := codexArgs(t.TempDir(), "danger-full-access", Agent{}, "Inspect the change")
	if containsString(args, "--add-dir") {
		t.Fatalf("danger-full-access should not receive redundant writable roots: %v", args)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsAdjacentArgs(args []string, flag, value string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

func chmodTree(t *testing.T, root string, dirMode, fileMode os.FileMode) {
	t.Helper()
	if err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return os.Chmod(path, dirMode)
		}
		return os.Chmod(path, fileMode)
	}); err != nil {
		t.Fatalf("change permissions below %q: %v", root, err)
	}
}

func TestPrefixAgentContext(t *testing.T) {
	agent := Agent{ContextPrompt: "Agent context"}
	project := Project{ContextPrompt: "Project context"}
	card := Card{ContextPrompt: "Workstream context"}
	want := "Agent context\n\nProject context:\nProject context\n\nWorkstream context:\nWorkstream context\n\n---\n\nTask text"
	if got := prefixAgentContext(agent, project, card, "Task text"); got != want {
		t.Fatalf("prefixed input = %q, want %q", got, want)
	}
	if got := prefixAgentContext(Agent{}, Project{}, Card{}, "Task text"); got != "Task text" {
		t.Fatalf("input without context = %q, want unchanged text", got)
	}
}

func TestCopilotArgsAreAutonomousButKeepOutboundCommandsDenied(t *testing.T) {
	args := copilotArgs(Agent{Model: "gpt-test"}, "Fix the bug")
	joined := strings.Join(args, " ")
	for _, required := range []string{
		"--autopilot",
		"--no-ask-user",
		"--allow-tool=read,write,shell",
		"--deny-tool=" + copilotDeniedTools,
		"--disable-builtin-mcps",
		"--no-remote",
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
	for _, forbidden := range []string{"--allow-all", "--yolo", "dangerously-skip-permissions"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("Copilot arguments should not contain %q: %v", forbidden, args)
		}
	}
	if len(args) < 2 || args[len(args)-2] != "-p" || args[len(args)-1] != "Fix the bug" {
		t.Fatalf("Copilot prompt arguments are malformed: %v", args)
	}
}

func TestAntigravityArgsUseSandbox(t *testing.T) {
	args := antigravityArgs(Agent{Model: "gemini-test"}, "/tmp/wt", "Fix the bug")
	joined := strings.Join(args, " ")
	for _, required := range []string{"--print", "--sandbox", "--print-timeout=60m", "--model gemini-test", "--add-dir /tmp/wt"} {
		if !strings.Contains(joined, required) {
			t.Fatalf("Antigravity arguments should contain %q: %v", required, args)
		}
	}
	if strings.Contains(joined, "dangerously-skip-permissions") {
		t.Fatalf("Antigravity must not bypass permissions: %v", args)
	}
	// --print prend le prompt en valeur : il doit rester juste avant lui,
	// sinon agy exécute le drapeau suivant comme prompt.
	if len(args) < 2 || args[len(args)-2] != "--print" || args[len(args)-1] != "Fix the bug" {
		t.Fatalf("Antigravity prompt arguments are malformed: %v", args)
	}
}

func TestBuildTranscript(t *testing.T) {
	msgs := []Message{
		{Author: "agent", AuthorName: "Otto", Text: "J'ai trouvé le bloc dans show_lead.html.twig lignes 78-80."},
		{Author: "agent", AuthorName: "", Text: "[reassigned:bolt]"},
		{Author: "user", AuthorName: "JF", Text: "réessaie"},
	}
	got := buildTranscript(msgs)
	if strings.Contains(got, "[reassigned:") {
		t.Fatalf("les marqueurs système ne doivent pas apparaître : %q", got)
	}
	if !strings.Contains(got, "[Otto] J'ai trouvé le bloc") || !strings.Contains(got, "[JF] réessaie") {
		t.Fatalf("transcript incomplet : %q", got)
	}

	// Budget total : seuls les messages récents survivent.
	var many []Message
	for i := 0; i < 50; i++ {
		many = append(many, Message{Author: "agent", AuthorName: "Bolt", Text: strings.Repeat("x", 500) + " fin" + string(rune('A'+i%26))})
	}
	out := buildTranscript(many)
	if len(out) > 6100 {
		t.Fatalf("transcript trop long : %d", len(out))
	}
	if buildTranscript(nil) != "" {
		t.Fatalf("transcript vide attendu pour aucun message")
	}
}

// Le socle doit rester agnostique : y remettre une commande de langage ferait
// hériter tout projet d'une allowlist écrite pour un autre (voir
// Project.AllowedTools). Ce test est le garde-fou de cette règle.
func TestClaudeSocleStaysLanguageAgnostic(t *testing.T) {
	for _, forbidden := range []string{"go build", "go test", "go vet", "gofmt", "npm", "pytest", "cargo", "node"} {
		if strings.Contains(claudeAllowedTools, forbidden) {
			t.Fatalf("le socle contient %q : les commandes de langage vont dans Project.AllowedTools", forbidden)
		}
	}
	// Le git du socle est en lecture seule : aucune sous-commande qui écrive.
	for _, forbidden := range []string{"git push", "git commit", "git merge", "git rebase", "Bash(git:*)"} {
		if strings.Contains(claudeAllowedTools, forbidden) {
			t.Fatalf("le socle contient %q : le git du socle est en lecture seule", forbidden)
		}
	}
}

// Invariant 1 : rien de capable de pousser, ni dans le socle ni hors du refus
// figé. Un agent qui pourrait écrire .claude/settings.json s'accorderait de
// nouveaux droits au run suivant, d'où sa présence dans le refus.
func TestClaudeDeniedToolsCoversOutboundAndSelfEscalation(t *testing.T) {
	for _, required := range []string{"git push", "gh", "glab", ".claude/settings"} {
		if !strings.Contains(claudeDeniedTools, required) {
			t.Fatalf("le refus figé ne couvre pas %q (invariant 1)", required)
		}
	}
	if strings.Contains(claudeAllowedTools, "push") {
		t.Fatalf("le socle ne doit jamais contenir d'entrée capable de pousser")
	}
}

func TestClaudeToolsAppendsProjectTools(t *testing.T) {
	if got := claudeTools(nil); got != claudeAllowedTools {
		t.Fatalf("sans outil de projet, la liste doit être le socle seul ; reçu %q", got)
	}
	// Les entrées vides viennent d'un champ saisi une ligne par outil.
	got := claudeTools([]string{" Bash(pytest:*) ", "", "Bash(ruff:*)"})
	want := claudeAllowedTools + ",Bash(pytest:*),Bash(ruff:*)"
	if got != want {
		t.Fatalf("liste attendue %q, reçue %q", want, got)
	}
}

// Le contenu d'un tool_result arrive tantôt en chaîne, tantôt en liste de blocs
// typés selon l'appel : les deux formes doivent rendre le même texte, sinon un
// refus de permission passe inaperçu et le marqueur n'est jamais posé.
func TestClaudeToolResultText(t *testing.T) {
	const denial = "Claude requested permissions to use Bash"
	if got := claudeToolResultText(json.RawMessage(`"` + denial + `"`)); got != denial {
		t.Fatalf("forme chaîne : attendu %q, reçu %q", denial, got)
	}
	blocks := json.RawMessage(`[{"type":"text","text":"` + denial + `"}]`)
	if got := claudeToolResultText(blocks); got != denial {
		t.Fatalf("forme liste de blocs : attendu %q, reçu %q", denial, got)
	}
	// Contenu absent ou forme inattendue : pas de texte, donc pas de marqueur.
	if got := claudeToolResultText(nil); got != "" {
		t.Fatalf("contenu absent : attendu vide, reçu %q", got)
	}
	if got := claudeToolResultText(json.RawMessage(`{"unexpected":1}`)); got != "" {
		t.Fatalf("forme inconnue : attendu vide, reçu %q", got)
	}
}

// Un marqueur de refus est une ligne système, pas un tour de conversation : il
// ne doit pas être rejoué au CLI quand la session est repartie de zéro (codex,
// tâche réassignée), sinon l'agent lit ses propres refus comme du contexte.
func TestBuildTranscriptExcludesToolDeniedMarker(t *testing.T) {
	got := buildTranscript([]Message{
		{Author: "agent", AuthorName: "", Text: "[tool-denied:Bash · pytest -q]"},
		{Author: "agent", AuthorName: "Bolt", Text: "je contourne autrement"},
	})
	if strings.Contains(got, "tool-denied") {
		t.Fatalf("le marqueur de refus ne doit pas être rejoué : %q", got)
	}
	if !strings.Contains(got, "je contourne autrement") {
		t.Fatalf("transcript incomplet : %q", got)
	}
}

// Un test rouge ou un fichier absent font partie du travail de l'agent : seul
// un refus de permission mérite un marqueur dans le fil.
func TestIsPermissionDenial(t *testing.T) {
	denied := []string{
		"Claude requested permissions to write to .claude/skills/x/SKILL.md, but you haven't granted it yet",
		"Claude requested permissions to use Bash",
		"Agent does not have permission to use Bash(pytest:*)",
	}
	for _, text := range denied {
		if !isPermissionDenial(text) {
			t.Fatalf("refus de permission non détecté : %q", text)
		}
	}
	ordinary := []string{
		"FAIL github.com/org/repo 0.12s",
		"cat: /tmp/nope: No such file or directory",
		"exit status 1",
		"",
	}
	for _, text := range ordinary {
		if isPermissionDenial(text) {
			t.Fatalf("erreur d'exécution ordinaire prise pour un refus : %q", text)
		}
	}
}
