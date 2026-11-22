package server

// Exécution en lecture seule avec suivi en direct : relecture d'un signal de la
// Vigie et autoconfiguration d'un projet. Mêmes garanties que la planification
// (planAllowedTools pour claude, sandbox read-only pour codex), mais la sortie
// est lue au fil de l'eau pour remonter chaque outil employé : un agent qui
// travaille dix minutes sans rien montrer ne se distingue pas d'un agent
// planté.

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

// readOnlyStepsMax : les dernières étapes gardées pour l'affichage.
const readOnlyStepsMax = 12

// appendStep ajoute une étape en gardant les readOnlyStepsMax dernières.
func appendStep(steps []string, line string) []string {
	steps = append(steps, line)
	if len(steps) > readOnlyStepsMax {
		steps = append([]string(nil), steps[len(steps)-readOnlyStepsMax:]...)
	}
	return steps
}

// runReadOnly lance agent en lecture seule dans dir et retourne sa réponse
// finale. onStep reçoit une ligne par outil employé (peut être nil). fake n'est
// pas géré ici : chaque appelant simule sa propre réponse.
func runReadOnly(ctx context.Context, agent Agent, dir, prompt string, onStep func(string)) (string, error) {
	if onStep == nil {
		onStep = func(string) {}
	}
	// Les chemins absolus du dépôt mangent la ligne : on les montre relatifs.
	prefix := strings.TrimRight(dir, "/") + "/"
	raw := onStep
	onStep = func(line string) { raw(strings.ReplaceAll(line, prefix, "")) }
	switch agent.Cli {
	case "claude":
		args := []string{"-p", "--output-format", "stream-json", "--verbose",
			"--allowedTools", planAllowedTools, "--disallowedTools", claudeDeniedTools}
		if agent.Model != "" {
			args = append(args, "--model", agent.Model)
		}
		args = append(args, prompt)
		return streamReadOnly(ctx, "claude", args, dir, claudeStreamParser(onStep))
	case "codex":
		args := []string{"exec", "--json", "--sandbox", "read-only", "-C", dir}
		if agent.Model != "" {
			args = append(args, "--model", agent.Model)
		}
		args = append(args, prompt)
		return streamReadOnly(ctx, "codex", args, dir, codexStreamParser(onStep))
	}
	return "", fmt.Errorf("%s cannot run read-only", agent.Name)
}

// streamParser lit une ligne de sortie ; final retourne la réponse retenue.
type streamParser struct {
	line  func(string)
	final func() string
}

func claudeStreamParser(onStep func(string)) streamParser {
	var result, lastText string
	var failed bool
	return streamParser{
		line: func(raw string) {
			var env claudeEnvelope
			if err := json.Unmarshal([]byte(raw), &env); err != nil {
				return
			}
			switch env.Type {
			case "assistant":
				if env.Message == nil {
					return
				}
				for _, b := range env.Message.Content {
					switch b.Type {
					case "text":
						if strings.TrimSpace(b.Text) != "" {
							lastText = b.Text
						}
					case "tool_use":
						onStep(summarizeToolUse(b.Name, b.Input))
					}
				}
			case "result":
				result, failed = env.Result, env.IsError
			}
		},
		final: func() string {
			if failed && result != "" {
				// is_error : quota épuisé, authentification... Le texte du
				// résultat est le message d'erreur, pas une réponse.
				return "\x00" + result
			}
			if result != "" {
				return result
			}
			return lastText
		},
	}
}

func codexStreamParser(onStep func(string)) streamParser {
	var last string
	return streamParser{
		line: func(raw string) {
			var m map[string]any
			if err := json.Unmarshal([]byte(raw), &m); err != nil {
				return
			}
			if text, ok := extractCodexMessage(m); ok {
				last = text
				return
			}
			if m["type"] == "item.started" {
				if item, ok := m["item"].(map[string]any); ok && item["type"] == "command_execution" {
					if cmd, ok := item["command"].(string); ok && cmd != "" {
						onStep("Bash · " + truncate(cmd, 80))
					}
				}
			}
		},
		final: func() string { return last },
	}
}

// streamReadOnly exécute bin, passe chaque ligne de stdout au parseur et
// retourne sa réponse finale. Setpgid + Cancel comme runPlanProcess : à
// l'expiration, tout le groupe meurt, enfants compris.
func streamReadOnly(ctx context.Context, bin string, args []string, dir string, p streamParser) (string, error) {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("%s is not installed (not found in PATH)", bin)
		}
		return "", err
	}
	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			p.line(line)
		}
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	reply := p.final()
	if strings.HasPrefix(reply, "\x00") {
		return "", fmt.Errorf("%s", truncate(strings.TrimPrefix(reply, "\x00"), 400))
	}
	if waitErr != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = waitErr.Error()
		}
		return "", fmt.Errorf("%s", truncate(msg, 400))
	}
	if strings.TrimSpace(reply) == "" {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", fmt.Errorf("%s", truncate(msg, 400))
		}
		return "", fmt.Errorf("%s produced no output", bin)
	}
	return reply, nil
}
