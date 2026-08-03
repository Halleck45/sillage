package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

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

	// pwMu protège passwordHash, qui peut être remplacé en mémoire après le
	// rapatriement (clone) d'un espace de travail, sans redémarrage.
	pwMu         sync.RWMutex
	passwordHash string

	// autoSyncMu protège autoSyncStop (goroutine de synchronisation
	// automatique périodique, voir workspace.go).
	autoSyncMu   sync.Mutex
	autoSyncStop chan struct{}

	// syncErrMu protège lastSyncErr, état en mémoire uniquement (jamais
	// persisté) du dernier échec de synchronisation automatique.
	syncErrMu   sync.Mutex
	lastSyncErr string
}

// NewServer construit le serveur Sillage. webFS doit pointer vers le sous-répertoire
// "web" du système de fichiers embarqué (frontend statique).
func NewServer(store *Store, passwordHash, dataDir string, webFS fs.FS) *Server {
	hub := NewHub()
	s := &Server{
		store:        store,
		hub:          hub,
		runner:       NewRunner(store, hub),
		sessions:     NewSessionManager(),
		loginLimiter: NewLoginLimiter(),
		passwordHash: passwordHash,
		dataDir:      dataDir,
		static:       staticWithRevalidation(webFS, http.FileServer(http.FS(webFS))),
	}
	if store.GetWorkspace().AutoSync {
		s.startAutoSync()
	}
	return s
}

// staticWithRevalidation ajoute un ETag (empreinte du contenu embarqué) et
// Cache-Control: no-cache aux fichiers statiques. Sans ça, embed.FS ne fournit
// aucune date de modification : la réponse n'a aucun validateur, le navigateur
// ne revalide jamais et continue de servir son ancienne copie de app.js après
// un rebuild. Les empreintes sont calculées une fois au démarrage (le frontend
// embarqué ne change pas en cours de route) ; http.ServeContent lit l'ETag déjà
// posé sur la réponse et répond 304 aux requêtes If-None-Match.
func staticWithRevalidation(webFS fs.FS, next http.Handler) http.Handler {
	etags := map[string]string{}
	_ = fs.WalkDir(webFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := fs.ReadFile(webFS, p)
		if err != nil {
			return nil
		}
		sum := sha256.Sum256(content)
		etags["/"+p] = `"` + hex.EncodeToString(sum[:8]) + `"`
		return nil
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if etag, ok := etags[path]; ok {
			w.Header().Set("Etag", etag)
			w.Header().Set("Cache-Control", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) getPasswordHash() string {
	s.pwMu.RLock()
	defer s.pwMu.RUnlock()
	return s.passwordHash
}

func (s *Server) setPasswordHash(hash string) {
	s.pwMu.Lock()
	defer s.pwMu.Unlock()
	s.passwordHash = hash
}

// Handler construit le mux HTTP complet avec ses middlewares.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.handleLogout)
	mux.HandleFunc("GET /api/state", s.handleState)
	mux.HandleFunc("GET /api/workspace", s.handleGetWorkspace)
	mux.HandleFunc("POST /api/workspace/setup", s.handleWorkspaceSetup)
	mux.HandleFunc("PATCH /api/workspace", s.handleUpdateWorkspace)
	mux.HandleFunc("POST /api/workspace/sync", s.handleWorkspaceSync)
	mux.HandleFunc("PATCH /api/settings", s.handleUpdateSettings)
	mux.HandleFunc("POST /api/agents", s.handleCreateAgent)
	mux.HandleFunc("PATCH /api/agents/{id}", s.handleUpdateAgent)
	mux.HandleFunc("DELETE /api/agents/{id}", s.handleDeleteAgent)
	mux.HandleFunc("POST /api/projects", s.handleCreateProject)
	mux.HandleFunc("PATCH /api/projects/{id}", s.handleUpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", s.handleDeleteProject)
	mux.HandleFunc("POST /api/cards", s.handleCreateCard)
	mux.HandleFunc("PATCH /api/cards/{id}", s.handleUpdateCard)
	mux.HandleFunc("DELETE /api/cards/{id}", s.handleDeleteCard)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", s.handleReassignTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTask)
	mux.HandleFunc("POST /api/tasks/{id}/messages", s.handlePostMessage)
	mux.HandleFunc("POST /api/tasks/{id}/interrupt", s.handleInterrupt)
	mux.HandleFunc("POST /api/tasks/{id}/ship", s.handleShip)
	mux.HandleFunc("POST /api/tasks/{id}/pr", s.handlePR)
	mux.HandleFunc("POST /api/tasks/{id}/finish", s.handleFinish)
	mux.HandleFunc("POST /api/tasks/{id}/cancel", s.handleCancel)
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
				writeError(w, http.StatusForbidden, "Content-Type application/json required")
				return
			}
		}

		if r.URL.Path != "/api/login" {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil || !s.sessions.Validate(cookie.Value) {
				writeError(w, http.StatusUnauthorized, "authentication required")
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

// statusForStoreError distingue "not found" (404) des autres erreurs de
// validation (400) renvoyées par les méthodes du Store.
func statusForStoreError(err error) int {
	if strings.Contains(err.Error(), "not found") {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}

// --- Auth ---

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if s.loginLimiter.Blocked(ip) {
		writeError(w, http.StatusTooManyRequests, "too many attempts, try again in a minute")
		return
	}

	var body struct {
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(s.getPasswordHash()), []byte(body.Password)) != nil {
		s.loginLimiter.RecordFailure(ip)
		writeError(w, http.StatusUnauthorized, "invalid password")
		return
	}

	token, err := s.sessions.Create()
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

// --- État global ---

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	state := s.store.Snapshot()
	state.Workspace = s.workspaceStatus()
	writeJSON(w, http.StatusOK, state)
}

// --- Réglages globaux ---

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DisplayName *string `json:"displayName"`
		Lang        *string `json:"lang"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	settings, err := s.store.UpdateSettings(body.DisplayName, body.Lang)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.runner.publishSettings(settings)
	writeJSON(w, http.StatusOK, settings)
}

// --- Espace de travail (synchronisation git de dataDir) ---

func (s *Server) handleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.workspaceStatus())
}

func (s *Server) handleWorkspaceSetup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Mode   string `json:"mode"`
		Remote string `json:"remote"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ws := s.store.GetWorkspace()
	// "init" reste autorisé après coup pour activer git sur un espace local ;
	// "local" et "clone" ne peuvent être joués qu'une seule fois.
	if body.Mode != "init" && ws.SetupDone {
		writeError(w, http.StatusBadRequest, "workspace setup already done")
		return
	}

	switch body.Mode {
	case "local":
		if _, err := s.store.UpdateWorkspace(func(w *Workspace) { w.SetupDone = true }); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save workspace")
			return
		}

	case "init":
		if err := InitWorkspaceGit(s.dataDir, body.Remote); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		if _, err := s.store.UpdateWorkspace(func(w *Workspace) {
			w.SetupDone = true
			if body.Remote != "" {
				w.SyncRemote = body.Remote
			}
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save workspace")
			return
		}

	case "clone":
		if body.Remote == "" {
			writeError(w, http.StatusBadRequest, "remote is required")
			return
		}
		cloneDir, err := CloneWorkspace(s.dataDir, body.Remote)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		if err := ReplaceWorkspaceFiles(s.dataDir, cloneDir); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if err := s.store.ReloadFromDisk(); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to reload workspace: "+err.Error())
			return
		}
		// Le mot de passe devient celui de l'espace rapatrié, sans redémarrage.
		if hash, ok, err := ReadPasswordHash(s.dataDir); err == nil && ok {
			s.setPasswordHash(hash)
		}
		if _, err := s.store.UpdateWorkspace(func(w *Workspace) {
			w.SetupDone = true
			w.SyncRemote = body.Remote
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save workspace")
			return
		}

	default:
		writeError(w, http.StatusBadRequest, "invalid mode: must be local, init or clone")
		return
	}

	status := s.workspaceStatus()
	s.runner.publishWorkspace(status)
	writeJSON(w, http.StatusOK, status)
}

// handleUpdateWorkspace applique remote et/ou autoSync (au moins un des deux
// requis). Activer autoSync exige git initialisé ET un remote déjà défini
// (celui fourni dans le même appel compte) : 400 sinon.
func (s *Server) handleUpdateWorkspace(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Remote   *string `json:"remote"`
		AutoSync *bool   `json:"autoSync"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Remote == nil && body.AutoSync == nil {
		writeError(w, http.StatusBadRequest, "remote or autoSync is required")
		return
	}

	if body.Remote != nil {
		if *body.Remote == "" {
			writeError(w, http.StatusBadRequest, "remote is required")
			return
		}
		if !WorkspaceGitEnabled(s.dataDir) {
			writeError(w, http.StatusBadRequest, "git is not initialized for this workspace")
			return
		}
		if err := SetWorkspaceRemote(s.dataDir, *body.Remote); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		if _, err := s.store.UpdateWorkspace(func(w *Workspace) { w.SyncRemote = *body.Remote }); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save workspace")
			return
		}
	}

	if body.AutoSync != nil {
		if *body.AutoSync {
			ws := s.store.GetWorkspace()
			if !WorkspaceGitEnabled(s.dataDir) || ws.SyncRemote == "" {
				writeError(w, http.StatusBadRequest, "git must be initialized with a remote before enabling automatic sync")
				return
			}
		}
		if _, err := s.store.UpdateWorkspace(func(w *Workspace) { w.AutoSync = *body.AutoSync }); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save workspace")
			return
		}
		if *body.AutoSync {
			s.startAutoSync()
		} else {
			s.stopAutoSync()
		}
	}

	status := s.workspaceStatus()
	s.runner.publishWorkspace(status)
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleWorkspaceSync(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeJSON(r, &body); err != nil || !body.Confirm {
		writeError(w, http.StatusBadRequest, "confirmation required")
		return
	}
	if !WorkspaceGitEnabled(s.dataDir) {
		writeError(w, http.StatusBadRequest, "git is not initialized for this workspace")
		return
	}

	// SyncPush n'opère que sur s.dataDir (le dataDir configuré côté serveur) :
	// jamais une valeur issue de la requête.
	output, err := SyncPush(s.dataDir)
	if err != nil {
		if errors.Is(err, ErrSyncConflict) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	now := time.Now().UTC()
	// Une synchronisation manuelle réussie efface toujours lastSyncError,
	// même si elle ne venait pas d'un conflit : elle "relance" l'auto-sync
	// mise en pause le cas échéant (voir Server.autoSyncTick).
	s.setLastSyncError("")
	if _, err := s.store.UpdateWorkspace(func(w *Workspace) { w.LastSyncAt = &now }); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save workspace")
		return
	}
	s.runner.publishWorkspace(s.workspaceStatus())
	writeJSON(w, http.StatusOK, SyncResponse{Output: output, LastSyncAt: now})
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
		Name          string `json:"name"`
		Path          string `json:"path"`
		Repos         []Repo `json:"repos"`
		Description   string `json:"description"`
		ContextPrompt string `json:"contextPrompt"`
		Links         []Link `json:"links"`
	}
	if err := decodeJSON(r, &body); err != nil || body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	repos := body.Repos
	if len(repos) == 0 && body.Path != "" {
		repos = []Repo{{Path: body.Path}}
	}
	normalized, err := NormalizeRepos(repos)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	for _, repo := range normalized {
		if err := ValidateRepoPath(repo.Path); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	links, err := NormalizeLinks(body.Links)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	links = fillMissingLinkTitles(links)
	project, err := s.store.AddProject(body.Name, body.Description, body.ContextPrompt, normalized, links)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create project")
		return
	}
	writeJSON(w, http.StatusOK, project)
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		CheckCmd      *string `json:"checkCmd"`
		ContextPrompt *string `json:"contextPrompt"`
		Repos         *[]Repo `json:"repos"`
		Links         *[]Link `json:"links"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Repos != nil {
		normalized, err := NormalizeRepos(*body.Repos)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		for _, repo := range normalized {
			if err := ValidateRepoPath(repo.Path); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
		}
		body.Repos = &normalized
	}
	if body.Links != nil {
		links, err := NormalizeLinks(*body.Links)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		links = fillMissingLinkTitles(links)
		body.Links = &links
	}
	project, err := s.store.UpdateProject(id, body.Name, body.Description, body.CheckCmd, body.ContextPrompt, body.Repos, body.Links)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	s.runner.publishProject(project)
	writeJSON(w, http.StatusOK, project)
}

// handleDeleteProject supprime un projet, avec cascade sur ses chantiers et
// tâches (interruption des agents en cours, retrait best-effort des
// worktrees). Ne publie que projectDeleted, le front recharge l'état complet.
func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeJSON(r, &body); err != nil || !body.Confirm {
		writeError(w, http.StatusBadRequest, "confirmation required")
		return
	}
	if err := s.runner.DeleteProject(id); err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Cartes ---

func (s *Server) handleCreateCard(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ProjectID     string `json:"projectId"`
		Title         string `json:"title"`
		Column        string `json:"column"`
		ContextPrompt string `json:"contextPrompt"`
	}
	if err := decodeJSON(r, &body); err != nil || body.ProjectID == "" || body.Title == "" {
		writeError(w, http.StatusBadRequest, "projectId and title are required")
		return
	}
	card, err := s.store.AddCard(body.ProjectID, body.Title, body.Column, body.ContextPrompt)
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
		Column        *string `json:"column"`
		Title         *string `json:"title"`
		ContextPrompt *string `json:"contextPrompt"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	card, err := s.store.UpdateCard(id, body.Column, body.Title, body.ContextPrompt)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	s.runner.publishCards(card.ProjectID)
	writeJSON(w, http.StatusOK, card)
}

// handleDeleteCard supprime une carte (chantier), avec cascade sur ses
// tâches (interruption des agents en cours, retrait best-effort des
// worktrees).
func (s *Server) handleDeleteCard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeJSON(r, &body); err != nil || !body.Confirm {
		writeError(w, http.StatusBadRequest, "confirmation required")
		return
	}
	if err := s.runner.DeleteCard(id); err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Tâches ---

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CardID   string `json:"cardId"`
		Title    string `json:"title"`
		AgentID  string `json:"agentId"`
		Prompt   string `json:"prompt"`
		RepoName string `json:"repoName"`
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
	repo, err := s.store.ResolveTaskRepo(project.ID, body.RepoName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, ref := s.store.ReserveTaskID()
	branch := fmt.Sprintf("sillage/%d-%s", ref, Slugify(body.Title))
	dir, base, err := CreateWorktree(repo.Path, s.dataDir, id, branch)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to create worktree: "+err.Error())
		return
	}

	task, err := s.store.CreateTask(id, ref, card.ID, project.ID, body.Title, body.AgentID, branch, base, dir, repo.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	s.runner.publishTask(task)
	s.runner.publishCards(project.ID)

	cliInput := contextualizeCliInput(body.Title, body.Prompt)
	if err := s.runner.Start(task.ID, true, cliInput); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start agent: "+err.Error())
		return
	}
	task, _ = s.store.GetTask(task.ID)
	writeJSON(w, http.StatusOK, task)
}

// handleReassignTask change l'agent assigné à une tâche : refusé si la tâche
// est en cours d'exécution ou si l'agent est inconnu. Ajoute un message
// marqueur au fil (détecté et localisé par le frontend) et republie
// task + message + agents.
func (s *Server) handleReassignTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		AgentID string `json:"agentId"`
	}
	if err := decodeJSON(r, &body); err != nil || body.AgentID == "" {
		writeError(w, http.StatusBadRequest, "agentId is required")
		return
	}
	task, err := s.store.ReassignTask(id, body.AgentID)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "task not found" {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	msg, task, err := s.store.AddMessage(task.ID, "agent", "", "[reassigned:"+body.AgentID+"]")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.runner.publishTask(task)
	s.runner.publishMessage(msg)
	s.runner.publishAgents()
	writeJSON(w, http.StatusOK, task)
}

// handleDeleteTask supprime une tâche : interrompt l'agent au préalable s'il
// tourne encore (même mécanique que cancel), retire le worktree (best-effort,
// jamais la branche), supprime la tâche et ses messages.
func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeJSON(r, &body); err != nil || !body.Confirm {
		writeError(w, http.StatusBadRequest, "confirmation required")
		return
	}
	if err := s.runner.DeleteTask(id); err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
	queued, err := s.runner.Message(id, body.Text)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]bool{"queued": queued})
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

// handleShip pousse la branche de la tâche (statut ready) : accepté depuis
// "review" (l'étape d'acceptation manuelle a disparu, v0.3.4). Ajoute un
// message marqueur "[shipped:<branch>]" et fournit branchUrl (GitHub
// uniquement, vide sinon) dans la réponse.
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
	if task.Status != "review" {
		writeError(w, http.StatusBadRequest, "task must be in review before shipping")
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
	msg, task, err := s.store.AddMessage(task.ID, "agent", "", "[shipped:"+task.Branch+"]")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.runner.publishTask(task)
	s.runner.publishMessage(msg)
	s.runner.publishCards(task.ProjectID)
	s.runner.publishAgents()
	branchUrl := githubBranchURL(task.WorktreeDir, task.Branch)
	writeJSON(w, http.StatusOK, ShipResponse{Task: task, Output: output, BranchUrl: branchUrl})
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

// handleFinish marque une tâche "done" (autorisé depuis review/ready/shipped).
func (s *Server) handleFinish(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := s.store.FinishTask(id)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	s.runner.publishTask(task)
	s.runner.publishCards(task.ProjectID)
	writeJSON(w, http.StatusOK, task)
}

// handleCancel annule une tâche (autorisé depuis running/review/ready) ;
// si elle est en cours d'exécution, l'agent est interrompu au préalable.
func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := s.runner.Cancel(id)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// handleReopen remet une tâche en revue (autorisé depuis shipped/done/cancelled).
func (s *Server) handleReopen(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := s.store.ReopenTask(id)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	s.runner.publishTask(task)
	s.runner.publishCards(task.ProjectID)
	writeJSON(w, http.StatusOK, task)
}

// handleRead marque une tâche comme lue. N'affecte jamais UpdatedAt : ouvrir
// une tâche ne doit pas la faire remonter dans une liste triée par date.
func (s *Server) handleRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := s.store.MarkTaskRead(id)
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
