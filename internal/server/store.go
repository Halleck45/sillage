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

	"golang.org/x/crypto/bcrypt"
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
	Users    map[string]User

	NextProjectN int
	NextCardN    int
	NextTaskN    int
	NextMessageN int
	NextRef      int
	NextUserN    int
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
	if s.Users == nil {
		s.Users = map[string]User{}
	}
}

func (s *Store) initEmpty() {
	s.ensureMaps()
	s.NextRef = 100
	s.Agents["bolt"] = Agent{
		ID: "bolt", Name: "Bolt", Emoji: "🐝", Color: "#f2b705",
		Model: "claude-sonnet-5", Cli: "claude",
		ContextPrompt: "You are a pragmatic backend developer, focused on quality and simplicity.",
	}
	s.Agents["muse"] = Agent{
		ID: "muse", Name: "Muse", Emoji: "🦊", Color: "#d0662f",
		Model: "claude-opus-5", Cli: "claude",
		ContextPrompt: "You are a product owner: specs, documentation, functional clarity.",
	}
	s.Agents["otto"] = Agent{
		ID: "otto", Name: "Otto", Emoji: "🦉", Color: "#4f7d2f",
		Model: "", Cli: "codex",
		ContextPrompt: "You are an infrastructure engineer: CI, deployment, tooling.",
	}
	s.Agents["echo"] = Agent{
		ID: "echo", Name: "Écho", Emoji: "🧪", Color: "#777777",
		Model: "", Cli: "fake",
		ContextPrompt: "Local test agent, free of charge, for demos and checks.",
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

// --- Utilisateurs ---

var validRole = map[string]bool{"admin": true, "member": true}

// GetUser retourne un utilisateur par son identifiant.
func (s *Store) GetUser(id string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.Users[id]
	return u, ok
}

// FindUserByName retourne un utilisateur par son nom (utilisé par le login).
func (s *Store) FindUserByName(name string) (User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.findUserByNameLocked(name)
}

// findUserByNameLocked suppose le verrou déjà tenu par l'appelant.
func (s *Store) findUserByNameLocked(name string) (User, bool) {
	for _, u := range s.Users {
		if u.Name == name {
			return u, true
		}
	}
	return User{}, false
}

// countAdminsLocked suppose le verrou déjà tenu par l'appelant.
func (s *Store) countAdminsLocked() int {
	n := 0
	for _, u := range s.Users {
		if u.Role == "admin" {
			n++
		}
	}
	return n
}

// ListUsers retourne tous les utilisateurs, triés par identifiant.
func (s *Store) ListUsers() []User {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]User, 0, len(s.Users))
	for _, u := range s.Users {
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return idNum(out[i].ID) < idNum(out[j].ID) })
	return out
}

// MigrateUsers assure la présence d'un compte "admin", à partir du hash de
// mot de passe hérité (config.json existant, ou mot de passe généré au tout
// premier lancement). Si state.json contient déjà des utilisateurs, ne fait
// rien, sauf si SILLAGE_PASSWORD est positionnée dans l'environnement : dans
// ce cas le mot de passe de "admin" est toujours remplacé (compte créé s'il
// est absent). C'est le point d'entrée de compatibilité avec les installations
// antérieures à la v0.2 (state.json sans utilisateurs).
func (s *Store) MigrateUsers(legacyHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureMaps()

	envSet := os.Getenv("SILLAGE_PASSWORD") != ""
	if !envSet && len(s.Users) > 0 {
		return nil
	}

	if admin, ok := s.findUserByNameLocked("admin"); ok {
		admin.PasswordHash = legacyHash
		s.Users[admin.ID] = admin
	} else {
		s.NextUserN++
		id := fmt.Sprintf("u%d", s.NextUserN)
		s.Users[id] = User{ID: id, Name: "admin", Role: "admin", PasswordHash: legacyHash}
	}
	return s.save()
}

// AddUser crée un utilisateur. name non vide et unique ; role par défaut
// "member" ; le mot de passe est haché avant stockage.
func (s *Store) AddUser(name, password, role string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(name) == "" {
		return User{}, fmt.Errorf("name is required")
	}
	if password == "" {
		return User{}, fmt.Errorf("password is required")
	}
	if role == "" {
		role = "member"
	}
	if !validRole[role] {
		return User{}, fmt.Errorf("invalid role: must be admin or member")
	}
	if _, exists := s.findUserByNameLocked(name); exists {
		return User{}, fmt.Errorf("a user with this name already exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	s.NextUserN++
	id := fmt.Sprintf("u%d", s.NextUserN)
	u := User{ID: id, Name: name, Role: role, PasswordHash: string(hash)}
	s.Users[id] = u
	if err := s.save(); err != nil {
		return User{}, err
	}
	return u, nil
}

// UpdateUser met à jour le mot de passe et/ou le rôle d'un utilisateur.
// Refuse de rétrograder le dernier administrateur.
func (s *Store) UpdateUser(id string, password, role *string) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.Users[id]
	if !ok {
		return User{}, fmt.Errorf("user not found")
	}
	if role != nil {
		if !validRole[*role] {
			return User{}, fmt.Errorf("invalid role: must be admin or member")
		}
		if u.Role == "admin" && *role != "admin" && s.countAdminsLocked() <= 1 {
			return User{}, fmt.Errorf("cannot demote the last admin")
		}
		u.Role = *role
	}
	if password != nil {
		if *password == "" {
			return User{}, fmt.Errorf("password is required")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, err
		}
		u.PasswordHash = string(hash)
	}
	s.Users[id] = u
	if err := s.save(); err != nil {
		return User{}, err
	}
	return u, nil
}

// DeleteUser supprime un utilisateur. Refusé pour le dernier administrateur
// (la garde "refus pour soi-même" est appliquée par l'appelant, qui connaît
// l'identité de l'utilisateur courant).
func (s *Store) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.Users[id]
	if !ok {
		return fmt.Errorf("user not found")
	}
	if u.Role == "admin" && s.countAdminsLocked() <= 1 {
		return fmt.Errorf("cannot delete the last admin")
	}
	delete(s.Users, id)
	return s.save()
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

// UpdateProject modifie le nom et/ou la commande de vérification d'un projet.
// Les champs nil ne sont pas modifiés.
func (s *Store) UpdateProject(id string, name, checkCmd *string) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.Projects[id]
	if !ok {
		return Project{}, fmt.Errorf("project not found")
	}
	if name != nil {
		if strings.TrimSpace(*name) == "" {
			return Project{}, fmt.Errorf("name is required")
		}
		p.Name = *name
	}
	if checkCmd != nil {
		p.CheckCmd = *checkCmd
	}
	s.Projects[id] = p
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
		return Card{}, fmt.Errorf("project not found")
	}
	if column == "" {
		column = "soon"
	}
	if !validColumns[column] {
		return Card{}, fmt.Errorf("invalid column")
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
		return Card{}, fmt.Errorf("card not found")
	}
	if !validColumns[column] {
		return Card{}, fmt.Errorf("invalid column")
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
		return Task{}, fmt.Errorf("task not found")
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
// authorName est le nom affiché dans le fil (nom de l'utilisateur pour author="user",
// nom de l'agent pour author="agent").
func (s *Store) AddMessage(taskID, author, authorName, text string) (Message, Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.Tasks[taskID]
	if !ok {
		return Message{}, Task{}, fmt.Errorf("task not found")
	}
	s.NextMessageN++
	m := Message{ID: fmt.Sprintf("m%d", s.NextMessageN), TaskID: taskID, Author: author, AuthorName: authorName, Text: text, CreatedAt: time.Now().UTC()}
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

// --- Agents (CRUD) ---

var validCli = map[string]bool{"claude": true, "codex": true, "fake": true}

// AddAgent crée un agent. name et cli sont obligatoires ; cli doit être
// claude, codex ou fake ; l'identifiant est le slug du nom (unique).
func (s *Store) AddAgent(name, emoji, color, cli, model, contextPrompt string) (Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(name) == "" || cli == "" {
		return Agent{}, fmt.Errorf("name and cli are required")
	}
	if !validCli[cli] {
		return Agent{}, fmt.Errorf("invalid cli: must be claude, codex or fake")
	}
	id := Slugify(name)
	if _, exists := s.Agents[id]; exists {
		return Agent{}, fmt.Errorf("an agent with this name already exists")
	}
	a := Agent{ID: id, Name: name, Emoji: emoji, Color: color, Cli: cli, Model: model, ContextPrompt: contextPrompt}
	s.Agents[id] = a
	if err := s.save(); err != nil {
		return Agent{}, err
	}
	return a, nil
}

// UpdateAgent met à jour les champs fournis (non nil) d'un agent existant.
// L'identifiant ne change jamais après création.
func (s *Store) UpdateAgent(id string, name, emoji, color, cli, model, contextPrompt *string) (Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.Agents[id]
	if !ok {
		return Agent{}, fmt.Errorf("agent not found")
	}
	if cli != nil {
		if !validCli[*cli] {
			return Agent{}, fmt.Errorf("invalid cli: must be claude, codex or fake")
		}
		a.Cli = *cli
	}
	if name != nil {
		if strings.TrimSpace(*name) == "" {
			return Agent{}, fmt.Errorf("name is required")
		}
		a.Name = *name
	}
	if emoji != nil {
		a.Emoji = *emoji
	}
	if color != nil {
		a.Color = *color
	}
	if model != nil {
		a.Model = *model
	}
	if contextPrompt != nil {
		a.ContextPrompt = *contextPrompt
	}
	s.Agents[id] = a
	if err := s.save(); err != nil {
		return Agent{}, err
	}
	return a, nil
}

// DeleteAgent supprime un agent. Refusé si une tâche le référence encore.
func (s *Store) DeleteAgent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Agents[id]; !ok {
		return fmt.Errorf("agent not found")
	}
	for _, t := range s.Tasks {
		if t.AgentID == id {
			return fmt.Errorf("agent is referenced by a task")
		}
	}
	delete(s.Agents, id)
	return s.save()
}
