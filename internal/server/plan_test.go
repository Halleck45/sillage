package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestPlanAllowedToolsIsReadOnly : la planification tourne dans le worktree du
// chantier ou dans le dépôt du projet, pas dans une branche jetable relue par
// un humain. Rien de ce qu'elle écrirait ne passerait par une revue, donc
// l'allowlist ne doit contenir aucun outil d'écriture ni la moindre sortie.
func TestPlanAllowedToolsIsReadOnly(t *testing.T) {
	for _, forbidden := range []string{"Edit", "Write", "push", "gh", "glab", "WebFetch",
		"git commit", "git add", "git checkout", "mkdir", "rm"} {
		if strings.Contains(planAllowedTools, forbidden) {
			t.Errorf("planAllowedTools ne doit pas contenir %q : %s", forbidden, planAllowedTools)
		}
	}
	for _, required := range []string{"Read", "Grep", "Glob"} {
		if !strings.Contains(planAllowedTools, required) {
			t.Errorf("planAllowedTools devrait contenir %q pour qu'un plan soit informé", required)
		}
	}
}

// TestCanPlanOnlyReadOnlyCLIs : un CLI dont Sillage ne peut pas garantir la
// lecture seule ne planifie pas.
func TestCanPlanOnlyReadOnlyCLIs(t *testing.T) {
	for _, cli := range []string{"claude", "codex", "fake"} {
		if !canPlan(cli) {
			t.Errorf("%s devrait savoir planifier", cli)
		}
	}
	for _, cli := range []string{"copilot", "agy", "kiro", "", "unknown"} {
		if canPlan(cli) {
			t.Errorf("%s ne devrait pas savoir planifier : Sillage n'y garantit pas la lecture seule", cli)
		}
	}
}

func TestExtractPlanStepsFromFencedJSON(t *testing.T) {
	reply := "Voici mon découpage.\n\n```json\n" +
		`{"steps":[{"title":"Poser le modèle","prompt":"Ajouter les structs."},` +
		`{"title":"Brancher l'UI","prompt":"Ajouter la modale."}]}` +
		"\n```\n\nDis-moi si ça convient.\n"
	steps, ok := extractPlanSteps(reply)
	if !ok {
		t.Fatalf("plan non trouvé dans %q", reply)
	}
	if len(steps) != 2 {
		t.Fatalf("2 étapes attendues, reçu %d", len(steps))
	}
	if steps[0].Title != "Poser le modèle" || steps[1].Prompt != "Ajouter la modale." {
		t.Fatalf("étapes mal décodées : %+v", steps)
	}
}

// TestExtractPlanStepsKeepsLastObject : plusieurs modèles recopient l'exemple
// du prompt avant de répondre. C'est le dernier objet exploitable qui compte.
func TestExtractPlanStepsKeepsLastObject(t *testing.T) {
	reply := `Le format demandé est {"steps":[{"title":"exemple","prompt":"exemple"}]}.` + "\n" +
		`Voici le vrai plan : {"steps":[{"title":"Étape réelle","prompt":"Faire le travail."}]}`
	steps, ok := extractPlanSteps(reply)
	if !ok {
		t.Fatal("plan non trouvé")
	}
	if len(steps) != 1 || steps[0].Title != "Étape réelle" {
		t.Fatalf("le dernier objet devait gagner, reçu %+v", steps)
	}
}

func TestExtractPlanStepsRejectsNonPlan(t *testing.T) {
	for _, reply := range []string{
		"",
		"Je ne peux pas planifier ce chantier.",
		`{"steps":[]}`,
		`{"steps":[{"title":"   ","prompt":"vide"}]}`,
		`{"autre":"chose"}`,
	} {
		if steps, ok := extractPlanSteps(reply); ok {
			t.Errorf("extractPlanSteps(%q) devrait échouer, reçu %+v", reply, steps)
		}
	}
}

// TestSanitizePlanStepsCapsAndTrims : tronquer plutôt que refuser, un plan
// presque bon restant éditable par l'humain.
func TestSanitizePlanStepsCapsAndTrims(t *testing.T) {
	raw := []PlanStep{{Title: "  Nettoyer  ", Prompt: "  faire  "}}
	for i := 0; i < planMaxSteps+5; i++ {
		raw = append(raw, PlanStep{Title: "Étape", Prompt: "Faire"})
	}
	raw = append(raw, PlanStep{Title: strings.Repeat("x", planTitleMax+50)})

	steps := sanitizePlanSteps(raw)
	if len(steps) != planMaxSteps {
		t.Fatalf("le plan devait être plafonné à %d étapes, reçu %d", planMaxSteps, len(steps))
	}
	if steps[0].Title != "Nettoyer" || steps[0].Prompt != "faire" {
		t.Fatalf("titre et prompt devaient être détourés : %+v", steps[0])
	}
}

func TestSanitizePlanStepsTruncatesLongTitle(t *testing.T) {
	steps := sanitizePlanSteps([]PlanStep{{Title: strings.Repeat("x", planTitleMax+50), Prompt: "p"}})
	if len(steps) != 1 {
		t.Fatalf("1 étape attendue, reçu %d", len(steps))
	}
	if len(steps[0].Title) != planTitleMax {
		t.Fatalf("titre attendu tronqué à %d, reçu %d", planTitleMax, len(steps[0].Title))
	}
}

// TestPlanInstructionsCarriesContext : un plan écrit sans les instructions du
// projet propose des étapes que le projet n'accepte pas.
func TestPlanInstructionsCarriesContext(t *testing.T) {
	card := Card{Title: "Refonte auth", ContextPrompt: "Ne pas toucher aux sessions"}
	project := Project{ContextPrompt: "Go 1.24, pas de dépendance externe"}
	prompt := planInstructions(card, project, "Commence par le stockage")

	for _, want := range []string{
		"Refonte auth",
		"Ne pas toucher aux sessions",
		"Go 1.24, pas de dépendance externe",
		"Commence par le stockage",
		`{"steps":`,
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("le prompt de planification devrait contenir %q :\n%s", want, prompt)
		}
	}
	if !strings.Contains(prompt, "NOT implementing") {
		t.Error("le prompt doit dire à l'agent qu'il planifie et n'écrit rien")
	}
}

// TestFakePlanReplyParses : l'agent simulé imite un vrai CLI, barrières
// comprises, pour éprouver le parseur autant que l'interface.
func TestFakePlanReplyParses(t *testing.T) {
	steps, ok := extractPlanSteps(fakePlanReply(Card{Title: "Chantier d'essai"}))
	if !ok {
		t.Fatal("la réponse de l'agent simulé devrait se décoder")
	}
	if len(steps) != 3 {
		t.Fatalf("3 étapes attendues de l'agent simulé, reçu %d", len(steps))
	}
	if !strings.Contains(steps[0].Prompt, "Chantier d'essai") {
		t.Errorf("le prompt simulé devrait citer le chantier : %q", steps[0].Prompt)
	}
}

// plan appelle POST /api/cards/{id}/plan sur la fixture de livraison.
func (f *deliveryFixture) plan(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.SetPathValue("id", f.card.ID)
	w := httptest.NewRecorder()
	f.srv.handleCardPlan(w, req)
	return w
}

// TestCardPlanWithFakeAgent : le flux complet de la route, sans process
// externe. La planification ne crée RIEN : ni tâche, ni branche de chantier.
func TestCardPlanWithFakeAgent(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge"})
	w := f.plan(t, `{"agentId":"echo"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var resp PlanResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if resp.CardID != f.card.ID || resp.AgentID != "echo" || resp.RepoName != "demo" {
		t.Fatalf("réponse mal renseignée : %+v", resp)
	}
	if len(resp.Steps) == 0 {
		t.Fatal("le plan devrait porter des étapes")
	}
	if tasks := f.srv.store.TasksByCard(f.card.ID); len(tasks) != 0 {
		t.Fatalf("planifier ne doit créer aucune tâche, reçu %d", len(tasks))
	}
	if _, ok := f.srv.store.GetCardBranch(f.card.ID, "demo"); ok {
		t.Fatal("planifier ne doit créer aucune branche de chantier : c'est une lecture")
	}
}

func TestCardPlanRefusesAgentThatCannotPlan(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge"})
	agent, err := f.srv.store.AddAgent("Copilote", "🧭", "#888888", "copilot", "", "")
	if err != nil {
		t.Fatalf("AddAgent: %v", err)
	}
	w := f.plan(t, `{"agentId":"`+agent.ID+`"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("attendu 400, reçu %d (%s)", w.Code, w.Body.String())
	}
}

func TestCardPlanRequiresAgent(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge"})
	if w := f.plan(t, `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("sans agent : attendu 400, reçu %d", w.Code)
	}
	if w := f.plan(t, `{"agentId":"inconnu"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("agent inconnu : attendu 400, reçu %d", w.Code)
	}
}
