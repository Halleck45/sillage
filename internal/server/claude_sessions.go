package server

// Sessions Claude Code menées hors de Sillage : celles qu'on lance à la main
// dans le dossier d'un dépôt, et que `claude --resume` liste. Sillage les
// montre en lecture seule, pour que le projet dise tout ce qui s'y passe ; il
// ne les reprend pas lui-même (une session est rangée par dossier de
// lancement, et une tâche Sillage tourne dans son propre worktree) et ne les
// transforme pas en tâches d'office : elles n'ont ni branche à elles ni diff
// à relire.
//
// Le format des fichiers (JSONL sous <config>/projects/<dossier encodé>/) est
// interne à Claude Code et change d'une version à l'autre : on n'en lit que
// quelques champs, et une ligne illisible est ignorée, jamais fatale.
//
// Les exécutions non interactives (`claude -p`, entrypoint "sdk-cli") sont
// écartées : ce sont les tâches, plans, relectures de la Vigie et
// configurations automatiques de Sillage lui-même.

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	// claudeSessionsMax : les plus récentes par projet.
	claudeSessionsMax = 30
	// claudeSessionsMaxAge : au-delà, une session n'apprend plus rien sur le projet.
	claudeSessionsMaxAge = 60 * 24 * time.Hour
	// claudeSessionActive : un fichier écrit depuis moins longtemps appartient
	// à une session probablement ouverte dans un terminal.
	claudeSessionActive = 3 * time.Minute
	// claudeSessionPrompts : les dernières demandes gardées, pour « en faire
	// un chantier ».
	claudeSessionPrompts   = 8
	claudeSessionPromptMax = 600
)

// ClaudeSession est une session Claude Code trouvée dans un dépôt du projet.
type ClaudeSession struct {
	ID         string    `json:"id"`
	RepoName   string    `json:"repoName"`
	Cwd        string    `json:"cwd"`
	Title      string    `json:"title"`
	LastPrompt string    `json:"lastPrompt"`
	Branch     string    `json:"branch"`
	StartedAt  time.Time `json:"startedAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Active     bool      `json:"active"`
	PromptsN   int       `json:"promptsCount"`
	Prompts    []string  `json:"prompts"` // les dernières demandes, la plus récente en dernier
	ConfigDir  string    `json:"configDir"`
	ResumeCmd  string    `json:"resumeCmd"`
}

// claudeSessionFile est ce qu'on retient d'un fichier, mis en cache tant que
// sa taille et sa date ne changent pas : relire 1 Mo par session à chaque
// rafraîchissement serait du gaspillage.
type claudeSessionFile struct {
	size     int64
	mod      time.Time
	session  ClaudeSession
	ok       bool
	headless bool
}

type claudeSessionCache struct {
	mu    sync.Mutex
	files map[string]claudeSessionFile
}

var claudeSessions = &claudeSessionCache{files: map[string]claudeSessionFile{}}

var nonAlnum = regexp.MustCompile(`[^a-zA-Z0-9]`)

// claudeProjectDirName : le nom de dossier qu'emploie Claude Code pour un
// chemin de lancement (tout ce qui n'est pas alphanumérique devient "-").
func claudeProjectDirName(path string) string {
	return nonAlnum.ReplaceAllString(path, "-")
}

// claudeConfigDirs : les profils Claude Code de la machine. CLAUDE_CONFIG_DIR,
// ~/.claude, et les ~/.claude-* qu'on crée pour séparer des comptes. Les
// sauvegardes (.bak, .backup, .old) sont écartées.
func claudeConfigDirs() []string {
	seen := map[string]bool{}
	var out []string
	add := func(d string) {
		if d == "" || seen[d] {
			return
		}
		if st, err := os.Stat(filepath.Join(d, "projects")); err != nil || !st.IsDir() {
			return
		}
		seen[d] = true
		out = append(out, d)
	}
	add(os.Getenv("CLAUDE_CONFIG_DIR"))
	home, err := os.UserHomeDir()
	if err != nil {
		return out
	}
	add(filepath.Join(home, ".claude"))
	matches, _ := filepath.Glob(filepath.Join(home, ".claude-*"))
	sort.Strings(matches)
	for _, m := range matches {
		base := strings.ToLower(filepath.Base(m))
		if strings.HasSuffix(base, ".bak") || strings.HasSuffix(base, ".backup") || strings.HasSuffix(base, ".old") {
			continue
		}
		add(m)
	}
	return out
}

// ListClaudeSessions retourne les sessions interactives lancées dans un dépôt
// du projet (ou un de ses sous-dossiers), les plus récentes d'abord.
func ListClaudeSessions(project Project) []ClaudeSession {
	return listClaudeSessions(project, claudeConfigDirs(), time.Now())
}

func listClaudeSessions(project Project, configDirs []string, now time.Time) []ClaudeSession {
	var out []ClaudeSession
	for _, cfg := range configDirs {
		entries, err := os.ReadDir(filepath.Join(cfg, "projects"))
		if err != nil {
			continue
		}
		for _, repo := range project.Repos {
			root := filepath.Clean(repo.Path)
			prefix := claudeProjectDirName(root)
			for _, e := range entries {
				// Le préfixe encodé est ambigu (« a-b » et « a/b » donnent le
				// même nom) : le cwd lu dans le fichier tranche.
				if !e.IsDir() || !strings.HasPrefix(e.Name(), prefix) {
					continue
				}
				dir := filepath.Join(cfg, "projects", e.Name())
				files, _ := filepath.Glob(filepath.Join(dir, "*.jsonl"))
				for _, f := range files {
					st, err := os.Stat(f)
					if err != nil || now.Sub(st.ModTime()) > claudeSessionsMaxAge {
						continue
					}
					sess, ok := claudeSessions.read(f, st)
					if !ok || (sess.Cwd != root && !strings.HasPrefix(sess.Cwd, root+string(filepath.Separator))) {
						continue
					}
					sess.RepoName = repo.Name
					sess.ConfigDir = cfg
					sess.Active = now.Sub(sess.UpdatedAt) < claudeSessionActive
					sess.ResumeCmd = claudeResumeCmd(sess.Cwd, cfg, sess.ID)
					out = append(out, sess)
				}
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	if len(out) > claudeSessionsMax {
		out = out[:claudeSessionsMax]
	}
	return out
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// claudeResumeCmd : la commande à coller dans un terminal. Le profil n'est
// précisé que s'il n'est pas celui par défaut.
func claudeResumeCmd(cwd, configDir, id string) string {
	cmd := "cd " + shellQuote(cwd) + " && "
	if home, err := os.UserHomeDir(); err != nil || configDir != filepath.Join(home, ".claude") {
		cmd += "CLAUDE_CONFIG_DIR=" + shellQuote(configDir) + " "
	}
	return cmd + "claude --resume " + id
}

func (c *claudeSessionCache) read(path string, st os.FileInfo) (ClaudeSession, bool) {
	c.mu.Lock()
	cached, hit := c.files[path]
	c.mu.Unlock()
	if hit && cached.size == st.Size() && cached.mod.Equal(st.ModTime()) {
		return cached.session, cached.ok && !cached.headless
	}
	sess, ok, headless := parseClaudeSessionFile(path)
	sess.UpdatedAt = st.ModTime().UTC()
	c.mu.Lock()
	c.files[path] = claudeSessionFile{size: st.Size(), mod: st.ModTime(), session: sess, ok: ok, headless: headless}
	c.mu.Unlock()
	return sess, ok && !headless
}

// claudeLine : les seuls champs lus. Tout le reste du format est ignoré.
type claudeLine struct {
	Type        string          `json:"type"`
	SessionID   string          `json:"sessionId"`
	Cwd         string          `json:"cwd"`
	GitBranch   string          `json:"gitBranch"`
	Entrypoint  string          `json:"entrypoint"`
	Timestamp   string          `json:"timestamp"`
	IsSidechain bool            `json:"isSidechain"`
	IsMeta      bool            `json:"isMeta"`
	AiTitle     string          `json:"aiTitle"`
	CustomTitle string          `json:"customTitle"`
	LastPrompt  string          `json:"lastPrompt"`
	Message     *claudeLineMsg  `json:"message"`
	ToolResult  json.RawMessage `json:"toolUseResult"`
}

type claudeLineMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// promptText : le texte tapé par l'humain, vide pour un retour d'outil ou une
// commande interne (« <command-name> », « <local-command-stdout> »...).
func promptText(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		var blocks []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(raw, &blocks); err != nil {
			return ""
		}
		var parts []string
		for _, b := range blocks {
			if b.Type == "text" {
				parts = append(parts, b.Text)
			}
		}
		s = strings.Join(parts, "\n")
	}
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "<") || strings.HasPrefix(s, "Caveat:") || strings.HasPrefix(s, "[Request interrupted") {
		return ""
	}
	return s
}

func parseClaudeSessionFile(path string) (sess ClaudeSession, ok, headless bool) {
	f, err := os.Open(path)
	if err != nil {
		return sess, false, false
	}
	defer f.Close()
	sess.ID = strings.TrimSuffix(filepath.Base(path), ".jsonl")
	sess.Prompts = []string{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 256*1024), 32*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		// Écarter vite les lignes volumineuses qui ne nous concernent pas
		// (réponses, instantanés de fichiers, requêtes API).
		if len(line) > 4096 && !strings.Contains(string(line[:min(len(line), 200)]), `"type":"user"`) &&
			!strings.Contains(string(line[:min(len(line), 200)]), `"parentUuid"`) {
			continue
		}
		var l claudeLine
		if err := json.Unmarshal(line, &l); err != nil {
			continue
		}
		if l.Entrypoint != "" && l.Entrypoint != "cli" {
			headless = true
		}
		switch l.Type {
		case "ai-title":
			if l.AiTitle != "" {
				sess.Title = l.AiTitle
			}
		case "custom-title":
			if l.CustomTitle != "" {
				sess.Title = l.CustomTitle
			}
		case "last-prompt":
			if l.LastPrompt != "" {
				sess.LastPrompt = truncate(strings.TrimSpace(l.LastPrompt), claudeSessionPromptMax)
			}
		case "user":
			if l.IsSidechain || l.IsMeta || l.Message == nil || len(l.ToolResult) > 0 {
				continue
			}
			if sess.Cwd == "" && l.Cwd != "" {
				sess.Cwd = filepath.Clean(l.Cwd)
				if ts, err := time.Parse(time.RFC3339, l.Timestamp); err == nil {
					sess.StartedAt = ts.UTC()
				}
			}
			if l.GitBranch != "" {
				sess.Branch = l.GitBranch
			}
			text := promptText(l.Message.Content)
			if text == "" {
				continue
			}
			sess.PromptsN++
			sess.Prompts = append(sess.Prompts, truncate(text, claudeSessionPromptMax))
			if len(sess.Prompts) > claudeSessionPrompts {
				sess.Prompts = sess.Prompts[1:]
			}
		}
	}
	if sess.Title == "" && len(sess.Prompts) > 0 {
		first := sess.Prompts[0]
		if i := strings.IndexByte(first, '\n'); i > 0 {
			first = first[:i]
		}
		sess.Title = truncate(first, 120)
	}
	if sess.LastPrompt == "" && len(sess.Prompts) > 0 {
		sess.LastPrompt = sess.Prompts[len(sess.Prompts)-1]
	}
	return sess, sess.Cwd != "" && sess.PromptsN > 0, headless
}

func (s *Server) handleClaudeSessions(w http.ResponseWriter, r *http.Request) {
	project, ok := s.store.GetProject(r.PathValue("id"))
	if !ok {
		writeError(w, http.StatusNotFound, "project not found")
		return
	}
	sessions := ListClaudeSessions(project)
	if sessions == nil {
		sessions = []ClaudeSession{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}
