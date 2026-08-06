package server

// Plan de tâches : demander à un agent de découper un chantier en une
// succession d'étapes. La planification est en LECTURE SEULE et ne crée rien :
// elle rend une proposition, l'humain la relit, la corrige, puis crée les
// tâches lui-même par les routes ordinaires. Sillage ne lance jamais un agent
// sur une tâche que personne n'a validée.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const (
	// planMaxSteps plafonne la proposition. Au-delà, ce n'est plus un découpage
	// relisable en un coup d'œil mais une liste de choses à faire, et chaque
	// étape est une tâche qu'un humain devra relire séparément.
	planMaxSteps = 8

	// planTitleMax : un titre de tâche tient sur une ligne de la liste. Un
	// modèle qui rédige une phrase entière est tronqué plutôt que refusé.
	planTitleMax = 120

	// planTimeout : lire un dépôt et écrire un plan est court. Passé ce délai,
	// l'agent s'est perdu ; personne n'attend un plan dix minutes.
	planTimeout = 5 * time.Minute
)

// planAllowedTools : la planification lit, elle n'écrit pas. Ni Edit, ni Write,
// ni la moindre commande git capable d'écrire. Ce n'est pas une précaution de
// confort : un planificateur tourne dans le worktree du chantier ou dans le
// dépôt du projet lui-même, pas dans la branche jetable d'une tâche, donc rien
// de ce qu'il écrirait ne passerait par une revue.
const planAllowedTools = "Read,Glob,Grep," +
	"Bash(git status:*),Bash(git diff:*),Bash(git log:*),Bash(git show:*)," +
	"Bash(ls:*),Bash(cat:*)"

// canPlan : seuls les CLI dont Sillage peut garantir le mode lecture seule
// savent planifier, par allowlist figée (claude) ou par sandbox (codex). Les
// autres adaptateurs n'ont pas de cran d'arrêt en écriture que Sillage pose
// lui-même, et un plan ne vaut pas d'y renoncer. L'agent simulé est de la
// partie : il permet d'éprouver le flux sans coût.
func canPlan(cli string) bool {
	switch cli {
	case "claude", "codex", "fake":
		return true
	}
	return false
}

// planPayload est le format demandé à l'agent : un seul objet JSON.
type planPayload struct {
	Steps []PlanStep `json:"steps"`
}

// planInstructions rédige la demande envoyée à l'agent. Le contexte du projet
// et celui du chantier y entrent comme partout ailleurs (voir contextParts) :
// un plan écrit sans les instructions du projet propose des étapes que le
// projet n'accepte pas. La langue n'est pas imposée par un réglage mais alignée
// sur le titre du chantier, qui est ce que l'humain vient d'écrire.
func planInstructions(card Card, project Project, extra string) string {
	var b strings.Builder
	b.WriteString("You are planning a workstream. You are NOT implementing it: do not modify, create or delete any file.\n\n")
	b.WriteString("Workstream: " + card.Title + "\n")
	if blocks := strings.Join(contextParts(project.ContextPrompt, card.ContextPrompt), "\n\n"); blocks != "" {
		b.WriteString("\n" + blocks + "\n")
	}
	if extra = strings.TrimSpace(extra); extra != "" {
		b.WriteString("\nAdditional instructions from the human:\n" + extra + "\n")
	}
	b.WriteString("\nRead the repository to understand how it is built, then split this workstream")
	b.WriteString(" into an ordered succession of implementation tasks. Each task will be handed to")
	b.WriteString(" a single coding agent working alone in its own branch, and reviewed by a human on")
	b.WriteString(" its own, so each one must be coherent and worth reviewing by itself. Each task")
	b.WriteString(" starts from the work accepted by the previous one, so a task may rely on what the")
	b.WriteString(" previous ones landed. Prefer few substantial steps over many small ones: aim for 2")
	b.WriteString(" to 5 tasks, never more than ")
	fmt.Fprintf(&b, "%d", planMaxSteps)
	b.WriteString(".\n\n")
	b.WriteString("End your reply with a single JSON object and nothing after it:\n")
	b.WriteString(`{"steps":[{"title":"...","prompt":"..."}]}` + "\n")
	b.WriteString("- title: a short imperative label, no numbering, no trailing period.\n")
	b.WriteString("- prompt: the complete brief for the agent doing that task: what to change, where, and how to check it worked.\n")
	b.WriteString("Write titles and prompts in the same language as the workstream title above.\n")
	return b.String()
}

// ProposePlan lance l'agent en lecture seule dans dir et retourne les étapes
// qu'il propose. N'écrit rien : ni tâche, ni worktree, ni état.
func ProposePlan(ctx context.Context, agent Agent, project Project, card Card, dir, extra string) ([]PlanStep, error) {
	if !canPlan(agent.Cli) {
		return nil, fmt.Errorf("%s cannot plan", agent.Name)
	}
	prompt := planInstructions(card, project, extra)

	ctx, cancel := context.WithTimeout(ctx, planTimeout)
	defer cancel()

	var reply string
	var err error
	switch agent.Cli {
	case "claude":
		reply, err = runPlanClaude(ctx, agent, dir, prompt)
	case "codex":
		reply, err = runPlanCodex(ctx, agent, dir, prompt)
	case "fake":
		reply, err = fakePlanReply(card), nil
	}
	if err != nil {
		return nil, err
	}
	steps, ok := extractPlanSteps(reply)
	if !ok {
		return nil, errors.New("the agent did not return a plan")
	}
	return steps, nil
}

func runPlanClaude(ctx context.Context, agent Agent, dir, prompt string) (string, error) {
	// Sortie texte : un plan n'a ni fil de conversation ni tokens à suivre, le
	// flux JSONL de l'adaptateur de tâche n'apporterait rien ici.
	args := []string{"-p", "--allowedTools", planAllowedTools, "--disallowedTools", claudeDeniedTools}
	if agent.Model != "" {
		args = append(args, "--model", agent.Model)
	}
	args = append(args, prompt)
	return runPlanProcess(ctx, "claude", args, dir)
}

func runPlanCodex(ctx context.Context, agent Agent, dir, prompt string) (string, error) {
	// read-only en dur, et pas SILLAGE_CODEX_SANDBOX : ce réglage existe pour
	// élargir les droits d'écriture d'une tâche, ce dont un plan n'a aucun
	// besoin. Sur une machine où le sandbox de codex ne démarre pas du tout
	// (AppArmor, voir agentWarning), planifier avec codex échoue comme
	// n'importe quelle tâche codex : planifier avec claude reste possible.
	args := []string{"exec", "--json", "--sandbox", "read-only", "-C", dir}
	if agent.Model != "" {
		args = append(args, "--model", agent.Model)
	}
	args = append(args, prompt)
	out, err := runPlanProcess(ctx, "codex", args, dir)
	if err != nil {
		return "", err
	}
	return lastCodexAgentMessage(out), nil
}

// lastCodexAgentMessage aplatit le flux --json de codex en gardant le dernier
// message de l'agent : c'est sa réponse finale, celle qui porte le JSON.
func lastCodexAgentMessage(stream string) string {
	last := ""
	for _, line := range strings.Split(stream, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var generic map[string]any
		if err := json.Unmarshal([]byte(line), &generic); err != nil {
			continue
		}
		if text, ok := extractCodexMessage(generic); ok {
			last = text
		}
	}
	if last == "" {
		// Un codex qui n'a rien émis en JSON a peut-être écrit en clair : le
		// parseur de plan saura toujours y chercher son objet.
		return stream
	}
	return last
}

// runPlanProcess exécute un CLI de planification et retourne sa sortie
// standard. Setpgid + Cancel : sans quoi l'expiration du délai tuerait le CLI
// mais laisserait tourner ses enfants.
func runPlanProcess(ctx context.Context, bin string, args []string, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", errors.New("the agent took too long to plan")
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", truncate(msg, 400))
	}
	return stdout.String(), nil
}

// extractPlanSteps retrouve l'objet JSON du plan dans la réponse de l'agent.
// Les CLI encadrent volontiers leur JSON de prose ou de barrières ```json, et
// certains modèles recopient l'exemple du prompt avant de répondre : on tente
// donc de décoder à partir de chaque accolade ouvrante et on garde le DERNIER
// objet exploitable, la réponse finale étant à la fin.
func extractPlanSteps(reply string) ([]PlanStep, bool) {
	var best []PlanStep
	for i := 0; i < len(reply); i++ {
		if reply[i] != '{' {
			continue
		}
		var payload planPayload
		// Un décodeur plutôt qu'Unmarshal : il s'arrête à la fin de la première
		// valeur et ignore ce qui suit, donc la prose de queue ne gêne pas.
		if err := json.NewDecoder(strings.NewReader(reply[i:])).Decode(&payload); err != nil {
			continue
		}
		if steps := sanitizePlanSteps(payload.Steps); len(steps) > 0 {
			best = steps
		}
	}
	return best, len(best) > 0
}

// sanitizePlanSteps met la proposition aux normes du produit : une étape sans
// titre n'existe pas, un titre tient sur une ligne, et le plan s'arrête à
// planMaxSteps. Tronquer plutôt que refuser : un plan presque bon reste
// éditable par l'humain, un plan refusé le renvoie à zéro.
func sanitizePlanSteps(raw []PlanStep) []PlanStep {
	steps := make([]PlanStep, 0, len(raw))
	for _, s := range raw {
		title := strings.TrimSpace(s.Title)
		if title == "" {
			continue
		}
		steps = append(steps, PlanStep{
			Title:  truncate(title, planTitleMax),
			Prompt: strings.TrimSpace(s.Prompt),
		})
		if len(steps) == planMaxSteps {
			break
		}
	}
	return steps
}

// fakePlanReply imite la réponse d'un vrai CLI, prose et barrières comprises,
// pour que l'agent simulé éprouve aussi le parseur et pas seulement l'UI.
func fakePlanReply(card Card) string {
	payload := planPayload{Steps: []PlanStep{
		{
			Title:  "Préparer le terrain",
			Prompt: "Chantier simulé « " + card.Title + " ». Première étape : repérer les fichiers concernés et poser les structures de données, sans changer de comportement.",
		},
		{
			Title:  "Implémenter le cœur",
			Prompt: "Chantier simulé « " + card.Title + " ». Deuxième étape : écrire le comportement attendu par-dessus l'étape précédente, avec ses tests.",
		},
		{
			Title:  "Brancher l'interface",
			Prompt: "Chantier simulé « " + card.Title + " ». Dernière étape : rendre le tout accessible depuis l'interface et mettre la documentation à jour.",
		},
	}}
	body, _ := json.Marshal(payload)
	return "Voici le découpage proposé.\n\n```json\n" + string(body) + "\n```\n"
}
