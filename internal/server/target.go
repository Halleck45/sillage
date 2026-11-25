package server

import (
	"net/http"
	"sync"
	"time"
)

// Retard de la branche de destination locale sur origin. Un chantier part de
// la version la plus fraîche (freshestStart), mais le main local de l'humain
// reste en retard tant que personne ne l'avance : l'accueil du projet le dit,
// et propose de l'avancer en fast-forward.

// targetFetchInterval espace les fetch d'un même dépôt : l'accueil d'un projet
// se rouvre souvent, le remote ne bouge pas à ce rythme.
const targetFetchInterval = 5 * time.Minute

// TargetStatus est l'état d'un dépôt vis-à-vis de sa branche de destination.
type TargetStatus struct {
	RepoName string `json:"repoName"`
	Target   string `json:"target"`
	Behind   int    `json:"behind"`
	Ahead    int    `json:"ahead"`
	Error    string `json:"error,omitempty"`

	// Provider et WebURL décrivent la page du dépôt sur sa forge (github|gitlab),
	// déduites du remote origin sans réseau. Vides pour une forge inconnue.
	Provider string `json:"provider,omitempty"`
	WebURL   string `json:"webUrl,omitempty"`
}

// targetFetches retient le dernier fetch de chaque dépôt (en mémoire seulement).
type targetFetches struct {
	mu   sync.Mutex
	last map[string]time.Time
}

// due réserve le fetch de repoPath s'il n'a pas eu lieu depuis
// targetFetchInterval. Réserver avant de fetcher évite deux fetch simultanés
// quand deux onglets ouvrent le même projet.
func (f *targetFetches) due(repoPath string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if time.Since(f.last[repoPath]) < targetFetchInterval {
		return false
	}
	if f.last == nil {
		f.last = map[string]time.Time{}
	}
	f.last[repoPath] = time.Now()
	return true
}

// repoTarget est la branche de destination d'un dépôt : celle du réglage de
// livraison, sinon sa branche courante (la même règle que CreateCardWorktree).
func repoTarget(project Project, repo Repo) (string, error) {
	if project.Delivery.Target != "" {
		return project.Delivery.Target, nil
	}
	return currentBranch(repo.Path)
}

// repoWebURL donne la page web du dépôt sur sa forge, déduite du remote origin
// (lecture de la config git, aucun réseau). Forge inconnue : deux chaînes vides.
func repoWebURL(dir string) (provider, webURL string) {
	info := DetectForge(dir)
	if info.Provider == "" {
		return "", ""
	}
	return info.Provider, "https://" + info.Host + "/" + info.Path
}

// targetStatuses calcule l'état de chaque dépôt du projet, après un fetch si
// fetch est vrai et que le dernier date de plus de targetFetchInterval. Un
// fetch en échec (hors ligne, remote qui demande des identifiants) n'empêche
// pas le calcul : les refs locales restent la meilleure information.
func (s *Server) targetStatuses(project Project, fetch bool) []TargetStatus {
	out := make([]TargetStatus, 0, len(project.Repos))
	for _, repo := range project.Repos {
		provider, webURL := repoWebURL(repo.Path)
		target, err := repoTarget(project, repo)
		if err != nil {
			out = append(out, TargetStatus{RepoName: repo.Name, Error: err.Error(), Provider: provider, WebURL: webURL})
			continue
		}
		st := TargetStatus{RepoName: repo.Name, Target: target, Provider: provider, WebURL: webURL}
		if fetch && s.targetFetches.due(repo.Path) {
			if err := FetchTarget(repo.Path, target); err != nil {
				st.Error = err.Error()
			}
		}
		st.Behind, st.Ahead, _ = TargetDrift(repo.Path, target)
		out = append(out, st)
	}
	return out
}

// handleTargetStatus : GET /api/projects/{id}/target-status. Fetch au plus une
// fois toutes les cinq minutes par dépôt (lecture réseau, rien n'est envoyé).
func (s *Server) handleTargetStatus(w http.ResponseWriter, r *http.Request) {
	project, ok := s.store.GetProject(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"repos": s.targetStatuses(project, true)})
}

// handleTargetUpdate : POST /api/projects/{id}/target-status/update. Avance en
// fast-forward chaque branche de destination en retard et sans commit local.
// Local (aucun réseau, rien n'est poussé), donc sans confirmation ; un dépôt
// divergent ou qu'un fast-forward abîmerait garde son erreur et ne bloque pas
// les autres.
func (s *Server) handleTargetUpdate(w http.ResponseWriter, r *http.Request) {
	project, ok := s.store.GetProject(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	statuses := s.targetStatuses(project, false)
	for i, st := range statuses {
		if st.Error != "" || st.Behind == 0 {
			continue
		}
		if st.Ahead > 0 {
			statuses[i].Error = "local " + st.Target + " has diverged from origin/" + st.Target
			continue
		}
		if err := FastForwardTarget(project.Repos[i].Path, st.Target); err != nil {
			statuses[i].Error = err.Error()
			continue
		}
		statuses[i].Behind, statuses[i].Ahead, _ = TargetDrift(project.Repos[i].Path, st.Target)
	}
	writeJSON(w, http.StatusOK, map[string]any{"repos": statuses})
}
