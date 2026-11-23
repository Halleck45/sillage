package server

// La Vigie : surveiller les PR et issues GitHub des projets qui l'ont activée,
// et faire analyser chaque nouvel élément par un agent en LECTURE SEULE avant
// d'en parler à l'humain. On ne le dérange pas pour lui demander de lire une
// PR, on le dérange avec une revue déjà rédigée.
//
// Rien ne sort de la machine. Sillage lit la forge avec gh (liste, diff), et
// n'y écrit jamais : ni commentaire, ni revue, ni label. Le verdict et la revue
// rédigés par l'agent restent dans state.json ; les publier est un geste de
// l'humain, hors de Sillage (copier, coller sur GitHub).
//
// Le code d'une PR n'est jamais exécuté ni même extrait : l'agent reçoit le
// diff en texte et lit le dépôt local tel qu'il est. Lancer les tests d'une PR
// venue d'un fork, c'est exécuter le code d'un inconnu sur ce poste.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	// watchInterval : une PR n'est pas une alerte de production. Dix minutes
	// suffisent pour qu'elle soit relue avant qu'on y pense.
	watchInterval = 10 * time.Minute

	// watchFirstDelay : laisser le serveur démarrer avant d'appeler gh.
	watchFirstDelay = 20 * time.Second

	// watchLookback : à l'activation, la Vigie remonte ce qui a été ouvert
	// pendant le dernier mois. Sans ce recul, elle resterait muette jusqu'à la
	// prochaine PR et on ne saurait pas si elle marche. Un mois, parce qu'un
	// petit projet open source reçoit une issue toutes les deux semaines : une
	// semaine laissait la Vigie muette sur un projet pourtant actif. Le
	// plafond par relève (watchMaxNewPerList) empêche le backlog de tomber
	// d'un coup.
	watchLookback = 30 * 24 * time.Hour

	// watchLogMax : lignes gardées dans le journal, en mémoire.
	watchLogMax = 80

	// watchListLimit : les éléments les plus récents de chaque liste. Le reste
	// est antérieur à Since dans tout dépôt raisonnable.
	watchListLimit = 30

	// watchMaxNewPerList plafonne ce qu'une relève ajoute par dépôt et par
	// type. Chaque signal coûte une analyse : un dépôt très actif est absorbé
	// sur plusieurs relèves plutôt que d'un coup.
	watchMaxNewPerList = 5

	// watchMaxAnalysesPerRun borne le coût d'une relève, tous projets confondus.
	watchMaxAnalysesPerRun = 10

	// watchGhTimeout : un appel gh, liste ou diff.
	watchGhTimeout = 30 * time.Second

	// signalDiffMax, signalBodyMax : le prompt part en argument du CLI, et un
	// argument est plafonné à 128 Kio sous Linux (MAX_ARG_STRLEN).
	signalDiffMax = 40000
	signalBodyMax = 4000

	// signalVerdictMax : un verdict se lit d'un coup d'œil, dans une ligne.
	signalVerdictMax = 400

	// signalCommentsMax, signalCommentMax : la discussion gardée et passée à
	// l'analyste. Les derniers messages suffisent à savoir où elle en est.
	signalCommentsMax = 10
	signalCommentMax  = 1500
)

// forgeItem est une PR ou une issue telle que gh la liste.
type forgeItem struct {
	Number    int
	Title     string
	URL       string
	Body      string
	Author    string
	CreatedAt string
	Labels    []string
	IsDraft   bool
	Comments  []SignalComment // commentaires et revues, du plus ancien au plus récent
}

// lastOtherComment : la date du dernier message écrit par quelqu'un d'autre
// que viewer (le compte gh de la machine) et qui n'est pas un robot. Vide s'il
// n'y en a pas. Répondre soi-même ne doit pas relancer une analyse.
func lastOtherComment(comments []SignalComment, viewer string) string {
	last := ""
	for _, c := range comments {
		if c.Author == "" || strings.EqualFold(c.Author, viewer) || strings.HasSuffix(c.Author, "[bot]") {
			continue
		}
		if c.CreatedAt > last {
			last = c.CreatedAt
		}
	}
	return last
}

// latestComment : la date du message le plus récent, quel qu'en soit l'auteur.
func latestComment(comments []SignalComment) string {
	last := ""
	for _, c := range comments {
		if c.CreatedAt > last {
			last = c.CreatedAt
		}
	}
	return last
}

// newOtherComments compte les messages de tiers postérieurs à after.
func newOtherComments(comments []SignalComment, viewer, after string) int {
	n := 0
	for _, c := range comments {
		if c.Author == "" || strings.EqualFold(c.Author, viewer) || strings.HasSuffix(c.Author, "[bot]") {
			continue
		}
		if c.CreatedAt > after {
			n++
		}
	}
	return n
}

// forgeLister liste les PR ("pr") ou issues ("issue") ouvertes du dépôt dir.
type forgeLister func(ctx context.Context, dir, kind string) ([]forgeItem, error)

// prDiffer retourne le diff d'une PR.
type prDiffer func(ctx context.Context, dir string, number int) (string, error)

// watcher porte l'état du veilleur. En mémoire uniquement.
type watcher struct {
	mu     sync.Mutex
	status WatchStatus
	stop   chan struct{}
	kick   chan struct{}
	list   forgeLister
	diff   prDiffer
	// whoami : le compte gh de la machine, pour ne pas prendre ses propres
	// réponses pour du nouveau. viewer le garde une fois connu.
	whoami func(ctx context.Context) string
	viewer string
	// eol : les cycles de vie publiés (endoflife.date), remplaçable en test.
	eol eolFetcher
	// cancel arrête la relecture en cours (Current), nil sinon.
	cancel context.CancelFunc
	// runMu sérialise les relèves : une relève manuelle pendant la relève
	// périodique attend son tour au lieu d'analyser deux fois le même signal.
	runMu sync.Mutex
}

func newWatcher() *watcher {
	return &watcher{
		status: WatchStatus{Errors: map[string]string{}},
		kick:   make(chan struct{}, 1),
		list:   ghListItems,
		diff:   ghPRDiff,
		whoami: ghViewer,
		eol:    fetchEOL,
	}
}

func (w *watcher) snapshot() WatchStatus {
	w.mu.Lock()
	defer w.mu.Unlock()
	out := w.status
	out.Log = append([]WatchLogEntry(nil), w.status.Log...)
	if w.status.Current != nil {
		c := *w.status.Current
		c.Steps = append([]string(nil), c.Steps...)
		out.Current = &c
	}
	out.Errors = make(map[string]string, len(w.status.Errors))
	for k, v := range w.status.Errors {
		out.Errors[k] = v
	}
	return out
}

// log ajoute une ligne en tête du journal.
func (w *watcher) log(e WatchLogEntry) {
	if e.At.IsZero() {
		e.At = time.Now().UTC()
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status.Log = append([]WatchLogEntry{e}, w.status.Log...)
	if len(w.status.Log) > watchLogMax {
		w.status.Log = w.status.Log[:watchLogMax]
	}
}

// setProgress modifie la relecture en cours ; sans effet s'il n'y en a pas.
func (w *watcher) setProgress(fn func(*AnalysisProgress)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status.Current != nil {
		fn(w.status.Current)
	}
}

func (w *watcher) eolTool() eolFetcher {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.eol
}

// stopUpkeep arrête la veille d'entretien en cours du projet.
func (w *watcher) stopUpkeep(projectID string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	c := w.status.Current
	if c == nil || c.Kind != "upkeep" || c.ProjectID != projectID || w.cancel == nil {
		return false
	}
	w.cancel()
	return true
}

// stopAnalysis arrête la relecture de signalID si c'est elle qui tourne.
func (w *watcher) stopAnalysis(signalID string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.status.Current == nil || w.status.Current.SignalID != signalID || w.cancel == nil {
		return false
	}
	w.cancel()
	return true
}

func (w *watcher) tools() (forgeLister, prDiffer) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.list, w.diff
}

// WatchStatus est l'état du veilleur exposé par l'API.
func (s *Server) WatchStatus() WatchStatus { return s.watch.snapshot() }

func (s *Server) publishWatch() {
	s.hub.Publish(Event{Name: "watch", Data: s.watch.snapshot()})
}

func (s *Server) publishSignal(sig Signal) {
	s.hub.Publish(Event{Name: "signal", Data: sig})
}

// startWatcher lance la relève périodique. Elle tourne toujours, mais ne fait
// rien tant qu'aucun projet n'a activé la Vigie : pas un appel à gh.
func (s *Server) startWatcher() {
	s.watch.mu.Lock()
	if s.watch.stop != nil {
		s.watch.mu.Unlock()
		return
	}
	stop := make(chan struct{})
	s.watch.stop = stop
	s.watch.mu.Unlock()

	go func() {
		timer := time.NewTimer(watchFirstDelay)
		defer timer.Stop()
		ticker := time.NewTicker(watchInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-timer.C:
			case <-ticker.C:
			case <-s.watch.kick:
			}
			s.watchOnce(context.Background())
		}
	}()
}

func (s *Server) stopWatcher() {
	s.watch.mu.Lock()
	defer s.watch.mu.Unlock()
	if s.watch.stop == nil {
		return
	}
	close(s.watch.stop)
	s.watch.stop = nil
}

// kickWatcher demande une relève sans attendre la prochaine échéance. Non
// bloquant : une relève déjà demandée suffit.
func (s *Server) kickWatcher() {
	select {
	case s.watch.kick <- struct{}{}:
	default:
	}
}

// watchOnce fait une relève complète : lister les forges, puis analyser ce qui
// attend. Synchrone, c'est ce que les tests appellent.
func (s *Server) watchOnce(ctx context.Context) {
	projects := s.store.WatchedProjects()
	if len(projects) == 0 && len(s.store.QueuedSignals()) == 0 {
		return
	}
	s.watch.runMu.Lock()
	defer s.watch.runMu.Unlock()

	s.watch.mu.Lock()
	s.watch.status.Checking = true
	s.watch.mu.Unlock()
	s.publishWatch()

	errs := map[string]string{}
	for _, p := range projects {
		if msg := s.pollProject(ctx, p); msg != "" {
			errs[p.ID] = msg
		}
	}

	now := time.Now().UTC()
	s.watch.mu.Lock()
	s.watch.status.Checking = false
	s.watch.status.CheckedAt = &now
	s.watch.status.Errors = errs
	s.watch.mu.Unlock()
	s.publishWatch()

	s.analyzeQueued(ctx)
	// La veille d'entretien passe après les PR et issues : quelqu'un attend
	// une réponse là, personne ici.
	s.upkeepDue(ctx)
}

// pollProject relève les PR et issues des dépôts GitHub d'un projet. Retourne
// la première erreur rencontrée, "" si tout s'est bien passé.
func (s *Server) pollProject(ctx context.Context, p Project) string {
	list, _ := s.watch.tools()
	viewer := s.watch.viewerLogin(ctx)
	since := p.Watch.Since.Add(-watchLookback)
	var firstErr string
	for _, repo := range p.Repos {
		if DetectForge(repo.Path).Provider != "github" {
			continue
		}
		for _, kind := range []string{"pr", "issue"} {
			items, err := list(ctx, repo.Path, kind)
			if err != nil {
				if firstErr == "" {
					firstErr = repo.Name + ": " + err.Error()
				}
				s.watch.log(WatchLogEntry{Kind: "error", ProjectID: p.ID, RepoName: repo.Name, Item: kind, Error: truncate(err.Error(), 300)})
				continue
			}
			entry := s.recordItems(p, repo.Name, kind, items, since, viewer)
			s.watch.log(entry)
		}
	}
	return firstErr
}

// recordItems ajoute les éléments nouveaux et postérieurs à since, les plus
// anciens d'abord pour que les identifiants suivent l'ordre d'ouverture. Rend
// la ligne de journal qui dit ce qui a été vu, et pourquoi le reste a été
// écarté : une Vigie muette doit pouvoir s'expliquer.
func (s *Server) recordItems(p Project, repoName, kind string, items []forgeItem, since time.Time, viewer string) WatchLogEntry {
	entry := WatchLogEntry{Kind: "list", ProjectID: p.ID, RepoName: repoName, Item: kind, Open: len(items), From: since}
	sinceStr := since.UTC().Format(time.RFC3339)
	var fresh []forgeItem
	for _, it := range items {
		if it.Number == 0 {
			continue
		}
		if it.IsDraft {
			entry.Drafts++
			continue
		}
		// Un élément déjà suivi : seule une discussion nouvelle compte.
		if existing, ok := s.store.FindSignal(p.ID, repoName, kind, it.Number); ok {
			if s.refreshSignalComments(existing, it, viewer) {
				entry.Updated++
			} else {
				entry.Known++
			}
			continue
		}
		// Un élément ancien entre quand même s'il vient d'être commenté par
		// un tiers : c'est de l'activité nouvelle sur le projet.
		created, err := time.Parse(time.RFC3339, it.CreatedAt)
		if (err != nil || created.Before(since)) && lastOtherComment(it.Comments, viewer) < sinceStr {
			entry.Older++
			continue
		}
		fresh = append(fresh, it)
	}
	sort.Slice(fresh, func(i, j int) bool { return fresh[i].CreatedAt < fresh[j].CreatedAt })
	for _, it := range fresh {
		if entry.Added == watchMaxNewPerList {
			break
		}
		sig, isNew, err := s.store.AddSignalIfNew(Signal{
			ProjectID: p.ID, RepoName: repoName, Kind: kind, Number: it.Number,
			Title: it.Title, URL: it.URL, Author: it.Author,
			Body: truncate(normalizeIssueBody(it.Body), signalBodyMax), Labels: it.Labels,
			CreatedAt: it.CreatedAt, Comments: it.Comments, SeenCommentAt: latestComment(it.Comments),
		})
		if err != nil {
			continue
		}
		if !isNew {
			entry.Known++
			continue
		}
		entry.Added++
		s.publishSignal(sig)
	}
	return entry
}

// refreshSignalComments met à jour la discussion d'un signal suivi. Un message
// d'un tiers plus récent que le dernier vu remet le signal en file, même classé :
// quelqu'un attend une réponse. Un signal devenu chantier ou en cours
// d'analyse n'est pas touché. Retourne true si l'analyse est relancée.
func (s *Server) refreshSignalComments(sig Signal, it forgeItem, viewer string) bool {
	if sig.Status == "handled" || sig.Status == "analyzing" {
		return false
	}
	latest := latestComment(it.Comments)
	if sig.SeenCommentAt == "" {
		// Signal antérieur au suivi des commentaires : on prend la
		// discussion telle qu'elle est, sans relancer tout l'historique.
		if latest == "" {
			return false
		}
		if updated, err := s.store.UpdateSignal(sig.ID, func(x *Signal) { x.Comments, x.SeenCommentAt = it.Comments, latest }); err == nil {
			s.publishSignal(updated)
		}
		return false
	}
	if lastOtherComment(it.Comments, viewer) <= sig.SeenCommentAt {
		// Rien de neuf d'un tiers, mais la discussion a pu bouger (une
		// réponse du mainteneur) : on la garde à jour sans relancer, pour que
		// la fiche et la prochaine relecture la voient.
		if latest != "" && latest > sig.SeenCommentAt {
			if updated, err := s.store.UpdateSignal(sig.ID, func(x *Signal) { x.Comments, x.SeenCommentAt = it.Comments, latest }); err == nil {
				s.publishSignal(updated)
			}
		}
		return false
	}
	n := newOtherComments(it.Comments, viewer, sig.SeenCommentAt)
	updated, err := s.store.UpdateSignal(sig.ID, func(x *Signal) {
		x.Comments, x.SeenCommentAt, x.NewComments = it.Comments, latest, n
		x.Title = it.Title
		x.Body = truncate(normalizeIssueBody(it.Body), signalBodyMax)
		x.Labels = it.Labels
		x.Status, x.Error = "queued", ""
	})
	if err != nil {
		return false
	}
	s.publishSignal(updated)
	return true
}

func (w *watcher) viewerLogin(ctx context.Context) string {
	w.mu.Lock()
	v, fn := w.viewer, w.whoami
	w.mu.Unlock()
	if v != "" || fn == nil {
		return v
	}
	v = fn(ctx)
	w.mu.Lock()
	w.viewer = v
	w.mu.Unlock()
	return v
}

// analyzeQueued analyse les signaux en file, un à la fois : une relève ne
// doit pas lancer dix agents en parallèle sur un portable.
func (s *Server) analyzeQueued(ctx context.Context) {
	for i, sig := range s.store.QueuedSignals() {
		if i == watchMaxAnalysesPerRun {
			return
		}
		s.analyzeOne(ctx, sig)
	}
}

func (s *Server) analyzeOne(ctx context.Context, sig Signal) {
	project, ok := s.store.GetProject(sig.ProjectID)
	if !ok || !project.Watch.Enabled {
		// Vigie coupée entre-temps : le signal attend qu'on la rallume.
		return
	}
	agent, ok := s.store.GetAgent(project.Watch.AgentID)
	if !ok {
		s.failSignal(sig.ID, "the analyst agent no longer exists")
		return
	}
	repo, err := s.store.ResolveTaskRepo(project.ID, sig.RepoName)
	if err != nil {
		s.failSignal(sig.ID, err.Error())
		return
	}
	if sig, err = s.store.UpdateSignal(sig.ID, func(x *Signal) { x.Status = "analyzing"; x.Error = "" }); err != nil {
		return
	}
	s.publishSignal(sig)

	started := time.Now()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s.watch.mu.Lock()
	s.watch.status.Current = &AnalysisProgress{Kind: "review", ProjectID: sig.ProjectID, RepoName: sig.RepoName,
		SignalID: sig.ID, StartedAt: started.UTC(), Phase: "diff",
		Steps: []string{}, TimeoutSec: int(signalAnalysisTimeout.Seconds())}
	s.watch.cancel = cancel
	s.watch.mu.Unlock()
	s.publishWatch()
	defer func() {
		s.watch.mu.Lock()
		s.watch.status.Current, s.watch.cancel = nil, nil
		s.watch.mu.Unlock()
		s.publishWatch()
	}()
	logDone := func(errMsg string) {
		s.watch.log(WatchLogEntry{Kind: "analysis", ProjectID: sig.ProjectID, RepoName: sig.RepoName, Item: sig.Kind,
			Number: sig.Number, DurationMs: time.Since(started).Milliseconds(), Error: truncate(errMsg, 300)})
	}

	diff := ""
	if sig.Kind == "pr" && agent.Cli != "fake" {
		_, differ := s.watch.tools()
		if diff, err = differ(ctx, repo.Path, sig.Number); err != nil {
			if ctx.Err() != nil {
				err = errors.New("stopped")
			}
			s.failSignal(sig.ID, err.Error())
			logDone(err.Error())
			return
		}
	}
	s.watch.setProgress(func(p *AnalysisProgress) { p.Phase = "reading" })
	s.publishWatch()
	onStep := func(line string) {
		s.watch.setProgress(func(p *AnalysisProgress) { p.Steps = appendStep(p.Steps, line) })
		s.publishWatch()
	}
	res, err := AnalyzeSignal(ctx, agent, project, sig, repo.Path, diff, s.store.GetSettings().Lang, s.watch.viewerLogin(ctx), onStep)
	if errors.Is(err, errAnalysisStopped) {
		// Arrêtée par l'humain : ce n'est pas une panne. Le signal attend
		// qu'on le relance, sans message d'erreur.
		if sig, err := s.store.UpdateSignal(sig.ID, func(x *Signal) {
			if x.Status == "analyzing" {
				x.Status, x.Error = "error", "stopped"
			}
		}); err == nil {
			s.publishSignal(sig)
		}
		logDone("stopped")
		return
	}
	if err != nil {
		s.failSignal(sig.ID, err.Error())
		logDone(err.Error())
		return
	}
	logDone("")
	now := time.Now().UTC()
	sig, err = s.store.UpdateSignal(sig.ID, func(x *Signal) {
		x.Status = "ready"
		x.Priority = res.Priority
		x.Decision = res.Decision
		x.Verdict = res.Verdict
		x.Review = res.Review
		x.AnalyzedAt = &now
	})
	if err == nil {
		s.publishSignal(sig)
	}
}

func (s *Server) failSignal(id, msg string) {
	sig, err := s.store.UpdateSignal(id, func(x *Signal) { x.Status = "error"; x.Error = truncate(msg, 300) })
	if err == nil {
		s.publishSignal(sig)
	}
}

// --- Lecture de la forge (gh, lecture seule) ---

func ghListItems(ctx context.Context, dir, kind string) ([]forgeItem, error) {
	sub := "issue"
	fields := "number,title,url,body,labels,author,createdAt,comments"
	if kind == "pr" {
		sub = "pr"
		fields += ",isDraft,reviews"
	}
	out, err := runGh(ctx, dir, sub, "list", "--state", "open",
		"--limit", strconv.Itoa(watchListLimit), "--json", fields)
	if err != nil {
		return nil, err
	}
	return parseForgeItems(out)
}

func parseForgeItems(raw []byte) ([]forgeItem, error) {
	var items []struct {
		Number    int    `json:"number"`
		Title     string `json:"title"`
		URL       string `json:"url"`
		Body      string `json:"body"`
		CreatedAt string `json:"createdAt"`
		IsDraft   bool   `json:"isDraft"`
		Labels    []struct {
			Name string `json:"name"`
		} `json:"labels"`
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		Comments []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			Body      string `json:"body"`
			CreatedAt string `json:"createdAt"`
		} `json:"comments"`
		Reviews []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			Body        string `json:"body"`
			State       string `json:"state"`
			SubmittedAt string `json:"submittedAt"`
		} `json:"reviews"`
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, errors.New("could not read the list returned by gh")
	}
	out := make([]forgeItem, 0, len(items))
	for _, it := range items {
		labels := make([]string, 0, len(it.Labels))
		for _, l := range it.Labels {
			if l.Name != "" {
				labels = append(labels, l.Name)
			}
		}
		var comments []SignalComment
		for _, c := range it.Comments {
			comments = append(comments, SignalComment{Author: c.Author.Login, Body: truncate(strings.TrimSpace(c.Body), signalCommentMax), CreatedAt: c.CreatedAt})
		}
		for _, r := range it.Reviews {
			// Une revue sans texte (une approbation sèche) dit quand même
			// quelque chose : son état.
			body := strings.TrimSpace(r.Body)
			if r.State != "" && r.State != "COMMENTED" {
				body = strings.TrimSpace("[" + strings.ToLower(strings.ReplaceAll(r.State, "_", " ")) + "] " + body)
			}
			if body == "" || r.SubmittedAt == "" {
				continue
			}
			comments = append(comments, SignalComment{Author: r.Author.Login, Body: truncate(body, signalCommentMax), CreatedAt: r.SubmittedAt})
		}
		sort.SliceStable(comments, func(i, j int) bool { return comments[i].CreatedAt < comments[j].CreatedAt })
		if len(comments) > signalCommentsMax {
			comments = comments[len(comments)-signalCommentsMax:]
		}
		out = append(out, forgeItem{
			Number: it.Number, Title: strings.TrimSpace(it.Title), URL: it.URL, Body: it.Body,
			Author: it.Author.Login, CreatedAt: it.CreatedAt, Labels: labels, IsDraft: it.IsDraft,
			Comments: comments,
		})
	}
	return out, nil
}

// ghViewer : le compte gh de la machine, vide si gh ne répond pas (toutes les
// réponses comptent alors comme celles d'un tiers).
func ghViewer(ctx context.Context) string {
	out, err := runGh(ctx, "", "api", "user", "--jq", ".login")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ghPRDiff(ctx context.Context, dir string, number int) (string, error) {
	out, err := runGh(ctx, dir, "pr", "diff", strconv.Itoa(number), "--color", "never")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// runGh lance une commande gh de lecture dans dir. Setpgid + Cancel comme pour
// l'import d'issues.
func runGh(ctx context.Context, dir string, args ...string) ([]byte, error) {
	if _, err := lookPath("gh"); err != nil {
		return nil, errors.New("gh not found in PATH")
	}
	ctx, cancel := context.WithTimeout(ctx, watchGhTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, errors.New("gh took too long to answer")
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", truncate(msg, 300))
	}
	return stdout.Bytes(), nil
}
