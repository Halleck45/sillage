package server

import (
	"context"
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
	dataDir      string
	static       http.Handler
}

// NewServer construit le serveur Sillage. webFS doit pointer vers le sous-répertoire
// "web" du système de fichiers embarqué (frontend statique).
func NewServer(store *Store, dataDir string, webFS fs.FS) *Server {
	hub := NewHub()
	return &Server{
		store:        store,
		hub:          hub,
		runner:       NewRunner(store, hub),
		sessions:     NewSessionManager(),
		loginLimiter: NewLoginLimiter(),
		dataDir:      dataDir,
		static:       http.FileServer(http.FS(webFS)),
	}
}

// Handler construit le mux HTTP complet avec ses middlewares.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.handleLogout)
	mux.HandleFunc("GET /api/me", s.handleMe)
	mux.HandleFunc("GET /api/state", s.handleState)
	mux.HandleFunc("GET /api/users", s.handleListUsers)
	mux.HandleFunc("POST /api/users", s.handleCreateUser)
	mux.HandleFunc("PATCH /api/users/{id}", s.handleUpdateUser)
	mux.HandleFunc("DELETE /api/users/{id}", s.handleDeleteUser)
	mux.HandleFunc("POST /api/agents", s.handleCreateAgent)
	mux.HandleFunc("PATCH /api/agents/{id}", s.handleUpdateAgent)
	mux.HandleFunc("DELETE /api/agents/{id}", s.handleDeleteAgent)
	mux.HandleFunc("POST /api/projects", s.handleCreateProject)
	mux.HandleFunc("PATCH /api/projects/{id}", s.handleUpdateProject)
	mux.HandleFunc("POST /api/cards", s.handleCreateCard)
	mux.HandleFunc("PATCH /api/cards/{id}", s.handleUpdateCard)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTask)
	mux.HandleFunc("POST /api/tasks/{id}/messages", s.handlePostMessage)
	mux.HandleFunc("POST /api/tasks/{id}/interrupt", s.handleInterrupt)
	mux.HandleFunc("POST /api/tasks/{id}/accept", s.handleAccept)
	mux.HandleFunc("POST /api/tasks/{id}/ship", s.handleShip)
	mux.HandleFunc("POST /api/tasks/{id}/pr", s.handlePR)
	mux.HandleFunc("POST /api/tasks/{id}/reopen", s.handleReopen)
	mux.HandleFunc("POST /api/tasks/{id}/read", s.handleRead)
	mux.HandleFunc("GET /api/tasks/{id}/diff", s.handleDiff)
	mux.HandleFunc("GET /api/tasks/{id}/deliverables", s.handleDeliverables)
	mux.HandleFunc("GET /api/events", s.hub.ServeSSE)
	mux.Handle("/", s.static)

	return s.withMiddleware(mux)
}

// --- Contexte utilisateur ---

type ctxKey int

const userCtxKey ctxKey = iota

// userFromContext retourne l'utilisateur authentifié attaché à la requête
// par withMiddleware.
func userFromContext(r *http.Request) (User, bool) {
	u, ok := r.Context().Value(userCtxKey).(User)
	return u, ok
}

// requireAdmin exige que l'utilisateur courant ait le rôle admin ; écrit une
// réponse 403 et retourne false sinon.
func requireAdmin(w http.ResponseWriter, r *http.Request) (User, bool) {
	user, ok := userFromContext(r)
	if !ok || user.Role != "admin" {
		writeError(w, http.StatusForbidden, "admin privileges required")
		return User{}, false
	}
	return user, true
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
				writeError(w, http.StatusForbidden, "Content-Type application/json required")
				return
			}
		}

		if r.URL.Path != "/api/login" {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			userID, ok := s.sessions.Validate(cookie.Value)
			if !ok {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			user, ok := s.store.GetUser(userID)
			if !ok {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), userCtxKey, user))
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
		writeError(w, http.StatusTooManyRequests, "too many attempts, try again in a minute")
		return
	}

	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &body); err != nil || body.Username == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	user, ok := s.store.FindUserByName(body.Username)
	if !ok || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)) != nil {
		s.loginLimiter.RecordFailure(ip)
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	token, err := s.sessions.Create(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session creation failed")
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

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, user.Public())
}

// --- Utilisateurs (admin uniquement) ---

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	users := s.store.ListUsers()
	out := make([]UserPublic, len(users))
	for i, u := range users {
		out[i] = u.Public()
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var body struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := decodeJSON(r, &body); err != nil || body.Name == "" || body.Password == "" {
		writeError(w, http.StatusBadRequest, "name and password are required")
		return
	}
	user, err := s.store.AddUser(body.Name, body.Password, body.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user.Public())
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	id := r.PathValue("id")
	var body struct {
		Password *string `json:"password"`
		Role     *string `json:"role"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := s.store.UpdateUser(id, body.Password, body.Role)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user.Public())
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	me, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	if id == me.ID {
		writeError(w, http.StatusBadRequest, "cannot delete your own account")
		return
	}
	if err := s.store.DeleteUser(id); err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	s.sessions.DeleteByUser(id)
	w.WriteHeader(http.StatusNoContent)
}

// statusForStoreError distingue "not found" (404) des autres erreurs de
// validation (400) renvoyées par les méthodes du Store.
func statusForStoreError(err error) int {
	if strings.Contains(err.Error(), "not found") {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

// --- État global ---

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	state := s.store.Snapshot()
	if user, ok := userFromContext(r); ok {
		state.Me = user.Public()
	}
	writeJSON(w, http.StatusOK, state)
}

// --- Agents ---

func (s *Server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name          string `json:"name"`
		Emoji         string `json:"emoji"`
		Color         string `json:"color"`
		Cli           string `json:"cli"`
		Model         string `json:"model"`
		ContextPrompt string `json:"contextPrompt"`
	}
	if err := decodeJSON(r, &body); err != nil || body.Name == "" || body.Cli == "" {
		writeError(w, http.StatusBadRequest, "name and cli are required")
		return
	}
	agent, err := s.store.AddAgent(body.Name, body.Emoji, body.Color, body.Cli, body.Model, body.ContextPrompt)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.runner.publishAgents()
	writeJSON(w, http.StatusOK, agent)
}

func (s *Server) handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := s.store.GetAgent(id); !ok {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	var body struct {
		Name          *string `json:"name"`
		Emoji         *string `json:"emoji"`
		Color         *string `json:"color"`
		Cli           *string `json:"cli"`
		Model         *string `json:"model"`
		ContextPrompt *string `json:"contextPrompt"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	agent, err := s.store.UpdateAgent(id, body.Name, body.Emoji, body.Color, body.Cli, body.Model, body.ContextPrompt)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.runner.publishAgents()
	writeJSON(w, http.StatusOK, agent)
}

func (s *Server) handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := s.store.GetAgent(id); !ok {
		writeError(w, http.StatusNotFound, "agent not found")
		return
	}
	if err := s.store.DeleteAgent(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.runner.publishAgents()
	w.WriteHeader(http.StatusNoContent)
}

// --- Projets ---

func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if err := decodeJSON(r, &body); err != nil || body.Name == "" || body.Path == "" {
		writeError(w, http.StatusBadRequest, "name and path are required")
		return
	}
	info, err := os.Stat(body.Path)
	if err != nil || !info.IsDir() {
		writeError(w, http.StatusBadRequest, "invalid project path: directory not found")
		return
	}
	if !IsGitRepo(body.Path) {
		writeError(w, http.StatusBadRequest, "invalid project path: not a git repository")
		return
	}
	project, err := s.store.AddProject(body.Name, body.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create project")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Name     *string `json:"name"`
		CheckCmd *string `json:"checkCmd"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	project, err := s.store.UpdateProject(id, body.Name, body.CheckCmd)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	s.runner.publishProject(project)
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
		writeError(w, http.StatusBadRequest, "projectId and title are required")
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
		writeError(w, http.StatusBadRequest, "column is required")
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
		writeError(w, http.StatusBadRequest, "cardId, title and agentId are required")
		return
	}
	card, ok := s.store.GetCard(body.CardID)
	if !ok {
		writeError(w, http.StatusBadRequest, "card not found")
		return
	}
	if _, ok := s.store.GetAgent(body.AgentID); !ok {
		writeError(w, http.StatusBadRequest, "agent not found")
		return
	}
	project, ok := s.store.GetProject(card.ProjectID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "project not found")
		return
	}

	id, ref := s.store.ReserveTaskID()
	branch := fmt.Sprintf("sillage/%d-%s", ref, Slugify(body.Title))
	dir, base, err := CreateWorktree(project.Path, s.dataDir, id, branch)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to create worktree: "+err.Error())
		return
	}

	task, err := s.store.CreateTask(id, ref, card.ID, project.ID, body.Title, body.AgentID, branch, base, dir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	s.runner.publishTask(task)
	s.runner.publishCards(project.ID)

	cliInput := body.Title
	if body.Prompt != "" {
		cliInput = body.Title + "\n\n" + body.Prompt
	}
	authorName := ""
	if me, ok := userFromContext(r); ok {
		authorName = me.Name
	}
	if err := s.runner.Start(task.ID, true, cliInput, authorName); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start agent: "+err.Error())
		return
	}
	task, _ = s.store.GetTask(task.ID)
	writeJSON(w, http.StatusOK, task)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
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
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	if err := decodeJSON(r, &body); err != nil || strings.TrimSpace(body.Text) == "" {
		writeError(w, http.StatusBadRequest, "text is required")
		return
	}
	authorName := ""
	if me, ok := userFromContext(r); ok {
		authorName = me.Name
	}
	if err := s.runner.Start(id, false, body.Text, authorName); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) handleInterrupt(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, ok := s.store.GetTask(id); !ok {
		writeError(w, http.StatusNotFound, "task not found")
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
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if task.Status != "review" {
		writeError(w, http.StatusBadRequest, "task is not in review")
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
		writeError(w, http.StatusBadRequest, "confirmation required")
		return
	}
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if task.Status != "ready" {
		writeError(w, http.StatusBadRequest, "task must be ready before shipping")
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

// handlePR ouvre une pull request pour une tâche déjà livrée (shipped).
// N'exécute jamais de push : la branche a déjà été poussée par Ship.
func (s *Server) handlePR(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeJSON(r, &body); err != nil || !body.Confirm {
		writeError(w, http.StatusBadRequest, "confirmation required")
		return
	}
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if task.Status != "shipped" {
		writeError(w, http.StatusBadRequest, "task must be shipped before opening a pull request")
		return
	}

	url, err := OpenPR(task.WorktreeDir, task.Branch, task.Base, task.Title)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, PRResponse{URL: url})
}

func (s *Server) handleReopen(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if task.Status != "shipped" {
		writeError(w, http.StatusBadRequest, "only a shipped task can be reopened")
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
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	files, err := Diff(task.WorktreeDir, task.Base)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute diff: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, DiffResponse{Branch: task.Branch, Base: task.Base, Files: files})
}

func (s *Server) handleDeliverables(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	resp := DeliverablesResponse{Code: []Item{}, Docs: []Item{}, Images: []Item{}}
	resp.Code = append(resp.Code, Item{Kind: "branch", Title: task.Branch, Meta: "base: " + task.Base})
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
