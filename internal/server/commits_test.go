package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// L'onglet Commits montre sujet, corps et auteur de chaque commit, y compris
// après acceptation, et compte ce qui reste hors commit.
func TestTaskCommitsListsMessagesAndUncommitted(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	id, _ := f.addTask(t, "Commits", "a.txt", "a\n")
	task, _ := f.srv.store.GetTask(id)
	runTestGit(t, task.WorktreeDir, "add", "-A")
	runTestGit(t, task.WorktreeDir, "commit", "-q", "-m", "feat: add a | pipe", "-m", "Body line one.\n\nBody line two.")
	if err := os.WriteFile(filepath.Join(task.WorktreeDir, "b.txt"), []byte("b\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	get := func() CommitsResponse {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.SetPathValue("id", id)
		w := httptest.NewRecorder()
		f.srv.handleTaskCommits(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("attendu 200, reçu %d (%s)", w.Code, w.Body.String())
		}
		var res CommitsResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatal(err)
		}
		return res
	}

	res := get()
	if len(res.Commits) != 1 || res.Uncommitted != 1 {
		t.Fatalf("1 commit et 1 fichier hors commit attendus, reçu %+v", res)
	}
	c := res.Commits[0]
	if c.Subject != "feat: add a | pipe" || c.Body != "Body line one.\n\nBody line two." || c.Author != "Test" || c.Hash == "" {
		t.Fatalf("commit mal lu : %+v", c)
	}

	if w := f.accept(t, id); w.Code != http.StatusOK {
		t.Fatalf("accept : attendu 200, reçu %d (%s)", w.Code, w.Body.String())
	}
	res = get()
	if len(res.Commits) != 2 || res.Commits[0].Subject != fallbackCommitMessage || res.Uncommitted != 0 {
		t.Fatalf("après acceptation : le commit de repli puis celui de l'agent attendus, reçu %+v", res)
	}
}
