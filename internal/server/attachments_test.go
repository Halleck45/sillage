package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// pngPixel est la plus petite image valide qu'on puisse joindre.
const pngPixel = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg=="

// newAttachmentFixture monte un serveur avec une tâche prête à recevoir des
// messages. Aucun dépôt git : rien ici n'a besoin d'un worktree.
func newAttachmentFixture(t *testing.T) (*Server, string, string) {
	t.Helper()
	dataDir := t.TempDir()
	store, err := NewStore(dataDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	srv := NewServer(store, "", dataDir, fstest.MapFS{})
	project, err := store.AddProject("demo", "", "", []Repo{{Name: "demo", Path: t.TempDir()}}, nil, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	card, err := store.AddCard(project.ID, "Chantier", "", "")
	if err != nil {
		t.Fatalf("AddCard: %v", err)
	}
	taskID, ref := store.ReserveTaskID()
	// Écho (cli "fake") : le message part par le chemin normal, sans process
	// externe. Les tests coupent court à la simulation avec Interrupt.
	if _, err := store.CreateTask(taskID, ref, card.ID, project.ID, "Tâche", "echo", "sillage/100-tache", "main", filepath.Join(dataDir, "worktrees", taskID), "demo"); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if _, err := store.UpdateTask(taskID, func(tk *Task) { tk.Status = "review" }); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}
	return srv, dataDir, taskID
}

func postMessage(t *testing.T, srv *Server, taskID string, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/tasks/"+taskID+"/messages", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	// La simulation d'Écho dure quelques secondes : rien ici n'attend son
	// travail, seulement le message déjà ajouté au fil.
	_, _ = srv.runner.Interrupt(taskID)
	return rec
}

// Une image jointe est écrite sur le disque, référencée par le message, et
// jamais stockée en base64 dans l'état.
func TestPostMessageWithImageAttachment(t *testing.T) {
	srv, dataDir, taskID := newAttachmentFixture(t)
	t.Cleanup(func() { srv.runner.waitTasks() })

	rec := postMessage(t, srv, taskID, `{"text":"regarde ça","attachments":[{"name":"capture.png","mime":"image/png","data":"`+pngPixel+`"}]}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("statut attendu 202, reçu %d (%s)", rec.Code, rec.Body.String())
	}

	msgs := srv.store.GetMessages(taskID)
	if len(msgs) == 0 {
		t.Fatal("aucun message ajouté")
	}
	msg := msgs[0]
	if len(msg.Attachments) != 1 {
		t.Fatalf("une pièce jointe attendue, reçu %d", len(msg.Attachments))
	}
	att := msg.Attachments[0]
	if att.Name != "capture.png" || att.Mime != "image/png" || att.ID == "" {
		t.Fatalf("descripteur inattendu : %+v", att)
	}
	if !strings.HasSuffix(att.Path, ".png") || !strings.HasPrefix(att.Path, filepath.Join(dataDir, "attachments", taskID)) {
		t.Fatalf("chemin inattendu : %q", att.Path)
	}
	raw, err := os.ReadFile(att.Path)
	if err != nil {
		t.Fatalf("l'image devrait être sur le disque : %v", err)
	}
	want, _ := base64.StdEncoding.DecodeString(pngPixel)
	if string(raw) != string(want) {
		t.Fatal("le contenu écrit ne correspond pas à l'image envoyée")
	}
	if strings.Contains(string(mustReadState(t, dataDir)), pngPixel[:32]) {
		t.Fatal("state.json ne doit jamais porter le contenu de l'image")
	}

	// L'image se relit par l'API, avec son type.
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/"+taskID+"/attachments/"+att.ID, nil)
	got := httptest.NewRecorder()
	srv.Handler().ServeHTTP(got, req)
	if got.Code != http.StatusOK {
		t.Fatalf("statut attendu 200, reçu %d", got.Code)
	}
	if ct := got.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("content-type attendu image/png, reçu %q", ct)
	}
	if got.Body.String() != string(want) {
		t.Fatal("l'octet à octet servi ne correspond pas à l'image")
	}

	// Une tâche supprimée emporte ses images.
	if err := srv.runner.DeleteTask(taskID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dataDir, "attachments", taskID)); !os.IsNotExist(err) {
		t.Fatalf("le répertoire des pièces jointes devrait avoir disparu (%v)", err)
	}
}

func mustReadState(t *testing.T, dataDir string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dataDir, "state.json"))
	if err != nil {
		t.Fatalf("lecture de state.json : %v", err)
	}
	return raw
}

// Une image seule, sans texte, est une instruction valable.
func TestPostMessageImageWithoutText(t *testing.T) {
	srv, _, taskID := newAttachmentFixture(t)
	t.Cleanup(func() { srv.runner.waitTasks() })

	rec := postMessage(t, srv, taskID, `{"text":"","attachments":[{"name":"x.png","mime":"image/png","data":"`+pngPixel+`"}]}`)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("statut attendu 202, reçu %d (%s)", rec.Code, rec.Body.String())
	}
	if msgs := srv.store.GetMessages(taskID); len(msgs) == 0 || len(msgs[0].Attachments) != 1 {
		t.Fatalf("message avec pièce jointe attendu, reçu %+v", msgs)
	}
}

// Un message sans texte ni image reste refusé.
func TestPostMessageEmptyRejected(t *testing.T) {
	srv, _, taskID := newAttachmentFixture(t)
	rec := postMessage(t, srv, taskID, `{"text":"   "}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("statut attendu 400, reçu %d", rec.Code)
	}
}

// Formats refusés, poids refusé, nombre refusé : rien n'est écrit.
func TestSaveAttachmentsRejects(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name   string
		inputs []attachmentInput
	}{
		{"type non supporté", []attachmentInput{{Name: "doc.pdf", Mime: "application/pdf", Data: pngPixel}}},
		{"base64 invalide", []attachmentInput{{Name: "x.png", Mime: "image/png", Data: "pas du base64 !"}}},
		{"image vide", []attachmentInput{{Name: "x.png", Mime: "image/png", Data: ""}}},
		{"image trop lourde", []attachmentInput{{Name: "x.png", Mime: "image/png", Data: base64.StdEncoding.EncodeToString(make([]byte, maxAttachmentBytes+1))}}},
		{"trop d'images", make([]attachmentInput, maxAttachmentsPerMessage+1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := saveAttachments(dir, "t1", tc.inputs); err == nil {
				t.Fatal("erreur attendue")
			}
			entries, _ := os.ReadDir(filepath.Join(dir, "attachments", "t1"))
			if len(entries) != 0 {
				t.Fatalf("aucun fichier ne devrait rester, reçu %d", len(entries))
			}
		})
	}
}

// Une pièce jointe refusée sur la deuxième image ne laisse pas la première.
func TestSaveAttachmentsRollsBack(t *testing.T) {
	dir := t.TempDir()
	_, err := saveAttachments(dir, "t1", []attachmentInput{
		{Name: "ok.png", Mime: "image/png", Data: pngPixel},
		{Name: "ko.pdf", Mime: "application/pdf", Data: pngPixel},
	})
	if err == nil {
		t.Fatal("erreur attendue")
	}
	entries, _ := os.ReadDir(filepath.Join(dir, "attachments", "t1"))
	if len(entries) != 0 {
		t.Fatalf("la première image aurait dû être retirée, reçu %d fichier(s)", len(entries))
	}
}

// Le nom d'origine n'est qu'une étiquette : il ne sert jamais de nom de fichier.
func TestAttachmentNameIsSanitized(t *testing.T) {
	dir := t.TempDir()
	atts, err := saveAttachments(dir, "t1", []attachmentInput{{Name: "../../evil.png", Mime: "image/png", Data: pngPixel}})
	if err != nil {
		t.Fatalf("saveAttachments: %v", err)
	}
	if atts[0].Name != "evil.png" {
		t.Fatalf("nom attendu 'evil.png', reçu %q", atts[0].Name)
	}
	if filepath.Dir(atts[0].Path) != filepath.Join(dir, "attachments", "t1") {
		t.Fatalf("le fichier est sorti de son répertoire : %q", atts[0].Path)
	}
}

// Ce que l'agent reçoit : le texte, puis les chemins, une seule fois.
func TestWithAttachmentPaths(t *testing.T) {
	atts := []Attachment{{Path: "/data/attachments/t1/a.png"}, {Path: "/data/attachments/t1/b.png"}}
	got := withAttachmentPaths("relis ça", atts)
	want := "relis ça\n\nAttached images (local files, read them):\n- /data/attachments/t1/a.png\n- /data/attachments/t1/b.png"
	if got != want {
		t.Fatalf("texte inattendu :\n%q", got)
	}
	if got := withAttachmentPaths("", atts[:1]); got != "Attached images (local files, read them):\n- /data/attachments/t1/a.png" {
		t.Fatalf("sans texte, seuls les chemins : %q", got)
	}
	if got := withAttachmentPaths("rien", nil); got != "rien" {
		t.Fatalf("sans image, le texte est intact : %q", got)
	}
}

// Une image jointe dès la création de la tâche : elle est écrite sous la
// nouvelle tâche et posée dans le fil, puisque le prompt initial, lui, n'est
// pas un message.
func TestCreateTaskWithImageAttachment(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "push"})
	body := `{"cardId":"` + f.card.ID + `","title":"Refaire l'en-tête","agentId":"echo","prompt":"comme sur l'image",` +
		`"attachments":[{"name":"maquette.png","mime":"image/png","data":"` + pngPixel + `"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	f.srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("statut attendu 200, reçu %d (%s)", rec.Code, rec.Body.String())
	}
	var task Task
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	_, _ = f.srv.runner.Interrupt(task.ID)

	msgs := f.srv.store.GetMessages(task.ID)
	if len(msgs) == 0 || len(msgs[0].Attachments) != 1 {
		t.Fatalf("le fil devrait porter le message des images, reçu %+v", msgs)
	}
	att := msgs[0].Attachments[0]
	if att.Name != "maquette.png" {
		t.Fatalf("nom inattendu : %q", att.Name)
	}
	if _, err := os.Stat(att.Path); err != nil {
		t.Fatalf("l'image devrait être sur le disque : %v", err)
	}
	if !strings.HasPrefix(att.Path, filepath.Join(f.dataDir, "attachments", task.ID)) {
		t.Fatalf("l'image devrait vivre sous sa tâche, reçu %q", att.Path)
	}
}

// Une tâche à démarrage différé garde les chemins de ses images avec son
// prompt, pour les remettre à l'agent le jour où elle démarre.
func TestCreateWaitingTaskKeepsPendingAttachments(t *testing.T) {
	f := newDeliveryFixture(t, Delivery{Mode: "push"})
	first, _ := f.addTask(t, "Première", "a.txt", "a")

	body := `{"cardId":"` + f.card.ID + `","title":"Ensuite","agentId":"echo","prompt":"suite","waitsForTaskId":"` + first + `",` +
		`"attachments":[{"name":"note.png","mime":"image/png","data":"` + pngPixel + `"}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	f.srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("statut attendu 200, reçu %d (%s)", rec.Code, rec.Body.String())
	}
	var task Task
	if err := json.Unmarshal(rec.Body.Bytes(), &task); err != nil {
		t.Fatalf("réponse illisible : %v", err)
	}
	if task.Status != "waiting" {
		t.Fatalf("statut attendu waiting, reçu %q", task.Status)
	}
	if len(task.PendingAttachments) != 1 || task.PendingPrompt != "suite" {
		t.Fatalf("prompt et images auraient dû attendre ensemble : %+v", task)
	}
	if msgs := f.srv.store.GetMessages(task.ID); len(msgs) != 1 || len(msgs[0].Attachments) != 1 {
		t.Fatalf("le fil devrait déjà montrer l'image, reçu %+v", msgs)
	}
}

// Un chemin qui sort du répertoire des pièces jointes n'est jamais servi.
func TestGetAttachmentRefusesEscapedPath(t *testing.T) {
	srv, dataDir, taskID := newAttachmentFixture(t)
	secret := filepath.Join(dataDir, "config.json")
	if _, _, err := srv.store.AddMessage(taskID, "user", "", "", Attachment{ID: "abc", Name: "x.png", Mime: "image/png", Path: secret}); err != nil {
		t.Fatalf("AddMessage: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/"+taskID+"/attachments/abc", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("statut attendu 404, reçu %d", rec.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] == "" {
		t.Fatalf("réponse d'erreur attendue, reçu %q", rec.Body.String())
	}
}
