package server

import (
	"strings"
	"testing"
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
