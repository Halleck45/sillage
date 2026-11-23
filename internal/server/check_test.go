package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setCheckCmd(t *testing.T, f *deliveryFixture, cmd string) {
	t.Helper()
	if _, err := f.srv.store.UpdateProject(f.project.ID, nil, nil, &cmd, nil, nil, nil, nil, nil, nil, nil); err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	t.Cleanup(f.srv.runner.waitChecks)
}

func postReview(t *testing.T, f *deliveryFixture, taskID, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", taskID)
	w := httptest.NewRecorder()
	f.srv.handleReview(w, req)
	return w
}

func TestCheckFailureKeepsOutputAndFingerprint(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	setCheckCmd(t, f, "echo first line; echo 'FAIL TestLogin' >&2; exit 3")
	id, _ := f.addTask(t, "Login", "login.go", "package login\n")

	if !f.srv.runner.StartCheck(id) {
		t.Fatal("la vérification aurait dû démarrer")
	}
	f.srv.runner.waitChecks()

	task, _ := f.srv.store.GetTask(id)
	if task.Checking || task.CheckStartedAt != nil {
		t.Fatal("Checking devrait être retombé")
	}
	if len(task.Checks) != 1 {
		t.Fatalf("un résultat attendu, reçu %+v", task.Checks)
	}
	c := task.Checks[0]
	if c.Ok || c.Reason != "exit 3" || !strings.Contains(c.Output, "FAIL TestLogin") || c.At == nil {
		t.Fatalf("échec mal rendu : %+v", c)
	}
	if c.Sha == "" || c.Sha != task.WorkSha {
		t.Fatalf("le résultat doit porter l'empreinte du travail : check %q, tâche %q", c.Sha, task.WorkSha)
	}
}

func TestCheckSuccessKeepsOutput(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	setCheckCmd(t, f, "echo all good")
	id, _ := f.addTask(t, "Login", "login.go", "package login\n")
	f.srv.runner.StartCheck(id)
	f.srv.runner.waitChecks()
	task, _ := f.srv.store.GetTask(id)
	if len(task.Checks) != 1 || !task.Checks[0].Ok || task.Checks[0].Output != "all good" || task.Checks[0].Reason != "" {
		t.Fatalf("succès mal rendu : %+v", task.Checks)
	}
}

func TestCancelCheckStopsProcessAndKeepsPreviousResult(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	setCheckCmd(t, f, "sleep 30")
	id, _ := f.addTask(t, "Login", "login.go", "package login\n")
	previous := []Check{{Label: "old", Ok: true}}
	if _, err := f.srv.store.PatchTask(id, func(tk *Task) { tk.Checks = previous }); err != nil {
		t.Fatal(err)
	}
	f.srv.runner.StartCheck(id)
	if task, _ := f.srv.store.GetTask(id); !task.Checking {
		t.Fatal("Checking devrait être vrai pendant la vérification")
	}
	start := time.Now()
	f.srv.runner.CancelCheck(id)
	if time.Since(start) > 8*time.Second {
		t.Fatal("l'annulation a attendu la fin du sleep")
	}
	task, _ := f.srv.store.GetTask(id)
	if task.Checking || len(task.Checks) != 1 || task.Checks[0].Label != "old" {
		t.Fatalf("une vérification annulée ne doit rien écrire : %+v", task)
	}
}

func TestCheckTimeoutReason(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	ok, reason, out := runCheckCommand(ctx, t.TempDir(), "echo started; sleep 5", nil)
	if ok || reason != "timeout" || !strings.Contains(out, "started") {
		t.Fatalf("délai dépassé mal rendu : ok=%v reason=%q out=%q", ok, reason, out)
	}
}

func TestReviewVerdictFollowsFingerprint(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "pr", Target: "main"})
	id, cb := f.addTask(t, "Login", "login.go", "package login\n")

	if w := postReview(t, f, id, `{"verdict":"approved"}`); w.Code != http.StatusOK {
		t.Fatalf("review: %d %s", w.Code, w.Body.String())
	}
	task, _ := f.srv.store.GetTask(id)
	if task.Review == nil || task.Review.Verdict != "approved" || task.Review.Sha == "" || task.Review.Sha != task.WorkSha {
		t.Fatalf("verdict mal posé : %+v", task.Review)
	}

	// Un sibling accepté fait avancer le chantier ; le rebase de fond ne
	// change pas le code de la tâche, la relecture reste valable.
	other, _ := f.addTask(t, "Signup", "signup.go", "package signup\n")
	if w := f.accept(t, other); w.Code != http.StatusOK {
		t.Fatalf("accept: %d %s", w.Code, w.Body.String())
	}
	task, _ = f.srv.store.GetTask(id)
	if !IsBranchMergedInto(task.WorktreeDir, cb.Branch, task.Branch) {
		t.Fatal("la tâche aurait dû être rebasée sur le chantier")
	}
	if task.WorkSha != task.Review.Sha {
		t.Fatalf("un rebase sans conflit ne doit pas périmer la relecture : %q != %q", task.WorkSha, task.Review.Sha)
	}

	// Le code change : l'empreinte aussi.
	if err := os.WriteFile(filepath.Join(task.WorktreeDir, "login.go"), []byte("package login\n\nfunc X() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := WorkFingerprint(task.WorktreeDir, task.Base); got == task.Review.Sha {
		t.Fatal("une modification du code doit changer l'empreinte")
	}

	if w := postReview(t, f, id, `{"verdict":""}`); w.Code != http.StatusOK {
		t.Fatalf("clear: %d", w.Code)
	}
	if task, _ = f.srv.store.GetTask(id); task.Review != nil {
		t.Fatal("le verdict aurait dû être effacé")
	}
	if w := postReview(t, f, id, `{"verdict":"approved","step":"test"}`); w.Code != http.StatusOK {
		t.Fatalf("recette: %d", w.Code)
	}
	if task, _ = f.srv.store.GetTask(id); task.Tested == nil || task.Tested.Verdict != "approved" || task.Review != nil {
		t.Fatalf("le verdict de recette doit vivre à part : %+v / %+v", task.Tested, task.Review)
	}
	if w := postReview(t, f, id, `{"verdict":"approved","step":"deploy"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("étape inconnue acceptée : %d", w.Code)
	}
	if w := postReview(t, f, id, `{"verdict":"lgtm"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("verdict inconnu accepté : %d", w.Code)
	}
}
