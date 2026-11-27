package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Adaptateur GitHub Copilot : `copilot --acp`, piloté par le client ACP
// (acp.go, acp_runner.go). Ce qui lui est propre tient ici : ses règles
// d'outils, son dossier de configuration isolé et la notification qui annonce
// le début d'un tour.

// Copilot deny rules override the selectively allowed read, write, and shell
// tools. They keep outbound Git operations and agent-controlled settings out
// of a task even when the CLI runs unattended.
const copilotDeniedTools = "shell(git push),shell(gh:*),shell(glab:*)," +
	"write(.github/copilot-instructions.md)," +
	"write(.github/copilot/settings.json)," +
	"write(.github/copilot/settings.local.json)"

// agentHomesDir : sous le dataDir, un dossier de configuration par tâche et
// par CLI (`agent-homes/<taskId>/copilot`), supprimé avec la tâche.
const agentHomesDir = "agent-homes"

// copilotSignInHint remplace l'erreur d'authentification du CLI, qui renvoie
// vers un bouton de connexion que Sillage n'a pas.
const copilotSignInHint = "copilot is not signed in: run `copilot login` in a terminal, or set COPILOT_GITHUB_TOKEN"

// copilotArgs lance Copilot en serveur ACP. Les autorisations sont fixées ici,
// d'avance : lecture, écriture et shell sont permis, les refus l'emportent
// (deny wins), et tout ce qui sort de ce cadre remonte en demande
// d'autorisation que Sillage refuse (acpRejectAll). Ni --allow-all ni mode
// autopilot, qui équivaut à tout autoriser.
func copilotArgs(agent Agent) []string {
	args := []string{
		"--acp",
		"--no-ask-user",
		"--allow-tool=read,write,shell",
		"--deny-tool=" + copilotDeniedTools,
		"--disable-builtin-mcps",
		"--no-auto-update",
		"--sync-session=off",
	}
	if agent.Model != "" {
		args = append(args, "--model="+agent.Model)
	}
	return args
}

// copilotSettings est le settings.json du dossier isolé, réécrit à chaque
// exécution : pas d'export des sessions vers GitHub, pas de mise à jour.
const copilotSettings = `{
  "remoteExport": false,
  "autoUpdate": false
}
`

// copilotUserHome est le dossier de configuration de l'utilisateur, celui où
// `copilot login` a enregistré le compte.
func copilotUserHome() string {
	if dir := strings.TrimSpace(os.Getenv("COPILOT_HOME")); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".copilot")
}

// readCopilotConfig lit un config.json de Copilot, qui commence par des lignes
// de commentaire (`// ...`) que encoding/json refuse.
func readCopilotConfig(path string) map[string]json.RawMessage {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var kept []string
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "//") {
			kept = append(kept, line)
		}
	}
	var config map[string]json.RawMessage
	if json.Unmarshal([]byte(strings.Join(kept, "\n")), &config) != nil {
		return nil
	}
	return config
}

// copilotAccountKeys : ce que le dossier isolé reprend de la configuration de
// l'utilisateur. Ce sont des noms de compte, pas des secrets : le jeton reste
// dans le trousseau du système, où le CLI le retrouve par compte.
var copilotAccountKeys = []string{"lastLoggedInUser", "loggedInUsers"}

// prepareCopilotHome crée (ou remet d'aplomb) le dossier de configuration
// d'une tâche et rend son chemin.
//
// Pourquoi un dossier à part : Copilot charge depuis son dossier les hooks,
// serveurs MCP et plugins de l'utilisateur, et un hook qui échoue vaut refus
// (fail-closed). Un seul hook de permission mal écrit pour Copilot suffit donc
// à refuser toutes les écritures d'une tâche, sans message utile. Le dossier
// isolé n'hérite de rien de tout cela. Il est propre à la tâche et persiste
// d'un tour à l'autre : c'est là que Copilot range la session à reprendre.
func prepareCopilotHome(dataDir, taskID, userHome string) (string, error) {
	home, err := filepath.Abs(filepath.Join(dataDir, agentHomesDir, taskID, "copilot"))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(home, 0o700); err != nil {
		return "", fmt.Errorf("cannot prepare the copilot home: %w", err)
	}
	if err := os.WriteFile(filepath.Join(home, "settings.json"), []byte(copilotSettings), 0o600); err != nil {
		return "", fmt.Errorf("cannot prepare the copilot home: %w", err)
	}

	// Le compte est recopié à chaque tour : l'utilisateur a pu en changer. Le
	// reste du fichier appartient au CLI, qui y tient son propre état.
	configPath := filepath.Join(home, "config.json")
	config := readCopilotConfig(configPath)
	if config == nil {
		config = map[string]json.RawMessage{}
	}
	user := map[string]json.RawMessage{}
	if userHome != "" && filepath.Clean(userHome) != home {
		user = readCopilotConfig(filepath.Join(userHome, "config.json"))
	}
	for _, key := range copilotAccountKeys {
		if value, ok := user[key]; ok {
			config[key] = value
		} else {
			delete(config, key)
		}
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(configPath, append(data, '\n'), 0o600); err != nil {
		return "", fmt.Errorf("cannot prepare the copilot home: %w", err)
	}
	return home, nil
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
//
// home, when set, becomes COPILOT_HOME (see prepareCopilotHome).
func copilotEnv(base []string, home string) []string {
	env := base
	if env == nil {
		env = os.Environ()
	}
	filtered := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if strings.HasPrefix(kv, "GITHUB_TOKEN=ghp_") {
			continue
		}
		if home != "" && strings.HasPrefix(kv, "COPILOT_HOME=") {
			continue
		}
		filtered = append(filtered, kv)
	}
	if home != "" {
		filtered = append(filtered, "COPILOT_HOME="+home)
	}
	return filtered
}

// copilotMeta : extensions annoncées à l'ouverture. interactive=false dit au
// CLI qu'aucun humain ne pilote la session ; events demande la notification de
// début de tour (voir copilotTurnStart).
var copilotMeta = map[string]any{
	"github.com/copilot": map[string]any{
		"interactive": false,
		"events":      []string{"assistant.turn_start"},
	},
}

// copilotTurnStart reconnaît le début d'un tour de l'agent principal. Copilot
// écrit ses avertissements de démarrage (« Run /login again to use
// Connectors ») dans le canal des messages, sans rien qui les distingue d'une
// réponse : seul cet événement dit où la réponse commence. Un événement qui
// porte agentId vient d'un sous-agent.
func copilotTurnStart(method string, params json.RawMessage) bool {
	if method != "github.com/copilot/sessionEvent" {
		return false
	}
	var event struct {
		Type    string `json:"type"`
		AgentID string `json:"agentId"`
	}
	return json.Unmarshal(params, &event) == nil && event.Type == "assistant.turn_start" && event.AgentID == ""
}

func (r *Runner) runCopilot(task *Task, agent Agent, project Project, card Card, handle *procHandle, cliInput string, runtime permissionRuntime) error {
	home, err := prepareCopilotHome(r.store.DataDir(), task.ID, copilotUserHome())
	if err != nil {
		return err
	}
	return r.runACP(task, agent, handle, acpAgentSpec{
		Binary:     "copilot",
		Args:       copilotArgs(agent),
		Env:        copilotEnv(runtime.Env, home),
		Meta:       copilotMeta,
		Permission: acpRejectAll,
		TurnStart:  copilotTurnStart,
		FreshPrompt: func(text string) string {
			return prefixAgentContext(agent, project, card, commitNote+"\n\n"+text)
		},
		SignInHint: copilotSignInHint,
	}, cliInput)
}
