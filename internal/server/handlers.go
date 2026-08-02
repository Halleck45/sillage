package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// Server assemble l'état applicatif et expose le handler HTTP racine.
type Server struct {
	store        *Store
	hub          *Hub
	runner       *Runner
	sessions     *SessionManager
	loginLimiter *LoginLimiter
	passwordHash string
	dataDir      string
	static       http.Handler
}

// NewServer construit le serveur Sillage. webFS doit pointer vers le sous-répertoire
// "web" du système de fichiers embarqué (frontend statique).
func NewServer(store *Store, passwordHash, dataDir string, webFS fs.FS) *Server {
	hub := NewHub()
	return &Server{
		store:        store,
		hub:          hub,
		runner:       NewRunner(store, hub),
		sessions:     NewSessionManager(),
		loginLimiter: NewLoginLimiter(),
		passwordHash: passwordHash,
		dataDir:      dataDir,
		static:       http.FileServer(http.FS(webFS)),
	}
}

// Handler construit le mux HTTP complet avec ses middlewares.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.handleLogout)
	mux.HandleFunc("GET /api/state", s.handleState)
	mux.HandleFunc("POST /api/projects", s.handleCreateProject)
	mux.HandleFunc("POST /api/cards", s.handleCreateCard)
	mux.HandleFunc("PATCH /api/cards/{id}", s.handleUpdateCard)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTask)
	mux.HandleFunc("POST /api/tasks/{id}/messages", s.handlePostMessage)
	mux.HandleFunc("POST /api/tasks/{id}/interrupt", s.handleInterrupt)
	mux.HandleFunc("POST /api/tasks/{id}/accept", s.handleAccept)
	mux.HandleFunc("POST /api/tasks/{id}/ship", s.handleShip)
	mux.HandleFunc("POST /api/tasks/{id}/reopen", s.handleReopen)
	mux.HandleFunc("POST /api/tasks/{id}/read", s.handleRead)
	mux.HandleFunc("GET /api/tasks/{id}/diff", s.handleDiff)
	mux.HandleFunc("GET /api/tasks/{id}/deliverables", s.handleDeliverables)
	mux.HandleFunc("GET /api/events", s.hub.ServeSSE)
	mux.Handle("/", s.static)

	return s.withMiddleware(mux)
}

// withMiddleware applique l'authentification et la protection CSRF sur /api/*.
func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		if isMutatingMethod(r.Method) {
			ct := r.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, "application/json") {
				writeError(w, http.StatusForbidden, "Content-Type application/json requis")
				return
			}
		}

		if r.URL.Path != "/api/login" {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil || !s.sessions.Validate(cookie.Value) {
				writeError(w, http.StatusUnauthorized, "authentification requise")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// --- Helpers JSON ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(v)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// --- Auth ---

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if s.loginLimiter.Blocked(ip) {
		writeError(w, http.StatusTooManyRequests, "trop de tentatives, réessayez dans une minute")
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "corps de requête invalide")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(s.passwordHash), []byte(body.Password)) != nil {
		s.loginLimiter.RecordFailure(ip)
		writeError(w, http.StatusUnauthorized, "mot de passe incorrect")
		return
	}

	token, err := s.sessions.Create()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "création de session impossible")
		return
	}
	setSessionCookie(w, r, token)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		s.sessions.Delete(cookie.Value)
	}
	clearSessionCookie(w, r)
	w.WriteHeader(http.StatusNoContent)
}

// --- État global ---

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Snapshot())
}

// --- Projets ---

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if err := decodeJSON(r, &body); err != nil || body.Name == "" || body.Path == "" {
		writeError(w, http.StatusBadRequest, "nom et chemin requis")
		return
	}
	info, err := os.Stat(body.Path)
	if err != nil || !info.IsDir() {
		writeError(w, http.StatusBadRequest, "chemin invalide : répertoire introuvable")
		return
	}
	if !IsGitRepo(body.Path) {
		writeError(w, http.StatusBadRequest, "chemin invalide : ce n'est pas un dépôt git")
		return
	}
	project, err := s.store.AddProject(body.Name, body.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "création du projet impossible")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// --- Cartes ---

func (s *Server) handleCreateCard(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProjectID string `json:"projectId"`
		Title     string `json:"title"`
		Column    string `json:"column"`
	}
	if err := decodeJSON(r, &body); err != nil || body.ProjectID == "" || body.Title == "" {
		writeError(w, http.StatusBadRequest, "projectId et title requis")
		return
	}
	card, err := s.store.AddCard(body.ProjectID, body.Title, body.Column)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.runner.publishCards(card.ProjectID)
	writeJSON(w, http.StatusOK, card)
}

func (s *Server) handleUpdateCard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Column string `json:"column"`
	}
	if err := decodeJSON(r, &body); err != nil || body.Column == "" {
		writeError(w, http.StatusBadRequest, "column requis")
		return
	}
	card, err := s.store.UpdateCardColumn(id, body.Column)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.runner.publishCards(card.ProjectID)
	writeJSON(w, http.StatusOK, card)
}

// --- Tâches ---

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CardID  string `json:"cardId"`
		Title   string `json:"title"`
		AgentID string `json:"agentId"`
		Prompt  string `json:"prompt"`
	}
	if err := decodeJSON(r, &body); err != nil || body.CardID == "" || body.Title == "" || body.AgentID == "" {
		writeError(w, http.StatusBadRequest, "cardId, title et agentId requis")
		return
	}
	card, ok := s.store.GetCard(body.CardID)
	if !ok {
		writeError(w, http.StatusBadRequest, "carte introuvable")
		return
	}
	if _, ok := s.store.GetAgent(body.AgentID); !ok {
		writeError(w, http.StatusBadRequest, "agent introuvable")
		return
	}
	project, ok := s.store.GetProject(card.ProjectID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "projet introuvable")
		return
	}

	id, ref := s.store.ReserveTaskID()
	branch := fmt.Sprintf("sillage/%d-%s", ref, Slugify(body.Title))
	dir, base, err := CreateWorktree(project.Path, s.dataDir, id, branch)
	if err != nil {
		writeError(w, http.StatusBadRequest, "création du worktree impossible : "+err.Error())
		return
	}

	task, err := s.store.CreateTask(id, ref, card.ID, project.ID, body.Title, body.AgentID, branch, base, dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "création de la tâche impossible")
		return
	}
	s.runner.publishTask(task)
	s.runner.publishCards(project.ID)

	cliInput := body.Title
	if body.Prompt != "" {
		cliInput = body.Title + "\n\n" + body.Prompt
	}
	if err := s.runner.Start(task.ID, true, cliInput); err != nil {
		writeError(w, http.StatusInternalServerError, "lancement de l'agent impossible : "+err.Error())
		return
	}
	task, _ = s.store.GetTask(task.ID)
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "tâche introuvable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"task":     task,
		"messages": s.store.GetMessages(id),
	})
}

func (s *Server) handlePostMessage(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := s.store.GetTask(id); !ok {
		writeError(w, http.StatusNotFound, "tâche introuvable")
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	if err := decodeJSON(r, &body); err != nil || strings.TrimSpace(body.Text) == "" {
		writeError(w, http.StatusBadRequest, "text requis")
		return
	}
	if err := s.runner.Start(id, false, body.Text); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleInterrupt(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := s.store.GetTask(id); !ok {
		writeError(w, http.StatusNotFound, "tâche introuvable")
		return
	}
	task, err := s.runner.Interrupt(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleAccept(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "tâche introuvable")
		return
	}
	if task.Status != "review" {
		writeError(w, http.StatusBadRequest, "la tâche n'est pas en revue")
		return
	}
	task, err := s.store.UpdateTask(id, func(t *Task) { t.Status = "ready" })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.runner.publishTask(task)
	s.runner.publishCards(task.ProjectID)
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleShip(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeJSON(r, &body); err != nil || !body.Confirm {
		writeError(w, http.StatusBadRequest, "confirmation requise")
		return
	}
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "tâche introuvable")
		return
	}
	if task.Status != "ready" {
		writeError(w, http.StatusBadRequest, "la tâche doit être prête avant l'envoi")
		return
	}

	output, err := Ship(task.WorktreeDir, task.Branch, task.Title)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, ShipResponse{Task: task, Output: output + "\n" + err.Error()})
		return
	}

	task, err = s.store.UpdateTask(id, func(t *Task) { t.Status = "shipped" })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.runner.publishTask(task)
	s.runner.publishCards(task.ProjectID)
	s.runner.publishAgents()
	writeJSON(w, http.StatusOK, ShipResponse{Task: task, Output: output})
}

func (s *Server) handleReopen(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "tâche introuvable")
		return
	}
	if task.Status != "shipped" {
		writeError(w, http.StatusBadRequest, "seule une tâche livrée peut être rouverte")
		return
	}
	task, err := s.store.UpdateTask(id, func(t *Task) { t.Status = "review" })
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.runner.publishTask(task)
	s.runner.publishCards(task.ProjectID)
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := s.store.UpdateTask(id, func(t *Task) { t.Unread = false })
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	s.runner.publishTask(task)
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "tâche introuvable")
		return
	}
	files, err := Diff(task.WorktreeDir, task.Base)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "calcul du diff impossible : "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, DiffResponse{Branch: task.Branch, Base: task.Base, Files: files})
}

func (s *Server) handleDeliverables(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "tâche introuvable")
		return
	}

	resp := DeliverablesResponse{Code: []Item{}, Docs: []Item{}, Images: []Item{}}
	resp.Code = append(resp.Code, Item{Kind: "branch", Title: task.Branch, Meta: "base : " + task.Base})
	if commits, err := Commits(task.WorktreeDir, task.Base); err == nil {
		for _, c := range commits {
			resp.Code = append(resp.Code, Item{Kind: "commit", Title: c.Subject, Meta: c.Hash + " · " + c.RelTime})
		}
	}
	if files, err := Diff(task.WorktreeDir, task.Base); err == nil {
		for _, f := range files {
			meta := fmt.Sprintf("+%d −%d", f.Additions, f.Deletions)
			switch {
			case isDocFile(f.Path):
				resp.Docs = append(resp.Docs, Item{Kind: "doc", Title: f.Path, Meta: meta, Path: f.Path})
			case isImageFile(f.Path):
				resp.Images = append(resp.Images, Item{Kind: "image", Title: f.Path, Meta: meta, Path: f.Path})
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
