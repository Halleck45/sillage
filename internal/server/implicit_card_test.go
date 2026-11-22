package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Créer une tâche sur un projet, sans chantier : le chantier est créé au nom
// de la tâche, et le signal de la Vigie qui l'a motivée passe en "handled".
func TestCreateTaskImplicitCard(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr"})
	sig, _, err := f.srv.store.AddSignalIfNew(Signal{ProjectID: f.project.ID, RepoName: "demo", Kind: "issue", Number: 9, Status: "ready"})
	if err != nil {
		t.Fatal(err)
	}
	before := len(f.srv.store.CardsByProject(f.project.ID))

	post := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		f.srv.Handler().ServeHTTP(w, req)
		return w
	}

	w := post(`{"projectId":"` + f.project.ID + `","title":"Fix temp files","agentId":"echo","signalId":"` + sig.ID + `"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("création : %d %s", w.Code, w.Body.String())
	}
	var task Task
	_ = json.Unmarshal(w.Body.Bytes(), &task)
	card, ok := f.srv.store.GetCard(task.CardID)
	if !ok || card.Title != "Fix temp files" || card.ProjectID != f.project.ID {
		t.Fatalf("chantier implicite attendu, reçu %+v", card)
	}
	if got, _ := f.srv.store.GetSignal(sig.ID); got.Status != "handled" || got.CardID != card.ID {
		t.Fatalf("signal non rattaché : %+v", got)
	}

	// Un échec avant la création de la tâche ne laisse pas de chantier orphelin.
	// Dépôt rendu illisible : la création du worktree échoue après celle du chantier.
	if err := os.Rename(filepath.Join(f.repo, ".git"), filepath.Join(f.repo, ".git.off")); err != nil {
		t.Fatal(err)
	}
	w = post(`{"projectId":"` + f.project.ID + `","title":"Nope","agentId":"echo"}`)
	_ = os.Rename(filepath.Join(f.repo, ".git.off"), filepath.Join(f.repo, ".git"))
	if w.Code == http.StatusOK {
		t.Fatal("création acceptée sur un dépôt illisible")
	}
	if n := len(f.srv.store.CardsByProject(f.project.ID)); n != before+1 {
		t.Fatalf("attendu %d chantiers, reçu %d", before+1, n)
	}
}
