package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// claudeAllowedTools est la liste figée d'outils autorisés à l'agent claude.
// Jamais de Bash(git push:*) ici : le push n'est déclenché que par Ship (git.go).
// Socle : uniquement ce qui ne dépend d'aucun langage. Aucune commande de
// build, de test ou de format n'y figure, quel que soit le langage : elles
// vont toutes dans Project.AllowedTools, sinon un projet Python hérite d'une
// allowlist écrite pour Go. Le git est en lecture seule et reste dans le socle
// parce qu'il est le modèle même du produit (une tâche vit dans un worktree),
// pas un choix de stack.
// Tout ce qui touche au push reste hors de cette liste (invariant 1) et est en
// plus explicitement refusé par claudeDeniedTools.
const claudeAllowedTools = "Read,Edit,Write,Glob,Grep,WebFetch," +
	"Bash(git status:*),Bash(git diff:*),Bash(git log:*),Bash(git show:*)," +
	"Bash(ls:*),Bash(cat:*),Bash(mkdir:*)"

// claudeDeniedTools est le refus figé, passé en --disallowedTools : le refus
// l'emporte sur l'autorisation, donc aucune entrée de Project.AllowedTools ne
// peut ouvrir ces portes. Deux familles : ce qui peut pousser (invariant 1) et
// les fichiers qui pilotent l'agent lui-même, qu'un agent capable de les écrire
// utiliserait pour s'accorder de nouveaux droits au run suivant.
// Ce refus attrape la maladresse, pas la détermination (Bash(sh:*) le
// contournerait) : le confinement réel reste le worktree dédié, la branche
// jetable et la revue humaine avant toute sortie. À faire grandir quand un
// nouveau binaire sortant apparaît, jamais à réduire.
const claudeDeniedTools = "Bash(git push:*),Bash(gh:*),Bash(glab:*)," +
	"Edit(.claude/settings*.json),Write(.claude/settings*.json)," +
	"Edit(.claude/hooks/**),Write(.claude/hooks/**)"

// Copilot deny rules override the selectively allowed read, write, and shell
// tools. They keep outbound Git operations and agent-controlled settings out
// of a task even when the CLI runs unattended.
const copilotDeniedTools = "shell(git push),shell(gh:*),shell(glab:*)," +
	"write(.github/copilot-instructions.md)," +
	"write(.github/copilot/settings.json)," +
	"write(.github/copilot/settings.local.json)"

// claudeTools assemble la liste d'outils réellement passée au CLI : le socle,
// puis les outils accordés au projet par l'humain dans les réglages.
func claudeTools(projectTools []string) string {
	extra := NormalizeAllowedTools(projectTools)
	if len(extra) == 0 {
		return claudeAllowedTools
	}
	return claudeAllowedTools + "," + strings.Join(extra, ",")
}

// commandLogLimit plafonne Task.CommandLog : un historique de débogage, pas
// un audit exhaustif, donc les entrées les plus anciennes tombent en premier.
const commandLogLimit = 500

// appendCommandLog ajoute une commande (déjà résumée par summarizeToolUse) à
// l'historique d'une tâche, plafonné à commandLogLimit.
func appendCommandLog(log []CommandLogEntry, text string) []CommandLogEntry {
	log = append(log, CommandLogEntry{Text: text, At: time.Now().UTC()})
	if len(log) > commandLogLimit {
		log = log[len(log)-commandLogLimit:]
	}
	return log
}

// procHandle représente un processus (ou une simulation) en cours pour une tâche.
type procHandle struct {
	cmd         *exec.Cmd
	cancel      context.CancelFunc // utilisé par l'adaptateur fake (pas de process réel)
	interrupted atomic.Bool
	done        chan struct{}
}

// Runner exécute au plus un agent par tâche à la fois. Les messages envoyés
// pendant qu'un agent tourne sont mis en file (pending) et transmis
// automatiquement à la fin de l'exécution en cours.
type Runner struct {
	mu      sync.Mutex
	procs   map[string]*procHandle
	pending map[string][]string
	store   *Store
	hub     *Hub

	// runWG compte les agents en cours. Interrompre un agent ne fait que lui
	// demander de s'arrêter : sa goroutine peut encore écrire dans le worktree
	// après le retour d'Interrupt. waitTasks donne aux tests le point d'attente
	// qui manque, sans quoi t.TempDir() supprime un répertoire pendant qu'un
	// agent y écrit encore.
	runWG sync.WaitGroup
}

// waitTasks attend la fin des agents en cours. Réservé aux tests : en
// production personne n'attend un agent, on l'interrompt.
func (r *Runner) waitTasks() { r.runWG.Wait() }

// NewRunner crée un runner lié à un store et un hub SSE.
func NewRunner(store *Store, hub *Hub) *Runner {
	return &Runner{procs: map[string]*procHandle{}, pending: map[string][]string{}, store: store, hub: hub}
}

// Message ajoute un message utilisateur à la tâche et le transmet à l'agent :
// immédiatement s'il est libre, sinon en file d'attente (queued=true), vidée
// automatiquement dès la fin de l'exécution en cours.
func (r *Runner) Message(taskID, text string) (queued bool, err error) {
	r.mu.Lock()
	_, running := r.procs[taskID]
	if running {
		r.pending[taskID] = append(r.pending[taskID], text)
	}
	r.mu.Unlock()

	if running {
		authorName := r.store.GetSettings().DisplayName
		msg, updated, err := r.store.AddMessage(taskID, "user", authorName, text)
		if err != nil {
			return false, err
		}
		r.publishMessage(msg)
		r.publishTask(updated)
		return true, nil
	}
	return false, r.Start(taskID, false, text)
}

// RunningCount retourne le nombre d'agents en cours d'exécution. Compte des
// process réels, pas des statuts persistés : un "running" laissé par un
// redémarrage brutal ne bloque donc rien.
func (r *Runner) RunningCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.procs)
}

func (r *Runner) publishTask(t Task)       { r.hub.Publish(Event{Name: "task", Data: t}) }
func (r *Runner) publishMessage(m Message) { r.hub.Publish(Event{Name: "message", Data: m}) }
func (r *Runner) publishAgents()           { r.hub.Publish(Event{Name: "agents", Data: r.store.ListAgents()}) }
func (r *Runner) publishTokens() {
	r.hub.Publish(Event{Name: "tokens", Data: r.store.TokensSnapshot()})
}
func (r *Runner) publishCards(projectID string) {
	r.hub.Publish(Event{Name: "cards", Data: r.store.CardsByProject(projectID)})
}
func (r *Runner) publishProject(p ProjectOut) { r.hub.Publish(Event{Name: "project", Data: p}) }
func (r *Runner) publishWorkspace(w WorkspaceStatus) {
	r.hub.Publish(Event{Name: "workspace", Data: w})
}
func (r *Runner) publishSettings(s Settings) { r.hub.Publish(Event{Name: "settings", Data: s}) }
func (r *Runner) publishUpdate(u UpdateStatus) {
	r.hub.Publish(Event{Name: "update", Data: u})
}
func (r *Runner) publishActivity(taskID string, line *string) {
	r.hub.Publish(Event{Name: "activity", Data: map[string]any{"taskId": taskID, "line": line}})
}
func (r *Runner) publishTaskDeleted(taskID, cardID, projectID string) {
	r.hub.Publish(Event{Name: "taskDeleted", Data: TaskDeletedEvent{TaskID: taskID, CardID: cardID, ProjectID: projectID}})
}
func (r *Runner) publishCardDeleted(cardID, projectID string) {
	r.hub.Publish(Event{Name: "cardDeleted", Data: CardDeletedEvent{CardID: cardID, ProjectID: projectID}})
}
func (r *Runner) publishProjectDeleted(projectID string) {
	r.hub.Publish(Event{Name: "projectDeleted", Data: ProjectDeletedEvent{ProjectID: projectID}})
}

// Start lance (ou relance) l'agent d'une tâche. initial=true correspond au
// lancement initial (text = titre + description, aucun Message ajouté) ;
// initial=false correspond à un nouveau message utilisateur (text = son
// contenu, ajouté comme Message puis transmis à l'agent). L'AuthorName du
// message utilisateur est le displayName des Settings (vide si non renseigné,
// le frontend affiche alors "Vous"/"You").
func (r *Runner) Start(taskID string, initial bool, text string) error {
	r.mu.Lock()
	if _, exists := r.procs[taskID]; exists {
		r.mu.Unlock()
		return fmt.Errorf("an agent is already running for this task")
	}
	r.mu.Unlock()

	task, ok := r.store.GetTask(taskID)
	if !ok {
		return fmt.Errorf("task not found")
	}
	agent, ok := r.store.GetAgent(task.AgentID)
	if !ok {
		return fmt.Errorf("agent not found")
	}

	task, err := r.store.UpdateTask(taskID, func(t *Task) { t.Status = "running"; t.LiveActivity = nil })
	if err != nil {
		return err
	}
	r.publishTask(task)
	r.publishAgents()

	cliInput := text
	if !initial {
		authorName := r.store.GetSettings().DisplayName
		msg, updated, err := r.store.AddMessage(taskID, "user", authorName, text)
		if err != nil {
			return err
		}
		r.publishMessage(msg)
		task = updated
		r.publishTask(task)

		cliInput = r.prepareCliInput(task, text)
	}

	handle := &procHandle{done: make(chan struct{})}
	r.mu.Lock()
	r.procs[taskID] = handle
	r.mu.Unlock()

	r.runWG.Add(1)
	go r.run(task, agent, handle, cliInput)
	return nil
}

// prepareCliInput construit le texte réellement envoyé au CLI. Avec une
// session à reprendre, le texte brut suffit. Sans session (codex, tâche
// réassignée), le CLI ne connaît rien : on rejoue le titre et l'historique de
// la conversation, en excluant les messages utilisateur en fin de fil (ce sont
// eux qui forment la nouvelle instruction, déjà portée par text).
func (r *Runner) prepareCliInput(task Task, text string) string {
	if task.SessionID != "" {
		return text
	}
	msgs := r.store.GetMessages(task.ID)
	for len(msgs) > 0 && msgs[len(msgs)-1].Author == "user" {
		msgs = msgs[:len(msgs)-1]
	}
	transcript := buildTranscript(msgs)
	if transcript == "" {
		return contextualizeCliInput(task.Title, text)
	}
	return "Task: " + task.Title +
		"\n\nPrevious conversation (for context):\n" + transcript +
		"\n\nNew instruction:\n" + text
}

// startQueued relance l'agent avec les messages accumulés pendant l'exécution
// précédente (déjà présents dans le fil : aucun Message n'est ajouté ici).
func (r *Runner) startQueued(taskID, text string) {
	task, ok := r.store.GetTask(taskID)
	if !ok {
		return
	}
	agent, ok := r.store.GetAgent(task.AgentID)
	if !ok {
		return
	}
	task, err := r.store.UpdateTask(taskID, func(t *Task) { t.Status = "running"; t.LiveActivity = nil })
	if err != nil {
		return
	}
	r.publishTask(task)
	r.publishCards(task.ProjectID)
	r.publishAgents()

	handle := &procHandle{done: make(chan struct{})}
	r.mu.Lock()
	if _, exists := r.procs[taskID]; exists {
		r.mu.Unlock()
		return
	}
	r.procs[taskID] = handle
	r.mu.Unlock()

	r.runWG.Add(1)
	go r.run(task, agent, handle, r.prepareCliInput(task, text))
}

// markerPrefixes liste les préfixes des messages marqueurs posés par le
// backend (le frontend les remplace par une ligne système localisée). Ce ne
// sont pas des tours de conversation : ils sont exclus du transcript rejoué.
var markerPrefixes = []string{"[reassigned:", "[accepted:", "[auto-accepted:", "[merge-conflict:", "[tool-denied:"}

func isMarkerMessage(text string) bool {
	for _, prefix := range markerPrefixes {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

// buildTranscript rejoue la conversation pour un agent démarré sans session :
// une ligne par message, marqueurs système exclus, chaque message tronqué à
// 600 caractères et l'ensemble limité aux ~6000 derniers caractères (les
// messages les plus récents priment).
func buildTranscript(msgs []Message) string {
	const perMsg, total = 600, 6000
	var lines []string
	for _, m := range msgs {
		if isMarkerMessage(m.Text) {
			continue
		}
		who := "assistant"
		if m.Author == "user" {
			who = "user"
		}
		if m.AuthorName != "" {
			who = m.AuthorName
		}
		lines = append(lines, "["+who+"] "+truncate(strings.TrimSpace(m.Text), perMsg))
	}
	// Garder la fin (messages récents) si l'ensemble dépasse le budget.
	out := strings.Join(lines, "\n")
	for len(out) > total && len(lines) > 1 {
		lines = lines[1:]
		out = strings.Join(lines, "\n")
	}
	return out
}

// contextualizeCliInput préfixe text par un rappel minimal de la tâche
// ("Task: <title>") : utilisé au lancement initial et chaque fois qu'un
// message est envoyé sans session CLI à reprendre (départ frais).
func contextualizeCliInput(title, text string) string {
	if text == "" {
		return "Task: " + title
	}
	return "Task: " + title + "\n\n" + text
}

// Interrupt arrête l'agent en cours pour une tâche : SIGINT au groupe de
// process, puis SIGKILL après 5 s si nécessaire. Le statut passe
// immédiatement à "review".
func (r *Runner) Interrupt(taskID string) (Task, error) {
	r.mu.Lock()
	handle, ok := r.procs[taskID]
	r.mu.Unlock()
	if !ok {
		return Task{}, fmt.Errorf("no agent is running for this task")
	}
	handle.interrupted.Store(true)
	r.mu.Lock()
	delete(r.pending, taskID) // interrompre = stopper : on ne relance pas la file
	r.mu.Unlock()

	task, err := r.store.UpdateTask(taskID, func(t *Task) { t.Status = "review" })
	if err != nil {
		return Task{}, err
	}
	r.publishTask(task)
	r.publishCards(task.ProjectID)
	r.publishAgents()

	go killProcessGroup(handle)
	return task, nil
}

// Cancel annule une tâche (autorisé depuis running/review/ready, voir
// Store.CancelTask). Si elle est en cours d'exécution, l'agent est interrompu
// au préalable (même mécanique qu'Interrupt) avant de passer au statut final
// "cancelled" (et non "review").
func (r *Runner) Cancel(taskID string) (Task, error) {
	task, ok := r.store.GetTask(taskID)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}

	var handle *procHandle
	if task.Status == "running" {
		r.mu.Lock()
		h, hasProc := r.procs[taskID]
		delete(r.pending, taskID) // annuler = stopper : on ne relance pas la file
		r.mu.Unlock()
		if hasProc {
			h.interrupted.Store(true)
			handle = h
		}
	}

	updated, err := r.store.CancelTask(taskID)
	if err != nil {
		return Task{}, err
	}
	r.publishTask(updated)
	r.publishCards(updated.ProjectID)
	r.publishAgents()

	if handle != nil {
		go killProcessGroup(handle)
	}
	return updated, nil
}

// deleteTaskQuiet supprime une tâche sans publier aucun événement SSE :
// interrompt l'agent au préalable s'il tourne encore (même mécanique que
// Cancel), retire le worktree du dépôt d'origine (best-effort, jamais la
// branche : elle peut avoir été poussée), puis supprime la tâche et ses
// messages. Utilisée directement par DeleteTask (qui publie ensuite ses
// événements) et en cascade par DeleteCard/DeleteProject (qui choisissent
// eux-mêmes quels événements publier une fois la cascade terminée).
func (r *Runner) deleteTaskQuiet(taskID string) (Task, error) {
	task, ok := r.store.GetTask(taskID)
	if !ok {
		return Task{}, fmt.Errorf("task not found")
	}

	r.mu.Lock()
	handle, hasProc := r.procs[taskID]
	delete(r.pending, taskID) // supprimer = stopper : on ne relance pas la file
	r.mu.Unlock()
	if hasProc {
		handle.interrupted.Store(true)
		go killProcessGroup(handle)
	}

	if project, ok := r.store.GetProject(task.ProjectID); ok {
		for _, repo := range project.Repos {
			if repo.Name == task.RepoName {
				RemoveWorktree(repo.Path, task.WorktreeDir)
				break
			}
		}
	}

	return r.store.DeleteTask(taskID)
}

// DeleteTask supprime une tâche isolément et publie les événements SSE requis
// (taskDeleted, cards, agents, tokens).
func (r *Runner) DeleteTask(taskID string) error {
	task, err := r.deleteTaskQuiet(taskID)
	if err != nil {
		return err
	}
	r.publishTaskDeleted(taskID, task.CardID, task.ProjectID)
	r.publishCards(task.ProjectID)
	r.publishAgents()
	r.publishTokens()
	return nil
}

// IsRunning indique qu'un process d'agent est en cours pour cette tâche.
// Utilisé avant toute décision automatique sur son worktree (voir
// l'acceptation automatique des branches déjà fusionnées).
func (r *Runner) IsRunning(taskID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.procs[taskID]
	return ok
}

// removeCardWorktrees retire les worktrees des branches de chantier d'une
// carte (un par dépôt touché). Best-effort, et la branche n'est JAMAIS
// supprimée : elle a pu être poussée ou fusionnée.
func (r *Runner) removeCardWorktrees(card Card) {
	project, ok := r.store.GetProject(card.ProjectID)
	if !ok {
		return
	}
	for _, b := range card.Branches {
		if repo, ok := repoByName(project, b.RepoName); ok {
			RemoveWorktree(repo.Path, b.WorktreeDir)
		}
	}
}

// DeleteCard supprime une carte (chantier) : cascade sur ses tâches (même
// traitement que deleteTaskQuiet, chacune), retire les worktrees de ses
// branches de chantier, puis supprime la carte elle-même.
// Publie taskDeleted par tâche, cards, agents, tokens, puis cardDeleted.
func (r *Runner) DeleteCard(cardID string) error {
	card, ok := r.store.GetCard(cardID)
	if !ok {
		return fmt.Errorf("card not found")
	}
	tasks := r.store.TasksByCard(cardID)
	for _, t := range tasks {
		_, _ = r.deleteTaskQuiet(t.ID) // best-effort : la cascade continue même si une tâche échoue
	}
	r.removeCardWorktrees(card)
	if _, err := r.store.DeleteCard(cardID); err != nil {
		return err
	}
	for _, t := range tasks {
		r.publishTaskDeleted(t.ID, cardID, card.ProjectID)
	}
	r.publishCards(card.ProjectID)
	r.publishAgents()
	r.publishTokens()
	r.publishCardDeleted(cardID, card.ProjectID)
	return nil
}

// DeleteProject supprime un projet : cascade sur ses chantiers puis leurs
// tâches (même traitement que deleteTaskQuiet), puis supprime le projet.
// Ne publie que projectDeleted (le front recharge l'état complet à cet
// événement, plus simple et sûr que de rejouer toute la cascade côté SSE).
func (r *Runner) DeleteProject(projectID string) error {
	if _, ok := r.store.GetProject(projectID); !ok {
		return fmt.Errorf("project not found")
	}
	for _, c := range r.store.CardsByProject(projectID) {
		for _, t := range r.store.TasksByCard(c.ID) {
			_, _ = r.deleteTaskQuiet(t.ID)
		}
		r.removeCardWorktrees(c)
		_, _ = r.store.DeleteCard(c.ID)
	}
	if _, err := r.store.DeleteProject(projectID); err != nil {
		return err
	}
	r.publishProjectDeleted(projectID)
	return nil
}

func killProcessGroup(handle *procHandle) {
	if handle.cmd != nil && handle.cmd.Process != nil {
		killGroup(handle.cmd, handle.done)
		return
	}
	if handle.cancel != nil {
		handle.cancel()
	}
}

// killGroup envoie SIGINT au groupe de process d'une commande, puis SIGKILL si
// elle n'a pas rendu la main au bout de 5 s. done doit être fermé quand le
// process est réellement mort. Partagé par les agents (procHandle) et les
// recettes manuelles (voir preview.go) : c'est le même geste, et il doit viser
// le groupe, pas le seul shell lancé par Sillage.
func killGroup(cmd *exec.Cmd, done <-chan struct{}) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	_ = syscall.Kill(-pid, syscall.SIGINT)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	}
}

// run exécute l'adaptateur adéquat puis finalise la tâche (checks, compteurs).
// À la fin, les messages arrivés pendant l'exécution (file pending) relancent
// automatiquement l'agent.
func (r *Runner) run(task Task, agent Agent, handle *procHandle, cliInput string) {
	// Déclaré en premier, donc exécuté en dernier (LIFO) : l'agent n'est compté
	// terminé qu'une fois tout son travail écrit. Un message mis en file relance
	// un agent depuis le defer ci-dessous, hors de ce compteur : un test qui
	// s'appuie sur la file doit se synchroniser lui-même.
	defer r.runWG.Done()
	defer func() {
		r.mu.Lock()
		delete(r.procs, task.ID)
		queued := r.pending[task.ID]
		delete(r.pending, task.ID)
		r.mu.Unlock()
		close(handle.done)
		if len(queued) > 0 && !handle.interrupted.Load() {
			go r.startQueued(task.ID, strings.Join(queued, "\n\n"))
		}
	}()

	project, _ := r.store.GetProject(task.ProjectID)
	card, _ := r.store.GetCard(task.CardID)

	var runErr error
	switch agent.Cli {
	case "claude":
		runErr = r.runClaude(&task, agent, project, card, handle, cliInput)
	case "codex":
		runErr = r.runCodex(&task, agent, project, card, handle, cliInput)
	case "copilot":
		runErr = r.runCopilot(&task, agent, project, card, handle, cliInput)
	case "agy":
		runErr = r.runAntigravity(&task, agent, project, card, handle, cliInput)
	case "fake":
		runErr = r.runFake(&task, agent, handle)
	default:
		runErr = fmt.Errorf("unknown cli: %s", agent.Cli)
	}

	if runErr != nil {
		msg, updated, err := r.store.AddMessage(task.ID, "agent", agent.Name, "⚠️ "+runErr.Error())
		if err == nil {
			r.publishMessage(msg)
			task = updated
		}
	}

	r.finalize(task.ID)
}

// finalize exécute le check éventuel du projet et clôt l'exécution.
func (r *Runner) finalize(taskID string) {
	task, ok := r.store.GetTask(taskID)
	if !ok {
		return
	}
	project, _ := r.store.GetProject(task.ProjectID)

	checks := []Check{}
	if project.CheckCmd != "" {
		ok, _ := runCheck(task.WorktreeDir, project.CheckCmd, 120*time.Second)
		checks = []Check{{Label: project.CheckCmd, Ok: ok}}
	}

	filesCount, docsCount := 0, 0
	if files, err := Diff(task.WorktreeDir, task.Base); err == nil {
		filesCount = len(files)
		for _, f := range files {
			if isDocFile(f.Path) {
				docsCount++
			}
		}
	}
	commitsCount := 0
	if commits, err := Commits(task.WorktreeDir, task.Base); err == nil {
		commitsCount = len(commits)
	}

	updated, err := r.store.UpdateTask(taskID, func(t *Task) {
		if t.Status == "running" {
			t.Status = "review"
			t.Unread = true
		}
		t.LiveActivity = nil
		t.Checks = checks
		t.FilesCount = filesCount
		t.DocsCount = docsCount
		t.CommitsCount = commitsCount
	})
	if err != nil {
		return
	}
	r.publishTask(updated)
	r.publishCards(updated.ProjectID)
	r.publishAgents()
	r.publishActivity(updated.ID, nil)
}

func runCheck(dir, command string, timeout time.Duration) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return err == nil, string(out)
}

func isDocFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown", ".txt", ".rst":
		return true
	default:
		return false
	}
}

func isImageFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

// contextParts assemble les blocs "Project context:"/"Workstream context:"
// (chacun omis s'il est vide, dans cet ordre). Partagé par l'adaptateur
// claude (buildSystemPrompt) et l'adaptateur codex (préfixe de prompt).
func contextParts(projectContext, workstreamContext string) []string {
	var parts []string
	if projectContext != "" {
		parts = append(parts, "Project context:\n"+projectContext)
	}
	if workstreamContext != "" {
		parts = append(parts, "Workstream context:\n"+workstreamContext)
	}
	return parts
}

// buildSystemPrompt combine le contexte de l'agent, celui du projet et celui
// du chantier (Card) pour --append-system-prompt (adaptateur claude) : chaque
// bloc n'est ajouté que s'il est non vide, séparés par des lignes vides.
// Retourne une chaîne vide si tout est vide (pas de flag ajouté).
func buildSystemPrompt(agentContext, projectContext, workstreamContext string) string {
	parts := contextParts(projectContext, workstreamContext)
	if agentContext != "" {
		parts = append([]string{agentContext}, parts...)
	}
	return strings.Join(parts, "\n\n")
}

// prefixAgentContext is used by one-shot CLIs that do not expose a dedicated
// system-prompt option. It keeps all Sillage context ahead of the task text.
func prefixAgentContext(agent Agent, project Project, card Card, cliInput string) string {
	context := buildSystemPrompt(agent.ContextPrompt, project.ContextPrompt, card.ContextPrompt)
	if context == "" {
		return cliInput
	}
	return context + "\n\n---\n\n" + cliInput
}

// summarizeToolUse construit la ligne d'activité affichée pour un tool_use claude.
func summarizeToolUse(name string, input map[string]any) string {
	for _, key := range []string{"file_path", "path", "command", "pattern", "query", "url"} {
		if v, ok := input[key]; ok {
			return name + " · " + truncate(fmt.Sprintf("%v", v), 80)
		}
	}
	b, _ := json.Marshal(input)
	return name + " · " + truncate(string(b), 80)
}

// --- Adaptateur claude ---

type claudeContentBlock struct {
	Type  string         `json:"type"`
	Text  string         `json:"text"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`

	// Blocs tool_use / tool_result : ID identifie l'appel côté assistant,
	// ToolUseID le rattache au résultat renvoyé dans l'enveloppe "user".
	// Content est tantôt une chaîne, tantôt une liste de blocs (voir
	// claudeToolResultText).
	ID        string          `json:"id"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
	IsError   bool            `json:"is_error"`
}

// claudeToolResultText aplatit le contenu d'un tool_result, que le CLI écrit
// soit en chaîne simple, soit en liste de blocs typés.
func claudeToolResultText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	var blocks []struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, block := range blocks {
		if block.Text != "" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// claudePermissionPhrases : formulations par lesquelles le CLI signale un outil
// refusé faute d'autorisation. Le flux ne porte pas de code d'erreur dédié, la
// détection se fait donc sur le texte du tool_result, et cette liste est à
// revoir si le CLI reformule. Une phrase manquée ne coûte que le marqueur : le
// refus lui-même reste appliqué par le CLI.
var claudePermissionPhrases = []string{
	"requested permissions",
	"permission to use",
	"haven't granted",
	"have not granted",
}

// isPermissionDenial distingue un outil refusé d'une erreur d'exécution
// ordinaire (test qui échoue, fichier absent), qui ne doit rien afficher.
func isPermissionDenial(text string) bool {
	lower := strings.ToLower(text)
	for _, phrase := range claudePermissionPhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

type claudeEnvelope struct {
	Type      string `json:"type"`
	Subtype   string `json:"subtype"`
	SessionID string `json:"session_id"`
	Message   *struct {
		Content []claudeContentBlock `json:"content"`
	} `json:"message"`
	Result       string  `json:"result"`
	TotalCostUSD float64 `json:"total_cost_usd"`
	Usage        *struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	} `json:"usage"`
	IsError bool `json:"is_error"`
}

func (r *Runner) runClaude(task *Task, agent Agent, project Project, card Card, handle *procHandle, cliInput string) error {
	args := []string{
		"-p", "--output-format", "stream-json", "--verbose",
		"--permission-mode", "acceptEdits",
		"--allowedTools", claudeTools(project.AllowedTools),
		"--disallowedTools", claudeDeniedTools,
	}
	if agent.Model != "" {
		args = append(args, "--model", agent.Model)
	}
	if systemPrompt := buildSystemPrompt(agent.ContextPrompt, project.ContextPrompt, card.ContextPrompt); systemPrompt != "" {
		args = append(args, "--append-system-prompt", systemPrompt)
	}
	if task.SessionID != "" {
		args = append(args, "--resume", task.SessionID)
	}
	args = append(args, cliInput)

	cmd := exec.Command("claude", args...)
	cmd.Dir = task.WorktreeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude: %w", err)
	}
	handle.cmd = cmd

	lastAgentText := ""
	// pendingTools rattache un tool_result à l'appel qui l'a produit : le
	// résultat ne porte que l'identifiant, pas le nom de l'outil demandé.
	// deniedTools garde un seul marqueur par outil et par exécution, sinon un
	// agent qui réessaie trois fois inonde le fil.
	pendingTools := map[string]string{}
	deniedTools := map[string]bool{}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var env claudeEnvelope
		if err := json.Unmarshal([]byte(line), &env); err != nil {
			continue
		}
		switch env.Type {
		case "system":
			if env.Subtype == "init" && env.SessionID != "" {
				if t, err := r.store.UpdateTask(task.ID, func(t *Task) { t.SessionID = env.SessionID }); err == nil {
					*task = t
				}
			}
		case "assistant":
			if env.Message == nil {
				continue
			}
			for _, block := range env.Message.Content {
				switch block.Type {
				case "text":
					if strings.TrimSpace(block.Text) == "" {
						continue
					}
					lastAgentText = block.Text
					msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, block.Text)
					if err == nil {
						r.publishMessage(msg)
						*task = t
						r.publishTask(*task)
					}
				case "tool_use":
					line := summarizeToolUse(block.Name, block.Input)
					if block.ID != "" {
						pendingTools[block.ID] = line
					}
					if t, err := r.store.UpdateTask(task.ID, func(t *Task) {
						t.LiveActivity = &line
						t.CommandLog = appendCommandLog(t.CommandLog, line)
					}); err == nil {
						*task = t
						r.publishActivity(task.ID, &line)
						r.publishTask(*task)
					}
				}
			}
		case "user":
			// Retours d'outils. Seul un refus de permission nous intéresse : une
			// erreur d'exécution ordinaire (test rouge, fichier absent) fait
			// partie du travail de l'agent et n'a rien à faire dans le fil.
			if env.Message == nil {
				continue
			}
			for _, block := range env.Message.Content {
				if block.Type != "tool_result" || !block.IsError {
					continue
				}
				if !isPermissionDenial(claudeToolResultText(block.Content)) {
					continue
				}
				// Sans le nom de l'outil, le marqueur n'apprendrait rien à
				// l'humain sur ce qu'il doit autoriser : mieux vaut rien.
				denied := strings.ReplaceAll(pendingTools[block.ToolUseID], "]", "")
				if denied == "" || deniedTools[denied] {
					continue
				}
				deniedTools[denied] = true
				msg, t, err := r.store.AddMessage(task.ID, "agent", "", "[tool-denied:"+denied+"]")
				if err == nil {
					r.publishMessage(msg)
					*task = t
					r.publishTask(*task)
				}
			}
		case "result":
			inputTok, outputTok := 0, 0
			if env.Usage != nil {
				inputTok = env.Usage.InputTokens + env.Usage.CacheCreationInputTokens
				outputTok = env.Usage.OutputTokens
			}
			cost := env.TotalCostUSD
			if t, err := r.store.UpdateTask(task.ID, func(t *Task) {
				t.Tokens.Input += inputTok
				t.Tokens.Output += outputTok
				t.Tokens.CostUsd += cost
			}); err == nil {
				*task = t
			}
			r.publishTokens()
			if env.Result != "" && env.Result != lastAgentText {
				msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, env.Result)
				if err == nil {
					r.publishMessage(msg)
					*task = t
					r.publishTask(*task)
				}
			}
			if env.IsError {
				return fmt.Errorf("the agent reported an error")
			}
		}
	}

	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if waitErr == nil && scanErr != nil {
		waitErr = scanErr
	}
	if waitErr != nil && !handle.interrupted.Load() {
		msg := strings.TrimSpace(stderrBuf.String())
		if msg == "" {
			msg = waitErr.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// --- Codex adapter (best-effort) ---

// codexSharedGitWriteDirs contains mutable Git storage that is shared by all
// linked worktrees. Hooks and repository configuration are intentionally not
// included: making them agent-writable could persist code outside the task's
// sandbox and affect later Git commands run by Sillage or the user.
var codexSharedGitWriteDirs = []string{"objects", "refs", "logs", "reftable", "lfs", "rr-cache"}

// codexGitWritableDirs returns only the external Git metadata directories that
// Codex needs for normal local writes such as commit, rebase, and stash. A
// linked worktree keeps its index and HEAD in its own git-dir while objects,
// refs, and reflogs live in the common git-dir; both sides must therefore be
// writable even though they sit outside Sillage's task worktree.
//
// Resolution is best-effort. If the directory is no longer a Git repository,
// Codex still starts and can report the repository problem itself.
func codexGitWritableDirs(worktreeDir string) []string {
	gitDir, err := resolveCodexGitDir(worktreeDir, "--absolute-git-dir")
	if err != nil {
		return nil
	}
	commonDir, err := resolveCodexGitDir(worktreeDir, "--git-common-dir")
	if err != nil {
		return nil
	}
	workspaceDir, err := canonicalExistingDir(worktreeDir)
	if err != nil {
		return nil
	}

	var dirs []string
	seen := make(map[string]bool)
	addExternalDir := func(path string) {
		resolved, resolveErr := canonicalExistingDir(path)
		if resolveErr != nil || isDirWithin(resolved, workspaceDir) || seen[resolved] {
			return
		}
		seen[resolved] = true
		dirs = append(dirs, resolved)
	}

	// The per-worktree git-dir holds the index, HEAD, and operation state.
	addExternalDir(gitDir)
	if gitDir == commonDir {
		// A repository created with --separate-git-dir has no narrower directory
		// that can receive index.lock, so its complete external git-dir is needed.
		return dirs
	}
	for _, name := range codexSharedGitWriteDirs {
		addExternalDir(filepath.Join(commonDir, name))
	}
	return dirs
}

func resolveCodexGitDir(worktreeDir, flag string) (string, error) {
	out, err := runGit(worktreeDir, 10*time.Second, "rev-parse", flag)
	if err != nil {
		return "", err
	}
	path := strings.TrimSuffix(strings.TrimSuffix(out, "\n"), "\r")
	if path == "" {
		return "", fmt.Errorf("git rev-parse %s returned an empty path", flag)
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(worktreeDir, path)
	}
	return canonicalExistingDir(path)
}

func canonicalExistingDir(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", resolved)
	}
	return filepath.Clean(resolved), nil
}

func isDirWithin(path, parent string) bool {
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func codexArgs(worktreeDir, sandbox string, agent Agent, cliInput string) []string {
	args := []string{"exec", "--json", "--sandbox", sandbox, "-C", worktreeDir}
	if sandbox == "workspace-write" {
		for _, dir := range codexGitWritableDirs(worktreeDir) {
			args = append(args, "--add-dir", dir)
		}
	}
	if agent.Model != "" {
		args = append(args, "--model", agent.Model)
	}
	return append(args, cliInput)
}

func (r *Runner) runCodex(task *Task, agent Agent, project Project, card Card, handle *procHandle, cliInput string) error {
	// workspace-write limits source edits to the task worktree plus the narrow
	// Git metadata roots added by codexArgs. Network access remains disabled, so
	// outbound Git operations still go through Ship after human validation.
	// SILLAGE_CODEX_SANDBOX can select another mode (for example,
	// danger-full-access when AppArmor blocks bwrap); Sillage's dedicated
	// worktree and review workflow are then the remaining containment.
	sandbox := os.Getenv("SILLAGE_CODEX_SANDBOX")
	if sandbox == "" {
		sandbox = "workspace-write"
	}
	if blocks := strings.Join(contextParts(project.ContextPrompt, card.ContextPrompt), "\n\n"); blocks != "" {
		cliInput = blocks + "\n\n---\n\n" + cliInput
	}
	args := codexArgs(task.WorktreeDir, sandbox, agent, cliInput)

	cmd := exec.Command("codex", args...)
	cmd.Dir = task.WorktreeDir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderrBuf, rawBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start codex: %w", err)
	}
	handle.cmd = cmd

	parsedAny := false
	var tokenAcc codexTokenAccumulator
	var threadID string
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		rawBuf.WriteString(line)
		rawBuf.WriteByte('\n')
		if strings.TrimSpace(line) == "" {
			continue
		}
		var generic map[string]any
		if err := json.Unmarshal([]byte(line), &generic); err != nil {
			continue
		}
		if text, ok := extractCodexMessage(generic); ok {
			parsedAny = true
			msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, text)
			if err == nil {
				r.publishMessage(msg)
				*task = t
				r.publishTask(*task)
			}
		}
		// token_count/turn.completed portent des TOTAUX CUMULÉS par exécution
		// (pas des deltas) : on retient juste le dernier vu, appliqué une seule
		// fois en fin de process (voir plus bas), pour ne pas sur-compter.
		if tokenAcc.observe(generic) {
			parsedAny = true
		}
		if generic["type"] == "thread.started" {
			if id, ok := generic["thread_id"].(string); ok {
				threadID = id
			}
		}
	}

	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if waitErr == nil && scanErr != nil {
		waitErr = scanErr
	}
	// Le quota de compte (rate_limits) n'est pas porté par le flux --json,
	// seulement par le fichier de session que codex écrit de son côté (même
	// en mode exec) : voir readCodexRateLimits. Best-effort, ne bloque jamais
	// la tâche.
	if windows, ok := readCodexRateLimits(threadID); ok {
		if err := r.store.SetCodexQuota(windows); err == nil {
			r.publishAgents()
		}
	}
	if tokenAcc.found {
		if t, err := r.store.UpdateTask(task.ID, func(t *Task) {
			t.Tokens.Input += tokenAcc.last.Input
			t.Tokens.Output += tokenAcc.last.Output
			t.Tokens.CostUsd += tokenAcc.last.CostUsd
		}); err == nil {
			*task = t
		}
		r.publishTokens()
	}
	if !parsedAny && rawBuf.Len() > 0 {
		msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, strings.TrimSpace(rawBuf.String()))
		if err == nil {
			r.publishMessage(msg)
			*task = t
			r.publishTask(*task)
		}
	}
	if waitErr != nil && !handle.interrupted.Load() {
		msg := strings.TrimSpace(stderrBuf.String())
		if msg == "" {
			msg = waitErr.Error()
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

// --- GitHub Copilot and Antigravity adapters ---

func copilotArgs(agent Agent, cliInput string) []string {
	args := []string{
		"--autopilot",
		"--max-autopilot-continues=10",
		"--no-ask-user",
		"--allow-tool=read,write,shell",
		"--deny-tool=" + copilotDeniedTools,
		"--disable-builtin-mcps",
		"--no-remote",
		"--no-remote-export",
		"--no-auto-update",
		"--no-color",
		"--stream=off",
		"--silent",
	}
	if agent.Model != "" {
		args = append(args, "--model="+agent.Model)
	}
	return append(args, "-p", cliInput)
}

// antigravityArgs construit la ligne de commande d'agy. Trois pièges du CLI :
//
//   - `--print` prend le prompt en valeur (c'est un alias de `--prompt`, pas un
//     booléen). Il doit donc rester en dernier : `--print --sandbox <prompt>`
//     fait exécuter le prompt « --sandbox » et l'agent répond à côté.
//   - en mode sandbox, agy ignore le répertoire de travail du process et
//     retombe sur son espace de travail interne (`~/.gemini/antigravity-cli/scratch`).
//     Sans `--add-dir <worktree>`, la tâche écrit ses fichiers hors du dépôt et
//     le diff reste vide.
//   - le sandbox ne montre que les répertoires ajoutés. Le worktree d'une tâche
//     est un worktree lié : son vrai dossier git est celui du dépôt d'origine,
//     hors du sandbox, donc tout git y échoue par « ce n'est pas un dépôt git ».
//     Le modèle relance alors la commande hors sandbox (`BypassSandbox`), ce que
//     le mode headless auto-refuse : la session s'arrête sans rien produire.
//     D'où `--add-dir <gitCommonDir>`.
//
// Ce dossier git n'ouvre aucune porte (invariant 1) : agy monte un `--add-dir`
// hors espace de travail principal **en lecture seule** (vérifié : `touch` y
// répond « système de fichiers accessible en lecture seulement », et `git
// commit` échoue sur `index.lock`), et son sandbox n'a pas de réseau (la
// résolution DNS y échoue). L'agent gagne donc exactement les lectures git de
// l'allowlist claude (status, log, diff, show), ni écriture de hooks ou de
// config, ni push.
func antigravityArgs(agent Agent, worktreeDir, gitCommonDir, cliInput string) []string {
	args := []string{"--sandbox", "--print-timeout=60m"}
	if worktreeDir != "" {
		args = append(args, "--add-dir", worktreeDir)
	}
	if gitCommonDir != "" {
		args = append(args, "--add-dir", gitCommonDir)
	}
	if agent.Model != "" {
		args = append(args, "--model", agent.Model)
	}
	return append(args, "--print", cliInput)
}

func (r *Runner) runCopilot(task *Task, agent Agent, project Project, card Card, handle *procHandle, cliInput string) error {
	cliInput = prefixAgentContext(agent, project, card, cliInput)
	return r.runTextCLI(task, agent, handle, "copilot", copilotArgs(agent, cliInput), copilotEnv())
}

// antigravityWorktreeNote prévient l'agent de la particularité de son
// répertoire qui l'a déjà fait échouer : une tâche travaille dans un worktree
// git lié, où `.git` est un fichier d'une ligne qui pointe ailleurs. Un modèle
// qui explore essaie de l'ouvrir ; agy traite ce chemin comme sensible et exige
// une confirmation, impossible en mode headless, donc la session s'arrête sur
// le champ sans rien produire. Aucune règle d'autorisation générale ne couvre
// le cas : `permissions.allow` n'accepte pas de motif de chemin
// (`read_file(<dossier>/*)` ne correspond à rien), seulement des chemins
// exacts, qui changent à chaque tâche. Le prévenir coûte deux lignes de prompt.
//
// La note couvre aussi la sortie de bac à sable, qui tue la session de la même
// façon : une commande bloquée dans le bac à sable (le réseau y est coupé)
// pousse le modèle à la relancer hors bac à sable, ce que le mode headless
// auto-refuse. Mieux vaut qu'il rapporte l'échec que de finir muet.
const antigravityWorktreeNote = "Environment: your working directory is a linked git worktree, so `.git` is a one-line pointer file, not a directory. Never open it: that read is refused and ends your session without any answer. Read-only git commands (`git status`, `git log`, `git diff`, `git show`) work here; `git add`, `git commit` and anything writing to git do not, and you never need them: your work is committed for you. Your sandbox also has no network access. Never re-run a command outside the sandbox (no bypass-sandbox, no unsandboxed execution): such a request is auto-denied and ends your session without any answer. Report the failure in your answer instead."

func (r *Runner) runAntigravity(task *Task, agent Agent, project Project, card Card, handle *procHandle, cliInput string) error {
	cliInput = prefixAgentContext(agent, project, card, antigravityWorktreeNote+"\n\n"+cliInput)
	// Best-effort : sans le dossier git commun, agy travaille quand même sur les
	// fichiers, il perd seulement les commandes git.
	gitCommonDir, _ := GitCommonDir(task.WorktreeDir)
	args := antigravityArgs(agent, task.WorktreeDir, gitCommonDir, cliInput)
	return r.runTextCLI(task, agent, handle, "agy", args, nil)
}

// copilotEnv strips a classic GitHub PAT (the "ghp_" prefix) from GITHUB_TOKEN
// before it reaches the copilot binary. The CLI refuses to start at all when
// it finds one ("Classic Personal Access Tokens are not supported"), even
// though this token is never needed inside the sandbox: copilotDeniedTools
// already blocks git push/gh/glab, and Sillage itself never reads this
// variable. Leaving it set otherwise breaks every copilot task on a machine
// where GITHUB_TOKEN happens to hold an old classic PAT for unrelated tools.
// Fine-grained PATs and an unset GITHUB_TOKEN pass through untouched, so
// gh's own auth (gh auth login, copilot's /login) keeps working.
func copilotEnv() []string {
	env := os.Environ()
	filtered := make([]string, 0, len(env))
	for _, kv := range env {
		if strings.HasPrefix(kv, "GITHUB_TOKEN=ghp_") {
			continue
		}
		filtered = append(filtered, kv)
	}
	return filtered
}

// runTextCLI runs a non-interactive CLI whose stdout is its final answer.
// Copilot and Antigravity currently do not expose token accounting in this
// mode, so the adapter deliberately records only the response and exit error.
// A nil env makes the child inherit the process environment unchanged.
func (r *Runner) runTextCLI(task *Task, agent Agent, handle *procHandle, binary string, args []string, env []string) error {
	cmd := exec.Command(binary, args...)
	cmd.Dir = task.WorktreeDir
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderrBuf, stdoutBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", binary, err)
	}
	handle.cmd = cmd

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		stdoutBuf.WriteString(scanner.Text())
		stdoutBuf.WriteByte('\n')
	}
	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if waitErr == nil && scanErr != nil {
		waitErr = scanErr
	}

	response := strings.TrimSpace(stdoutBuf.String())
	if response != "" {
		msg, updated, err := r.store.AddMessage(task.ID, "agent", agent.Name, response)
		if err == nil {
			r.publishMessage(msg)
			*task = updated
			r.publishTask(*task)
		}
	}
	if handle.interrupted.Load() {
		return nil
	}
	if waitErr != nil {
		message := strings.TrimSpace(stderrBuf.String())
		if message == "" {
			message = waitErr.Error()
		}
		return fmt.Errorf("%s", message)
	}
	// Un adaptateur texte qui sort en succès sans rien écrire est un échec : la
	// conversation resterait vide, sans le moindre indice. C'est le cas quand
	// agy auto-refuse une confirmation d'outil (mode print) : il explique alors
	// la marche à suivre sur stderr, que le succès de sortie ferait perdre.
	if response == "" {
		if message := strings.TrimSpace(stderrBuf.String()); message != "" {
			return fmt.Errorf("%s", message)
		}
		return fmt.Errorf("%s produced no output", binary)
	}
	return nil
}

func extractCodexMessage(m map[string]any) (string, bool) {
	if msg, ok := m["msg"].(map[string]any); ok {
		if msg["type"] == "agent_message" {
			if text, ok := msg["message"].(string); ok && text != "" {
				return text, true
			}
		}
	}
	if m["type"] == "item.completed" {
		if item, ok := m["item"].(map[string]any); ok && item["type"] == "agent_message" {
			if text, ok := item["text"].(string); ok && text != "" {
				return text, true
			}
		}
	}
	return "", false
}

func extractCodexTokens(m map[string]any) (Tokens, bool) {
	raw, ok := m["token_count"]
	if !ok {
		// Forme récente : {"type":"turn.completed","usage":{"input_tokens":N,...}}.
		if m["type"] == "turn.completed" {
			raw, ok = m["usage"]
		}
		if !ok {
			return Tokens{}, false
		}
	}
	tc, ok := raw.(map[string]any)
	if !ok {
		return Tokens{}, false
	}
	tok := Tokens{}
	found := false
	if v, ok := tc["input_tokens"].(float64); ok {
		tok.Input = int(v)
		found = true
	}
	if v, ok := tc["output_tokens"].(float64); ok {
		tok.Output = int(v)
		found = true
	}
	if v, ok := tc["cost_usd"].(float64); ok {
		tok.CostUsd = v
		found = true
	}
	return tok, found
}

// codexTokenAccumulator retient le dernier total de tokens vu dans le flux
// JSONL d'une exécution codex. Les événements token_count/turn.completed
// portent des TOTAUX CUMULÉS par exécution (pas des deltas) : il ne faut
// jamais les additionner entre eux, seul le dernier compte.
type codexTokenAccumulator struct {
	last  Tokens
	found bool
}

// observe traite un événement JSONL générique et retourne true s'il portait
// des tokens (auquel cas last est mis à jour).
func (a *codexTokenAccumulator) observe(m map[string]any) bool {
	tok, ok := extractCodexTokens(m)
	if !ok {
		return false
	}
	a.last = tok
	a.found = true
	return true
}

// parseCodexTokenStream traite une suite de lignes JSONL (une par événement
// codex) et retourne le dernier total de tokens vu, et si au moins un
// événement de tokens a été trouvé. Fonction pure, sans exec, pour tester la
// logique de non-cumul indépendamment du process codex réel.
func parseCodexTokenStream(lines []string) (Tokens, bool) {
	var acc codexTokenAccumulator
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		acc.observe(m)
	}
	return acc.last, acc.found
}

// quotaWindowLabel nomme une fenêtre de quota codex d'après sa durée en
// minutes : 300 = 5h, 10080 = une semaine (7*24*60). Toute autre valeur
// (changement côté OpenAI) retombe sur une étiquette générique plutôt que de
// perdre l'information.
func quotaWindowLabel(minutes int) string {
	switch minutes {
	case 300:
		return "5h"
	case 10080:
		return "week"
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

// parseCodexRateLimits extrait les fenêtres de quota d'un événement JSONL du
// fichier de session codex, de la forme :
//
//	{"type":"event_msg","payload":{"type":"token_count","rate_limits":{
//	  "primary":{"used_percent":44.0,"window_minutes":300,"resets_at":1771005432},
//	  "secondary":{"used_percent":49.0,"window_minutes":10080,"resets_at":1771409457}}}}
//
// Fonction pure, sans I/O, pour tester le parsing indépendamment du disque.
func parseCodexRateLimits(m map[string]any) ([]AgentQuotaWindow, bool) {
	payload, ok := m["payload"].(map[string]any)
	if !ok || payload["type"] != "token_count" {
		return nil, false
	}
	rl, ok := payload["rate_limits"].(map[string]any)
	if !ok {
		return nil, false
	}
	var windows []AgentQuotaWindow
	for _, key := range []string{"primary", "secondary"} {
		w, ok := rl[key].(map[string]any)
		if !ok {
			continue
		}
		used, _ := w["used_percent"].(float64)
		minutes, _ := w["window_minutes"].(float64)
		resets, _ := w["resets_at"].(float64)
		windows = append(windows, AgentQuotaWindow{
			Label:       quotaWindowLabel(int(minutes)),
			UsedPercent: used,
			ResetsAt:    time.Unix(int64(resets), 0),
		})
	}
	if len(windows) == 0 {
		return nil, false
	}
	return windows, true
}

// codexSessionsDir est le répertoire des fichiers de session codex, une
// indirection testable (évite de dépendre de $HOME dans les tests).
var codexSessionsDir = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".codex", "sessions")
}

// readCodexRateLimits localise le fichier de session écrit par codex pour
// threadID (structure sessions/AAAA/MM/JJ/rollout-...-<threadID>.jsonl,
// écrite par codex lui-même, y compris en mode `exec --json`) et en extrait
// le dernier instantané de quota connu. Ce n'est PAS porté par le flux --json
// que Sillage lit déjà pour les tokens (voir extractCodexTokens) : seul le
// fichier de session le contient. Best-effort : ok=false si le fichier est
// introuvable ou ne portait encore aucun rate_limits.
func readCodexRateLimits(threadID string) ([]AgentQuotaWindow, bool) {
	if threadID == "" {
		return nil, false
	}
	dir := codexSessionsDir()
	if dir == "" {
		return nil, false
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*", "*", "*", "rollout-*-"+threadID+".jsonl"))
	if err != nil || len(matches) == 0 {
		return nil, false
	}
	f, err := os.Open(matches[0])
	if err != nil {
		return nil, false
	}
	defer f.Close()

	var last []AgentQuotaWindow
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		var generic map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &generic); err != nil {
			continue
		}
		if windows, ok := parseCodexRateLimits(generic); ok {
			last = windows
		}
	}
	_ = scanner.Err() // best-effort : une lecture tronquée garde le dernier instantané trouvé
	return last, last != nil
}

// --- Adaptateur fake (test/démo, sans exec) ---

func (r *Runner) runFake(task *Task, agent Agent, handle *procHandle) error {
	ctx, cancel := context.WithCancel(context.Background())
	handle.cancel = cancel
	defer cancel()

	steps := []string{"Analyse du projet", "Rédaction des modifications", "Finalisation"}
	for _, step := range steps {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}
		line := step
		if t, err := r.store.UpdateTask(task.ID, func(t *Task) { t.LiveActivity = &line }); err == nil {
			*task = t
			r.publishActivity(task.ID, &line)
			r.publishTask(*task)
		}
	}

	content := fmt.Sprintf("# Sillage : tâche de test\n\nGénéré le %s par l'agent %s.\n", time.Now().UTC().Format(time.RFC3339), agent.Name)
	_ = os.WriteFile(filepath.Join(task.WorktreeDir, "SILLAGE-TEST.md"), []byte(content), 0o644)

	msg, t, err := r.store.AddMessage(task.ID, "agent", agent.Name, "Tâche simulée terminée : SILLAGE-TEST.md mis à jour.")
	if err == nil {
		r.publishMessage(msg)
		*task = t
		r.publishTask(*task)
	}

	if t, err := r.store.UpdateTask(task.ID, func(t *Task) {
		t.Tokens.Input += 1200
		t.Tokens.Output += 340
		t.Tokens.CostUsd += 0.004
	}); err == nil {
		*task = t
	}
	r.publishTokens()
	return nil
}
