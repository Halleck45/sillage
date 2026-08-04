package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// --- Recette manuelle (voir docs/SPEC-RECETTE.md) ---

// setPreviewCmd renseigne la commande (et l'URL) de recette du dépôt du projet,
// par le même chemin que la modale de réglages : PATCH /api/projects/{id}.
func (f *deliveryFixture) setPreviewCmd(t *testing.T, cmd, url string) {
	t.Helper()
	body := fmt.Sprintf(`{"repos":[{"name":"demo","path":%q,"previewCmd":%q,"previewUrl":%q}]}`, f.repo, cmd, url)
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", f.project.ID)
	w := httptest.NewRecorder()
	f.srv.handleUpdateProject(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("PATCH project = %d, corps %s", w.Code, w.Body.String())
	}
	project, ok := f.srv.store.GetProject(f.project.ID)
	if !ok {
		t.Fatal("projet introuvable après PATCH")
	}
	f.project = project
}

// startCardPreview lance la recette du chantier et retourne le run créé.
func (f *deliveryFixture) startCardPreview(t *testing.T) PreviewRun {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.SetPathValue("id", f.card.ID)
	w := httptest.NewRecorder()
	f.srv.handleCardPreview(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST card preview = %d, corps %s", w.Code, w.Body.String())
	}
	var run PreviewRun
	if err := json.Unmarshal(w.Body.Bytes(), &run); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	return run
}

// waitPreviewStatus attend qu'un run quitte l'état "running".
func waitPreviewStatus(t *testing.T, sup *PreviewSupervisor, runID string) PreviewRun {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		for _, run := range sup.List() {
			if run.ID == runID && run.Status != "running" {
				return run
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("le run %s est encore en cours après 10 s", runID)
	return PreviewRun{}
}

func previewLogText(t *testing.T, sup *PreviewSupervisor, runID string) string {
	t.Helper()
	lines, ok := sup.Log(runID)
	if !ok {
		t.Fatalf("journal du run %s introuvable", runID)
	}
	return strings.Join(lines, "\n")
}

// Une commande qui finit : le journal porte sa sortie, le statut est "exited" et
// le code de retour est celui du process.
func TestPreviewRunExitsWithOutput(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, "echo bonjour; exit 3", "")

	run := f.startCardPreview(t)
	if run.Status != "running" && run.Status != "exited" {
		t.Fatalf("statut initial inattendu : %q", run.Status)
	}
	final := waitPreviewStatus(t, f.srv.previews, run.ID)
	if final.Status != "exited" {
		t.Errorf("statut = %q, attendu exited", final.Status)
	}
	if final.ExitCode != 3 {
		t.Errorf("exitCode = %d, attendu 3", final.ExitCode)
	}
	if final.EndedAt == nil {
		t.Error("endedAt devrait être renseigné une fois le run terminé")
	}
	if log := previewLogText(t, f.srv.previews, run.ID); !strings.Contains(log, "bonjour") {
		t.Errorf("journal = %q, devrait contenir la sortie de la commande", log)
	}
}

// stderr et stdout partagent le même tuyau : une erreur doit se voir dans le
// journal, c'est la moitié de l'intérêt de la recette.
func TestPreviewRunCapturesStderr(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, "echo vers-stderr 1>&2", "")

	run := f.startCardPreview(t)
	waitPreviewStatus(t, f.srv.previews, run.ID)
	if log := previewLogText(t, f.srv.previews, run.ID); !strings.Contains(log, "vers-stderr") {
		t.Errorf("journal = %q, devrait contenir la sortie d'erreur", log)
	}
}

// Les cinq variables de recette, et le répertoire d'exécution : le worktree du
// chantier, jamais le dépôt du projet.
func TestPreviewInjectsVariablesAndRunsInWorktree(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	_, cb := f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, `echo "$SILLAGE_ID|$SILLAGE_N|$SILLAGE_PORT|$SILLAGE_DIR|$SILLAGE_BRANCH|$(pwd)"`, "")

	run := f.startCardPreview(t)
	waitPreviewStatus(t, f.srv.previews, run.ID)
	log := strings.TrimSpace(previewLogText(t, f.srv.previews, run.ID))

	want := fmt.Sprintf("ws-%d|%d|%d|%s|%s|%s", f.card.Ref, f.card.Ref, previewPortBase+f.card.Ref, cb.WorktreeDir, cb.Branch, cb.WorktreeDir)
	if log != want {
		t.Errorf("variables = %q, attendu %q", log, want)
	}
	if run.Dir == f.repo {
		t.Error("une recette ne doit jamais tourner dans le dépôt de travail de l'utilisateur")
	}
}

// SILLAGE_PORT reste au-dessus de la plage privilégiée (< 1024) même pour une
// référence minuscule : c'est tout le problème que la variable résout (un
// $((4000 + SILLAGE_N)) oublié dans la commande retombe sur un port qui exige
// les droits root).
func TestPreviewPortStaysAboveBase(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, `echo "$SILLAGE_PORT"`, "")

	run := f.startCardPreview(t)
	waitPreviewStatus(t, f.srv.previews, run.ID)
	log := strings.TrimSpace(previewLogText(t, f.srv.previews, run.ID))

	want := strconv.Itoa(previewPortBase + f.card.Ref)
	if log != want {
		t.Errorf("SILLAGE_PORT = %q, attendu %q", log, want)
	}
}

// L'identité d'une tâche est la sienne (t-<ref>), pas celle du chantier : deux
// recettes du même chantier ne se marchent donc pas dessus.
func TestPreviewTaskUsesItsOwnIdentity(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	taskID, _ := f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, `echo "$SILLAGE_ID|$SILLAGE_N"`, "")

	task, _ := f.srv.store.GetTask(taskID)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.SetPathValue("id", taskID)
	w := httptest.NewRecorder()
	f.srv.handleTaskPreview(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("POST task preview = %d, corps %s", w.Code, w.Body.String())
	}
	var run PreviewRun
	if err := json.Unmarshal(w.Body.Bytes(), &run); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if run.TaskID != taskID {
		t.Errorf("taskId = %q, attendu %q", run.TaskID, taskID)
	}
	waitPreviewStatus(t, f.srv.previews, run.ID)
	log := strings.TrimSpace(previewLogText(t, f.srv.previews, run.ID))
	want := fmt.Sprintf("t-%d|%d", task.Ref, task.Ref)
	if log != want {
		t.Errorf("variables = %q, attendu %q", log, want)
	}
	if run.Dir != task.WorktreeDir {
		t.Errorf("dir = %q, attendu le worktree de la tâche %q", run.Dir, task.WorktreeDir)
	}
}

// L'URL de recette accepte les mêmes variables que la commande, arithmétique
// comprise : c'est le shell qui la développe.
func TestPreviewExpandsURL(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, "true", "http://127.0.0.1:$((4000 + SILLAGE_N))/app?id=$SILLAGE_ID")

	run := f.startCardPreview(t)
	want := fmt.Sprintf("http://127.0.0.1:%d/app?id=ws-%d", 4000+f.card.Ref, f.card.Ref)
	if run.URL != want {
		t.Errorf("url = %q, attendu %q", run.URL, want)
	}
}

// Arrêter un run tue le groupe de process : le serveur de recette ne doit pas
// survivre au bouton Arrêter (sinon son port reste pris).
func TestPreviewStopKillsProcessGroup(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, "sleep 60", "")

	run := f.startCardPreview(t)
	if run.Status != "running" {
		t.Fatalf("statut = %q, attendu running", run.Status)
	}
	pid := previewPID(t, f.srv.previews, run.ID)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.SetPathValue("id", run.ID)
	w := httptest.NewRecorder()
	f.srv.handlePreviewStop(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("POST stop = %d, corps %s", w.Code, w.Body.String())
	}

	final := waitPreviewStatus(t, f.srv.previews, run.ID)
	if final.Status != "stopped" {
		t.Errorf("statut = %q, attendu stopped (arrêt humain, pas sortie naturelle)", final.Status)
	}
	if syscall.Kill(-pid, 0) == nil {
		t.Error("le groupe de process est encore vivant après l'arrêt")
	}
}

// Relancer remplace le run en cours du même worktree : deux serveurs sur le même
// port ne servent à personne.
func TestPreviewRelaunchReplacesRun(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, "sleep 60", "")

	first := f.startCardPreview(t)
	second := f.startCardPreview(t)
	if first.ID == second.ID {
		t.Fatal("le relancement devrait créer un nouveau run")
	}
	runs := f.srv.previews.List()
	if len(runs) != 1 {
		t.Fatalf("%d runs listés, attendu 1 (un seul par worktree)", len(runs))
	}
	if runs[0].ID != second.ID || runs[0].Status != "running" {
		t.Errorf("run listé = %q/%q, attendu %q/running", runs[0].ID, runs[0].Status, second.ID)
	}
	f.srv.previews.StopAll()
}

// StopAll est ce qui garantit qu'un Ctrl+C sur Sillage ne laisse pas de serveur
// de recette derrière lui.
func TestPreviewStopAllKillsEverything(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, "sleep 60", "")

	run := f.startCardPreview(t)
	pid := previewPID(t, f.srv.previews, run.ID)
	f.srv.Shutdown()

	if syscall.Kill(-pid, 0) == nil {
		t.Error("un process de recette a survécu à l'arrêt du serveur")
	}
	for _, r := range f.srv.previews.List() {
		if r.Status == "running" {
			t.Errorf("le run %s est encore annoncé en cours après l'arrêt", r.ID)
		}
	}
}

// Sans commande configurée, il n'y a rien à lancer : l'interface propose alors le
// chemin du worktree, elle n'appelle pas cet endpoint.
func TestPreviewWithoutCommandIsRefused(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.addTask(t, "Recette", "a.txt", "a")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.SetPathValue("id", f.card.ID)
	w := httptest.NewRecorder()
	f.srv.handleCardPreview(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("POST preview sans commande = %d, attendu 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "no preview command") {
		t.Errorf("erreur = %s, devrait dire qu'aucune commande n'est configurée", w.Body.String())
	}
}

// Un chantier sans branche (aucune tâche encore créée) n'a pas de worktree à
// recetter : le message doit dire quoi faire.
func TestPreviewWithoutCardBranchIsRefused(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.setPreviewCmd(t, "true", "")

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	req.SetPathValue("id", f.card.ID)
	w := httptest.NewRecorder()
	f.srv.handleCardPreview(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("POST preview sans branche = %d, attendu 400", w.Code)
	}
	if !strings.Contains(w.Body.String(), "workstream branch") {
		t.Errorf("erreur = %s, devrait parler de la branche de chantier", w.Body.String())
	}
}

// Une URL de recette non http(s) deviendrait un lien cliquable : refusée à
// l'enregistrement, comme les liens épinglés.
func TestPreviewURLMustBeHTTP(t *testing.T) {
	if _, err := NormalizeRepos([]Repo{{Path: "/tmp/demo", PreviewURL: "javascript:alert(1)"}}); err == nil {
		t.Error("une URL de recette non http(s) devrait être refusée")
	}
	repos, err := NormalizeRepos([]Repo{{Path: "/tmp/demo", PreviewCmd: " make serve ", PreviewURL: " http://127.0.0.1:3000 "}})
	if err != nil {
		t.Fatalf("NormalizeRepos: %v", err)
	}
	if repos[0].PreviewCmd != "make serve" || repos[0].PreviewURL != "http://127.0.0.1:3000" {
		t.Errorf("commande/URL non normalisées : %+v", repos[0])
	}
}

// Le journal est un tampon circulaire : un serveur bavard ne fait pas grossir la
// mémoire indéfiniment, et rien n'est écrit dans state.json.
func TestPreviewLogIsCapped(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "merge", Target: "main"})
	f.addTask(t, "Recette", "a.txt", "a")
	f.setPreviewCmd(t, fmt.Sprintf("seq 1 %d", previewLogLines+50), "")

	run := f.startCardPreview(t)
	waitPreviewStatus(t, f.srv.previews, run.ID)
	lines, _ := f.srv.previews.Log(run.ID)
	if len(lines) > previewLogLines {
		t.Errorf("%d lignes gardées, le tampon devrait plafonner à %d", len(lines), previewLogLines)
	}
	if len(lines) > 0 && lines[len(lines)-1] != fmt.Sprint(previewLogLines+50) {
		t.Errorf("dernière ligne = %q, le tampon devrait garder la fin", lines[len(lines)-1])
	}
}

// previewPID retourne le pid du process d'un run en cours.
func previewPID(t *testing.T, sup *PreviewSupervisor, runID string) int {
	t.Helper()
	sup.mu.Lock()
	defer sup.mu.Unlock()
	for _, proc := range sup.runs {
		if proc.run.ID == runID {
			if proc.cmd == nil || proc.cmd.Process == nil {
				t.Fatalf("le run %s n'a pas de process", runID)
			}
			return proc.cmd.Process.Pid
		}
	}
	t.Fatalf("run %s introuvable", runID)
	return 0
}
