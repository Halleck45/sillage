package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

// newTestServer construit un Server minimal (store en mémoire, webFS vide)
// pour appeler les handlers directement, sans passer par le mux ni
// l'authentification (hors scope de ces tests, voir withMiddleware).
func newTestServer(t *testing.T) *Server {
	t.Helper()
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return NewServer(s, "", t.TempDir(), fstest.MapFS{})
}

// TestDeleteHandlersRequireConfirm couvre le refus (400 "confirmation
// required") des trois suppressions destructives (tâche, carte, projet)
// quand le corps ne porte pas {"confirm":true} : corps absent ou
// confirm=false. Vérifie aussi qu'aucune suppression partielle n'a eu lieu.
func TestDeleteHandlersRequireConfirm(t *testing.T) {
	srv := newTestServer(t)
	project, err := srv.store.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := srv.store.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	taskID := mkTaskWithStatus(t, srv.store, card.ID, project.ID, "done")

	cases := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		id      string
		body    string
	}{
		{"task/no body", srv.handleDeleteTask, taskID, ""},
		{"task/confirm false", srv.handleDeleteTask, taskID, `{"confirm":false}`},
		{"card/no body", srv.handleDeleteCard, card.ID, ""},
		{"card/confirm false", srv.handleDeleteCard, card.ID, `{"confirm":false}`},
		{"project/no body", srv.handleDeleteProject, project.ID, ""},
		{"project/confirm false", srv.handleDeleteProject, project.ID, `{"confirm":false}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(c.body))
			req.SetPathValue("id", c.id)
			w := httptest.NewRecorder()
			c.handler(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("attendu 400, reçu %d (body=%s)", w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), "confirmation required") {
				t.Fatalf("message d'erreur attendu 'confirmation required', reçu %q", w.Body.String())
			}
		})
	}

	// Aucun refus ne doit avoir déclenché de suppression partielle.
	if _, ok := srv.store.GetTask(taskID); !ok {
		t.Fatalf("la tâche ne devrait pas avoir été supprimée")
	}
	if _, ok := srv.store.GetCard(card.ID); !ok {
		t.Fatalf("la carte ne devrait pas avoir été supprimée")
	}
	if _, ok := srv.store.GetProject(project.ID); !ok {
		t.Fatalf("le projet ne devrait pas avoir été supprimé")
	}
}

// TestDeleteTaskHandlerConfirmedSucceeds couvre le chemin nominal du handler
// HTTP : confirm=true supprime effectivement la tâche (204 No Content).
func TestDeleteTaskHandlerConfirmedSucceeds(t *testing.T) {
	srv := newTestServer(t)
	project, err := srv.store.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := srv.store.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	taskID := mkTaskWithStatus(t, srv.store, card.ID, project.ID, "done")

	req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(`{"confirm":true}`))
	req.SetPathValue("id", taskID)
	w := httptest.NewRecorder()
	srv.handleDeleteTask(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("attendu 204, reçu %d (body=%s)", w.Code, w.Body.String())
	}
	if _, ok := srv.store.GetTask(taskID); ok {
		t.Fatalf("la tâche devrait avoir été supprimée")
	}
}
