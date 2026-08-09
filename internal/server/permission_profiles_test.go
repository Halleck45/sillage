package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizePermissionProfilesValidatesAndOrdersSelections(t *testing.T) {
	got, err := NormalizePermissionProfiles([]string{" rust ", "php", "ruby", "rust", ""})
	if err != nil {
		t.Fatalf("normalize profiles: %v", err)
	}
	if strings.Join(got, ",") != "ruby,php,rust" {
		t.Fatalf("stable unique profiles expected ruby,php,rust, got %v", got)
	}
	if _, err := NormalizePermissionProfiles([]string{"unknown"}); err == nil {
		t.Fatal("an unknown profile should be rejected")
	}
}

func TestPHPPermissionProfileUsesWritableComposerCache(t *testing.T) {
	dataDir := t.TempDir()
	project := Project{ID: "p1", PermissionProfiles: []string{"php"}}
	task := Task{ID: "t42"}
	runtime, err := preparePermissionRuntime(dataDir, task, project)
	if err != nil {
		t.Fatalf("prepare PHP runtime: %v", err)
	}
	wantRoot := filepath.Join(dataDir, "package-cache", "t42", "php")
	if len(runtime.WritableDirs) != 1 || runtime.WritableDirs[0] != wantRoot {
		t.Fatalf("one writable PHP cache expected at %q, got %v", wantRoot, runtime.WritableDirs)
	}

	env := environmentMap(runtime.Env)
	for _, key := range []string{"COMPOSER_CACHE_DIR", "COMPOSER_HOME"} {
		if !isDirWithin(filepath.Clean(env[key]), wantRoot) {
			t.Errorf("%s should stay inside the Sillage cache, got %q", key, env[key])
		}
	}
	wantBin := filepath.Join(wantRoot, "composer", "vendor", "bin")
	if !strings.HasPrefix(env["PATH"], wantBin+string(os.PathListSeparator)) {
		t.Fatalf("global Composer executables should be first on PATH, got %q", env["PATH"])
	}
}

func TestRubyPermissionProfileUsesWritableSillageCache(t *testing.T) {
	dataDir := t.TempDir()
	project := Project{ID: "p1", PermissionProfiles: []string{"ruby"}}
	task := Task{ID: "t42"}
	runtime, err := preparePermissionRuntime(dataDir, task, project)
	if err != nil {
		t.Fatalf("prepare Ruby runtime: %v", err)
	}
	if len(runtime.WritableDirs) != 1 {
		t.Fatalf("one writable Ruby cache expected, got %v", runtime.WritableDirs)
	}
	wantRoot := filepath.Join(dataDir, "package-cache", "t42", "ruby")
	if runtime.WritableDirs[0] != wantRoot {
		t.Fatalf("Ruby cache expected %q, got %q", wantRoot, runtime.WritableDirs[0])
	}
	if info, err := os.Stat(wantRoot); err != nil || !info.IsDir() {
		t.Fatalf("Ruby cache was not created: info=%v err=%v", info, err)
	}

	env := environmentMap(runtime.Env)
	for _, key := range []string{"BUNDLE_APP_CONFIG", "BUNDLE_BIN", "BUNDLE_PATH", "BUNDLE_USER_CACHE", "GEM_HOME"} {
		if !isDirWithin(filepath.Clean(env[key]), wantRoot) {
			t.Errorf("%s should stay inside the Sillage cache, got %q", key, env[key])
		}
		if strings.HasPrefix(env[key], "/usr/local") {
			t.Errorf("%s must never target /usr/local, got %q", key, env[key])
		}
	}
	if !strings.HasPrefix(env["PATH"], filepath.Join(wantRoot, "bin")+string(os.PathListSeparator)) {
		t.Fatalf("Ruby executables should be first on PATH, got %q", env["PATH"])
	}
}

func TestPermissionProfilesExtendClaudeAndCodexExplicitly(t *testing.T) {
	tools := claudeTools([]string{"Bash(custom:*)"}, []string{"ruby", "php"})
	for _, required := range []string{"Bash(bundle:*)", "Bash(rake:*)", "Bash(composer:*)", "Bash(php:*)", "Bash(custom:*)"} {
		if !strings.Contains(tools, required) {
			t.Errorf("Claude tools should contain %q: %s", required, tools)
		}
	}

	cache := filepath.Join(t.TempDir(), "ruby")
	if err := os.MkdirAll(cache, 0o700); err != nil {
		t.Fatalf("create cache: %v", err)
	}
	args := codexArgs(t.TempDir(), "workspace-write", []string{cache}, Agent{}, "Install dependencies")
	if !containsAdjacentArgs(args, "--add-dir", cache) {
		t.Fatalf("Codex should receive the selected package cache: %v", args)
	}
	lockedDown := codexArgs(t.TempDir(), "read-only", []string{cache}, Agent{}, "Inspect")
	if containsAdjacentArgs(lockedDown, "--add-dir", cache) {
		t.Fatalf("a read-only Codex run must not gain writable roots: %v", lockedDown)
	}
}

func TestUpdateProjectPermissionProfiles(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	project, err := store.AddProject("demo", "", "", []Repo{{Path: "/tmp/demo"}}, nil, nil)
	if err != nil {
		t.Fatalf("add project: %v", err)
	}
	if project.PermissionProfiles == nil || len(project.PermissionProfiles) != 0 {
		t.Fatalf("a new project should start with an explicit empty selection: %v", project.PermissionProfiles)
	}

	profiles := []string{"rust", "ruby"}
	updated, err := store.UpdateProject(project.ID, nil, nil, nil, nil, nil, nil, nil, nil, nil, &profiles)
	if err != nil {
		t.Fatalf("update profiles: %v", err)
	}
	if strings.Join(updated.PermissionProfiles, ",") != "ruby,rust" {
		t.Fatalf("normalized profiles expected ruby,rust, got %v", updated.PermissionProfiles)
	}

	invalid := []string{"shell-everything"}
	if _, err := store.UpdateProject(project.ID, nil, nil, nil, nil, nil, nil, nil, nil, nil, &invalid); err == nil {
		t.Fatal("an unknown permission profile should be rejected")
	}
}

func TestDeleteTaskRemovesItsPackageCache(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	project, err := store.AddProject("demo", "", "", []Repo{{Name: "demo", Path: t.TempDir()}}, nil, nil)
	if err != nil {
		t.Fatalf("add project: %v", err)
	}
	card, err := store.AddCard(project.ID, "Work", "", "")
	if err != nil {
		t.Fatalf("add card: %v", err)
	}
	taskID := mkTaskWithStatus(t, store, card.ID, project.ID, "review")
	cacheDir := filepath.Join(store.DataDir(), "package-cache", taskID, "ruby")
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		t.Fatalf("create package cache: %v", err)
	}

	runner := NewRunner(store, NewHub())
	if err := runner.DeleteTask(taskID); err != nil {
		t.Fatalf("delete task: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(cacheDir)); !os.IsNotExist(err) {
		t.Fatalf("task package cache should be removed, got %v", err)
	}
}

func environmentMap(env []string) map[string]string {
	out := map[string]string{}
	for _, entry := range env {
		key, value, found := strings.Cut(entry, "=")
		if found {
			out[key] = value
		}
	}
	return out
}
