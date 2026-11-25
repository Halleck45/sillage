package server

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// permissionProfile bundles the command permissions and writable package
// cache required by a development stack. Profiles are enabled explicitly in
// project settings; repository contents never grant command permissions.
type permissionProfile struct {
	ID           string
	AllowedTools []string
	// Denied s'ajoute au refus passé au CLI quand le profil est actif.
	Denied []string
	// Environment est nil pour un profil sans cache (git, gh).
	Environment func(string) map[string]string
}

// ghWriteDenied remplace Bash(gh:*) dans le refus quand le profil gh est
// actif : un refus l'emporte sur une autorisation, Bash(gh:*) bloquerait aussi
// la lecture. Toute écriture sur la forge reste refusée, gh api compris (il
// sait faire un POST). Une sous-commande absente d'ici n'est pas autorisée
// pour autant : seule l'allowlist l'est, ce refus garde le cas où l'humain
// ajouterait Bash(gh:*) à Project.AllowedTools.
var ghWriteDenied = []string{
	"Bash(gh api:*)", "Bash(gh browse:*)",
	"Bash(gh pr create:*)", "Bash(gh pr merge:*)", "Bash(gh pr close:*)", "Bash(gh pr reopen:*)",
	"Bash(gh pr edit:*)", "Bash(gh pr comment:*)", "Bash(gh pr review:*)", "Bash(gh pr ready:*)",
	"Bash(gh pr checkout:*)", "Bash(gh pr lock:*)", "Bash(gh pr unlock:*)", "Bash(gh pr revert:*)",
	"Bash(gh pr update-branch:*)",
	"Bash(gh issue create:*)", "Bash(gh issue close:*)", "Bash(gh issue reopen:*)",
	"Bash(gh issue edit:*)", "Bash(gh issue comment:*)", "Bash(gh issue delete:*)",
	"Bash(gh issue transfer:*)", "Bash(gh issue pin:*)", "Bash(gh issue unpin:*)",
	"Bash(gh issue lock:*)", "Bash(gh issue unlock:*)", "Bash(gh issue develop:*)",
	"Bash(gh run rerun:*)", "Bash(gh run cancel:*)", "Bash(gh run delete:*)", "Bash(gh run download:*)",
	"Bash(gh workflow run:*)", "Bash(gh workflow enable:*)", "Bash(gh workflow disable:*)",
	"Bash(gh release create:*)", "Bash(gh release edit:*)", "Bash(gh release delete:*)",
	"Bash(gh release upload:*)", "Bash(gh release delete-asset:*)", "Bash(gh release download:*)",
	"Bash(gh repo create:*)", "Bash(gh repo delete:*)", "Bash(gh repo edit:*)", "Bash(gh repo fork:*)",
	"Bash(gh repo rename:*)", "Bash(gh repo archive:*)", "Bash(gh repo unarchive:*)",
	"Bash(gh repo sync:*)", "Bash(gh repo clone:*)", "Bash(gh repo set-default:*)",
	"Bash(gh repo deploy-key:*)", "Bash(gh repo autolink:*)",
	"Bash(gh secret:*)", "Bash(gh variable:*)", "Bash(gh auth:*)", "Bash(gh gist:*)",
	"Bash(gh extension:*)", "Bash(gh label:*)", "Bash(gh ssh-key:*)", "Bash(gh gpg-key:*)",
	"Bash(gh codespace:*)", "Bash(gh cache:*)", "Bash(gh ruleset:*)", "Bash(gh project:*)",
	"Bash(gh org:*)", "Bash(gh alias:*)", "Bash(gh config:*)", "Bash(gh agent-task:*)",
	"Bash(gh copilot:*)", "Bash(gh attestation:*)",
}

// claudeDenied est le refus réellement passé au CLI claude d'une tâche : le
// refus figé, où Bash(gh:*) cède la place à ghWriteDenied si le profil gh est
// actif, plus les refus propres aux profils actifs. Il ne fait que grandir,
// sauf pour gh, et seulement vers la lecture.
func claudeDenied(profileIDs []string) string {
	denied := claudeDeniedTools
	var extra []string
	for _, id := range profileIDs {
		profile, ok := permissionProfileByID(id)
		if !ok {
			continue
		}
		if profile.ID == "gh" {
			denied = strings.Replace(denied, "Bash(gh:*),", strings.Join(ghWriteDenied, ",")+",", 1)
		}
		extra = append(extra, profile.Denied...)
	}
	if len(extra) > 0 {
		denied += "," + strings.Join(extra, ",")
	}
	return denied
}

var permissionProfiles = []permissionProfile{
	{
		// git au-delà du socle (qui lit et commite) : rejouer, fusionner,
		// défaire dans le worktree de la tâche, dont la branche est jetable, et
		// lire l'historique plus finement. Ni push (refus figé), ni amend, ni
		// changement ou suppression de branche : la tâche reste sur la sienne
		// (voir IsOnBranch). fetch est une lecture réseau.
		ID: "git",
		AllowedTools: []string{
			"Bash(git rebase:*)", "Bash(git merge:*)", "Bash(git cherry-pick:*)",
			"Bash(git reset:*)", "Bash(git restore:*)", "Bash(git stash:*)",
			"Bash(git rm:*)", "Bash(git mv:*)", "Bash(git fetch:*)",
			"Bash(git blame:*)", "Bash(git grep:*)", "Bash(git rev-parse:*)",
			"Bash(git rev-list:*)", "Bash(git merge-base:*)", "Bash(git ls-files:*)",
			"Bash(git show-ref:*)", "Bash(git describe:*)", "Bash(git shortlog:*)",
			"Bash(git reflog:*)", "Bash(git range-diff:*)",
			"Bash(git branch --show-current:*)", "Bash(git branch --list:*)",
			"Bash(git worktree list:*)",
		},
		Denied: []string{"Bash(git rebase --exec:*)", "Bash(git rebase -x:*)"},
	},
	{
		// gh en lecture seule : relire une PR, ses commentaires, sa CI, une
		// issue. Rien n'est écrit sur la forge hors livraison (invariant 3) :
		// les sous-commandes d'écriture et gh api restent refusées (ghWriteDenied).
		ID: "gh",
		AllowedTools: []string{
			"Bash(gh pr view:*)", "Bash(gh pr list:*)", "Bash(gh pr diff:*)",
			"Bash(gh pr checks:*)", "Bash(gh pr status:*)",
			"Bash(gh issue view:*)", "Bash(gh issue list:*)", "Bash(gh issue status:*)",
			"Bash(gh run view:*)", "Bash(gh run list:*)", "Bash(gh run watch:*)",
			"Bash(gh workflow list:*)", "Bash(gh workflow view:*)",
			"Bash(gh release list:*)", "Bash(gh release view:*)",
			"Bash(gh repo view:*)", "Bash(gh search:*)", "Bash(gh status:*)",
		},
	},
	{
		ID: "ruby",
		AllowedTools: []string{
			"Bash(bundle:*)", "Bash(bundler:*)", "Bash(gem:*)",
			"Bash(rake:*)", "Bash(ruby:*)",
		},
		Environment: func(root string) map[string]string {
			return map[string]string{
				"BUNDLE_APP_CONFIG": filepath.Join(root, "config"),
				"BUNDLE_BIN":        filepath.Join(root, "bin"),
				"BUNDLE_PATH":       filepath.Join(root, "bundle"),
				"BUNDLE_USER_CACHE": filepath.Join(root, "cache"),
				"GEM_HOME":          filepath.Join(root, "gems"),
			}
		},
	},
	{
		ID: "node",
		AllowedTools: []string{
			"Bash(node:*)", "Bash(npm:*)", "Bash(npx:*)",
			"Bash(pnpm:*)", "Bash(yarn:*)",
		},
		Environment: func(root string) map[string]string {
			return map[string]string{
				"COREPACK_HOME":      filepath.Join(root, "corepack"),
				"NPM_CONFIG_CACHE":   filepath.Join(root, "npm"),
				"PNPM_HOME":          filepath.Join(root, "pnpm"),
				"YARN_CACHE_FOLDER":  filepath.Join(root, "yarn"),
				"YARN_GLOBAL_FOLDER": filepath.Join(root, "yarn-global"),
			}
		},
	},
	{
		ID: "php",
		AllowedTools: []string{
			"Bash(composer:*)", "Bash(php:*)",
		},
		Environment: func(root string) map[string]string {
			return map[string]string{
				"COMPOSER_CACHE_DIR": filepath.Join(root, "cache"),
				"COMPOSER_HOME":      filepath.Join(root, "composer"),
			}
		},
	},
	{
		ID: "python",
		AllowedTools: []string{
			"Bash(pip:*)", "Bash(pip3:*)", "Bash(pytest:*)",
			"Bash(python:*)", "Bash(python3:*)", "Bash(ruff:*)", "Bash(uv:*)",
		},
		Environment: func(root string) map[string]string {
			return map[string]string{
				"PIP_CACHE_DIR":                filepath.Join(root, "pip"),
				"PIPENV_CACHE_DIR":             filepath.Join(root, "pipenv"),
				"POETRY_CACHE_DIR":             filepath.Join(root, "poetry"),
				"UV_CACHE_DIR":                 filepath.Join(root, "uv"),
				"VIRTUALENV_OVERRIDE_APP_DATA": filepath.Join(root, "virtualenv"),
			}
		},
	},
	{
		ID: "go",
		AllowedTools: []string{
			"Bash(go:*)", "Bash(gofmt:*)", "Bash(golangci-lint:*)",
		},
		Environment: func(root string) map[string]string {
			return map[string]string{
				"GOBIN":      filepath.Join(root, "bin"),
				"GOCACHE":    filepath.Join(root, "build"),
				"GOMODCACHE": filepath.Join(root, "modules"),
			}
		},
	},
	{
		ID: "rust",
		AllowedTools: []string{
			"Bash(cargo:*)", "Bash(clippy:*)", "Bash(rustc:*)", "Bash(rustfmt:*)",
		},
		Environment: func(root string) map[string]string {
			return map[string]string{
				"CARGO_HOME":       filepath.Join(root, "cargo"),
				"CARGO_TARGET_DIR": filepath.Join(root, "target"),
			}
		},
	},
}

func permissionProfileByID(id string) (permissionProfile, bool) {
	for _, profile := range permissionProfiles {
		if profile.ID == id {
			return profile, true
		}
	}
	return permissionProfile{}, false
}

// NormalizePermissionProfiles validates user-selected profile IDs, removes
// duplicates, and returns them in the stable order used by the UI.
func NormalizePermissionProfiles(ids []string) ([]string, error) {
	selected := make(map[string]bool, len(ids))
	for _, raw := range ids {
		id := strings.TrimSpace(raw)
		if id == "" {
			continue
		}
		if _, ok := permissionProfileByID(id); !ok {
			return nil, fmt.Errorf("unknown permission profile: %s", id)
		}
		selected[id] = true
	}
	out := []string{}
	for _, profile := range permissionProfiles {
		if selected[profile.ID] {
			out = append(out, profile.ID)
		}
	}
	return out, nil
}

func permissionProfileTools(ids []string) []string {
	var tools []string
	for _, id := range ids {
		if profile, ok := permissionProfileByID(id); ok {
			tools = append(tools, profile.AllowedTools...)
		}
	}
	return tools
}

type permissionRuntime struct {
	Env          []string
	WritableDirs []string
}

// preparePermissionRuntime creates task-scoped package caches outside Git
// worktrees. This keeps downloads reusable across follow-up turns without
// letting one task poison another task's toolchain or polluting reviewed diffs.
// WritableDirs lets the Codex sandbox expose only the selected caches.
func preparePermissionRuntime(dataDir string, task Task, project Project) (permissionRuntime, error) {
	if len(project.PermissionProfiles) == 0 {
		return permissionRuntime{}, nil
	}
	cacheRoot, err := filepath.Abs(filepath.Join(dataDir, "package-cache", task.ID))
	if err != nil {
		return permissionRuntime{}, err
	}
	overrides := map[string]string{}
	var pathDirs, writableDirs []string
	for _, id := range project.PermissionProfiles {
		profile, ok := permissionProfileByID(id)
		if !ok {
			continue
		}
		if profile.Environment == nil {
			continue
		}
		root := filepath.Join(cacheRoot, profile.ID)
		if err := os.MkdirAll(root, 0o700); err != nil {
			return permissionRuntime{}, fmt.Errorf("cannot prepare %s permission profile: %w", profile.ID, err)
		}
		writableDirs = append(writableDirs, root)
		for key, value := range profile.Environment(root) {
			overrides[key] = value
		}
		switch profile.ID {
		case "ruby":
			pathDirs = append(pathDirs, filepath.Join(root, "bin"), filepath.Join(root, "gems", "bin"))
		case "node":
			pathDirs = append(pathDirs, filepath.Join(root, "pnpm"))
		case "php":
			pathDirs = append(pathDirs, filepath.Join(root, "composer", "vendor", "bin"))
		case "go":
			pathDirs = append(pathDirs, filepath.Join(root, "bin"))
		case "rust":
			pathDirs = append(pathDirs, filepath.Join(root, "cargo", "bin"))
		}
	}
	if len(pathDirs) > 0 {
		overrides["PATH"] = strings.Join(append(pathDirs, os.Getenv("PATH")), string(os.PathListSeparator))
	}
	return permissionRuntime{
		Env:          environmentWithOverrides(os.Environ(), overrides),
		WritableDirs: writableDirs,
	}, nil
}

func environmentWithOverrides(base []string, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return base
	}
	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	blocked := make(map[string]bool, len(keys))
	for _, key := range keys {
		blocked[key] = true
	}
	out := make([]string, 0, len(base)+len(keys))
	for _, entry := range base {
		key, _, found := strings.Cut(entry, "=")
		if !found || !blocked[key] {
			out = append(out, entry)
		}
	}
	for _, key := range keys {
		out = append(out, key+"="+overrides[key])
	}
	return out
}
