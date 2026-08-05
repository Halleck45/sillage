package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
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
	previews     *PreviewSupervisor
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

	// updates : ce que la dernière vérification de mise à jour a appris, plus
	// la goroutine périodique (voir update.go). En mémoire uniquement.
	updates *updateTracker

	// rebaseMu sérialise les rebases automatiques déclenchés par une
	// acceptation (voir rebaseSiblingTasks) : deux acceptations rapprochées ne
	// doivent pas rejouer la même branche en parallèle. rebaseWG permet aux
	// tests d'attendre la fin de ces rebases, qui sont lancés en tâche de fond
	// pour ne pas retarder la réponse de l'acceptation.
	rebaseMu sync.Mutex
	rebaseWG sync.WaitGroup
}

// waitRebases attend la fin des rebases automatiques en cours. Réservé aux
// tests : en production personne n'attend une opération de fond.
func (s *Server) waitRebases() { s.rebaseWG.Wait() }

// NewServer construit le serveur Sillage. webFS doit pointer vers le sous-répertoire
// "web" du système de fichiers embarqué (frontend statique).
func NewServer(store *Store, passwordHash, dataDir string, webFS fs.FS) *Server {
	hub := NewHub()
	s := &Server{
		store:        store,
		hub:          hub,
		runner:       NewRunner(store, hub),
		previews:     NewPreviewSupervisor(hub),
		sessions:     NewSessionManager(),
		loginLimiter: NewLoginLimiter(),
		passwordHash: passwordHash,
		dataDir:      dataDir,
		static:       staticWithRevalidation(webFS, http.FileServer(http.FS(webFS))),
		updates:      &updateTracker{},
	}
	if store.GetWorkspace().AutoSync {
		s.startAutoSync()
	}
	// Sans effet sur une compilation locale (version "dev") ni si le réglage
	// est coupé : aucun appel réseau n'est fait dans ces deux cas.
	s.startUpdateChecker()
	return s
}

// Shutdown arrête ce qui tourne pour le compte de Sillage : les recettes
// manuelles (voir docs/SPEC-RECETTE.md §3, rien ne survit à la fermeture) et la
// synchronisation automatique de l'espace de travail. Les agents ne sont pas
// interrompus : leur travail est dans un worktree, il survit au redémarrage.
func (s *Server) Shutdown() {
	s.previews.StopAll()
	s.stopAutoSync()
	s.stopUpdateChecker()
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
	mux.HandleFunc("GET /api/update", s.handleGetUpdate)
	mux.HandleFunc("POST /api/update/check", s.handleUpdateCheck)
	mux.HandleFunc("POST /api/update/apply", s.handleUpdateApply)
	mux.HandleFunc("POST /api/agents", s.handleCreateAgent)
	mux.HandleFunc("PATCH /api/agents/{id}", s.handleUpdateAgent)
	mux.HandleFunc("DELETE /api/agents/{id}", s.handleDeleteAgent)
	mux.HandleFunc("POST /api/projects", s.handleCreateProject)
	mux.HandleFunc("PATCH /api/projects/{id}", s.handleUpdateProject)
	mux.HandleFunc("DELETE /api/projects/{id}", s.handleDeleteProject)
	mux.HandleFunc("POST /api/projects/{id}/mark-all-read", s.handleMarkProjectAllRead)
	mux.HandleFunc("POST /api/cards", s.handleCreateCard)
	mux.HandleFunc("PATCH /api/cards/{id}", s.handleUpdateCard)
	mux.HandleFunc("DELETE /api/cards/{id}", s.handleDeleteCard)
	mux.HandleFunc("GET /api/cards/{id}/delivery", s.handleCardDelivery)
	mux.HandleFunc("POST /api/cards/{id}/ship", s.handleCardShip)
	mux.HandleFunc("POST /api/cards/{id}/catch-up", s.handleCardCatchUp)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", s.handleReassignTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTask)
	mux.HandleFunc("POST /api/tasks/{id}/messages", s.handlePostMessage)
	mux.HandleFunc("POST /api/tasks/{id}/interrupt", s.handleInterrupt)
	mux.HandleFunc("POST /api/tasks/{id}/start", s.handleStartWaitingTask)
	mux.HandleFunc("POST /api/tasks/{id}/accept", s.handleAccept)
	mux.HandleFunc("POST /api/tasks/{id}/cancel", s.handleCancel)
	mux.HandleFunc("POST /api/tasks/{id}/reopen", s.handleReopen)
	mux.HandleFunc("POST /api/tasks/{id}/read", s.handleRead)
	mux.HandleFunc("GET /api/tasks/{id}/diff", s.handleDiff)
	mux.HandleFunc("GET /api/tasks/{id}/deliverables", s.handleDeliverables)
	mux.HandleFunc("POST /api/cards/{id}/preview", s.handleCardPreview)
	mux.HandleFunc("POST /api/tasks/{id}/preview", s.handleTaskPreview)
	mux.HandleFunc("POST /api/previews/{id}/stop", s.handlePreviewStop)
	mux.HandleFunc("GET /api/previews/{id}/log", s.handlePreviewLog)
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

		if r.URL.Path != "/api/login" && s.getPasswordHash() != "" {
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
	state.Previews = s.previews.List()
	state.Update = s.UpdateStatus()
	writeJSON(w, http.StatusOK, state)
}

// --- Réglages globaux ---

func (s *Server) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DisplayName *string `json:"displayName"`
		Lang        *string `json:"lang"`
		UpdateCheck *bool   `json:"updateCheck"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	settings, err := s.store.UpdateSettings(body.DisplayName, body.Lang, body.UpdateCheck)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if body.UpdateCheck != nil {
		if *body.UpdateCheck {
			s.startUpdateChecker()
		} else {
			s.stopUpdateChecker()
		}
		s.runner.publishUpdate(s.UpdateStatus())
	}
	s.runner.publishSettings(settings)
	writeJSON(w, http.StatusOK, settings)
}

// --- Mises à jour de Sillage (voir update.go) ---

func (s *Server) handleGetUpdate(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.UpdateStatus())
}

// handleUpdateCheck force une vérification immédiate : le geste explicite de
// celui qui ne veut pas attendre le tick de 24 h (ou qui a coupé le réglage et
// veut regarder une fois).
func (s *Server) handleUpdateCheck(w http.ResponseWriter, r *http.Request) {
	if !isReleaseVersion(buildVersion) {
		writeError(w, http.StatusBadRequest, "this build has no version to compare")
		return
	}
	// La vérification explicite est aussi le moment de reregarder le lancement
	// à l'ouverture de session : c'est le seul geste qui suit, dans le temps, un
	// `brew services start` joué dans un terminal.
	s.refreshServiceStatus()
	ctx, cancel := context.WithTimeout(r.Context(), updateHTTPTimeout)
	defer cancel()
	st := s.checkForUpdate(ctx)
	if st.Error != "" {
		writeError(w, http.StatusBadGateway, st.Error)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// handleUpdateApply installe la nouvelle version puis remplace le process
// courant. Action sortante : {"confirm": true} obligatoire.
func (s *Server) handleUpdateApply(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeJSON(r, &body); err != nil || !body.Confirm {
		writeError(w, http.StatusBadRequest, "confirmation required")
		return
	}

	version := s.UpdateStatus().Latest
	output, execPath, err := s.applyUpdate()
	if err != nil {
		if output != "" {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error(), "output": output})
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	resp := UpdateApplyResponse{Output: output, Version: version, Restarting: execPath != ""}
	if execPath == "" {
		resp.Note = "update installed, restart Sillage to run it"
	}
	writeJSON(w, http.StatusOK, resp)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
	if execPath == "" {
		return
	}

	// Le redémarrage se fait après la réponse : la nouvelle version prend la
	// place du process courant (même PID, même terminal), les navigateurs
	// ouverts se reconnectent tout seuls.
	go func() {
		time.Sleep(updateRestartGrace)
		log.Printf("Sillage updated to %s: restarting", version)
		s.Shutdown()
		if err := restartInPlace(execPath); err != nil {
			log.Printf("update installed but restart failed (%v): restart Sillage by hand", err)
		}
	}()
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
		Name          string    `json:"name"`
		Path          string    `json:"path"`
		Repos         []Repo    `json:"repos"`
		Description   string    `json:"description"`
		ContextPrompt string    `json:"contextPrompt"`
		Links         []Link    `json:"links"`
		Delivery      *Delivery `json:"delivery"`
	}
	// Le nom est optionnel : sans lui, AddProject prend celui du premier dépôt.
	// Créer un projet ne demande donc qu'un chemin.
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
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
	delivery := body.Delivery
	if delivery == nil {
		// Le réglage de livraison n'est jamais une question posée à froid :
		// sans valeur explicite, il est déduit des remotes des dépôts.
		detected := detectDelivery(normalized)
		delivery = &detected
	}
	project, err := s.store.AddProject(body.Name, body.Description, body.ContextPrompt, normalized, links, delivery)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, projectOut(project))
}

// detectDelivery déduit le mode de livraison par défaut d'un projet : "pr" si
// tous les dépôts pointent vers une forge connue (GitHub ou GitLab), sinon
// fusion locale dans la branche courante du premier dépôt.
func detectDelivery(repos []Repo) Delivery {
	for _, repo := range repos {
		if DetectForge(repo.Path).Provider == "" {
			target := ""
			if len(repos) > 0 {
				target, _ = currentBranch(repos[0].Path)
			}
			return Delivery{Mode: "merge", Target: target}
		}
	}
	return Delivery{Mode: "pr"}
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Name          *string   `json:"name"`
		Description   *string   `json:"description"`
		CheckCmd      *string   `json:"checkCmd"`
		ContextPrompt *string   `json:"contextPrompt"`
		AllowedTools  *[]string `json:"allowedTools"`
		Repos         *[]Repo   `json:"repos"`
		Links         *[]Link   `json:"links"`
		Delivery      *Delivery `json:"delivery"`
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
	project, err := s.store.UpdateProject(id, body.Name, body.Description, body.CheckCmd, body.ContextPrompt, body.AllowedTools, body.Repos, body.Links, body.Delivery)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	out := projectOut(project)
	s.runner.publishProject(out)
	writeJSON(w, http.StatusOK, out)
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

// handleMarkProjectAllRead marque comme lues toutes les tâches non lues d'un
// projet (menu "..." d'un projet dans la sidebar). Action locale et
// réversible (une nouvelle activité repasse une tâche à non lue) : aucune
// confirmation.
func (s *Server) handleMarkProjectAllRead(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	tasks, err := s.store.MarkAllTasksReadForProject(id)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	for _, t := range tasks {
		s.runner.publishTask(t)
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
		CardID         string `json:"cardId"`
		Title          string `json:"title"`
		AgentID        string `json:"agentId"`
		Prompt         string `json:"prompt"`
		RepoName       string `json:"repoName"`
		WaitsForTaskID string `json:"waitsForTaskId"`
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
	// waitsForTaskId : la tâche reste "waiting" (agent non lancé) jusqu'à
	// l'acceptation de la tâche référencée (voir Server.startDependentTasks).
	// Seule une tâche du même chantier, pas déjà terminale, a un sens ici : une
	// tâche accepted/cancelled ne fera plus jamais rien.
	if body.WaitsForTaskID != "" {
		dep, ok := s.store.GetTask(body.WaitsForTaskID)
		if !ok || dep.CardID != body.CardID {
			writeError(w, http.StatusBadRequest, "waitsForTaskId must reference a task of the same workstream")
			return
		}
		if dep.Status == "accepted" || dep.Status == "cancelled" {
			writeError(w, http.StatusBadRequest, "waitsForTaskId task is already finished")
			return
		}
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

	// La tâche part de la branche du chantier (créée à la première tâche du
	// chantier sur ce dépôt) : une tâche créée après une acceptation démarre
	// donc sur le travail déjà accepté.
	cardBranch, err := s.ensureCardBranch(card, project, repo)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, ref := s.store.ReserveTaskID()
	branch := fmt.Sprintf("sillage/%d-%s", ref, Slugify(body.Title))
	dir, base, err := CreateWorktree(repo.Path, s.dataDir, id, branch, cardBranch.Branch)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to create worktree: "+err.Error())
		return
	}

	var task Task
	if body.WaitsForTaskID != "" {
		task, err = s.store.CreateWaitingTask(id, ref, card.ID, project.ID, body.Title, body.AgentID, branch, base, dir, repo.Name, body.WaitsForTaskID, body.Prompt)
	} else {
		task, err = s.store.CreateTask(id, ref, card.ID, project.ID, body.Title, body.AgentID, branch, base, dir, repo.Name)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	s.runner.publishTask(task)
	s.runner.publishCards(project.ID)

	if task.Status == "waiting" {
		writeJSON(w, http.StatusOK, task)
		return
	}

	cliInput := contextualizeCliInput(body.Title, body.Prompt)
	if err := s.runner.Start(task.ID, true, cliInput); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start agent: "+err.Error())
		return
	}
	task, _ = s.store.GetTask(task.ID)
	writeJSON(w, http.StatusOK, task)
}

// handleStartWaitingTask lance manuellement une tâche "waiting", sans
// attendre l'acceptation de sa dépendance : débloque une tâche dont la
// dépendance a été refusée ou supprimée, ou permet simplement de changer
// d'avis sur l'ordre.
func (s *Server) handleStartWaitingTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if task.Status != "waiting" {
		writeError(w, http.StatusBadRequest, "only a waiting task can be started")
		return
	}
	if err := s.startWaitingTask(task); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to start agent: "+err.Error())
		return
	}
	task, _ = s.store.GetTask(id)
	writeJSON(w, http.StatusOK, task)
}

// startWaitingTask lance l'agent d'une tâche "waiting" : rejoue son worktree
// sur la tête courante de la branche du chantier (qui a pu avancer depuis la
// création de la tâche) puis démarre l'agent avec le prompt initial mémorisé
// à la création. Le rebase est best-effort et ignoré en cas d'échec : une
// tâche jamais démarrée n'a que sa base comme contenu, donc rien à perdre, et
// démarrer légèrement en retard reste rattrapable comme pour toute tâche
// (voir le badge "en retard").
func (s *Server) startWaitingTask(task Task) error {
	if cardBranch, ok := s.store.GetCardBranch(task.CardID, task.RepoName); ok && task.WorktreeDir != "" {
		if !IsBranchMergedInto(task.WorktreeDir, cardBranch.Branch, task.Branch) {
			_, _, _ = RebaseOnto(task.WorktreeDir, cardBranch.Branch)
		}
	}
	prompt := task.PendingPrompt
	if _, err := s.store.UpdateTask(task.ID, func(t *Task) {
		t.WaitsForTaskID = ""
		t.PendingPrompt = ""
	}); err != nil {
		return err
	}
	cliInput := contextualizeCliInput(task.Title, prompt)
	return s.runner.Start(task.ID, true, cliInput)
}

// startDependentTasks démarre les tâches "waiting" du chantier qui
// attendaient l'acceptation de acceptedTaskID (voir Task.WaitsForTaskID).
// Appelée après une acceptation réussie ; best-effort, une tâche qui échoue à
// démarrer n'empêche pas les autres.
func (s *Server) startDependentTasks(cardID, acceptedTaskID string) {
	for _, t := range s.store.TasksByCard(cardID) {
		if t.Status != "waiting" || t.WaitsForTaskID != acceptedTaskID {
			continue
		}
		_ = s.startWaitingTask(t)
	}
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

// ensureCardBranch retourne la branche de feature du chantier sur un dépôt, en
// la créant (branche + worktree dédié) si c'est la première tâche du chantier
// sur ce dépôt. La base est la branche de destination du projet
// (project.delivery.target), ou la branche courante du dépôt si elle est vide.
//
// Appelée à chaque création de tâche : si le chantier avait déjà été livré, il
// repasse en « non livré » (tout n'est plus livré dès qu'il reste du travail à
// faire), ce qui permet de continuer à travailler sur un chantier livré.
func (s *Server) ensureCardBranch(card Card, project Project, repo Repo) (CardBranch, error) {
	if existing, ok := s.store.GetCardBranch(card.ID, repo.Name); ok {
		if existing.ShippedAt != nil {
			if _, err := s.store.MarkCardBranchPending(card.ID, repo.Name); err != nil {
				return CardBranch{}, err
			}
			existing.ShippedAt = nil
			s.runner.publishCards(card.ProjectID)
		}
		return existing, nil
	}
	branch := fmt.Sprintf("sillage/ws-%d-%s", card.Ref, Slugify(card.Title))
	dir, base, err := CreateCardWorktree(repo.Path, s.dataDir, card.ID, repo.Name, branch, project.Delivery.Target)
	if err != nil {
		return CardBranch{}, fmt.Errorf("failed to create workstream branch: %w", err)
	}
	cb := CardBranch{RepoName: repo.Name, Branch: branch, Base: base, WorktreeDir: dir}
	if _, err := s.store.SetCardBranch(card.ID, cb); err != nil {
		return CardBranch{}, err
	}
	s.runner.publishCards(card.ProjectID)
	return cb, nil
}

// handleAccept accepte une tâche en revue : commite ce qui reste dans son
// worktree, puis fusionne sa branche dans celle du chantier (aucun réseau,
// donc aucune confirmation : l'action est locale et réversible par /reopen).
// En cas de conflit, la tâche RESTE en revue et un message marqueur
// "[merge-conflict:<fichiers>]" est ajouté au fil.
func (s *Server) handleAccept(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, ok := s.store.GetTask(id)
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if task.Status != "review" {
		writeError(w, http.StatusBadRequest, "only a task in review can be accepted")
		return
	}
	// Un rebase automatique est en train de réécrire la branche de la tâche :
	// fusionner maintenant reviendrait à courir contre lui. C'est l'affaire de
	// quelques secondes, et l'UI désactive déjà le bouton.
	if task.Rebasing {
		writeError(w, http.StatusConflict, "task is being rebased, retry in a moment")
		return
	}
	cardBranch, ok := s.store.GetCardBranch(task.CardID, task.RepoName)
	if !ok {
		// Tâche antérieure aux branches de chantier (créée avant cette
		// version) : la branche du chantier est créée à la volée plutôt que de
		// refuser l'acceptation. Les deux branches descendent de la même base,
		// la fusion a donc le même sens qu'une acceptation ordinaire, et le
		// travail en cours n'est pas perdu.
		card, project, found := s.cardWithProject(w, task.CardID)
		if !found {
			return
		}
		repo, found := repoByName(project, task.RepoName)
		if !found {
			writeError(w, http.StatusBadRequest, "repository "+task.RepoName+" is no longer part of the project")
			return
		}
		created, err := s.ensureCardBranch(card, project, repo)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		cardBranch = created
	}
	s.acceptTaskInto(w, task, cardBranch)
}

// acceptTaskInto commite le worktree de la tâche puis fusionne sa branche dans
// celle du chantier, et écrit la réponse HTTP.
func (s *Server) acceptTaskInto(w http.ResponseWriter, task Task, cardBranch CardBranch) {
	id := task.ID

	output, err := CommitAll(task.WorktreeDir, "Sillage: "+task.Title)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	shaBefore := HeadSha(cardBranch.WorktreeDir)
	mergeOut, conflicts, err := MergeBranch(cardBranch.WorktreeDir, task.Branch, "Sillage: "+task.Title)
	output += mergeOut
	if err != nil {
		if len(conflicts) > 0 {
			msg, updated, addErr := s.store.AddMessage(task.ID, "agent", "", "[merge-conflict:"+strings.Join(conflicts, " ")+"]")
			if addErr == nil {
				s.runner.publishTask(updated)
				s.runner.publishMessage(msg)
			}
			// L'agent reprend la main tout de suite avec l'instruction de rebase :
			// le conflit se résout souvent tout seul, sans attendre que l'humain
			// remarque le marqueur et clique sur « Demander le rebase ».
			_, _ = s.runner.Message(task.ID, conflictRebasePrompt(cardBranch.Branch, conflicts))
			writeJSON(w, http.StatusConflict, map[string]string{
				"error":             "merge conflict with the workstream branch",
				"conflictFilePaths": strings.Join(conflicts, " "),
			})
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	// Des commits ont réellement rejoint la branche du chantier (comparaison de
	// SHA, la sortie de git étant localisée) : le chantier n'est plus à jour
	// avec ce qui a été livré, il redevient livrable.
	if HeadSha(cardBranch.WorktreeDir) != shaBefore {
		if _, err := s.store.MarkCardBranchPending(task.CardID, task.RepoName); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	task, err = s.store.AcceptTask(id)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	msg, task, err := s.store.AddMessage(task.ID, "agent", "", "[accepted:"+cardBranch.Branch+"]")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.runner.publishTask(task)
	s.runner.publishMessage(msg)
	s.runner.publishCards(task.ProjectID)
	s.runner.publishAgents()
	// La branche du chantier vient d'avancer : les autres tâches en revue sont
	// désormais en retard, et le seraient encore à leur propre acceptation. On
	// les remet d'aplomb sans faire attendre la réponse : l'UI a déjà tout ce
	// qu'il lui faut, le rebase s'annonce ensuite par SSE.
	s.rebaseWG.Add(1)
	go s.rebaseSiblingTasks(task.CardID, task.ID, cardBranch)
	// Tâches "waiting" sur celle-ci : leur tour est venu.
	s.startDependentTasks(task.CardID, task.ID)
	writeJSON(w, http.StatusOK, AcceptResponse{Task: task, WorkstreamBranch: cardBranch.Branch, Output: output})
}

// rebaseSiblingTasks rejoue les autres tâches en revue du chantier au-dessus de
// la branche du chantier, après qu'une acceptation l'a fait avancer. Sans ça,
// chaque tâche suivante se découvre en retard et conflicte à son tour à
// l'acceptation, alors que la reprise est presque toujours mécanique.
//
// Quatre garde-fous, dans cet ordre (une opération automatique doit être
// conservatrice, surtout quand elle réécrit un historique) :
//
//  1. même dépôt et statut "review" : une tâche en cours (running) appartient à
//     son agent, une tâche terminale n'a plus rien à rejouer ;
//  2. aucun agent en cours pour cette tâche, même en file d'attente ;
//  3. réellement en retard : rien à faire si la branche contient déjà celle du
//     chantier (et rien à annoncer dans le fil) ;
//  4. worktree propre APRÈS commit du travail en attente : les agents ne
//     commitent pas toujours, et un rebase perdrait leur travail. Le commit est
//     le même que celui de l'acceptation (`Sillage: <titre>`), et ne change
//     rien à la relecture : le diff d'une tâche se calcule contre sa base, que
//     le travail soit commité ou non.
//
// Un conflit annule le rebase (l'arbre revient intact) et pose un marqueur dans
// le fil : c'est alors à l'agent de reprendre la base, comme avant.
func (s *Server) rebaseSiblingTasks(cardID, acceptedTaskID string, cardBranch CardBranch) {
	defer s.rebaseWG.Done()
	s.rebaseMu.Lock()
	defer s.rebaseMu.Unlock()

	for _, task := range s.store.TasksByCard(cardID) {
		if task.ID == acceptedTaskID || task.Status != "review" || task.RepoName != cardBranch.RepoName {
			continue
		}
		if task.WorktreeDir == "" || s.runner.IsRunning(task.ID) {
			continue
		}
		if IsBranchMergedInto(task.WorktreeDir, cardBranch.Branch, task.Branch) {
			continue // déjà à jour : la branche du chantier est contenue dans la tâche
		}
		if _, err := CommitAll(task.WorktreeDir, "Sillage: "+task.Title); err != nil {
			continue
		}
		if !IsWorktreeClean(task.WorktreeDir) {
			continue
		}
		s.rebaseOneTask(task, cardBranch)
	}
}

// rebaseOneTask rejoue une tâche sur la branche du chantier, en annonçant
// l'opération (Task.Rebasing, pour le fuseau de l'UI) puis son résultat.
func (s *Server) rebaseOneTask(task Task, cardBranch CardBranch) {
	if updated, err := s.store.SetTaskRebasing(task.ID, true); err == nil {
		s.runner.publishTask(updated)
	}
	_, conflicts, err := RebaseOnto(task.WorktreeDir, cardBranch.Branch)
	if updated, setErr := s.store.SetTaskRebasing(task.ID, false); setErr == nil {
		s.runner.publishTask(updated)
	}

	marker := "[rebased:" + cardBranch.Branch + "]"
	if err != nil {
		if len(conflicts) == 0 {
			// Échec sans conflit (worktree cassé, git indisponible) : rien à
			// raconter dans le fil, le badge de retard reste et le bouton
			// « Demander le rebase » fait toujours son travail.
			return
		}
		marker = "[rebase-conflict:" + strings.Join(conflicts, " ") + "]"
	}
	msg, updated, addErr := s.store.AddMessage(task.ID, "agent", "", marker)
	if addErr != nil {
		return
	}
	s.runner.publishTask(updated)
	s.runner.publishMessage(msg)
	s.runner.publishCards(updated.ProjectID)
	if err != nil {
		// Même logique qu'à l'acceptation (voir acceptTaskInto) : l'agent
		// reprend la main avec l'instruction de rebase plutôt que d'attendre
		// que l'humain remarque le marqueur.
		_, _ = s.runner.Message(task.ID, conflictRebasePrompt(cardBranch.Branch, conflicts))
	}
}

// conflictRebasePrompt construit l'instruction envoyée à l'agent après un
// conflit de fusion ou de rebase : Sillage n'en résout jamais un à sa place
// (voir MergeBranch/RebaseOnto dans git.go), seul l'agent sait reprendre la
// base et régler les conflits.
func conflictRebasePrompt(branch string, conflicts []string) string {
	return fmt.Sprintf(
		"Your branch conflicts with %s on: %s. Rebase your branch onto %s, resolve the conflicts, then rerun the project checks.",
		branch, strings.Join(conflicts, ", "), branch,
	)
}

// autoAcceptMergedTasks marque « acceptées » les tâches en revue dont la
// branche est déjà entièrement contenue dans la branche du chantier : le cas
// d'une fusion faite à la main, hors de Sillage. Appelée à chaque lecture de
// l'aperçu de livraison (ouverture de la vue chantier, puis rafraîchissement
// périodique côté frontend).
//
// Trois garde-fous, dans cet ordre : aucun agent en cours pour la tâche, son
// worktree propre (du travail non commité ne serait par définition pas fusionné
// et serait perdu de vue), et sa branche effectivement contenue dans celle du
// chantier. Aucune écriture git : la fusion a déjà eu lieu, on ne fait que
// constater. Ne remet donc jamais le chantier en « non livré ».
func (s *Server) autoAcceptMergedTasks(card Card) bool {
	accepted := false
	for _, task := range s.store.TasksByCard(card.ID) {
		if task.Status != "review" || s.runner.IsRunning(task.ID) {
			continue
		}
		cb, ok := s.store.GetCardBranch(card.ID, task.RepoName)
		if !ok || task.WorktreeDir == "" {
			continue
		}
		// Une tâche qui n'a rien produit est contenue dans la branche du
		// chantier par construction : la fusionner « automatiquement » ne
		// constaterait rien. Elle reste à relire (accepter ou refuser).
		if task.FilesCount == 0 {
			continue
		}
		if !IsWorktreeClean(task.WorktreeDir) {
			continue
		}
		if !IsBranchMergedInto(cb.WorktreeDir, task.Branch, cb.Branch) {
			continue
		}
		updated, err := s.store.AcceptTask(task.ID)
		if err != nil {
			continue
		}
		msg, updated, err := s.store.AddMessage(updated.ID, "agent", "", "[auto-accepted:"+cb.Branch+"]")
		if err == nil {
			s.runner.publishMessage(msg)
		}
		s.runner.publishTask(updated)
		accepted = true
	}
	if accepted {
		s.runner.publishCards(card.ProjectID)
		s.runner.publishAgents()
	}
	return accepted
}

// handleCardDelivery est l'aperçu de livraison d'un chantier : ce qui va se
// passer, sur quels dépôts, et ce qui bloque le cas échéant. Aucune commande
// d'écriture git ; l'appel constate au passage les branches déjà fusionnées à
// la main (voir autoAcceptMergedTasks) avant de calculer l'aperçu.
func (s *Server) handleCardDelivery(w http.ResponseWriter, r *http.Request) {
	card, project, ok := s.cardWithProject(w, r.PathValue("id"))
	if !ok {
		return
	}
	if s.autoAcceptMergedTasks(card) {
		card, _ = s.store.GetCard(card.ID)
	}
	writeJSON(w, http.StatusOK, s.deliveryPreview(card, project))
}

// cardWithProject résout une carte et son projet, en écrivant la réponse
// d'erreur adéquate si l'un des deux manque.
func (s *Server) cardWithProject(w http.ResponseWriter, cardID string) (Card, Project, bool) {
	card, ok := s.store.GetCard(cardID)
	if !ok {
		writeError(w, http.StatusNotFound, "card not found")
		return Card{}, Project{}, false
	}
	project, ok := s.store.GetProject(card.ProjectID)
	if !ok {
		writeError(w, http.StatusInternalServerError, "project not found")
		return Card{}, Project{}, false
	}
	return card, project, true
}

// deliveryPreview construit l'aperçu de livraison d'un chantier : mode, cible,
// fournisseur détecté, état du bouton, avertissements de santé, compteurs de
// tâches et, par dépôt, le nombre de commits et de fichiers à livrer.
func (s *Server) deliveryPreview(card Card, project Project) DeliveryPreview {
	prev := DeliveryPreview{
		Mode:     project.Delivery.Mode,
		Target:   project.Delivery.Target,
		Ready:    card.ShipReady,
		Blocker:  card.ShipBlocker,
		Warnings: []string{},
		Repos:    []DeliveryRepoPreview{},
		Behind:   map[string]int{},
	}
	for _, t := range s.store.TasksByCard(card.ID) {
		switch t.Status {
		case "accepted":
			prev.Counts.Accepted++
		case "cancelled":
			prev.Counts.Refused++
		default:
			prev.Counts.Pending++
		}
		if behind := taskBehind(t); behind > 0 {
			prev.Behind[t.ID] = behind
		}
	}
	if warning := deliveryWarning(project); warning != "" {
		prev.Warnings = append(prev.Warnings, warning)
	}
	for _, b := range card.Branches {
		row := DeliveryRepoPreview{
			RepoName: b.RepoName, Branch: b.Branch, Base: b.Base,
			PrURL: b.PrURL, ShippedAt: b.ShippedAt,
		}
		if prev.Provider == "" {
			prev.Provider = DetectForge(b.WorktreeDir).Provider
		}
		if commits, err := Commits(b.WorktreeDir, b.Base); err == nil {
			row.Commits = len(commits)
		}
		if files, err := Diff(b.WorktreeDir, b.Base); err == nil {
			row.Files = len(files)
		}
		row.Pending = pendingCommits(project, b, row.Commits)
		if behind, err := CountCommits(b.WorktreeDir, b.Branch, b.Base); err == nil {
			row.Behind = behind
		}
		// Où en est la branche du chantier par rapport à sa destination : déjà
		// arrivée (plus rien à livrer), ou fusionnable en fast-forward. Les deux
		// se lisent avec merge-base, sans rien écrire.
		target := deliveryTarget(project, b)
		row.MergedIntoTarget = IsBranchMergedInto(b.WorktreeDir, b.Branch, target)
		row.FastForwardable = IsBranchMergedInto(b.WorktreeDir, target, b.Branch)
		prev.Repos = append(prev.Repos, row)
	}
	return prev
}

// taskBehind compte les commits que la branche du chantier a et que la branche
// de la tâche n'a pas : le retard qui provoquera un conflit à l'acceptation.
// Zéro pour une tâche terminale (rien à rebaser, worktree possiblement retiré)
// ou quand la révision de comparaison manque : mieux vaut ne rien annoncer
// qu'annoncer faux.
func taskBehind(t Task) int {
	if t.Status != "running" && t.Status != "review" {
		return 0
	}
	if t.WorktreeDir == "" || t.Base == "" {
		return 0
	}
	n, err := CountCommits(t.WorktreeDir, t.Branch, t.Base)
	if err != nil {
		return 0
	}
	return n
}

// pendingCommits compte ce qu'il reste réellement à livrer sur un dépôt : les
// commits non poussés dans les modes de branche ("pr", "push"), ceux pas encore
// fusionnés dans la branche de destination dans les modes de fusion ("merge",
// "merge-push"). Repli sur total (tout est à livrer) quand la révision de
// comparaison n'existe pas encore (branche jamais poussée).
func pendingCommits(project Project, b CardBranch, total int) int {
	from := "origin/" + b.Branch
	if deliveryModeMerges(project.Delivery.Mode) {
		from = deliveryTarget(project, b)
	}
	n, err := CountCommits(b.WorktreeDir, from, b.Branch)
	if err != nil {
		return total
	}
	return n
}

// deliveryModeMerges indique si un mode de livraison fusionne dans la branche
// de destination plutôt que de livrer la branche du chantier elle-même.
func deliveryModeMerges(mode string) bool {
	return mode == "merge" || mode == "merge-push"
}

// deliveryTarget résout la branche de destination d'une livraison : le réglage
// du projet, ou la base de la branche de chantier quand il est vide (« branche
// par défaut du dépôt »).
func deliveryTarget(project Project, b CardBranch) string {
	if target := project.Delivery.Target; target != "" {
		return target
	}
	return b.Base
}

// handleCardShip livre un chantier : pour chaque dépôt touché, ce que le mode
// de livraison du projet prescrit (voir Delivery.Mode et shipCardBranch).
// C'est la SEULE action sortante du produit, d'où {"confirm":true}.
//
// Un dépôt en échec n'annule pas les autres : chaque ligne de la réponse porte
// sa propre erreur, et la livraison est rejouable telle quelle.
func (s *Server) handleCardShip(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Confirm bool `json:"confirm"`
	}
	if err := decodeJSON(r, &body); err != nil || !body.Confirm {
		writeError(w, http.StatusBadRequest, "confirmation required")
		return
	}
	card, project, ok := s.cardWithProject(w, r.PathValue("id"))
	if !ok {
		return
	}
	if !card.ShipReady {
		writeError(w, http.StatusConflict, shipBlockerMessage(card.ShipBlocker))
		return
	}

	results := make([]ShipRepoResult, 0, len(card.Branches))
	for _, b := range card.Branches {
		results = append(results, s.shipCardBranch(card, project, b))
	}
	card, _ = s.store.GetCard(card.ID)
	s.runner.publishCards(card.ProjectID)
	writeJSON(w, http.StatusOK, ShipResponse{Card: card, Mode: project.Delivery.Mode, Repos: results})
}

// shipCardBranch livre la branche de chantier d'un seul dépôt et retourne son
// résultat (jamais d'erreur remontée : elle est portée par le résultat).
func (s *Server) shipCardBranch(card Card, project Project, b CardBranch) ShipRepoResult {
	res := ShipRepoResult{RepoName: b.RepoName, Branch: b.Branch, Base: b.Base}

	commits, err := Commits(b.WorktreeDir, b.Base)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	// Rien à livrer : aucun commit, ou déjà tout poussé/fusionné.
	if len(commits) == 0 || pendingCommits(project, b, len(commits)) == 0 {
		res.Skipped = true
		return res
	}

	if deliveryModeMerges(project.Delivery.Mode) {
		repo, ok := repoByName(project, b.RepoName)
		if !ok {
			res.Error = "repository " + b.RepoName + " is no longer part of the project"
			return res
		}
		target := deliveryTarget(project, b)
		merge := MergeLocal
		if project.Delivery.Mode == "merge-push" {
			merge = MergeAndPush
		}
		out, err := merge(repo.Path, s.dataDir, card.ID, b.RepoName, target, b.Branch)
		res.Output = out
		if err != nil {
			res.Error = err.Error()
			return res
		}
		res.Merged = true
		res.Pushed = project.Delivery.Mode == "merge-push"
		if _, err := s.store.MarkCardBranchShipped(card.ID, b.RepoName, "", time.Now().UTC()); err != nil {
			res.Error = err.Error()
		}
		return res
	}

	out, err := Ship(b.WorktreeDir, b.Branch, card.Title)
	res.Output = out
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.Pushed = true
	// La branche est poussée, la livraison a eu lieu. L'ouverture de la pull
	// request est un bonus : sur GitHub et GitLab elle aboutit toujours (repli
	// sur une URL pré-remplie), et sur une forge inconnue son absence n'est pas
	// une erreur (l'avertissement de santé du projet le dit déjà).
	//
	// Une pull request déjà ouverte est réutilisée telle quelle : le push
	// suffit à la mettre à jour, il n'y a rien de nouveau à ouvrir. Le mode
	// "push" s'arrête ici : pousser sans ouvrir de pull request est justement
	// ce qu'il promet.
	if b.PrURL != "" {
		res.PrURL = b.PrURL
	} else if project.Delivery.Mode == "pr" {
		if prURL, err := OpenPR(b.WorktreeDir, b.Branch, b.Base, card.Title); err == nil {
			res.PrURL = prURL
		}
	}
	if _, err := s.store.MarkCardBranchShipped(card.ID, b.RepoName, res.PrURL, time.Now().UTC()); err != nil {
		res.Error = err.Error()
	}
	return res
}

// handleCardCatchUp rattrape la branche de destination dans la branche du
// chantier : `git merge <destination>` dans le worktree du chantier, un dépôt
// après l'autre. C'est ce qui débloque la livraison quand la destination a
// avancé de son côté : la branche du chantier contient alors la destination, et
// la fusion redevient possible en fast-forward.
//
// Une fusion, pas un rebase : les branches des tâches déjà acceptées descendent
// de la branche du chantier, réécrire son historique les laisserait orphelines.
// Aucun réseau (la destination est une branche locale), aucune confirmation
// (rien ne sort de la machine), et un conflit annule tout (worktree intact).
func (s *Server) handleCardCatchUp(w http.ResponseWriter, r *http.Request) {
	card, project, ok := s.cardWithProject(w, r.PathValue("id"))
	if !ok {
		return
	}

	results := make([]CatchUpRepoResult, 0, len(card.Branches))
	for _, b := range card.Branches {
		results = append(results, s.catchUpCardBranch(card, project, b))
	}
	card, _ = s.store.GetCard(card.ID)
	s.runner.publishCards(card.ProjectID)
	writeJSON(w, http.StatusOK, CatchUpResponse{Card: card, Repos: results})
}

// catchUpCardBranch fusionne la destination dans la branche de chantier d'un
// dépôt. Jamais d'erreur remontée : elle est portée par le résultat, pour qu'un
// dépôt en échec n'empêche pas les autres d'être rattrapés.
func (s *Server) catchUpCardBranch(card Card, project Project, b CardBranch) CatchUpRepoResult {
	target := deliveryTarget(project, b)
	res := CatchUpRepoResult{RepoName: b.RepoName, Target: target}

	if IsBranchMergedInto(b.WorktreeDir, target, b.Branch) {
		res.UpToDate = true
		return res
	}
	if !IsWorktreeClean(b.WorktreeDir) {
		res.Error = "workstream worktree has uncommitted changes"
		return res
	}

	out, conflicts, err := MergeBranch(b.WorktreeDir, target, "Sillage: catch up with "+target)
	res.Output = out
	if err != nil {
		if len(conflicts) > 0 {
			res.ConflictFilePaths = strings.Join(conflicts, " ")
		}
		res.Error = err.Error()
		return res
	}
	res.Merged = true
	// La branche du chantier a de nouveaux commits : un chantier déjà livré ne
	// l'est plus tout à fait (même règle qu'à l'acceptation).
	if _, err := s.store.MarkCardBranchPending(card.ID, b.RepoName); err != nil {
		res.Error = err.Error()
	}
	return res
}

// repoByName retrouve un dépôt du projet par son nom court.
func repoByName(project Project, name string) (Repo, bool) {
	for _, repo := range project.Repos {
		if repo.Name == name {
			return repo, true
		}
	}
	return Repo{}, false
}

// shipBlockerMessage traduit un blocage de livraison en message d'API.
func shipBlockerMessage(blocker string) string {
	switch blocker {
	case "no-tasks":
		return "workstream has no task to ship"
	case "nothing-accepted":
		return "workstream has no accepted task"
	case "nothing-to-ship":
		return "workstream has no branch to ship"
	}
	return "workstream cannot be shipped"
}

// handleCancel refuse (ou annule) une tâche : autorisé depuis running/review ;
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
