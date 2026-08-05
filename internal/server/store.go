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

// stateFormatVersion est le format de state.json que ce binaire écrit et sait
// lire. **À incrémenter dès qu'un champ persisté est ajouté, renommé ou change
// de sens.** C'est la seule protection contre un binaire plus ancien : comme
// les champs exportés du Store sont sérialisés tels quels, une version qui ne
// connaît pas un champ le fait disparaître du fichier à sa première
// sauvegarde, en silence (voir ErrStateTooNew).
const stateFormatVersion = 2

// agentSeedVersion tracks one-time additions to the built-in agent profiles.
// Unlike checking for an ID at every startup, this lets users delete a seeded
// agent without Sillage recreating it on the next launch.
const agentSeedVersion = 1

// ErrStateTooNew : le fichier vient d'une version plus récente de Sillage.
// Refuser de démarrer est le seul comportement sûr, parce que charger puis
// sauvegarder détruirait tout ce que ce binaire ne connaît pas.
var ErrStateTooNew = errors.New("state.json comes from a newer Sillage")

// Store est l'état en mémoire de l'application, persisté dans state.json.
// Tous les champs exportés sont sérialisés ; les champs non exportés
// (verrou, répertoire de données) sont ignorés par encoding/json.
type Store struct {
	mu      sync.Mutex
	dataDir string

	// FormatVersion est le format de ce fichier (voir stateFormatVersion).
	// Écrit à chaque sauvegarde ; un fichier plus récent que le binaire fait
	// échouer le chargement au lieu d'être réécrit en perdant des champs.
	FormatVersion int

	// WrittenBy est la version de Sillage qui a écrit ce fichier en dernier
	// ("dev" pour une compilation locale). Sert au diagnostic : c'est la seule
	// trace de « quel binaire a touché mes données ».
	WrittenBy string

	Projects  map[string]Project
	Cards     map[string]Card
	Tasks     map[string]Task
	Messages  map[string][]Message
	Agents    map[string]Agent
	Workspace Workspace
	Settings  Settings

	// AgentSeedVersion records which built-in agent additions have already been
	// offered to this workspace. It does not prevent users from editing or
	// deleting those agents afterwards.
	AgentSeedVersion int

	// CodexQuota est le dernier instantané de quota codex connu (voir
	// AgentQuota), mis à jour en best-effort après chaque exécution codex
	// (runner.go readCodexRateLimits). Nil tant qu'aucune tâche codex n'a
	// encore tourné, ou si le fichier de session ne portait pas l'info.
	CodexQuota *AgentQuota

	NextProjectN int
	NextCardN    int
	NextTaskN    int
	NextMessageN int
	NextRef      int

	// previousWriter est la valeur de WrittenBy telle qu'elle était sur disque
	// au chargement, avant que ce binaire ne s'y inscrive. Non sérialisée.
	previousWriter string

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
	// Avant toute chose, et surtout avant la moindre sauvegarde : un fichier
	// d'un format plus récent n'est pas lisible sans perte.
	if s.FormatVersion > stateFormatVersion {
		return nil, fmt.Errorf("%w (format %d, this binary handles %d): upgrade Sillage (brew upgrade sillage, or grab the latest release) or point -data elsewhere; refusing to start so nothing is overwritten",
			ErrStateTooNew, s.FormatVersion, stateFormatVersion)
	}
	s.previousWriter = s.WrittenBy
	s.ensureMaps()
	migrateLegacyRepos(data, s)
	migrateLegacyWorkspace(data, s)
	migrateTaskStatuses(s)
	migrateLegacyDelivery(s)
	migrateProjectAllowedTools(s)
	migrateCardRefs(s)
	migrateAgentSeeds(s)
	resetTransientTaskFlags(s)
	return s, nil
}

// migrateAgentSeeds adds profiles introduced after a workspace was created.
// Existing agents with the same IDs always win so user customizations remain
// untouched.
func migrateAgentSeeds(s *Store) {
	if s.AgentSeedVersion >= agentSeedVersion {
		return
	}
	for id, agent := range newExternalAgentSeeds() {
		if _, exists := s.Agents[id]; !exists {
			s.Agents[id] = agent
		}
	}
	s.AgentSeedVersion = agentSeedVersion
}

// resetTransientTaskFlags remet à zéro les états de tâche qui ne décrivent
// qu'une opération en cours dans le processus : ils sont faux dès que ce
// processus n'existe plus (arrêt brutal, redémarrage, mise à jour en place).
//
// Un rebase interrompu ne doit pas laisser un fuseau tourner indéfiniment côté
// UI. Et une tâche « running » sans agent est un cul-de-sac : aucune sortie
// n'arrivera plus, « Interrompre l'agent » échoue (le Runner ne connaît aucun
// processus pour elle), sa colonne reste figée et ses compteurs restent à zéro
// alors que le travail est peut-être commité, puisqu'ils ne sont écrits qu'à la
// fin d'une exécution. Elle repasse donc en revue, non lue, avec un marqueur
// dans le fil qui dit pourquoi et des compteurs relus depuis git : ce que
// l'agent avait commité est relisible et acceptable tout de suite.
func resetTransientTaskFlags(s *Store) {
	var interrupted []string
	for id, t := range s.Tasks {
		if t.Rebasing {
			t.Rebasing = false
			s.Tasks[id] = t
		}
		if t.Status == "running" {
			interrupted = append(interrupted, id)
		}
	}
	if len(interrupted) == 0 {
		return
	}
	sort.Slice(interrupted, func(i, j int) bool { return idNum(interrupted[i]) < idNum(interrupted[j]) })
	for _, id := range interrupted {
		t := s.Tasks[id]
		t.Status = "review"
		t.Unread = true
		t.LiveActivity = nil
		t.FilesCount, t.DocsCount, t.CommitsCount = TaskWorkCounts(t.WorktreeDir, t.Base)
		s.Tasks[id] = t
		s.appendMessage(id, "agent", "", "[interrupted:server-restart]")
	}
	s.recomputeAll()
}

// migrateCardRefs donne une référence aux chantiers antérieurs au champ Ref :
// sans elle, leur branche s'appellerait `sillage/ws-0-<slug>` et deux chantiers
// de même titre entreraient en collision. Attribution dans l'ordre des
// identifiants pour rester déterministe.
func migrateCardRefs(s *Store) {
	var ids []string
	for id, c := range s.Cards {
		if c.Ref == 0 {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return idNum(ids[i]) < idNum(ids[j]) })
	for _, id := range ids {
		c := s.Cards[id]
		s.NextRef++
		c.Ref = s.NextRef
		s.Cards[id] = c
	}
}

// migrateTaskStatuses migre les statuts de tâche disparus :
//   - "ready" (antérieur à la v0.3.4) devient "review" ;
//   - "shipped" et "done" deviennent "accepted" : une tâche ne se livre plus
//     seule, elle est acceptée dans la branche de son chantier, et c'est le
//     chantier qui se livre (voir docs/SPEC-LIVRAISON.md).
func migrateTaskStatuses(s *Store) {
	for id, t := range s.Tasks {
		switch t.Status {
		case "ready":
			t.Status = "review"
		case "shipped", "done":
			t.Status = "accepted"
		default:
			continue
		}
		s.Tasks[id] = t
	}
}

// migrateLegacyDelivery donne un mode de livraison aux projets antérieurs à
// l'introduction de Project.Delivery : "pr", le comportement historique
// (pousser la branche puis ouvrir la pull request).
func migrateLegacyDelivery(s *Store) {
	for id, p := range s.Projects {
		if p.Delivery.Mode == "" {
			p.Delivery.Mode = "pr"
			s.Projects[id] = p
		}
	}
}

// legacyGoAllowedTools est la chaîne Go qui vivait dans le socle du binaire
// avant Project.AllowedTools. Elle en est sortie : le socle ne contient que ce
// qui ne dépend d'aucun langage, sinon tout projet Python ou Rust hérite d'une
// allowlist écrite pour un autre. Les projets existants la récupèrent ici pour
// garder exactement leur comportement d'avant la mise à jour.
var legacyGoAllowedTools = []string{"Bash(go build:*)", "Bash(go test:*)", "Bash(go vet:*)"}

// migrateProjectAllowedTools amorce les projets antérieurs au champ. nil veut
// dire « antérieur » ; une liste vide non nulle est un choix explicite de
// l'utilisateur (voir AddProject et NormalizeAllowedTools) et n'est pas
// réamorcée à chaque démarrage.
func migrateProjectAllowedTools(s *Store) {
	for id, p := range s.Projects {
		if p.AllowedTools == nil {
			p.AllowedTools = append([]string{}, legacyGoAllowedTools...)
			s.Projects[id] = p
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
	s.FormatVersion = fresh.FormatVersion
	s.WrittenBy = fresh.WrittenBy
	s.previousWriter = fresh.previousWriter
	s.Projects = fresh.Projects
	s.Cards = fresh.Cards
	s.Tasks = fresh.Tasks
	s.Messages = fresh.Messages
	s.Agents = fresh.Agents
	s.Workspace = fresh.Workspace
	s.Settings = fresh.Settings
	s.AgentSeedVersion = fresh.AgentSeedVersion
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

// DowngradeWarning décrit le risque que le format ne sait pas voir : deux
// versions publiées du même format, dont celle qui tourne est la plus ancienne.
// Elle chargera le fichier sans erreur, puis supprimera à la première
// sauvegarde les champs apparus entre les deux. Vide quand il n'y a rien à
// dire, notamment dès qu'une compilation locale est en jeu (une version "dev"
// ne se compare à rien).
func (s *Store) DowngradeWarning() string {
	previous := s.previousWriter
	if !isReleaseVersion(previous) || !isReleaseVersion(buildVersion) {
		return ""
	}
	if compareVersions(previous, buildVersion) <= 0 {
		return ""
	}
	return fmt.Sprintf("state.json was last written by Sillage %s and this binary is %s: fields it does not know will be dropped on the next save. Upgrade Sillage, or point -data elsewhere.", previous, buildVersion)
}

// StateFileFormatVersion lit le seul champ FormatVersion d'un state.json, sans
// charger le reste : sert à refuser un espace de travail distant plus récent
// avant de toucher au répertoire de données (voir CloneWorkspace).
func StateFileFormatVersion(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var head struct {
		FormatVersion int
	}
	if err := json.Unmarshal(data, &head); err != nil {
		return 0, fmt.Errorf("cannot read %s: %w", filepath.Base(path), err)
	}
	return head.FormatVersion, nil
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
	for id, agent := range newExternalAgentSeeds() {
		s.Agents[id] = agent
	}
	s.AgentSeedVersion = agentSeedVersion
}

func newExternalAgentSeeds() map[string]Agent {
	const contextPrompt = "You are a pragmatic developer, focused on quality and simplicity."
	return map[string]Agent{
		"github-copilot": {
			ID: "github-copilot", Name: "Octo", Emoji: "🐙", Color: "#24292f",
			Cli: "copilot", ContextPrompt: contextPrompt,
		},
		"antigravity": {
			ID: "antigravity", Name: "Astro", Emoji: "🚀", Color: "#4285f4",
			Cli: "agy", ContextPrompt: contextPrompt,
		},
	}
}

// save écrit l'état sur disque de façon atomique (fichier temp + rename).
// Doit être appelée avec le verrou déjà tenu.
func (s *Store) save() error {
	if err := os.MkdirAll(s.dataDir, 0o700); err != nil {
		return err
	}
	// Le fichier porte toujours la signature de qui l'a écrit en dernier : le
	// format pour interdire une relecture par un binaire plus ancien, la
	// version pour pouvoir répondre à « quel Sillage a touché mes données ».
	s.FormatVersion = stateFormatVersion
	s.WrittenBy = buildVersion
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
// tâches, l'état de son bouton de livraison (ShipReady/ShipBlocker) et sa
// colonne.
//
// Colonne : une carte qui a des tâches passe en "done" quand elle a été
// livrée ET que toutes ses tâches sont terminales (accepted/cancelled) ; sinon
// elle passe (ou reste) en "doing". La colonne "Terminé" veut donc dire livré,
// pas seulement relu. Le déplacement manuel reste indépendant. Les cartes
// antérieures aux branches de chantier (aucune Branches) sont considérées
// livrées : leur colonne garde le comportement historique.
//
// Les tâches "cancelled" (refusées) sont exclues de
// tasksTotal/tasksDone/progress mais comptent comme terminales.
// Doit être appelée avec le verrou tenu.
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

		terminal := t.Status == "accepted" || t.Status == "cancelled"
		if !terminal {
			allTerminal = false
		}
		if t.Status == "cancelled" {
			continue // exclue de tasksTotal/tasksDone/progress
		}
		total++
		switch t.Status {
		case "accepted":
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
	if c.Branches == nil {
		// Cartes antérieures aux branches de chantier : toujours exposer un
		// tableau, jamais null (voir le modèle Card de SPEC-API.md).
		c.Branches = []CardBranch{}
	}
	c.ShipReady, c.ShipBlocker = shipReadiness(c, hasTasks, done)

	shipped := len(c.Branches) == 0 // carte historique : comportement d'avant
	for _, b := range c.Branches {
		if b.ShippedAt != nil {
			shipped = true
			break
		}
	}
	c.AwaitingShip = hasTasks && allTerminal && done > 0 && !shipped
	if hasTasks {
		if allTerminal && shipped {
			c.Column = "done"
		} else if c.Column != "doing" {
			// Du travail actif (ou relu mais pas encore livré) : une carte
			// encore en "soon" (ou redescendue de "done") rejoint "En cours".
			c.Column = "doing"
		}
	}
	s.Cards[cardID] = c
}

// shipReadiness calcule si un chantier peut être livré, et sinon pourquoi.
// Les trois conditions de docs/SPEC-LIVRAISON.md, dans l'ordre : au moins une
// tâche, au moins une tâche acceptée, au moins une branche de chantier. Le
// nombre de commits à livrer n'est pas vérifié ici (il coûterait un appel git à
// chaque mutation) : c'est l'aperçu de livraison qui le rapporte, dépôt par
// dépôt.
//
// Une tâche encore en cours ou à relire ne bloque PAS la livraison : la branche
// du chantier ne contient que le travail accepté, donc livrer un chantier
// inachevé n'envoie jamais rien de non relu. On peut ainsi livrer en cours de
// route et relivrer ensuite (voir CardBranch.ShippedAt). La colonne « Terminé »,
// elle, reste conditionnée à « tout terminal ET livré » (voir recomputeCard).
func shipReadiness(c Card, hasTasks bool, accepted int) (bool, string) {
	switch {
	case !hasTasks:
		return false, "no-tasks"
	case accepted == 0:
		return false, "nothing-accepted"
	case len(c.Branches) == 0:
		return false, "nothing-to-ship"
	}
	return true, ""
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

// sortedProjectsWithWarnings retourne les projets triés par ID, avec leur
// avertissement de santé de livraison calculé (voir deliveryWarning). Coûteux
// (LookPath + git remote par dépôt) : à n'appeler qu'à la lecture de l'état ou
// après une mutation de projet, jamais depuis recompute*.
func sortedProjectsWithWarnings(m map[string]Project) []ProjectOut {
	sorted := sortedProjects(m)
	out := make([]ProjectOut, len(sorted))
	for i, p := range sorted {
		out[i] = projectOut(p)
	}
	return out
}

// projectOut habille un projet de son avertissement de livraison.
func projectOut(p Project) ProjectOut {
	return ProjectOut{Project: p, DeliveryWarning: deliveryWarning(p)}
}

// deliveryWarning calcule un avertissement de santé de la livraison (chaîne
// vide si tout va bien) : CLI de la forge absent du PATH, ou dépôt sans remote
// origin alors que le mode exige un push. N'empêche jamais rien : le repli est
// une URL de pull request pré-remplie. Jamais persisté : voir ProjectOut.
func deliveryWarning(p Project) string {
	if !deliveryModePushes(p.Delivery.Mode) {
		return "" // fusion locale : aucun binaire externe, aucun remote requis
	}
	needsGh, needsGlab := false, false
	for _, repo := range p.Repos {
		info := DetectForge(repo.Path)
		if info.RemoteURL == "" {
			return "no 'origin' remote on repository " + repo.Name + "; nothing can be pushed"
		}
		// Seul le mode "pr" a besoin d'une forge reconnue et de son CLI : les
		// modes "push" et "merge-push" poussent, point, quelle que soit la forge.
		if p.Delivery.Mode != "pr" {
			continue
		}
		switch info.Provider {
		case "github":
			needsGh = true
		case "gitlab":
			needsGlab = true
		default:
			return "unknown forge on repository " + repo.Name + "; the branch will be pushed without opening a pull request"
		}
	}
	if needsGh {
		if _, err := lookPath("gh"); err != nil {
			return "gh not found in PATH; Sillage will fall back to a prefilled pull request URL"
		}
	}
	if needsGlab {
		if _, err := lookPath("glab"); err != nil {
			return "glab not found in PATH; Sillage will fall back to a prefilled merge request URL"
		}
	}
	return ""
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
// codexQuota est le dernier instantané connu (Store.CodexQuota), attaché aux
// seuls agents cli=codex : c'est un quota de compte, partagé entre eux.
func sortedAgentsWithWarnings(m map[string]Agent, codexQuota *AgentQuota, worktreesDir string) []AgentOut {
	sorted := sortedAgents(m)
	out := make([]AgentOut, len(sorted))
	for i, a := range sorted {
		out[i] = AgentOut{Agent: a, Warning: agentWarning(a, worktreesDir)}
		if a.Cli == "codex" {
			out[i].Quota = codexQuota
		}
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

// antigravitySettingsPath est le fichier de configuration de la CLI agy, seul
// endroit où se règle sa politique d'exécution des commandes (aucun drapeau
// n'expose ce réglage). Indirection testable.
var antigravitySettingsPath = defaultAntigravitySettingsPath()

const antigravityPolicyWarning = `agy cannot work headlessly with its current permissions; see ~/.gemini/antigravity-cli/settings.json`

func defaultAntigravitySettingsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".gemini", "antigravity-cli", "settings.json")
}

// agentWarning calcule un avertissement de santé pour un agent (chaîne vide
// si tout va bien) : sandbox codex bloqué par AppArmor sans
// SILLAGE_CODEX_SANDBOX défini, politique d'exécution d'agy incompatible avec
// le mode headless, ou binaire cli introuvable dans le PATH.
// Jamais persisté : voir AgentOut.
func agentWarning(a Agent, worktreesDir string) string {
	switch a.Cli {
	case "codex":
		if _, err := lookPath("codex"); err != nil {
			return "codex CLI not found in PATH"
		}
		if apparmorRestrictsUserNamespaces() && os.Getenv("SILLAGE_CODEX_SANDBOX") == "" {
			return "codex sandbox is blocked on this machine (AppArmor); see README (SILLAGE_CODEX_SANDBOX)"
		}
	case "agy":
		if _, err := lookPath("agy"); err != nil {
			return "agy CLI not found in PATH"
		}
		if !antigravityWorksHeadlessly(worktreesDir) {
			return antigravityPolicyWarning
		}
	case "claude", "copilot":
		if _, err := lookPath(a.Cli); err != nil {
			return a.Cli + " CLI not found in PATH"
		}
	}
	return ""
}

// antigravitySettings est la partie du fichier de configuration d'agy dont
// dépend le travail sans humain devant l'écran.
type antigravitySettings struct {
	ToolPermission string `json:"toolPermission"`
	Permissions    struct {
		Allow []string `json:"allow"`
	} `json:"permissions"`
}

// antigravityWorksHeadlessly dit si la CLI agy peut travailler sans personne
// pour l'autoriser. En mode print, toute demande de confirmation est
// auto-refusée et la session s'arrête aussitôt, sans rien écrire sur stdout :
// une seule demande suffit à rendre une tâche muette, sans diff ni message.
// Deux choses sont donc nécessaires :
//
//   - une politique d'exécution des commandes qui ne demande rien :
//     "proceed-in-sandbox" (l'accord vient du bac à sable, que Sillage force
//     toujours) ou "always-proceed" ; le défaut "request-review" ne marche pas ;
//   - des autorisations de lecture et d'écriture sur les worktrees. Les fichiers
//     ordinaires du worktree passent sans règle, mais certains chemins sont
//     traités comme sensibles (`.git`, qui dans un worktree lié est un fichier
//     qu'un modèle en exploration essaie d'ouvrir) et certaines modifications
//     demandent un accord. Une règle sur un **dossier** couvre tout ce qu'il
//     contient (vérifié) : deux règles sur la racine des worktrees suffisent, et
//     ne donnent rien de plus que le terrain de jeu de l'agent.
//
// Fichier illisible ou absent : politique par défaut, donc non.
func antigravityWorksHeadlessly(worktreesDir string) bool {
	if antigravitySettingsPath == "" {
		return true // chemin du home inconnu : rien à diagnostiquer.
	}
	data, err := os.ReadFile(antigravitySettingsPath)
	if err != nil {
		return false
	}
	var cfg antigravitySettings
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false
	}
	switch cfg.ToolPermission {
	case "proceed-in-sandbox", "always-proceed":
	default:
		if !antigravityRuleCovers(cfg.Permissions.Allow, "command", "") {
			return false
		}
	}
	for _, tool := range antigravityFileTools {
		if !antigravityRuleCovers(cfg.Permissions.Allow, tool, worktreesDir) {
			return false
		}
	}
	return true
}

// antigravityFileTools sont les outils de fichier dont un agent de code a besoin
// dans son worktree.
var antigravityFileTools = []string{"read_file", "write_file"}

// antigravityRuleCovers dit si permissions.allow autorise tool sur tout le
// contenu de dir. Les règles d'agy ne connaissent pas les motifs de chemin :
// `read_file(<dossier>/*)` ne correspond à rien. Seuls comptent le joker global
// `*`, un chemin exact, et un dossier, qui couvre ce qu'il contient. dir vide
// ne demande que l'existence d'une règle pour cet outil.
func antigravityRuleCovers(rules []string, tool, dir string) bool {
	if dir != "" {
		dir = filepath.Clean(dir)
	}
	for _, rule := range rules {
		if !strings.HasPrefix(rule, tool+"(") || !strings.HasSuffix(rule, ")") {
			continue
		}
		target := rule[len(tool)+1 : len(rule)-1]
		if target == "*" || dir == "" {
			return true
		}
		target = filepath.Clean(target)
		if target == dir || strings.HasPrefix(dir, target+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// antigravityToolPermissionFix est la politique posée par le correctif : les
// commandes partent sans confirmation, mais uniquement dans le bac à sable, que
// Sillage force toujours pour cet agent.
const antigravityToolPermissionFix = "proceed-in-sandbox"

// fixAntigravityToolPermission règle la politique d'exécution et les deux
// autorisations de fichier sur la racine des worktrees dans le fichier de
// configuration de la CLI agy. C'est le seul endroit où ces réglages
// s'expriment, et Sillage n'y touche que sur demande explicite de l'humain
// (bouton de l'avertissement de l'agent).
//
// Les autres clés, et les règles déjà présentes, sont relues et réécrites
// telles quelles : le fichier appartient à l'utilisateur, et il y garde par
// exemple ses espaces de travail de confiance. Un fichier illisible n'est jamais
// écrasé : mieux vaut refuser et le dire que perdre une configuration.
func fixAntigravityToolPermission(worktreesDir string) error {
	if antigravitySettingsPath == "" {
		return fmt.Errorf("cannot locate the agy settings file")
	}
	settings := map[string]any{}
	mode := os.FileMode(0o644)
	data, err := os.ReadFile(antigravitySettingsPath)
	switch {
	case err == nil:
		if len(strings.TrimSpace(string(data))) > 0 {
			if err := json.Unmarshal(data, &settings); err != nil {
				return fmt.Errorf("agy settings file is not valid JSON, fix it by hand: %w", err)
			}
		}
		if info, err := os.Stat(antigravitySettingsPath); err == nil {
			mode = info.Mode().Perm()
		}
	case os.IsNotExist(err):
		if err := os.MkdirAll(filepath.Dir(antigravitySettingsPath), 0o755); err != nil {
			return err
		}
	default:
		return err
	}

	settings["toolPermission"] = antigravityToolPermissionFix
	settings["permissions"] = antigravityAllowWorktrees(settings["permissions"], worktreesDir)
	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')

	tmp := antigravitySettingsPath + ".sillage-tmp"
	if err := os.WriteFile(tmp, out, mode); err != nil {
		return err
	}
	if err := os.Rename(tmp, antigravitySettingsPath); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// antigravityAllowWorktrees complète le bloc permissions relu du fichier avec
// les règles manquantes sur la racine des worktrees, sans toucher au reste
// (règles déjà là, refus, autres clés).
func antigravityAllowWorktrees(existing any, worktreesDir string) map[string]any {
	perms := map[string]any{}
	if m, ok := existing.(map[string]any); ok {
		for k, v := range m {
			perms[k] = v
		}
	}
	var allow []string
	var raw []any
	if list, ok := perms["allow"].([]any); ok {
		raw = list
	}
	for _, item := range raw {
		if s, ok := item.(string); ok {
			allow = append(allow, s)
		}
	}
	for _, tool := range antigravityFileTools {
		if !antigravityRuleCovers(allow, tool, worktreesDir) {
			allow = append(allow, tool+"("+filepath.Clean(worktreesDir)+")")
		}
	}
	perms["allow"] = allow
	return perms
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
		Projects: sortedProjectsWithWarnings(s.Projects),
		Cards:    sortedCards(s.Cards),
		Tasks:    sortedTasks(s.Tasks),
		Agents:   sortedAgentsWithWarnings(s.Agents, s.CodexQuota, s.WorktreesDir()),
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
	return sortedAgentsWithWarnings(s.Agents, s.CodexQuota, s.WorktreesDir())
}

// WorktreesDir est la racine des worktrees (un dossier par tâche et par branche
// de chantier), sous dataDir. Voir git.go pour les chemins eux-mêmes.
func (s *Store) WorktreesDir() string {
	return filepath.Join(s.dataDir, "worktrees")
}

// SetCodexQuota met à jour l'instantané de quota codex (voir AgentQuota),
// partagé par tous les agents cli=codex. Appelé en best-effort après chaque
// exécution codex (runner.go readCodexRateLimits) : windows vide n'efface pas
// un instantané précédent, l'appelant ne l'invoque que s'il a trouvé quelque
// chose.
func (s *Store) SetCodexQuota(windows []AgentQuotaWindow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CodexQuota = &AgentQuota{UpdatedAt: time.Now(), Windows: windows}
	return s.save()
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

// UpdateSettings met à jour displayName, lang et/ou updateCheck (champs nil
// non modifiés). lang doit être "", "fr" ou "en".
func (s *Store) UpdateSettings(displayName, lang *string, updateCheck *bool) (Settings, error) {
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
	if updateCheck != nil {
		v := *updateCheck
		s.Settings.UpdateCheck = &v
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
		previewURL := strings.TrimSpace(r.PreviewURL)
		// L'URL de recette devient un lien cliquable dans l'interface : seuls
		// http(s) sont acceptés, comme pour les liens épinglés. Elle n'est pas
		// validée par url.Parse, parce que c'est un gabarit qui contient encore
		// ses variables (http://127.0.0.1:$((4000 + SILLAGE_N))).
		if previewURL != "" && !strings.HasPrefix(previewURL, "http://") && !strings.HasPrefix(previewURL, "https://") {
			return nil, fmt.Errorf("invalid preview url %q: only http(s) URLs are allowed", previewURL)
		}
		out[i] = Repo{
			Name:       name,
			Path:       path,
			PreviewCmd: strings.TrimSpace(r.PreviewCmd),
			PreviewURL: previewURL,
		}
	}
	return out, nil
}

var validDeliveryModes = map[string]bool{"pr": true, "push": true, "merge": true, "merge-push": true}

// deliveryModePushes indique si un mode de livraison exécute un push : les deux
// modes de branche, plus la fusion suivie du push de la branche de destination.
// Le mode "merge" est le seul à ne jamais toucher au réseau.
func deliveryModePushes(mode string) bool {
	return mode != "merge"
}

// NormalizeDelivery valide et normalise le réglage de livraison d'un projet.
// nil (ou un mode vide) vaut le défaut "pr" : pousser la branche du chantier
// et ouvrir la pull/merge request. Target vide signifie « branche par défaut
// du dépôt », et n'est donc pas une erreur.
func NormalizeDelivery(d *Delivery) (Delivery, error) {
	if d == nil {
		return Delivery{Mode: "pr"}, nil
	}
	out := Delivery{Mode: strings.TrimSpace(d.Mode), Target: strings.TrimSpace(d.Target), StackedPrs: d.StackedPrs}
	if out.Mode == "" {
		out.Mode = "pr"
	}
	if !validDeliveryModes[out.Mode] {
		return Delivery{}, fmt.Errorf("invalid delivery mode: must be pr, push, merge or merge-push")
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
// description et contextPrompt peuvent être vides ; delivery nil vaut le mode
// par défaut ("pr"). Un nom vide est déduit du premier dépôt : créer un projet
// ne demande donc qu'un chemin, tout le reste se règle ensuite.
func (s *Store) AddProject(name, description, contextPrompt string, repos []Repo, links []Link, delivery *Delivery) (Project, error) {
	normalized, err := NormalizeRepos(repos)
	if err != nil {
		return Project{}, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = normalized[0].Name
	}
	normalizedLinks, err := NormalizeLinks(links)
	if err != nil {
		return Project{}, err
	}
	normalizedDelivery, err := NormalizeDelivery(delivery)
	if err != nil {
		return Project{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.NextProjectN++
	p := Project{
		ID: fmt.Sprintf("p%d", s.NextProjectN), Name: name, Description: description,
		ContextPrompt: contextPrompt, Repos: normalized, Links: normalizedLinks,
		Delivery: normalizedDelivery,
		// Liste vide non nulle : un projet neuf n'a rien à migrer, alors qu'un
		// nil serait relu comme « antérieur au champ » au prochain démarrage.
		AllowedTools: []string{},
	}
	s.Projects[p.ID] = p
	if err := s.save(); err != nil {
		return Project{}, err
	}
	return p, nil
}

// NormalizeAllowedTools nettoie la liste saisie par l'humain : une entrée par
// ligne dans l'UI, donc espaces superflus et lignes vides à retirer. Aucune
// validation du contenu : c'est le refus figé passé au CLI (claudeDeniedTools)
// qui garantit l'invariant, pas un filtre de saisie qui donnerait une fausse
// impression de sûreté (Bash(sh:*) contournerait n'importe quelle liste noire).
// Retourne une liste vide non nulle plutôt que nil : nil veut dire « projet
// antérieur au champ » pour migrateProjectAllowedTools.
func NormalizeAllowedTools(tools []string) []string {
	out := []string{}
	for _, tool := range tools {
		if trimmed := strings.TrimSpace(tool); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// UpdateProject modifie le nom, la description, la commande de vérification,
// le contexte agent, les outils autorisés, la liste des dépôts et/ou les liens
// épinglés d'un projet. Les champs nil ne sont pas modifiés ; allowedTools,
// repos et links, s'ils sont fournis (même vides), remplacent entièrement la
// liste existante (mêmes validations que AddProject). Retirer un repo ne casse
// pas les tâches existantes : leur worktree déjà créé vit sa vie indépendamment.
func (s *Store) UpdateProject(id string, name, description, checkCmd, contextPrompt *string, allowedTools *[]string, repos *[]Repo, links *[]Link, delivery *Delivery) (Project, error) {
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
	var normalizedDelivery Delivery
	if delivery != nil {
		var err error
		normalizedDelivery, err = NormalizeDelivery(delivery)
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
	if allowedTools != nil {
		p.AllowedTools = NormalizeAllowedTools(*allowedTools)
	}
	if repos != nil {
		p.Repos = normalized
	}
	if links != nil {
		p.Links = normalizedLinks
	}
	if delivery != nil {
		p.Delivery = normalizedDelivery
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
	// Le compteur de référence est partagé avec les tâches : une référence
	// courte, unique et stable, utilisée dans le nom de la branche du chantier.
	s.NextRef++
	c := Card{
		ID: fmt.Sprintf("c%d", s.NextCardN), ProjectID: projectID, Ref: s.NextRef,
		Column: column, Title: title, ContextPrompt: contextPrompt, Branches: []CardBranch{},
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

// GetCardBranch retourne la branche de feature d'un chantier sur un dépôt
// donné, si elle a déjà été créée (première tâche du chantier sur ce dépôt).
func (s *Store) GetCardBranch(cardID, repoName string) (CardBranch, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range s.Cards[cardID].Branches {
		if b.RepoName == repoName {
			return b, true
		}
	}
	return CardBranch{}, false
}

// SetCardBranch enregistre (ou remplace) la branche de feature d'un chantier
// sur un dépôt. Appelée après la création effective de la branche et de son
// worktree par git (voir CreateCardWorktree).
func (s *Store) SetCardBranch(cardID string, b CardBranch) (Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.Cards[cardID]
	if !ok {
		return Card{}, fmt.Errorf("card not found")
	}
	replaced := false
	for i, existing := range c.Branches {
		if existing.RepoName == b.RepoName {
			c.Branches[i] = b
			replaced = true
			break
		}
	}
	if !replaced {
		c.Branches = append(c.Branches, b)
	}
	s.Cards[cardID] = c
	s.recomputeCard(cardID)
	if err := s.save(); err != nil {
		return Card{}, err
	}
	return s.Cards[cardID], nil
}

// MarkCardBranchShipped enregistre le résultat d'une livraison réussie sur un
// dépôt : URL de la pull/merge request (vide en mode fusion locale) et date.
func (s *Store) MarkCardBranchShipped(cardID, repoName, prURL string, at time.Time) (Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.Cards[cardID]
	if !ok {
		return Card{}, fmt.Errorf("card not found")
	}
	for i, b := range c.Branches {
		if b.RepoName != repoName {
			continue
		}
		when := at
		c.Branches[i].ShippedAt = &when
		if prURL != "" {
			c.Branches[i].PrURL = prURL
		}
	}
	s.Cards[cardID] = c
	s.recomputeCard(cardID)
	if err := s.save(); err != nil {
		return Card{}, err
	}
	return s.Cards[cardID], nil
}

// MarkCardBranchPending remet la branche de chantier d'un dépôt à l'état « non
// livré » : appelée dès que du travail nouveau apparaît (acceptation qui ajoute
// des commits, nouvelle tâche sur ce dépôt), pour qu'un chantier déjà livré
// puisse continuer à vivre. PrURL est conservée : la pull request existe
// toujours, et une nouvelle livraison la mettra à jour.
func (s *Store) MarkCardBranchPending(cardID, repoName string) (Card, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.Cards[cardID]
	if !ok {
		return Card{}, fmt.Errorf("card not found")
	}
	changed := false
	for i, b := range c.Branches {
		if b.RepoName == repoName && b.ShippedAt != nil {
			c.Branches[i].ShippedAt = nil
			changed = true
		}
	}
	if !changed {
		return c, nil
	}
	s.Cards[cardID] = c
	s.recomputeCard(cardID)
	if err := s.save(); err != nil {
		return Card{}, err
	}
	return s.Cards[cardID], nil
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
	return s.createTask(id, ref, cardID, projectID, title, agentID, branch, base, worktreeDir, repoName, "", "")
}

// CreateWaitingTask enregistre une nouvelle tâche "waiting" : le worktree est
// créé comme pour une tâche ordinaire, mais l'agent n'est pas lancé.
// pendingPrompt mémorise le texte à lui envoyer une fois waitsForTaskID
// (obligatoire) accepté ; voir Server.startWaitingTask.
func (s *Store) CreateWaitingTask(id string, ref int, cardID, projectID, title, agentID, branch, base, worktreeDir, repoName, waitsForTaskID, pendingPrompt string) (Task, error) {
	return s.createTask(id, ref, cardID, projectID, title, agentID, branch, base, worktreeDir, repoName, waitsForTaskID, pendingPrompt)
}

func (s *Store) createTask(id string, ref int, cardID, projectID, title, agentID, branch, base, worktreeDir, repoName, waitsForTaskID, pendingPrompt string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Task{
		ID: id, CardID: cardID, ProjectID: projectID, Ref: ref, Title: title,
		AgentID: agentID, RepoName: repoName, Branch: branch, Status: "running", Checks: []Check{},
		Unread: false, UpdatedAt: time.Now().UTC(), Tokens: Tokens{},
		Base: base, WorktreeDir: worktreeDir,
	}
	if waitsForTaskID != "" {
		t.Status = "waiting"
		t.WaitsForTaskID = waitsForTaskID
		t.PendingPrompt = pendingPrompt
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

// SetTaskRebasing arme (ou désarme) le témoin de rebase automatique en cours,
// sans toucher à UpdatedAt : une opération technique ne doit pas réordonner la
// liste des tâches sous le curseur de l'utilisateur.
func (s *Store) SetTaskRebasing(id string, rebasing bool) (Task, error) {
	return s.updateTask(id, func(t *Task) { t.Rebasing = rebasing }, false)
}

// MarkTaskRead marque une tâche comme lue (Unread=false) sans mettre à jour
// UpdatedAt : ouvrir une tâche ne doit jamais la faire remonter dans une
// liste triée par updatedAt (le tri sauterait sous le curseur).
func (s *Store) MarkTaskRead(id string) (Task, error) {
	return s.updateTask(id, func(t *Task) { t.Unread = false }, false)
}

// MarkAllTasksReadForProject marque comme lues toutes les tâches non lues
// d'un projet (menu "..." d'un projet dans la sidebar), sans toucher à
// UpdatedAt pour la même raison que MarkTaskRead. Renvoie les tâches
// effectivement modifiées, pour que l'appelant publie leurs événements SSE.
func (s *Store) MarkAllTasksReadForProject(projectID string) ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.Projects[projectID]; !ok {
		return nil, fmt.Errorf("project not found")
	}
	var changed []Task
	for id, t := range s.Tasks {
		if t.ProjectID == projectID && t.Unread {
			t.Unread = false
			s.Tasks[id] = t
			changed = append(changed, t)
		}
	}
	if len(changed) == 0 {
		return nil, nil
	}
	for _, t := range changed {
		s.recomputeCard(t.CardID)
		s.recomputeAgent(t.AgentID)
	}
	s.recomputeProject(projectID)
	if err := s.save(); err != nil {
		return nil, err
	}
	result := make([]Task, len(changed))
	for i, t := range changed {
		result[i] = s.Tasks[t.ID]
	}
	return result, nil
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

// AcceptTask marque une tâche "accepted" : son travail a été fusionné dans la
// branche du chantier. Autorisé depuis "review" uniquement ; la fusion git
// elle-même est faite par l'appelant AVANT cet appel (voir handleAccept), pour
// qu'un conflit laisse la tâche en revue.
func (s *Store) AcceptTask(id string) (Task, error) {
	t, ok := s.GetTask(id)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}
	switch t.Status {
	case "review":
	case "running":
		return Task{}, fmt.Errorf("interrupt the agent before accepting")
	default:
		return Task{}, fmt.Errorf("only a task in review can be accepted")
	}
	return s.UpdateTask(id, func(t *Task) { t.Status = "accepted"; t.Unread = false })
}

// CancelTask marque une tâche "cancelled" (refusée). Autorisé depuis
// waiting/running/review. N'interrompt PAS l'agent : c'est la responsabilité
// de l'appelant pour une tâche "running" (voir Runner.Cancel, qui interrompt
// le process puis appelle CancelTask).
func (s *Store) CancelTask(id string) (Task, error) {
	t, ok := s.GetTask(id)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}
	switch t.Status {
	case "waiting", "running", "review":
	default:
		return Task{}, fmt.Errorf("task cannot be cancelled from its current status")
	}
	return s.UpdateTask(id, func(t *Task) { t.Status = "cancelled"; t.Unread = false })
}

// ReopenTask remet une tâche en revue ("review"). Autorisé depuis
// accepted/cancelled. Le merge déjà fait dans la branche du chantier n'est pas
// annulé : la prochaine acceptation fusionnera les nouveaux commits.
func (s *Store) ReopenTask(id string) (Task, error) {
	t, ok := s.GetTask(id)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}
	switch t.Status {
	case "accepted", "cancelled":
	default:
		return Task{}, fmt.Errorf("only an accepted or refused task can be reopened")
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
	m, ok := s.appendMessage(taskID, author, authorName, text)
	if !ok {
		return Message{}, Task{}, fmt.Errorf("task not found")
	}
	s.recomputeCard(s.Tasks[taskID].CardID)
	if err := s.save(); err != nil {
		return Message{}, Task{}, err
	}
	return m, s.Tasks[taskID], nil
}

// appendMessage ajoute un message au fil d'une tâche et met à jour son
// compteur, sans recalcul dérivé ni sauvegarde : à n'appeler que verrou tenu
// (AddMessage) ou pendant le chargement, avant que le Store ne soit visible.
func (s *Store) appendMessage(taskID, author, authorName, text string) (Message, bool) {
	t, ok := s.Tasks[taskID]
	if !ok {
		return Message{}, false
	}
	s.NextMessageN++
	m := Message{ID: fmt.Sprintf("m%d", s.NextMessageN), TaskID: taskID, Author: author, AuthorName: authorName, Text: text, CreatedAt: time.Now().UTC()}
	s.Messages[taskID] = append(s.Messages[taskID], m)
	t.MessagesCount = len(s.Messages[taskID])
	t.UpdatedAt = time.Now().UTC()
	s.Tasks[taskID] = t
	return m, true
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

var validCli = map[string]bool{
	"agy":     true,
	"claude":  true,
	"codex":   true,
	"copilot": true,
	"fake":    true,
}

const validCliDescription = "claude, codex, copilot, agy or fake"

// AddAgent crée un agent. name et cli sont obligatoires ; cli doit être
// claude, codex, copilot, agy ou fake ; l'identifiant est le slug du nom
// (unique).
func (s *Store) AddAgent(name, emoji, color, cli, model, contextPrompt string) (Agent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(name) == "" || cli == "" {
		return Agent{}, fmt.Errorf("name and cli are required")
	}
	if !validCli[cli] {
		return Agent{}, fmt.Errorf("invalid cli: must be %s", validCliDescription)
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
			return Agent{}, fmt.Errorf("invalid cli: must be %s", validCliDescription)
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
