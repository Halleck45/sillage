package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtractPRDraft(t *testing.T) {
	reply := "Here is the example: {\"title\":\"...\",\"body\":\"...\"}\nAnd my answer:\n```json\n" +
		`{"title":"fix(report): keep tooltips visible in tables\nsecond line","body":"Tooltips were clipped.\n\n🤖 Generated with Claude Code"}` + "\n```"
	title, body, ok := extractPRDraft(reply)
	if !ok || title != "fix(report): keep tooltips visible in tables" || body != "Tooltips were clipped." {
		t.Fatalf("brouillon inattendu : ok=%v title=%q body=%q", ok, title, body)
	}
	if _, _, ok := extractPRDraft("no json here"); ok {
		t.Fatal("une réponse sans JSON ne doit pas donner de brouillon")
	}
}

// Sans rédaction, la pull request d'un chantier à un seul commit prend le
// sujet de ce commit, pas le titre du chantier ; l'aperçu le préremplit.
func TestPRDefaultsAndDraft(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	id, cb := f.addTask(t, "issue 197 - tooltip", "c.txt", "c\n")
	task, _ := f.srv.store.GetTask(id)
	runTestGit(t, task.WorktreeDir, "add", "-A")
	runTestGit(t, task.WorktreeDir, "commit", "-q", "-m", "fix: keep tooltips visible")
	if w := f.accept(t, id); w.Code != http.StatusOK {
		t.Fatalf("accept : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	if got := defaultPRTitle(cb.WorktreeDir, "main", cb.Branch, "issue 197 - tooltip"); got != "fix: keep tooltips visible" {
		t.Fatalf("titre par défaut inattendu : %q", got)
	}
	row := f.delivery(t).Repos[0]
	if row.PrTitle != "fix: keep tooltips visible" || row.PrBody != "- fix: keep tooltips visible" {
		t.Fatalf("aperçu sans titre ni corps par défaut : %+v", row)
	}

	// La tâche a été faite par l'agent simulé : c'est lui qui rédige.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"repoName":"`+cb.RepoName+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", f.card.ID)
	w := httptest.NewRecorder()
	f.srv.handleCardPRDraft(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("pr-draft : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	var d PRDraft
	if err := json.Unmarshal(w.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d.Title != "fix: keep tooltips visible" || d.AgentID != task.AgentID {
		t.Fatalf("brouillon inattendu : %+v", d)
	}
}

func TestPrefillQuery(t *testing.T) {
	got := GithubCompareURL("github.com", "acme/demo", "main", "b") + prefillQuery("title", "body", "fix: a & b", "line\nnext")
	if want := "https://github.com/acme/demo/compare/main...b?expand=1&title=fix%3A+a+%26+b&body=line%0Anext"; got != want {
		t.Fatalf("URL = %q, attendu %q", got, want)
	}
}
