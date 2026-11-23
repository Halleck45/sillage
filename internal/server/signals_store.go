package server

// Persistance des signaux de la Vigie et du réglage de surveillance des
// projets. Tout passe par le verrou du Store, comme le reste de l'état.

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// signalTerminal : un signal classé ou transformé en chantier ne réapparaît
// plus dans l'écran de la Vigie, mais reste en mémoire pour le dédoublonnage.
func signalTerminal(status string) bool {
	return status == "dismissed" || status == "handled" || status == "closed"
}

// activeSignals est ce que l'écran de la Vigie montre : tout sauf les signaux
// classés, du plus récent au plus ancien.
func activeSignals(m map[string]Signal) []Signal {
	out := make([]Signal, 0, len(m))
	for _, sig := range m {
		// Les signaux fermés sur la forge restent exposés : le kanban les
		// montre dans « Livré » jusqu'à ce qu'on les archive (dismissed).
		if sig.Status != "dismissed" && sig.Status != "handled" {
			out = append(out, sig)
		}
	}
	sort.Slice(out, func(i, j int) bool { return idNum(out[i].ID) > idNum(out[j].ID) })
	return out
}

// resetTransientSignals : une analyse en cours au moment de l'arrêt n'aboutira
// jamais. Le signal repasse en file, le veilleur la relancera.
func resetTransientSignals(s *Store) {
	for id, sig := range s.Signals {
		if sig.Status == "analyzing" {
			sig.Status = "queued"
			s.Signals[id] = sig
		}
	}
}

// SetProjectWatch active ou coupe la Vigie d'un projet. Since n'est posé qu'au
// passage de coupé à actif : changer d'agent ne doit pas faire oublier ce qui
// a déjà été vu, ni relâcher ce qui ne l'a pas été.
func (s *Store) SetProjectWatch(id string, enabled bool, agentID string) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.Projects[id]
	if !ok {
		return Project{}, fmt.Errorf("project not found")
	}
	if enabled {
		a, ok := s.Agents[agentID]
		if !ok {
			return Project{}, fmt.Errorf("agent not found")
		}
		if !canPlan(a.Cli) {
			return Project{}, fmt.Errorf("this agent cannot analyze")
		}
		if !p.Watch.Enabled {
			p.Watch.Since = time.Now().UTC()
		}
		p.Watch.AgentID = agentID
	}
	p.Watch.Enabled = enabled
	s.Projects[id] = p
	if err := s.save(); err != nil {
		return Project{}, err
	}
	return p, nil
}

// WatchedProjects liste les projets dont la Vigie est active.
func (s *Store) WatchedProjects() []Project {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Project
	for _, p := range s.Projects {
		if p.Watch.Enabled {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return idNum(out[i].ID) < idNum(out[j].ID) })
	return out
}

// AddSignalIfNew enregistre sig s'il n'existe aucun signal pour le même
// élément de forge (projet, dépôt, type, numéro). Retourne false sinon : une
// PR vue une fois, même classée, ne remonte plus.
func (s *Store) AddSignalIfNew(sig Signal) (Signal, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.Signals {
		if existing.ProjectID == sig.ProjectID && existing.RepoName == sig.RepoName &&
			existing.Kind == sig.Kind && existing.Number == sig.Number {
			return existing, false, nil
		}
	}
	s.NextSignalN++
	sig.ID = fmt.Sprintf("s%d", s.NextSignalN)
	if sig.DetectedAt.IsZero() {
		sig.DetectedAt = time.Now().UTC()
	}
	if sig.Status == "" {
		sig.Status = "queued"
	}
	s.Signals[sig.ID] = sig
	if err := s.save(); err != nil {
		return Signal{}, false, err
	}
	return sig, true, nil
}

// AddUpkeepSignalIfNew enregistre un constat d'entretien, sauf s'il existe
// déjà un signal d'entretien de même titre pour ce dépôt, quel que soit son
// statut : un constat classé ne revient pas chaque semaine.
func (s *Store) AddUpkeepSignalIfNew(sig Signal) (Signal, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.Signals {
		if existing.ProjectID == sig.ProjectID && existing.RepoName == sig.RepoName &&
			existing.Kind == "upkeep" && strings.EqualFold(strings.TrimSpace(existing.Title), strings.TrimSpace(sig.Title)) {
			return existing, false, nil
		}
	}
	s.NextSignalN++
	sig.ID = fmt.Sprintf("s%d", s.NextSignalN)
	sig.DetectedAt = time.Now().UTC()
	s.Signals[sig.ID] = sig
	if err := s.save(); err != nil {
		return Signal{}, false, err
	}
	return sig, true, nil
}

// SetUpkeepAt note la date de la dernière veille d'entretien.
func (s *Store) SetUpkeepAt(projectID string, at time.Time) error {
	return s.setUpkeepAt(projectID, &at)
}

// ResetUpkeepAt redemande la veille à la prochaine relève.
func (s *Store) ResetUpkeepAt(projectID string) error {
	return s.setUpkeepAt(projectID, nil)
}

func (s *Store) setUpkeepAt(projectID string, at *time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.Projects[projectID]
	if !ok {
		return fmt.Errorf("project not found")
	}
	p.Watch.UpkeepAt = at
	s.Projects[projectID] = p
	return s.save()
}

// FindSignal retrouve le signal d'un élément de forge.
func (s *Store) FindSignal(projectID, repoName, kind string, number int) (Signal, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.Signals {
		if existing.ProjectID == projectID && existing.RepoName == repoName &&
			existing.Kind == kind && existing.Number == number {
			return existing, true
		}
	}
	return Signal{}, false
}

// OpenSignals : les signaux encore suivis (ni classés, ni devenus chantier, ni
// fermés) d'un dépôt et d'un type, pour constater ceux qui ont quitté la liste
// des éléments ouverts de la forge.
func (s *Store) OpenSignals(projectID, repoName, kind string) []Signal {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Signal
	for _, sig := range s.Signals {
		if sig.ProjectID == projectID && sig.RepoName == repoName && sig.Kind == kind && !signalTerminal(sig.Status) {
			out = append(out, sig)
		}
	}
	return out
}

func (s *Store) GetSignal(id string) (Signal, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sig, ok := s.Signals[id]
	return sig, ok
}

func (s *Store) UpdateSignal(id string, fn func(sig *Signal)) (Signal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sig, ok := s.Signals[id]
	if !ok {
		return Signal{}, fmt.Errorf("signal not found")
	}
	fn(&sig)
	s.Signals[id] = sig
	if err := s.save(); err != nil {
		return Signal{}, err
	}
	return sig, nil
}

// QueuedSignals : les signaux qui attendent leur analyse, dans l'ordre
// d'arrivée.
func (s *Store) QueuedSignals() []Signal {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []Signal
	for _, sig := range s.Signals {
		if sig.Status == "queued" {
			out = append(out, sig)
		}
	}
	sort.Slice(out, func(i, j int) bool { return idNum(out[i].ID) < idNum(out[j].ID) })
	return out
}
