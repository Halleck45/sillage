package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

// newTestServerWithWeb construit un Server dont le frontend statique est
// fourni par files. fstest.MapFS a la même propriété qu'embed.FS pour ce qui
// nous intéresse : une date de modification nulle, donc pas de Last-Modified.
func newTestServerWithWeb(t *testing.T, files fstest.MapFS) *Server {
	t.Helper()
	s, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return NewServer(s, "", t.TempDir(), files)
}

// TestStaticFilesRevalidate couvre la mise en cache du frontend embarqué :
// chaque fichier statique porte un ETag et Cache-Control: no-cache, répond 304
// à un If-None-Match qui correspond, et voit son ETag changer quand le contenu
// change (sinon un rebuild de web/ resterait invisible dans le navigateur).
func TestStaticFilesRevalidate(t *testing.T) {
	files := fstest.MapFS{
		"index.html": {Data: []byte("<!doctype html><title>Sillage</title>")},
		"app.js":     {Data: []byte("(function(){})();")},
	}
	srv := newTestServerWithWeb(t, files)
	handler := srv.Handler()

	for _, path := range []string{"/", "/app.js"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("attendu 200, reçu %d", w.Code)
			}
			etag := w.Header().Get("Etag")
			if etag == "" {
				t.Fatalf("ETag absent : le navigateur n'a aucun validateur")
			}
			if got := w.Header().Get("Cache-Control"); got != "no-cache" {
				t.Fatalf("Cache-Control attendu \"no-cache\", reçu %q", got)
			}

			// Même contenu : le serveur doit répondre 304 sans corps.
			req = httptest.NewRequest(http.MethodGet, path, nil)
			req.Header.Set("If-None-Match", etag)
			w = httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusNotModified {
				t.Fatalf("attendu 304, reçu %d", w.Code)
			}
			if w.Body.Len() != 0 {
				t.Fatalf("un 304 ne doit pas avoir de corps, reçu %d octets", w.Body.Len())
			}
		})
	}

	// Un contenu différent (l'équivalent d'un rebuild de web/) doit produire un
	// ETag différent, sans quoi le navigateur garderait son ancienne copie.
	rebuilt := newTestServerWithWeb(t, fstest.MapFS{
		"index.html": {Data: []byte("<!doctype html><title>Sillage</title>")},
		"app.js":     {Data: []byte("(function(){ /* nouvel onglet */ })();")},
	})

	etagOf := func(h http.Handler) string {
		req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		return w.Header().Get("Etag")
	}
	if before, after := etagOf(handler), etagOf(rebuilt.Handler()); before == after {
		t.Fatalf("l'ETag doit changer avec le contenu, inchangé : %s", before)
	}
}

// TestCreateProjectFromPathOnly couvre la promesse de la modale « Nouveau
// projet » : un seul champ, le chemin d'un dépôt git. Le nom du projet vient du
// dépôt et le mode de livraison des remotes, sans qu'aucune question soit posée
// à froid.
func TestCreateProjectFromPathOnly(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git non disponible dans cet environnement")
	}
	srv := newTestServer(t)
	repo := filepath.Join(t.TempDir(), "mon-projet")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	initTestRepo(t, repo)

	body := `{"repos":[{"path":` + strconv.Quote(repo) + `}]}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.handleCreateProject(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200, reçu %d (body=%s)", w.Code, w.Body.String())
	}
	var out ProjectOut
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if out.Name != "mon-projet" {
		t.Fatalf("nom déduit du dépôt attendu 'mon-projet', reçu %q", out.Name)
	}
	if out.Delivery.Mode == "" {
		t.Fatalf("le mode de livraison devrait être déduit, reçu vide")
	}

	// Sans dépôt, il ne reste rien à déduire : la création doit être refusée.
	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	srv.handleCreateProject(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("attendu 400 sans dépôt, reçu %d (body=%s)", w.Code, w.Body.String())
	}
}

// TestDeleteHandlersRequireConfirm couvre le refus (400 "confirmation
// required") des trois suppressions destructives (tâche, carte, projet)
// quand le corps ne porte pas {"confirm":true} : corps absent ou
// confirm=false. Vérifie aussi qu'aucune suppression partielle n'a eu lieu.
func TestDeleteHandlersRequireConfirm(t *testing.T) {
	srv := newTestServer(t)
	project, err := srv.store.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := srv.store.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	taskID := mkTaskWithStatus(t, srv.store, card.ID, project.ID, "accepted")

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
// TestFixAgentWarningHandler couvre le bouton de l'avertissement Antigravity :
// il règle la politique d'exécution d'agy et renvoie l'agent sans son
// avertissement. Aucun autre cli n'a de correctif automatique (installer une
// CLI ou lancer un sudo n'est pas à Sillage de le faire).
func TestFixAgentWarningHandler(t *testing.T) {
	srv := newTestServer(t)

	origLookPath := lookPath
	defer func() { lookPath = origLookPath }()
	lookPath = func(file string) (string, error) { return "/usr/bin/" + file, nil }

	origSettings := antigravitySettingsPath
	defer func() { antigravitySettingsPath = origSettings }()
	antigravitySettingsPath = filepath.Join(t.TempDir(), "settings.json")

	agents := srv.store.ListAgents()
	var agyID, otherID string
	for _, a := range agents {
		if a.Cli == "agy" && agyID == "" {
			agyID = a.ID
		}
		if a.Cli == "claude" && otherID == "" {
			otherID = a.ID
		}
	}
	if agyID == "" || otherID == "" {
		t.Fatalf("seed should provide an agy agent and a claude agent, got %+v", agents)
	}

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetPathValue("id", agyID)
	w := httptest.NewRecorder()
	srv.handleFixAgentWarning(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("attendu 200, reçu %d (body=%s)", w.Code, w.Body.String())
	}
	var out AgentOut
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if out.Warning != "" {
		t.Fatalf("the returned agent should carry no warning anymore, got %q", out.Warning)
	}
	if !antigravityWorksHeadlessly(srv.store.WorktreesDir()) {
		t.Fatal("the fix should have been written to the settings file")
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetPathValue("id", otherID)
	w = httptest.NewRecorder()
	srv.handleFixAgentWarning(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("attendu 400 pour un autre cli, reçu %d (body=%s)", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetPathValue("id", "nobody")
	w = httptest.NewRecorder()
	srv.handleFixAgentWarning(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("attendu 404 pour un agent inconnu, reçu %d", w.Code)
	}
}

func TestDeleteTaskHandlerConfirmedSucceeds(t *testing.T) {
	srv := newTestServer(t)
	project, err := srv.store.AddProject("p", "", "", []Repo{{Name: "r", Path: "/tmp/x"}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := srv.store.AddCard(project.ID, "Carte", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	taskID := mkTaskWithStatus(t, srv.store, card.ID, project.ID, "accepted")

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
