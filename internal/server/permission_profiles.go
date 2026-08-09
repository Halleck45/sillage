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
	Environment  func(string) map[string]string
}

var permissionProfiles = []permissionProfile{
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
