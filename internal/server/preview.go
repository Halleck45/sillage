package server

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// previewLogLines est la taille du tampon circulaire du journal d'un run : de
// quoi voir une installation de dépendances et son erreur, pas de quoi garder
// l'historique d'un serveur qui tourne depuis une heure. Rien n'est écrit sur
// disque (voir docs/SPEC-RECETTE.md §3).
const previewLogLines = 2000

// previewPortBase fixe le port de départ de $SILLAGE_PORT (voir previewEnv) :
// au-dessus de la plage privilégiée (< 1024, réservée au superutilisateur sur
// Linux) et des ports enregistrés IANA usuels, pour qu'une référence de
// chantier ou de tâche, aussi petite soit-elle, ne retombe jamais sur un port
// qui exigerait les droits root ou entrerait en conflit avec un service connu.
const previewPortBase = 4000

// PreviewSupervisor lance les recettes manuelles : la commande d'un dépôt
// (Repo.PreviewCmd) exécutée dans le worktree d'un chantier ou d'une tâche. Un
// seul run par worktree à la fois ; relancer arrête le précédent.
//
// Rien n'est persisté : un redémarrage du serveur ne doit laisser derrière lui
// ni process (voir StopAll, appelé à l'arrêt) ni run fantôme.
type PreviewSupervisor struct {
	mu   sync.Mutex
	runs map[string]*previewProc // clé : le worktree, un run par worktree
	seq  int

	hub *Hub
}

type previewProc struct {
	run  PreviewRun
	cmd  *exec.Cmd
	done chan struct{}

	// stopped distingue un arrêt demandé par l'humain d'une sortie naturelle :
	// le statut final n'est pas le même (stopped vs exited).
	stopped atomic.Bool

	logMu sync.Mutex
	log   []string
}

func NewPreviewSupervisor(hub *Hub) *PreviewSupervisor {
	return &PreviewSupervisor{runs: map[string]*previewProc{}, hub: hub}
}

// previewTarget décrit ce qu'on recette : un worktree, la commande à y lancer,
// et l'identité (SILLAGE_ID / SILLAGE_N) qui permet à cette commande de
// s'isoler. Les deux points d'entrée (chantier, tâche) construisent la même
// structure, le reste du superviseur ne sait pas les distinguer.
type previewTarget struct {
	projectID string
	cardID    string
	taskID    string
	repoName  string

	dir    string
	branch string
	cmd    string
	url    string

	id string // SILLAGE_ID : ws-107 (chantier) ou t-482 (tâche)
	n  int    // SILLAGE_N : la référence courte ; SILLAGE_PORT en dérive (previewEnv)
}

// Start lance (ou relance) la recette d'une cible. Le run précédent du même
// worktree est arrêté d'abord : deux serveurs sur le même port ne servent à
// personne.
func (p *PreviewSupervisor) Start(target previewTarget) (PreviewRun, error) {
	if strings.TrimSpace(target.cmd) == "" {
		return PreviewRun{}, fmt.Errorf("no preview command configured for this repository")
	}
	if target.dir == "" {
		return PreviewRun{}, fmt.Errorf("no worktree to run the preview in")
	}
	if _, err := os.Stat(target.dir); err != nil {
		return PreviewRun{}, fmt.Errorf("preview worktree is missing: %s", target.dir)
	}

	p.stopByDir(target.dir)

	env := previewEnv(target)
	proc := &previewProc{done: make(chan struct{})}

	p.mu.Lock()
	p.seq++
	runID := "pv" + strconv.Itoa(p.seq)
	p.mu.Unlock()

	proc.run = PreviewRun{
		ID:        runID,
		ProjectID: target.projectID,
		CardID:    target.cardID,
		TaskID:    target.taskID,
		RepoName:  target.repoName,
		Cmd:       target.cmd,
		URL:       expandPreviewURL(target.url, env, target.dir),
		Dir:       target.dir,
		Status:    "running",
		StartedAt: time.Now(),
	}

	cmd := exec.Command("sh", "-c", target.cmd)
	cmd.Dir = target.dir
	cmd.Env = env
	// Setpgid : la commande de recette lance souvent des enfants (npm → node,
	// make → serveur). Sans groupe de process, Arrêter ne tuerait que le shell
	// et laisserait le serveur en vie sur son port.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Un seul tuyau pour stdout et stderr : le journal doit montrer les erreurs
	// dans l'ordre où elles arrivent, mélangées à la sortie normale.
	pr, pw, err := os.Pipe()
	if err != nil {
		return PreviewRun{}, fmt.Errorf("cannot create preview pipe: %w", err)
	}
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		proc.run.Status = "failed"
		proc.run.Error = err.Error()
		ended := time.Now()
		proc.run.EndedAt = &ended
		p.mu.Lock()
		p.runs[target.dir] = proc
		p.mu.Unlock()
		p.publishRun(proc.run)
		return proc.run, nil
	}
	pw.Close()
	proc.cmd = cmd

	p.mu.Lock()
	p.runs[target.dir] = proc
	p.mu.Unlock()
	p.publishRun(proc.run)

	go p.pump(proc, runID, pr)
	go p.wait(proc, cmd)

	return p.snapshot(proc), nil
}

// pump lit la sortie du process ligne par ligne : chaque ligne va dans le
// tampon circulaire et part en SSE. Le hub abandonne les événements des clients
// lents (envoi non bloquant), donc une commande très bavarde ne bloque jamais
// le process qu'on est en train de recetter.
func (p *PreviewSupervisor) pump(proc *previewProc, runID string, pr *os.File) {
	defer pr.Close()
	scanner := bufio.NewScanner(pr)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		proc.logMu.Lock()
		proc.log = append(proc.log, line)
		if len(proc.log) > previewLogLines {
			proc.log = proc.log[len(proc.log)-previewLogLines:]
		}
		proc.logMu.Unlock()
		p.hub.Publish(Event{Name: "previewLog", Data: PreviewLogEvent{RunID: runID, Line: line}})
	}
}

func (p *PreviewSupervisor) wait(proc *previewProc, cmd *exec.Cmd) {
	err := cmd.Wait()
	close(proc.done)

	p.mu.Lock()
	proc.run.Status = "exited"
	if proc.stopped.Load() {
		proc.run.Status = "stopped"
	}
	if cmd.ProcessState != nil {
		proc.run.ExitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		proc.run.Error = err.Error()
	}
	ended := time.Now()
	proc.run.EndedAt = &ended
	run := proc.run
	p.mu.Unlock()

	p.publishRun(run)
}

// Stop arrête un run par son identifiant : SIGINT au groupe de process, puis
// SIGKILL après 5 s (même mécanique que l'interruption d'un agent).
func (p *PreviewSupervisor) Stop(runID string) error {
	p.mu.Lock()
	var proc *previewProc
	var running bool
	for _, candidate := range p.runs {
		if candidate.run.ID == runID {
			proc = candidate
			running = candidate.run.Status == "running"
			break
		}
	}
	p.mu.Unlock()
	if proc == nil {
		return fmt.Errorf("preview run not found")
	}
	if !running {
		return nil // déjà terminé : arrêter est sans effet, pas une erreur
	}
	proc.stopped.Store(true)
	killGroup(proc.cmd, proc.done)
	return nil
}

// stopByDir arrête le run en cours d'un worktree et attend sa mort, pour que le
// port soit libre avant que le nouveau run ne démarre.
func (p *PreviewSupervisor) stopByDir(dir string) {
	p.mu.Lock()
	proc, ok := p.runs[dir]
	running := ok && proc.run.Status == "running"
	p.mu.Unlock()
	if !running {
		return
	}
	proc.stopped.Store(true)
	killGroup(proc.cmd, proc.done)
	<-proc.done
}

// StopAll arrête tous les runs, en parallèle : appelé à l'arrêt du serveur.
// Rien ne doit survivre à la fermeture de Sillage.
func (p *PreviewSupervisor) StopAll() {
	p.mu.Lock()
	procs := make([]*previewProc, 0, len(p.runs))
	for _, proc := range p.runs {
		if proc.run.Status == "running" {
			procs = append(procs, proc)
		}
	}
	p.mu.Unlock()

	var wg sync.WaitGroup
	for _, proc := range procs {
		wg.Add(1)
		go func(proc *previewProc) {
			defer wg.Done()
			proc.stopped.Store(true)
			killGroup(proc.cmd, proc.done)
			<-proc.done
		}(proc)
	}
	wg.Wait()
}

// RunningCount retourne le nombre de recettes en cours.
func (p *PreviewSupervisor) RunningCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := 0
	for _, proc := range p.runs {
		if proc.run.Status == "running" {
			n++
		}
	}
	return n
}

// List retourne les runs connus, triés par date de lancement (le plus récent
// en dernier), pour l'hydratation du frontend.
func (p *PreviewSupervisor) List() []PreviewRun {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]PreviewRun, 0, len(p.runs))
	for _, proc := range p.runs {
		out = append(out, proc.run)
	}
	sortPreviewRuns(out)
	return out
}

// Log retourne le journal d'un run (tampon circulaire), à l'ouverture du
// panneau de recette : la suite arrive en SSE.
func (p *PreviewSupervisor) Log(runID string) ([]string, bool) {
	p.mu.Lock()
	var proc *previewProc
	for _, candidate := range p.runs {
		if candidate.run.ID == runID {
			proc = candidate
			break
		}
	}
	p.mu.Unlock()
	if proc == nil {
		return nil, false
	}
	proc.logMu.Lock()
	defer proc.logMu.Unlock()
	return append([]string{}, proc.log...), true
}

func (p *PreviewSupervisor) snapshot(proc *previewProc) PreviewRun {
	p.mu.Lock()
	defer p.mu.Unlock()
	return proc.run
}

func (p *PreviewSupervisor) publishRun(run PreviewRun) {
	p.hub.Publish(Event{Name: "preview", Data: run})
}

func sortPreviewRuns(runs []PreviewRun) {
	for i := 1; i < len(runs); i++ {
		for j := i; j > 0 && runs[j].StartedAt.Before(runs[j-1].StartedAt); j-- {
			runs[j], runs[j-1] = runs[j-1], runs[j]
		}
	}
}

// previewEnv construit l'environnement du run : celui de Sillage, plus les
// cinq variables de recette. SILLAGE_ID et SILLAGE_N dérivent de la référence
// courte du chantier ou de la tâche : stables dans le temps (la base de recette
// survit entre deux sessions) et uniques dans un projet (le compteur de
// références est partagé entre chantiers et tâches). SILLAGE_PORT est
// `previewPortBase + SILLAGE_N` précalculé : la commande de recette peut
// écrire `PORT=$SILLAGE_PORT` sans refaire l'arithmétique elle-même, et donc
// sans risquer d'oublier le décalage et de retomber sur un port privilégié.
func previewEnv(target previewTarget) []string {
	return append(os.Environ(),
		"SILLAGE_ID="+target.id,
		"SILLAGE_N="+strconv.Itoa(target.n),
		"SILLAGE_PORT="+strconv.Itoa(previewPortBase+target.n),
		"SILLAGE_DIR="+target.dir,
		"SILLAGE_BRANCH="+target.branch,
	)
}

// expandPreviewURL substitue les variables de l'URL de recette. C'est le shell
// qui fait le travail, avec le même environnement que la commande : l'URL
// accepte donc exactement la même syntaxe, arithmétique comprise
// (http://127.0.0.1:$((4000 + SILLAGE_N))). Réimplémenter $(( )) en Go
// donnerait deux syntaxes à documenter et à ne pas confondre.
func expandPreviewURL(template string, env []string, dir string) string {
	template = strings.TrimSpace(template)
	if template == "" {
		return ""
	}
	cmd := exec.Command("sh", "-c", "printf '%s' "+shellDoubleQuote(template))
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.Output()
	if err != nil {
		return template
	}
	return strings.TrimSpace(string(out))
}

// shellDoubleQuote entoure s de guillemets doubles en échappant ce qui casserait
// la chaîne, mais pas le `$` : les variables et l'arithmétique doivent être
// développées, c'est tout l'intérêt.
func shellDoubleQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"', '\\', '`':
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}
