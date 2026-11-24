package server

// Titre et description de la pull request d'un chantier. Sillage en propose un
// par défaut (le sujet du commit quand il n'y en a qu'un, sinon le titre du
// chantier), et un agent en lecture seule rédige une proposition à l'ouverture
// de la modale de livraison. L'humain relit, corrige, et c'est ce qu'il valide
// qui part avec la livraison : rien n'est écrit sur la forge avant.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
)

// prTitleMax : GitHub accepte 256 caractères, un titre lisible bien moins.
const prTitleMax = 200

// PRDraft est une proposition de titre et de description pour la pull request
// d'un dépôt du chantier.
type PRDraft struct {
	RepoName string `json:"repoName"`
	AgentID  string `json:"agentId,omitempty"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

// prLog lit les commits du chantier, hors merge commits, du plus ancien au
// plus récent : sujet seul, ou sujet et corps.
func prLog(dir, base, branch string, withBody bool) []string {
	format := "%s"
	if withBody {
		format = "%B"
	}
	out, err := runGit(dir, gitDefaultTimeout, "log", "--no-merges", "--reverse", "--format="+format+"%x1e",
		freshestRef(dir, base)+".."+branch)
	if err != nil {
		return nil
	}
	var entries []string
	for _, e := range strings.Split(out, "\x1e") {
		if e = strings.TrimSpace(e); e != "" {
			entries = append(entries, e)
		}
	}
	return entries
}

// defaultPRTitle est le titre sans rédaction : le sujet du commit quand le
// chantier n'en a qu'un (il suit déjà les conventions du projet), sinon le
// titre du chantier.
func defaultPRTitle(dir, base, branch, cardTitle string) string {
	if subjects := prLog(dir, base, branch, false); len(subjects) == 1 {
		return subjects[0]
	}
	return cardTitle
}

// prDraftAgent choisit qui rédige : l'agent de la dernière tâche acceptée du
// dépôt, qui connaît le travail, s'il sait tourner en lecture seule ; sinon le
// premier agent claude, puis codex, correctement installé.
func (s *Server) prDraftAgent(card Card, repoName string) (Agent, bool) {
	tasks := s.store.TasksByCard(card.ID)
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt) })
	for _, t := range tasks {
		if t.Status != "accepted" || t.RepoName != repoName {
			continue
		}
		if a, ok := s.store.GetAgent(t.AgentID); ok && canPlan(a.Cli) {
			return a, true
		}
	}
	agents := s.store.ListAgents()
	for _, cli := range []string{"claude", "codex"} {
		for _, a := range agents {
			if a.Cli == cli && a.Warning == "" {
				return a.Agent, true
			}
		}
	}
	return Agent{}, false
}

// prDraftInstructions rédige la demande envoyée à l'agent.
func prDraftInstructions(project Project, card Card, b CardBranch, commits []string, stat string) string {
	var sb strings.Builder
	sb.WriteString("You are writing the title and description of a pull request. Do not modify, create or delete any file.\n\n")
	sb.WriteString("Workstream: " + card.Title + "\n")
	sb.WriteString("Branch " + b.Branch + " into " + b.Base + ".\n")
	if blocks := strings.Join(contextParts(project.ContextPrompt, card.ContextPrompt), "\n\n"); blocks != "" {
		sb.WriteString("\n" + blocks + "\n")
	}
	sb.WriteString("\nCommits, oldest first:\n")
	for _, c := range commits {
		sb.WriteString("---\n" + c + "\n")
	}
	if stat != "" {
		sb.WriteString("\nChanged files:\n" + stat + "\n")
	}
	sb.WriteString("\nRead the diff (git diff " + freshestRef(b.WorktreeDir, b.Base) + "..." + b.Branch + ") as needed.")
	sb.WriteString(" Write the pull request a careful human maintainer of this repository would write:\n")
	sb.WriteString("- title: one line saying what the change does for its users. Follow the project's conventions (a conventional commit subject when its commits use them), never a branch name or an issue label.\n")
	sb.WriteString("- body: Markdown, short. Why the change is needed, then what it changes. Mention how it was checked only if the commits or the code show it. Reference an issue (\"Fixes #123\") only when the workstream or the commits name one.\n")
	sb.WriteString("- Write in the language of the commit messages. Never mention an AI, an agent, Sillage or any tool, and add no signature.\n\n")
	sb.WriteString("End your reply with a single JSON object and nothing after it:\n")
	sb.WriteString(`{"title":"...","body":"..."}` + "\n")
	return sb.String()
}

// extractPRDraft retrouve le dernier objet {"title","body"} exploitable de la
// réponse (même tolérance que extractPlanSteps).
func extractPRDraft(reply string) (string, string, bool) {
	var title, body string
	for i := 0; i < len(reply); i++ {
		if reply[i] != '{' {
			continue
		}
		var p struct {
			Title string `json:"title"`
			Body  string `json:"body"`
		}
		if err := json.NewDecoder(strings.NewReader(reply[i:])).Decode(&p); err != nil {
			continue
		}
		if t := strings.TrimSpace(strings.SplitN(p.Title, "\n", 2)[0]); t != "" && t != "..." {
			title, body = t, p.Body
		}
	}
	if title == "" {
		return "", "", false
	}
	if len(title) > prTitleMax {
		title = strings.TrimSpace(title[:prTitleMax])
	}
	return title, stripToolSignature(body), true
}

// stripToolSignature retire les signatures que certains CLI ajoutent d'office
// (« Generated with … », co-auteur) : la pull request est celle de l'humain.
func stripToolSignature(body string) string {
	lines := strings.Split(body, "\n")
	kept := lines[:0]
	for _, l := range lines {
		low := strings.ToLower(l)
		if strings.Contains(low, "generated with") || strings.Contains(low, "co-authored-by") || strings.Contains(l, "🤖") {
			continue
		}
		kept = append(kept, l)
	}
	return strings.TrimSpace(strings.Join(kept, "\n"))
}

// ProposePRDraft lance l'agent en lecture seule dans le worktree du chantier.
// N'écrit rien : ni état, ni forge.
func ProposePRDraft(ctx context.Context, agent Agent, project Project, card Card, b CardBranch) (string, string, error) {
	commits := prLog(b.WorktreeDir, b.Base, b.Branch, true)
	if len(commits) == 0 {
		return "", "", errors.New("nothing to describe: the workstream has no commits")
	}
	stat, _ := runGit(b.WorktreeDir, gitDefaultTimeout, "diff", "--stat", freshestRef(b.WorktreeDir, b.Base)+"..."+b.Branch)
	prompt := prDraftInstructions(project, card, b, commits, truncate(strings.TrimSpace(stat), 4000))

	ctx, cancel := context.WithTimeout(ctx, planTimeout)
	defer cancel()
	var reply string
	var err error
	switch agent.Cli {
	case "claude":
		reply, err = runPlanClaude(ctx, agent, b.WorktreeDir, prompt)
	case "codex":
		reply, err = runPlanCodex(ctx, agent, b.WorktreeDir, prompt)
	case "fake":
		raw, _ := json.Marshal(map[string]string{
			"title": defaultPRTitle(b.WorktreeDir, b.Base, b.Branch, card.Title),
			"body":  prBody(b.WorktreeDir, b.Base, b.Branch),
		})
		reply = string(raw)
	default:
		return "", "", errors.New(agent.Name + " cannot run read-only")
	}
	if err != nil {
		return "", "", err
	}
	title, body, ok := extractPRDraft(reply)
	if !ok {
		return "", "", errors.New("the agent did not return a pull request draft")
	}
	return title, body, nil
}

// handleCardPRDraft : POST /api/cards/{id}/pr-draft {"repoName"}. Synchrone,
// lecture seule, jusqu'à planTimeout.
func (s *Server) handleCardPRDraft(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RepoName string `json:"repoName"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	card, project, ok := s.cardWithProject(w, r.PathValue("id"))
	if !ok {
		return
	}
	b, ok := s.store.GetCardBranch(card.ID, body.RepoName)
	if !ok || b.WorktreeDir == "" {
		writeError(w, http.StatusNotFound, "workstream branch not found")
		return
	}
	agent, ok := s.prDraftAgent(card, b.RepoName)
	if !ok {
		writeError(w, http.StatusUnprocessableEntity, "no agent can draft a pull request")
		return
	}
	title, text, err := ProposePRDraft(r.Context(), agent, project, card, b)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, PRDraft{RepoName: b.RepoName, AgentID: agent.ID, Title: title, Body: text})
}
