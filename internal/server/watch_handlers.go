package server

// Routes de la Vigie. Aucune n'écrit sur la forge : activer la surveillance,
// demander une relève et classer un signal sont des actions locales.

import (
	"net/http"
	"sort"
)

// defaultAnalystAgent choisit l'analyste quand l'UI n'en impose pas : un vrai
// CLI capable de lecture seule (claude, puis codex), l'agent simulé en dernier
// recours.
func (s *Server) defaultAnalystAgent() string {
	agents := s.store.ListAgents()
	sort.Slice(agents, func(i, j int) bool { return idNum(agents[i].ID) < idNum(agents[j].ID) })
	for _, cli := range []string{"claude", "codex", "fake"} {
		for _, a := range agents {
			if a.Cli == cli && a.Warning == "" {
				return a.ID
			}
		}
	}
	return ""
}

func (s *Server) handleProjectWatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled bool   `json:"enabled"`
		AgentID string `json:"agentId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := r.PathValue("id")
	if body.Enabled && body.AgentID == "" {
		if p, ok := s.store.GetProject(id); ok && p.Watch.AgentID != "" {
			body.AgentID = p.Watch.AgentID
		} else {
			body.AgentID = s.defaultAnalystAgent()
		}
		if body.AgentID == "" {
			writeError(w, http.StatusBadRequest, "no agent can analyze")
			return
		}
	}
	project, err := s.store.SetProjectWatch(id, body.Enabled, body.AgentID)
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	out := projectOut(project)
	s.runner.publishProject(out)
	if project.Watch.Enabled {
		s.kickWatcher()
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleWatchCheck(w http.ResponseWriter, r *http.Request) {
	s.kickWatcher()
	w.WriteHeader(http.StatusAccepted)
}

// handleUpdateSignal change le statut d'un signal : le classer ("dismissed"),
// le marquer traité par un chantier ("handled" + cardId), ou relancer son
// analyse ("queued").
func (s *Server) handleUpdateSignal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
		CardID string `json:"cardId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	switch body.Status {
	case "dismissed", "queued":
	case "handled":
		if _, ok := s.store.GetCard(body.CardID); !ok {
			writeError(w, http.StatusBadRequest, "card not found")
			return
		}
	default:
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	current, ok := s.store.GetSignal(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "signal not found")
		return
	}
	if body.Status == "queued" && current.Status == "analyzing" {
		writeError(w, http.StatusConflict, "analysis already running")
		return
	}
	if body.Status == "dismissed" {
		// Classer un signal en cours de relecture arrête l'agent : il ne
		// travaillerait plus pour personne.
		s.watch.stopAnalysis(current.ID)
	}
	sig, err := s.store.UpdateSignal(current.ID, func(x *Signal) {
		x.Status = body.Status
		if body.Status == "handled" {
			x.CardID = body.CardID
		}
		if body.Status == "queued" {
			x.Error = ""
		}
	})
	if err != nil {
		writeError(w, statusForStoreError(err), err.Error())
		return
	}
	s.publishSignal(sig)
	if sig.Status == "queued" {
		s.kickWatcher()
	}
	writeJSON(w, http.StatusOK, sig)
}

// handleRefreshSignal relit tout de suite l'état d'une PR ou d'une issue sur la
// forge (lecture seule), sans attendre la prochaine relève : l'interface
// l'appelle au retour d'un passage sur GitHub, où l'on vient peut-être de
// merger. Rend le signal, fermé s'il l'est sur la forge.
func (s *Server) handleRefreshSignal(w http.ResponseWriter, r *http.Request) {
	sig, ok := s.store.GetSignal(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "signal not found")
		return
	}
	if signalTerminal(sig.Status) || sig.Number == 0 || (sig.Kind != "pr" && sig.Kind != "issue") {
		writeJSON(w, http.StatusOK, sig)
		return
	}
	project, ok := s.store.GetProject(sig.ProjectID)
	if !ok {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	repo, ok := repoByName(project, sig.RepoName)
	if !ok {
		writeError(w, http.StatusNotFound, "repo not found")
		return
	}
	sig, _ = s.closeIfGone(r.Context(), s.watch.stateTool(), repo.Path, sig)
	writeJSON(w, http.StatusOK, sig)
}

// handleStopSignal arrête la relecture en cours d'un signal. Le signal passe en
// "error" avec l'erreur "stopped", et se relance comme une analyse échouée.
func (s *Server) handleStopSignal(w http.ResponseWriter, r *http.Request) {
	if !s.watch.stopAnalysis(r.PathValue("id")) {
		writeError(w, http.StatusConflict, "no analysis is running for this signal")
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
