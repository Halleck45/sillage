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
