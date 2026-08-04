package server

import (
	"net/http"
	"strconv"
)

// --- Recette manuelle (voir docs/SPEC-RECETTE.md) ---
//
// Lancer ou arrêter une recette est local et réversible : ces mutations
// n'exigent donc pas {"confirm": true}, contrairement aux actions sortantes
// (livraison d'un chantier, synchronisation de l'espace de travail).

// handleCardPreview lance la recette d'un chantier dans le worktree de sa
// branche, pour un dépôt donné. repoName est optionnel quand le chantier ne
// touche qu'un seul dépôt : c'est le cas courant, autant ne pas le demander.
func (s *Server) handleCardPreview(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RepoName string `json:"repoName"`
	}
	_ = decodeJSON(r, &body)

	card, ok := s.store.GetCard(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "card not found")
		return
	}
	project, ok := s.store.GetProject(card.ProjectID)
	if !ok {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	if len(card.Branches) == 0 {
		writeError(w, http.StatusBadRequest, "no workstream branch yet: create a task first")
		return
	}
	branch, ok := cardBranchByRepo(card, body.RepoName)
	if !ok {
		writeError(w, http.StatusBadRequest, "repoName is required")
		return
	}
	repo, ok := repoByName(project, branch.RepoName)
	if !ok {
		writeError(w, http.StatusNotFound, "repository not found")
		return
	}

	run, err := s.previews.Start(previewTarget{
		projectID: project.ID,
		cardID:    card.ID,
		repoName:  branch.RepoName,
		dir:       branch.WorktreeDir,
		branch:    branch.Branch,
		cmd:       repo.PreviewCmd,
		url:       repo.PreviewURL,
		id:        "ws-" + strconv.Itoa(card.Ref),
		n:         card.Ref,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, run)
}

// handleTaskPreview lance la recette dans le worktree d'une tâche : on éprouve
// un incrément avant de l'accepter. Même mécanique que le chantier, autre
// worktree et autre identité (t-<ref> au lieu de ws-<ref>).
func (s *Server) handleTaskPreview(w http.ResponseWriter, r *http.Request) {
	task, ok := s.store.GetTask(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	project, ok := s.store.GetProject(task.ProjectID)
	if !ok {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	repo, ok := repoByName(project, task.RepoName)
	if !ok {
		writeError(w, http.StatusNotFound, "repository not found")
		return
	}

	run, err := s.previews.Start(previewTarget{
		projectID: project.ID,
		cardID:    task.CardID,
		taskID:    task.ID,
		repoName:  task.RepoName,
		dir:       task.WorktreeDir,
		branch:    task.Branch,
		cmd:       repo.PreviewCmd,
		url:       repo.PreviewURL,
		id:        "t-" + strconv.Itoa(task.Ref),
		n:         task.Ref,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *Server) handlePreviewStop(w http.ResponseWriter, r *http.Request) {
	if err := s.previews.Stop(r.PathValue("id")); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handlePreviewLog rend le tampon du journal à l'ouverture du panneau de
// recette ; la suite arrive en SSE (événement previewLog).
func (s *Server) handlePreviewLog(w http.ResponseWriter, r *http.Request) {
	lines, ok := s.previews.Log(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "preview run not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runId": r.PathValue("id"), "lines": lines})
}

// cardBranchByRepo retrouve la branche de chantier d'un dépôt. Sans nom de
// dépôt, la branche unique du chantier fait office de réponse évidente ; avec
// plusieurs branches, il faut choisir.
func cardBranchByRepo(card Card, repoName string) (CardBranch, bool) {
	if repoName == "" {
		if len(card.Branches) == 1 {
			return card.Branches[0], true
		}
		return CardBranch{}, false
	}
	for _, b := range card.Branches {
		if b.RepoName == repoName {
			return b, true
		}
	}
	return CardBranch{}, false
}
