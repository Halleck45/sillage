package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Store est l'état en mémoire de l'application, persisté dans state.json.
// Tous les champs exportés sont sérialisés ; les champs non exportés
// (verrou, répertoire de données) sont ignorés par encoding/json.
type Store struct {
	mu      sync.Mutex
	dataDir string

	Projects map[string]Project
	Cards    map[string]Card
	Tasks    map[string]Task
	Messages map[string][]Message
	Agents   map[string]Agent

	NextProjectN int
	NextCardN    int
	NextTaskN    int
	NextMessageN int
	NextRef      int
}

// NewStore charge state.json s'il existe, sinon initialise un état neuf
// avec les agents seedés.
func NewStore(dataDir string) (*Store, error) {
	s := &Store{dataDir: dataDir}
	data, err := os.ReadFile(s.statePath())
	if err == nil {
		if err := json.Unmarshal(data, s); err != nil {
			return nil, fmt.Errorf("lecture de state.json impossible : %w", err)
		}
		s.ensureMaps()
		return s, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	s.initEmpty()
	if err := s.save(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) statePath() string {
	return filepath.Join(s.dataDir, "state.json")
}

func (s *Store) ensureMaps() {
	if s.Projects == nil {
		s.Projects = map[string]Project{}
	}
	if s.Cards == nil {
		s.Cards = map[string]Card{}
	}
	if s.Tasks == nil {
		s.Tasks = map[string]Task{}
	}
	if s.Messages == nil {
		s.Messages = map[string][]Message{}
	}
	if s.Agents == nil {
		s.Agents = map[string]Agent{}
	}
}

func (s *Store) initEmpty() {
	s.ensureMaps()
	s.NextRef = 100
	s.Agents["bolt"] = Agent{
		ID: "bolt", Name: "Bolt", Emoji: "🐝", Color: "#f2b705",
		Model: "claude-sonnet-5", Cli: "claude",
		ContextPrompt: "Tu es un développeur backend pragmatique, orienté qualité et simplicité.",
	}
	s.Agents["muse"] = Agent{
		ID: "muse", Name: "Muse", Emoji: "🦊", Color: "#d0662f",
		Model: "claude-opus-5", Cli: "claude",
		ContextPrompt: "Tu es responsable produit : specs, documentation, clarté fonctionnelle.",
	}
	s.Agents["otto"] = Agent{
		ID: "otto", Name: "Otto", Emoji: "🦉", Color: "#4f7d2f",
		Model: "", Cli: "codex",
		ContextPrompt: "Tu es un ingénieur infra : CI, déploiement, outillage.",
	}
	s.Agents["echo"] = Agent{
		ID: "echo", Name: "Écho", Emoji: "🧪", Color: "#777777",
		Model: "", Cli: "fake",
		ContextPrompt: "Agent de test local, gratuit, pour la démo et les vérifications.",
	}
}

// save écrit l'état sur disque de façon atomique (fichier temp + rename).
// Doit être appelée avec le verrou déjà tenu.
func (s *Store) save() error {
	if err := os.MkdirAll(s.dataDir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(s.dataDir, "state-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, s.statePath())
}

// idNum extrait la partie numérique d'un identifiant ("t12" -> 12) pour un tri stable.
func idNum(id string) int {
	i := 0
	for i < len(id) && (id[i] < '0' || id[i] > '9') {
		i++
	}
	n, _ := strconv.Atoi(id[i:])
	return n
}

// recomputeCard recalcule les compteurs dérivés d'une carte à partir des tâches.
// Doit être appelée avec le verrou tenu.
func (s *Store) recomputeCard(cardID string) {
	c, ok := s.Cards[cardID]
	if !ok {
		return
	}
	var total, done, review, docs, msgs int
	var live *string
	for _, t := range s.Tasks {
		if t.CardID != cardID {
			continue
		}
		total++
		switch t.Status {
		case "shipped":
			done++
		case "review":
			review++
		}
		docs += t.DocsCount
		msgs += t.MessagesCount
		if live == nil && t.Status == "running" && t.LiveActivity != nil {
			v := *t.LiveActivity
			live = &v
		}
	}
	progress := 0
	if total > 0 {
		progress = done * 100 / total
	}
	c.TasksTotal, c.TasksDone, c.ReviewCount = total, done, review
	c.DocsCount, c.MessagesCount, c.Progress = docs, msgs, progress
	c.LiveActivity = live
	s.Cards[cardID] = c
}

// recomputeProject recalcule le nombre de tâches non lues et le total de tokens du projet.
func (s *Store) recomputeProject(projectID string) {
	p, ok := s.Projects[projectID]
	if !ok {
		return
	}
	var unread int
	tok := Tokens{}
	for _, t := range s.Tasks {
		if t.ProjectID != projectID {
			continue
		}
		if t.Unread {
			unread++
		}
		tok.Input += t.Tokens.Input
		tok.Output += t.Tokens.Output
		tok.CostUsd += t.Tokens.CostUsd
	}
	p.Unread = unread
	p.Tokens = tok
	s.Projects[projectID] = p
}

// recomputeAgent recalcule le flag "active" d'un agent (tâche en cours d'exécution).
func (s *Store) recomputeAgent(agentID string) {
	a, ok := s.Agents[agentID]
	if !ok {
		return
	}
	active := false
	for _, t := range s.Tasks {
		if t.AgentID == agentID && t.Status == "running" {
			active = true
			break
		}
	}
	a.Active = active
	s.Agents[agentID] = a
}

// recomputeAll recalcule tous les compteurs dérivés (cartes, projets, agents).
func (s *Store) recomputeAll() {
	for id := range s.Cards {
		s.recomputeCard(id)
	}
	for id := range s.Projects {
		s.recomputeProject(id)
	}
	for id := range s.Agents {
		s.recomputeAgent(id)
	}
}

func sortedProjects(m map[string]Project) []Project {
	out := make([]Project, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return idNum(out[i].ID) < idNum(out[j].ID) })
	return out
}

func sortedCards(m map[string]Card) []Card {
	out := make([]Card, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return idNum(out[i].ID) < idNum(out[j].ID) })
	return out
}

func sortedTasks(m map[string]Task) []Task {
	out := make([]Task, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return idNum(out[i].ID) < idNum(out[j].ID) })
	return out
}

func sortedAgents(m map[string]Agent) []Agent {
	out := make([]Agent, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Snapshot recalcule les compteurs dérivés et retourne l'état complet pour GET /api/state.
func (s *Store) Snapshot() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recomputeAll()

	global := Tokens{}
	for _, t := range s.Tasks {
		global.Input += t.Tokens.Input
		global.Output += t.Tokens.Output
		global.CostUsd += t.Tokens.CostUsd
	}

	st := State{
		Projects: sortedProjects(s.Projects),
		Cards:    sortedCards(s.Cards),
		Tasks:    sortedTasks(s.Tasks),
		Agents:   sortedAgents(s.Agents),
	}
	st.Tokens.Global = global
	return st
}

// TokensSnapshot construit le contenu de l'événement SSE "tokens".
func (s *Store) TokensSnapshot() TokensEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recomputeAll()

	ev := TokensEvent{Projects: map[string]Tokens{}, Tasks: map[string]Tokens{}}
	for id, p := range s.Projects {
		ev.Projects[id] = p.Tokens
		ev.Global.Input += p.Tokens.Input
		ev.Global.Output += p.Tokens.Output
		ev.Global.CostUsd += p.Tokens.CostUsd
	}
	for id, t := range s.Tasks {
		ev.Tasks[id] = t.Tokens
	}
	return ev
}

// ListAgents retourne les agents (compteur "active" recalculé), triés par ID.
func (s *Store) ListAgents() []Agent {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recomputeAll()
	return sortedAgents(s.Agents)
}

// CardsByProject retourne les cartes d'un projet, recalculées.
func (s *Store) CardsByProject(projectID string) []Card {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recomputeAll()
	var out []Card
	for _, c := range s.Cards {
		if c.ProjectID == projectID {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return idNum(out[i].ID) < idNum(out[j].ID) })
	if out == nil {
		out = []Card{}
	}
	return out
}

func (s *Store) GetProject(id string) (Project, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.Projects[id]
	return p, ok
}

func (s *Store) GetCard(id string) (Card, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.Cards[id]
	return c, ok
}

func (s *Store) GetAgent(id string) (Agent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.Agents[id]
	return a, ok
}

func (s *Store) GetTask(id string) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.Tasks[id]
	return t, ok
}

func (s *Store) GetMessages(taskID string) []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	msgs := s.Messages[taskID]
	out := make([]Message, len(msgs))
	copy(out, msgs)
	return out
}

// AddProject crée un projet. La validation du chemin (existence, dépôt git)
// est faite par l'appelant (handlers.go) avant l'appel.
func (s *Store) AddProject(name, path string) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.NextProjectN++
	p := Project{ID: fmt.Sprintf("p%d", s.NextProjectN), Name: name, Path: path}
	s.Projects[p.ID] = p
	if err := s.save(); err != nil {
		return Project{}, err
	}
	return p, nil
}

var validColumns = map[string]bool{"soon": true, "doing": true, "done": true}

// AddCard crée une carte dans un projet. column vide => "soon".
func (s *Store) AddCard(projectID, title, column string) (Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Projects[projectID]; !ok {
		return Card{}, fmt.Errorf("projet %s introuvable", projectID)
	}
	if column == "" {
		column = "soon"
	}
	if !validColumns[column] {
		return Card{}, fmt.Errorf("colonne invalide : %s", column)
	}
	s.NextCardN++
	c := Card{ID: fmt.Sprintf("c%d", s.NextCardN), ProjectID: projectID, Column: column, Title: title}
	s.Cards[c.ID] = c
	s.recomputeCard(c.ID)
	if err := s.save(); err != nil {
		return Card{}, err
	}
	return s.Cards[c.ID], nil
}

// UpdateCardColumn déplace une carte vers une autre colonne du kanban.
func (s *Store) UpdateCardColumn(id, column string) (Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.Cards[id]
	if !ok {
		return Card{}, fmt.Errorf("carte %s introuvable", id)
	}
	if !validColumns[column] {
		return Card{}, fmt.Errorf("colonne invalide : %s", column)
	}
	c.Column = column
	s.Cards[id] = c
	if err := s.save(); err != nil {
		return Card{}, err
	}
	return c, nil
}

// ReserveTaskID réserve un identifiant de tâche et une référence globale.
func (s *Store) ReserveTaskID() (id string, ref int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.NextTaskN++
	s.NextRef++
	return fmt.Sprintf("t%d", s.NextTaskN), s.NextRef
}

// CreateTask enregistre une nouvelle tâche (statut initial "running").
func (s *Store) CreateTask(id string, ref int, cardID, projectID, title, agentID, branch, base, worktreeDir string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Task{
		ID: id, CardID: cardID, ProjectID: projectID, Ref: ref, Title: title,
		AgentID: agentID, Branch: branch, Status: "running", Checks: []Check{},
		Unread: false, UpdatedAt: time.Now().UTC(), Tokens: Tokens{},
		Base: base, WorktreeDir: worktreeDir,
	}
	s.Tasks[id] = t
	s.recomputeCard(cardID)
	s.recomputeProject(projectID)
	s.recomputeAgent(agentID)
	if err := s.save(); err != nil {
		return Task{}, err
	}
	return s.Tasks[id], nil
}

// UpdateTask applique fn à une copie de la tâche puis persiste le résultat.
// C'est le point d'entrée générique utilisé par les handlers et le runner.
func (s *Store) UpdateTask(id string, fn func(t *Task)) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.Tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("tâche %s introuvable", id)
	}
	fn(&t)
	t.UpdatedAt = time.Now().UTC()
	s.Tasks[id] = t
	s.recomputeCard(t.CardID)
	s.recomputeProject(t.ProjectID)
	s.recomputeAgent(t.AgentID)
	if err := s.save(); err != nil {
		return Task{}, err
	}
	return s.Tasks[id], nil
}

// AddMessage ajoute un message au fil d'une tâche et retourne le message et la tâche mises à jour.
func (s *Store) AddMessage(taskID, author, text string) (Message, Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.Tasks[taskID]
	if !ok {
		return Message{}, Task{}, fmt.Errorf("tâche %s introuvable", taskID)
	}
	s.NextMessageN++
	m := Message{ID: fmt.Sprintf("m%d", s.NextMessageN), TaskID: taskID, Author: author, Text: text, CreatedAt: time.Now().UTC()}
	s.Messages[taskID] = append(s.Messages[taskID], m)
	t.MessagesCount = len(s.Messages[taskID])
	t.UpdatedAt = time.Now().UTC()
	s.Tasks[taskID] = t
	s.recomputeCard(t.CardID)
	if err := s.save(); err != nil {
		return Message{}, Task{}, err
	}
	return m, s.Tasks[taskID], nil
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

var slugAccents = strings.NewReplacer(
	"à", "a", "â", "a", "ä", "a", "é", "e", "è", "e", "ê", "e", "ë", "e",
	"î", "i", "ï", "i", "ô", "o", "ö", "o", "ù", "u", "û", "u", "ü", "u",
	"ç", "c", "œ", "oe", "æ", "ae",
)

// Slugify normalise un titre de tâche en segment de nom de branche.
func Slugify(title string) string {
	s := slugAccents.Replace(strings.ToLower(title))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = strings.Trim(s[:40], "-")
	}
	if s == "" {
		s = "tache"
	}
	return s
}
