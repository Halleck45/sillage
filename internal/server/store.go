package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

	Projects  map[string]Project
	Cards     map[string]Card
	Tasks     map[string]Task
	Messages  map[string][]Message
	Agents    map[string]Agent
	Workspace Workspace
	Settings  Settings

	NextProjectN int
	NextCardN    int
	NextTaskN    int
	NextMessageN int
	NextRef      int

	// commitTimer pilote le commit automatique throttlé de l'espace de travail
	// après chaque sauvegarde. Non sérialisé (non exporté).
	commitTimer *time.Timer

	// commitInterval surcharge workspaceCommitInterval quand elle est non
	// nulle (uniquement les tests : attendre 15 minutes n'est pas une option).
	commitInterval time.Duration
}

// NewStore charge state.json s'il existe, sinon initialise un état neuf
// avec les agents seedés.
func NewStore(dataDir string) (*Store, error) {
	s, err := loadStoreFile(dataDir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		s = &Store{dataDir: dataDir}
		s.initEmpty()
	}
	if err := s.save(); err != nil {
		return nil, err
	}
	return s, nil
}

// loadStoreFile lit et parse state.json depuis dataDir, migrations comprises
// (dépôts legacy path->repos, workspace absent -> setupDone=true mode local).
// Retourne une erreur enveloppant os.ErrNotExist si le fichier n'existe pas.
func loadStoreFile(dataDir string) (*Store, error) {
	s := &Store{dataDir: dataDir}
	data, err := os.ReadFile(s.statePath())
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("cannot read state.json: %w", err)
	}
	s.ensureMaps()
	migrateLegacyRepos(data, s)
	migrateLegacyWorkspace(data, s)
	migrateReadyStatus(s)
	return s, nil
}

// migrateReadyStatus migre les tâches antérieures à la v0.3.4 : le statut
// "ready" a disparu (ship est désormais accepté directement depuis review),
// toute tâche encore "ready" devient "review".
func migrateReadyStatus(s *Store) {
	for id, t := range s.Tasks {
		if t.Status == "ready" {
			t.Status = "review"
			s.Tasks[id] = t
		}
	}
}

// migrateLegacyRepos migre les projets antérieurs à la v0.3 (champ "path"
// unique) vers la liste "repos" : si repos est vide et qu'un path legacy est
// présent, repos devient [{name: basename(path), path}].
func migrateLegacyRepos(data []byte, s *Store) {
	var legacy struct {
		Projects map[string]struct {
			Path string `json:"path"`
		} `json:"Projects"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return
	}
	for id, lp := range legacy.Projects {
		if lp.Path == "" {
			continue
		}
		p, ok := s.Projects[id]
		if !ok || len(p.Repos) > 0 {
			continue
		}
		p.Repos = []Repo{{Name: filepath.Base(lp.Path), Path: lp.Path}}
		s.Projects[id] = p
	}
}

// migrateLegacyWorkspace migre les installations antérieures à la v0.3 (state
// déjà rempli, aucun champ "Workspace") : setupDone=true, mode local, pour ne
// jamais afficher l'onboarding sur un espace déjà utilisé.
func migrateLegacyWorkspace(data []byte, s *Store) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return
	}
	if _, ok := raw["Workspace"]; !ok {
		s.Workspace = Workspace{SetupDone: true}
	}
}

// ReloadFromDisk relit state.json depuis dataDir et remplace l'état en
// mémoire, sans changer le pointeur Store : les sessions actives et les
// abonnements SSE restent valides. Utilisé après le rapatriement (clone)
// d'un espace de travail.
func (s *Store) ReloadFromDisk() error {
	fresh, err := loadStoreFile(s.dataDir)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Projects = fresh.Projects
	s.Cards = fresh.Cards
	s.Tasks = fresh.Tasks
	s.Messages = fresh.Messages
	s.Agents = fresh.Agents
	s.Workspace = fresh.Workspace
	s.Settings = fresh.Settings
	s.NextProjectN = fresh.NextProjectN
	s.NextCardN = fresh.NextCardN
	s.NextTaskN = fresh.NextTaskN
	s.NextMessageN = fresh.NextMessageN
	s.NextRef = fresh.NextRef
	return nil
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
	// Les agents par défaut ne ciblent plus des rôles distincts (backend,
	// produit, infra) : Bolt, Muse, Otto et Fably reçoivent tous le même
	// contexte générique. Écho (agent de test local, fake) garde son texte.
	const defaultContextPrompt = "You are a pragmatic developer, focused on quality and simplicity."
	s.Agents["bolt"] = Agent{
		ID: "bolt", Name: "Bolt", Emoji: "🐝", Color: "#f2b705",
		Model: "claude-sonnet-5", Cli: "claude",
		ContextPrompt: defaultContextPrompt,
	}
	s.Agents["muse"] = Agent{
		ID: "muse", Name: "Muse", Emoji: "🦊", Color: "#d0662f",
		Model: "claude-opus-5", Cli: "claude",
		ContextPrompt: defaultContextPrompt,
	}
	s.Agents["otto"] = Agent{
		ID: "otto", Name: "Otto", Emoji: "🦉", Color: "#4f7d2f",
		Model: "", Cli: "codex",
		ContextPrompt: defaultContextPrompt,
	}
	s.Agents["fably"] = Agent{
		ID: "fably", Name: "Fably", Emoji: "🪶", Color: "#6b4fbb",
		Model: "claude-fable-5", Cli: "claude",
		ContextPrompt: defaultContextPrompt,
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
	if err := os.Rename(tmpName, s.statePath()); err != nil {
		return err
	}
	s.scheduleWorkspaceCommit()
	return nil
}

// workspaceCommitInterval borne la fréquence du commit automatique de
// l'espace de travail : au plus un commit par intervalle, quel que soit le
// nombre de sauvegardes (un agent actif en déclenche plusieurs par seconde :
// lignes d'activité, messages, tokens). Un commit stocke un blob complet de
// state.json ; commiter à chaque sauvegarde gonfle le dépôt en objets libres
// pour aucun gain, state.json étant de toute façon déjà écrit atomiquement
// sur disque à chaque mutation. Le commit n'est qu'un point de restauration.
const workspaceCommitInterval = 15 * time.Minute

// scheduleWorkspaceCommit arme le minuteur de commit automatique de l'espace
// de travail s'il ne l'est pas déjà. C'est un throttle, pas un debounce : un
// minuteur en attente n'est jamais repoussé, sinon une activité continue
// (agent qui tourne) empêcherait indéfiniment tout commit.
//
// Doit être appelée avec le verrou déjà tenu (comme save()) ; commitWorkspace
// est sans effet si dataDir n'est pas un dépôt git. Un arrêt du process avant
// l'échéance ne perd que le commit, pas l'état : la sauvegarde faite par
// NewStore au démarrage suivant réarme le minuteur, et une synchronisation
// manuelle (SyncPush) commite de toute façon ce qui est en attente.
func (s *Store) scheduleWorkspaceCommit() {
	if s.commitTimer != nil {
		return
	}
	interval := s.commitInterval
	if interval <= 0 {
		interval = workspaceCommitInterval
	}
	dataDir := s.dataDir
	s.commitTimer = time.AfterFunc(interval, func() {
		// Libérer le créneau avant de commiter (git tourne hors verrou) :
		// une sauvegarde survenant pendant le commit réarme un minuteur.
		s.mu.Lock()
		s.commitTimer = nil
		s.mu.Unlock()
		commitWorkspace(dataDir)
	})
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

// recomputeCard recalcule les compteurs dérivés d'une carte à partir des
// tâches, et déplace automatiquement la carte entre "doing" et "done" :
// si la carte a au moins une tâche et qu'elles sont toutes terminales
// (shipped/done/cancelled), la carte passe (ou reste) en "done" ; si une
// tâche redevient active (reopen, nouvelle tâche) alors que la carte est en
// "done", elle repasse en "doing". Les tâches "cancelled" sont exclues de
// tasksTotal/tasksDone/progress mais comptent comme terminales pour ce
// déplacement automatique. Doit être appelée avec le verrou tenu.
func (s *Store) recomputeCard(cardID string) {
	c, ok := s.Cards[cardID]
	if !ok {
		return
	}
	var total, done, review, docs, msgs int
	var live *string
	hasTasks := false
	allTerminal := true
	for _, t := range s.Tasks {
		if t.CardID != cardID {
			continue
		}
		hasTasks = true
		docs += t.DocsCount
		msgs += t.MessagesCount
		if live == nil && t.Status == "running" && t.LiveActivity != nil {
			v := *t.LiveActivity
			live = &v
		}

		terminal := t.Status == "shipped" || t.Status == "done" || t.Status == "cancelled"
		if !terminal {
			allTerminal = false
		}
		if t.Status == "cancelled" {
			continue // exclue de tasksTotal/tasksDone/progress
		}
		total++
		switch t.Status {
		case "shipped", "done":
			done++
		case "review":
			review++
		}
	}
	progress := 0
	if total > 0 {
		progress = done * 100 / total
	}
	c.TasksTotal, c.TasksDone, c.ReviewCount = total, done, review
	c.DocsCount, c.MessagesCount, c.Progress = docs, msgs, progress
	c.LiveActivity = live

	if hasTasks {
		if allTerminal {
			c.Column = "done"
		} else if c.Column != "doing" {
			// Du travail actif : une carte encore en "soon" (ou redescendue en
			// "done") rejoint la colonne "En cours".
			c.Column = "doing"
		}
	}
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

// sortedAgentsWithWarnings retourne les agents triés par ID, avec leur
// avertissement de santé calculé (voir agentWarning). Coûteux (LookPath,
// lecture /proc) : à n'appeler qu'à la liste des agents (ListAgents/Snapshot),
// jamais depuis recomputeAgent/recomputeAll qui tournent à chaque mutation.
func sortedAgentsWithWarnings(m map[string]Agent) []AgentOut {
	sorted := sortedAgents(m)
	out := make([]AgentOut, len(sorted))
	for i, a := range sorted {
		out[i] = AgentOut{Agent: a, Warning: agentWarning(a)}
	}
	return out
}

// lookPath est une indirection testable de exec.LookPath.
var lookPath = exec.LookPath

// apparmorRestrictPath est le chemin lu pour détecter si les espaces de noms
// utilisateur non privilégiés sont restreints par AppArmor (bloque le sandbox
// par défaut de codex sur certaines machines, ex : Ubuntu 23.10+).
// Indirection testable.
var apparmorRestrictPath = "/proc/sys/kernel/apparmor_restrict_unprivileged_userns"

// agentWarning calcule un avertissement de santé pour un agent (chaîne vide
// si tout va bien) : sandbox codex bloqué par AppArmor sans
// SILLAGE_CODEX_SANDBOX défini, ou binaire cli introuvable dans le PATH.
// Jamais persisté : voir AgentOut.
func agentWarning(a Agent) string {
	switch a.Cli {
	case "codex":
		if apparmorRestrictsUserNamespaces() && os.Getenv("SILLAGE_CODEX_SANDBOX") == "" {
			return "codex sandbox is blocked on this machine (AppArmor); see README (SILLAGE_CODEX_SANDBOX)"
		}
		if _, err := lookPath("codex"); err != nil {
			return "codex CLI not found in PATH"
		}
	case "claude":
		if _, err := lookPath("claude"); err != nil {
			return "claude CLI not found in PATH"
		}
	}
	return ""
}

// apparmorRestrictsUserNamespaces lit apparmorRestrictPath : "1" signifie que
// les espaces de noms utilisateur non privilégiés sont restreints.
func apparmorRestrictsUserNamespaces() bool {
	data, err := os.ReadFile(apparmorRestrictPath)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "1"
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
		Agents:   sortedAgentsWithWarnings(s.Agents),
		Settings: s.Settings,
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

// ListAgents retourne les agents (compteur "active" recalculé et
// avertissement de santé calculé, voir agentWarning), triés par ID.
func (s *Store) ListAgents() []AgentOut {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recomputeAll()
	return sortedAgentsWithWarnings(s.Agents)
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

// TasksByCard retourne les tâches d'une carte (chantier), recalculées.
func (s *Store) TasksByCard(cardID string) []Task {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recomputeAll()
	var out []Task
	for _, t := range s.Tasks {
		if t.CardID == cardID {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return idNum(out[i].ID) < idNum(out[j].ID) })
	if out == nil {
		out = []Task{}
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

// GetWorkspace retourne l'état persisté de synchronisation git de l'espace de travail.
func (s *Store) GetWorkspace() Workspace {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Workspace
}

// UpdateWorkspace applique fn à l'état persisté de l'espace de travail et
// sauvegarde le résultat (déclenche le commit automatique debounced).
func (s *Store) UpdateWorkspace(fn func(w *Workspace)) (Workspace, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(&s.Workspace)
	if err := s.save(); err != nil {
		return Workspace{}, err
	}
	return s.Workspace, nil
}

var validLang = map[string]bool{"": true, "fr": true, "en": true}

// GetSettings retourne les préférences globales persistées.
func (s *Store) GetSettings() Settings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Settings
}

// UpdateSettings met à jour displayName et/ou lang (champs nil non modifiés).
// lang doit être "", "fr" ou "en".
func (s *Store) UpdateSettings(displayName, lang *string) (Settings, error) {
	if lang != nil && !validLang[*lang] {
		return Settings{}, fmt.Errorf("invalid lang: must be fr, en or empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if displayName != nil {
		s.Settings.DisplayName = *displayName
	}
	if lang != nil {
		s.Settings.Lang = *lang
	}
	if err := s.save(); err != nil {
		return Settings{}, err
	}
	return s.Settings, nil
}

// NormalizeRepos valide et normalise une liste de dépôts : au moins un repo,
// chemins et noms non vides, noms uniques (défaut : basename du chemin).
// Ne vérifie PAS que le chemin existe ni qu'il s'agit d'un dépôt git : cette
// validation coûteuse (filesystem + exec git) reste à la charge de l'appelant
// (handlers.go), voir ValidateRepoPath.
func NormalizeRepos(repos []Repo) ([]Repo, error) {
	if len(repos) == 0 {
		return nil, fmt.Errorf("at least one repository is required")
	}
	out := make([]Repo, len(repos))
	seen := map[string]bool{}
	for i, r := range repos {
		path := strings.TrimSpace(r.Path)
		if path == "" {
			return nil, fmt.Errorf("repository path is required")
		}
		name := strings.TrimSpace(r.Name)
		if name == "" {
			name = filepath.Base(path)
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate repository name: %s", name)
		}
		seen[name] = true
		out[i] = Repo{Name: name, Path: path}
	}
	return out, nil
}

// ValidateRepoPath vérifie que path existe et est un dépôt git valide.
// Validation coûteuse (filesystem + exec git) : appelée par les handlers
// avant de persister un projet, jamais par le Store lui-même (pour que les
// tests puissent construire des projets avec des chemins de test simples).
func ValidateRepoPath(path string) error {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("invalid repository path %q: directory not found", path)
	}
	if !IsGitRepo(path) {
		return fmt.Errorf("invalid repository path %q: not a git repository", path)
	}
	return nil
}

// AddProject crée un projet avec un ou plusieurs dépôts git (repos).
// description et contextPrompt peuvent être vides.
func (s *Store) AddProject(name, description, contextPrompt string, repos []Repo, links []Link) (Project, error) {
	normalized, err := NormalizeRepos(repos)
	if err != nil {
		return Project{}, err
	}
	normalizedLinks, err := NormalizeLinks(links)
	if err != nil {
		return Project{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.NextProjectN++
	p := Project{
		ID: fmt.Sprintf("p%d", s.NextProjectN), Name: name, Description: description,
		ContextPrompt: contextPrompt, Repos: normalized, Links: normalizedLinks,
	}
	s.Projects[p.ID] = p
	if err := s.save(); err != nil {
		return Project{}, err
	}
	return p, nil
}

// UpdateProject modifie le nom, la description, la commande de vérification,
// le contexte agent, la liste des dépôts et/ou les liens épinglés d'un
// projet. Les champs nil ne sont pas modifiés ; repos et links, s'ils sont
// fournis (même vides), remplacent entièrement la liste existante (mêmes
// validations que AddProject). Retirer un repo ne casse pas les tâches
// existantes : leur worktree déjà créé vit sa vie indépendamment.
func (s *Store) UpdateProject(id string, name, description, checkCmd, contextPrompt *string, repos *[]Repo, links *[]Link) (Project, error) {
	var normalized []Repo
	if repos != nil {
		var err error
		normalized, err = NormalizeRepos(*repos)
		if err != nil {
			return Project{}, err
		}
	}
	var normalizedLinks []Link
	if links != nil {
		var err error
		normalizedLinks, err = NormalizeLinks(*links)
		if err != nil {
			return Project{}, err
		}
	}
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
	if description != nil {
		p.Description = *description
	}
	if checkCmd != nil {
		p.CheckCmd = *checkCmd
	}
	if contextPrompt != nil {
		p.ContextPrompt = *contextPrompt
	}
	if repos != nil {
		p.Repos = normalized
	}
	if links != nil {
		p.Links = normalizedLinks
	}
	s.Projects[id] = p
	if err := s.save(); err != nil {
		return Project{}, err
	}
	return p, nil
}

// ResolveTaskRepo détermine le dépôt à utiliser pour une nouvelle tâche.
// repoName est optionnel si le projet n'a qu'un seul dépôt (il est alors
// ignoré) ; obligatoire et validé sinon.
func (s *Store) ResolveTaskRepo(projectID, repoName string) (Repo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.Projects[projectID]
	if !ok {
		return Repo{}, fmt.Errorf("project not found")
	}
	if len(p.Repos) == 0 {
		return Repo{}, fmt.Errorf("project has no repositories")
	}
	if len(p.Repos) == 1 {
		return p.Repos[0], nil
	}
	if repoName == "" {
		return Repo{}, fmt.Errorf("repoName required (project has several repositories)")
	}
	for _, r := range p.Repos {
		if r.Name == repoName {
			return r, nil
		}
	}
	return Repo{}, fmt.Errorf("unknown repository")
}

var validColumns = map[string]bool{"soon": true, "doing": true, "done": true}

// AddCard crée une carte (chantier) dans un projet. column vide => "soon".
// contextPrompt est optionnel (texte libre transmis aux agents).
func (s *Store) AddCard(projectID, title, column, contextPrompt string) (Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Projects[projectID]; !ok {
		return Card{}, fmt.Errorf("project not found")
	}
	if column == "" {
		column = "soon"
	}
	if column != "soon" {
		return Card{}, fmt.Errorf("cards are created in the soon column")
	}
	s.NextCardN++
	c := Card{
		ID: fmt.Sprintf("c%d", s.NextCardN), ProjectID: projectID, Column: column, Title: title,
		ContextPrompt: contextPrompt,
	}
	s.Cards[c.ID] = c
	s.recomputeCard(c.ID)
	if err := s.save(); err != nil {
		return Card{}, err
	}
	return s.Cards[c.ID], nil
}

// UpdateCard modifie la colonne, le titre et/ou le contexte agent d'une
// carte. Les champs nil ne sont pas modifiés ; title, s'il est fourni, doit
// être non vide.
func (s *Store) UpdateCard(id string, column, title, contextPrompt *string) (Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.Cards[id]
	if !ok {
		return Card{}, fmt.Errorf("card not found")
	}
	if column != nil {
		if !validColumns[*column] {
			return Card{}, fmt.Errorf("invalid column")
		}
		c.Column = *column
	}
	if title != nil {
		if strings.TrimSpace(*title) == "" {
			return Card{}, fmt.Errorf("title is required")
		}
		c.Title = *title
	}
	if contextPrompt != nil {
		c.ContextPrompt = *contextPrompt
	}
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
func (s *Store) CreateTask(id string, ref int, cardID, projectID, title, agentID, branch, base, worktreeDir, repoName string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Task{
		ID: id, CardID: cardID, ProjectID: projectID, Ref: ref, Title: title,
		AgentID: agentID, RepoName: repoName, Branch: branch, Status: "running", Checks: []Check{},
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

// UpdateTask applique fn à une copie de la tâche puis persiste le résultat,
// en mettant à jour UpdatedAt. C'est le point d'entrée générique utilisé par
// les handlers et le runner pour toute mutation qui doit faire remonter la
// tâche dans un tri par updatedAt. Voir MarkTaskRead pour l'exception.
func (s *Store) UpdateTask(id string, fn func(t *Task)) (Task, error) {
	return s.updateTask(id, fn, true)
}

// MarkTaskRead marque une tâche comme lue (Unread=false) sans mettre à jour
// UpdatedAt : ouvrir une tâche ne doit jamais la faire remonter dans une
// liste triée par updatedAt (le tri sauterait sous le curseur).
func (s *Store) MarkTaskRead(id string) (Task, error) {
	return s.updateTask(id, func(t *Task) { t.Unread = false }, false)
}

// updateTask est l'implémentation commune de UpdateTask/MarkTaskRead.
// bumpUpdatedAt contrôle si UpdatedAt est rafraîchi après fn.
func (s *Store) updateTask(id string, fn func(t *Task), bumpUpdatedAt bool) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.Tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}
	fn(&t)
	if bumpUpdatedAt {
		t.UpdatedAt = time.Now().UTC()
	}
	s.Tasks[id] = t
	s.recomputeCard(t.CardID)
	s.recomputeProject(t.ProjectID)
	s.recomputeAgent(t.AgentID)
	if err := s.save(); err != nil {
		return Task{}, err
	}
	return s.Tasks[id], nil
}

// FinishTask marque une tâche "done". Autorisé depuis review/shipped ;
// message dédié depuis "running" (l'agent tourne encore). Le statut "ready"
// a disparu (v0.3.4) : ship est désormais accepté directement depuis review.
func (s *Store) FinishTask(id string) (Task, error) {
	t, ok := s.GetTask(id)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}
	switch t.Status {
	case "review", "shipped":
	case "running":
		return Task{}, fmt.Errorf("task must be reviewed before finishing")
	default:
		return Task{}, fmt.Errorf("task cannot be finished from its current status")
	}
	return s.UpdateTask(id, func(t *Task) { t.Status = "done" })
}

// CancelTask marque une tâche "cancelled". Autorisé depuis running/review.
// N'interrompt PAS l'agent : c'est la responsabilité de l'appelant pour une
// tâche "running" (voir Runner.Cancel, qui interrompt le process puis appelle
// CancelTask).
func (s *Store) CancelTask(id string) (Task, error) {
	t, ok := s.GetTask(id)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}
	switch t.Status {
	case "running", "review":
	default:
		return Task{}, fmt.Errorf("task cannot be cancelled from its current status")
	}
	return s.UpdateTask(id, func(t *Task) { t.Status = "cancelled" })
}

// ReopenTask remet une tâche en revue ("review"). Autorisé depuis
// shipped/done/cancelled.
func (s *Store) ReopenTask(id string) (Task, error) {
	t, ok := s.GetTask(id)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}
	switch t.Status {
	case "shipped", "done", "cancelled":
	default:
		return Task{}, fmt.Errorf("only a shipped, done or cancelled task can be reopened")
	}
	return s.UpdateTask(id, func(t *Task) { t.Status = "review" })
}

// ReassignTask change l'agent assigné à une tâche. Refusé si la tâche est en
// cours d'exécution (l'agent doit d'abord être interrompu), ou si l'agent
// cible est inconnu. sessionId est vidé : le nouvel agent ne peut pas
// reprendre la session CLI de l'ancien (voir Runner.Start pour le rappel de
// contexte envoyé au prochain message).
func (s *Store) ReassignTask(id, agentID string) (Task, error) {
	t, ok := s.GetTask(id)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}
	if t.Status == "running" {
		return Task{}, fmt.Errorf("interrupt the agent before reassigning")
	}
	if _, ok := s.GetAgent(agentID); !ok {
		return Task{}, fmt.Errorf("agent not found")
	}
	return s.UpdateTask(id, func(t *Task) {
		t.AgentID = agentID
		t.SessionID = ""
	})
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

// --- Suppressions (tâches, cartes, projets) ---

// DeleteTask supprime une tâche et ses messages du store, puis recalcule les
// compteurs dérivés (carte, projet, agent). Ne gère ni l'interruption d'un
// agent en cours, ni le retrait du worktree : voir Runner.DeleteTask pour
// l'orchestration complète.
func (s *Store) DeleteTask(id string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.Tasks[id]
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}
	delete(s.Tasks, id)
	delete(s.Messages, id)
	s.recomputeCard(t.CardID)
	s.recomputeProject(t.ProjectID)
	s.recomputeAgent(t.AgentID)
	if err := s.save(); err != nil {
		return Task{}, err
	}
	return t, nil
}

// DeleteCard supprime une carte (chantier). L'appelant doit avoir supprimé au
// préalable les tâches qu'elle contient (voir Runner.DeleteCard : cascade
// orchestrée tâche par tâche via Runner.DeleteTask).
func (s *Store) DeleteCard(id string) (Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.Cards[id]
	if !ok {
		return Card{}, fmt.Errorf("card not found")
	}
	delete(s.Cards, id)
	if err := s.save(); err != nil {
		return Card{}, err
	}
	return c, nil
}

// DeleteProject supprime un projet. L'appelant doit avoir supprimé au
// préalable ses cartes et tâches (voir Runner.DeleteProject : cascade
// orchestrée chantier par chantier, puis tâche par tâche).
func (s *Store) DeleteProject(id string) (Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.Projects[id]
	if !ok {
		return Project{}, fmt.Errorf("project not found")
	}
	delete(s.Projects, id)
	if err := s.save(); err != nil {
		return Project{}, err
	}
	return p, nil
}
