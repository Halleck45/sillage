package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExtractAutoconfigSanitizes(t *testing.T) {
	reply := "Example: {\"check\":\"...\"}\n```json\n" +
		`{"check":"npm ci &&\nnpm test","previewCmd":"","previewUrl":"javascript:alert(1)","profiles":["node","cobol"],"notes":" ok "}` +
		"\n```"
	a, ok := extractAutoconfig(reply)
	if !ok {
		t.Fatal("configuration non trouvée")
	}
	if a.Check != "npm ci && npm test" {
		t.Fatalf("commande sur une ligne attendue, reçu %q", a.Check)
	}
	if a.PreviewURL != "" {
		t.Fatalf("une URL non http(s) ou sans commande doit être écartée, reçu %q", a.PreviewURL)
	}
	if len(a.Profiles) != 1 || a.Profiles[0] != "node" {
		t.Fatalf("profils inconnus à écarter, reçu %v", a.Profiles)
	}
}

func TestAutoconfigProposesWithoutSaving(t *testing.T) {
	srv := newTestServer(t)
	p, err := srv.store.AddProject("widget", "", "", []Repo{{Name: "widget", Path: t.TempDir()}}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+p.ID+"/autoconfigure", strings.NewReader(`{"agentId":"echo"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Fatalf("attendu 202, reçu %d : %s", w.Code, w.Body.String())
	}
	deadline := time.Now().Add(5 * time.Second)
	var run AutoconfigRun
	for time.Now().Before(deadline) {
		runs := srv.autoconf.list()
		if len(runs) == 1 && runs[0].Status != "running" {
			run = runs[0]
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if run.Status != "ready" || run.Proposal == nil {
		t.Fatalf("proposition attendue, reçu %+v", run)
	}
	if run.Proposal.CheckCmd == "" || len(run.Proposal.Repos) != 1 || run.Proposal.Repos[0].PreviewCmd == "" {
		t.Fatalf("proposition incomplète : %+v", run.Proposal)
	}
	if len(run.Steps) == 0 {
		t.Fatal("les étapes de l'agent doivent être remontées")
	}
	got, _ := srv.store.GetProject(p.ID)
	if got.CheckCmd != "" || got.Repos[0].PreviewCmd != "" {
		t.Fatal("rien ne doit être enregistré sans l'humain")
	}
}
