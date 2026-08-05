package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"
)

// withVersion fait passer le binaire de test pour une version publiée. À
// appeler APRÈS la construction du serveur : sinon NewServer démarrerait la
// vérification périodique, et un test sortirait sur le réseau.
func withVersion(t *testing.T, v string) {
	t.Helper()
	prev := buildVersion
	buildVersion = v
	t.Cleanup(func() { buildVersion = prev })
}

func withInstall(t *testing.T, info installInfo) {
	t.Helper()
	prev := detectInstallFn
	detectInstallFn = func() installInfo { return info }
	t.Cleanup(func() { detectInstallFn = prev })
}

func withLatestRelease(t *testing.T, tag string, err error) {
	t.Helper()
	prev := fetchLatestReleaseFn
	fetchLatestReleaseFn = func(context.Context) (string, string, error) {
		return tag, "https://github.com/" + updateRepo + "/releases/tag/" + tag, err
	}
	t.Cleanup(func() { fetchLatestReleaseFn = prev })
}

func TestVersionParsing(t *testing.T) {
	for _, v := range []string{"0.8.0", "v0.8.0", " v1.2.3 ", "v1.0.0-rc1"} {
		if !isReleaseVersion(v) {
			t.Errorf("isReleaseVersion(%q) = false, attendu true", v)
		}
	}
	for _, v := range []string{"", "dev", "(devel)", "0.8", "v0.8.x", "main"} {
		if isReleaseVersion(v) {
			t.Errorf("isReleaseVersion(%q) = true, attendu false", v)
		}
	}

	cases := []struct {
		a, b string
		want int
	}{
		{"0.9.0", "0.8.0", 1},
		{"0.8.0", "0.9.0", -1},
		{"0.8.0", "0.8.0", 0},
		{"v0.10.0", "v0.9.9", 1}, // comparaison numérique, pas lexicographique
		{"1.0.0", "0.99.99", 1},
		{"1.0.0-rc1", "1.0.0", 0}, // une rc de la même version n'est pas une mise à jour
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, attendu %d", c.a, c.b, got, c.want)
		}
	}
}

func TestParseChecksumsFixture(t *testing.T) {
	linux := strings.Repeat("a", 64)
	darwin := strings.Repeat("b", 64)
	// Deux lignes valides, plus trois formes à ignorer : une ligne vide, de la
	// prose en deux mots, et une empreinte tronquée.
	body := linux + "  sillage_linux_amd64\n" + darwin + "  sillage_darwin_arm64\n\nligne cassée\nabc123  sillage_linux_arm64\n"
	sums := parseChecksums(body)
	if sums["sillage_linux_amd64"] != linux || sums["sillage_darwin_arm64"] != darwin {
		t.Fatalf("checksums mal lus : %#v", sums)
	}
	if len(sums) != 2 {
		t.Fatalf("attendu 2 entrées, reçu %d : %#v", len(sums), sums)
	}
}

// TestUpdateStatusDevBuild : une compilation locale ne propose jamais rien et
// ne sort jamais sur le réseau. C'est le cas de tous les autres tests, qui
// construisent un serveur sans jamais appeler SetVersion.
func TestUpdateStatusDevBuild(t *testing.T) {
	srv := newTestServer(t)
	called := false
	prev := fetchLatestReleaseFn
	fetchLatestReleaseFn = func(context.Context) (string, string, error) {
		called = true
		return "", "", nil
	}
	t.Cleanup(func() { fetchLatestReleaseFn = prev })

	srv.startUpdateChecker()
	st := srv.UpdateStatus()
	if st.Method != "dev" || st.Blocker != "dev" {
		t.Fatalf("attendu method/blocker \"dev\", reçu %q/%q", st.Method, st.Blocker)
	}
	if st.Available || st.SelfUpdatable {
		t.Fatalf("une compilation locale ne propose pas de mise à jour : %#v", st)
	}
	if called {
		t.Fatalf("startUpdateChecker a interrogé GitHub pour une compilation locale")
	}
}

func TestUpdateStatusAvailable(t *testing.T) {
	srv := newTestServer(t)
	withVersion(t, "0.8.0")
	withInstall(t, installInfo{Method: "binary", Path: "/home/x/.local/bin/sillage", Writable: true, Command: "curl ... | sh"})
	withLatestRelease(t, "v0.9.0", nil)

	st := srv.checkForUpdate(context.Background())
	if !st.Available || st.Latest != "0.9.0" || st.Current != "0.8.0" {
		t.Fatalf("mise à jour non détectée : %#v", st)
	}
	if !st.SelfUpdatable || st.Blocker != "" {
		t.Fatalf("un binaire dans un dossier écrivable est modifiable en un clic : %#v", st)
	}
	if st.CheckedAt == nil {
		t.Fatalf("checkedAt non renseigné")
	}

	// Même version publiée : plus rien à proposer.
	withLatestRelease(t, "v0.8.0", nil)
	if st := srv.checkForUpdate(context.Background()); st.Available {
		t.Fatalf("la version courante ne doit pas être proposée : %#v", st)
	}
}

func TestUpdateStatusBlockers(t *testing.T) {
	withLatestRelease(t, "v0.9.0", nil)

	cases := []struct {
		name    string
		install installInfo
		want    string
		selfOK  bool
	}{
		{"go install", installInfo{Method: "go", Command: "go install ..."}, "goInstall", false},
		{"dossier non écrivable", installInfo{Method: "binary", Path: "/usr/local/bin/sillage", Command: "curl ... | sh"}, "notWritable", false},
		{"brew absent du PATH", installInfo{Method: "brew", Path: "/opt/homebrew/bin/sillage", Command: "brew upgrade sillage"}, "brewMissing", false},
		{"mode inconnu", installInfo{Method: "unknown", Command: "curl ... | sh"}, "unknownMethod", false},
		{"brew disponible", installInfo{Method: "brew", Path: "/opt/homebrew/bin/sillage", Writable: true, Command: "brew upgrade sillage"}, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := newTestServer(t)
			withVersion(t, "0.8.0")
			withInstall(t, c.install)
			st := srv.checkForUpdate(context.Background())
			if st.Blocker != c.want {
				t.Fatalf("blocker attendu %q, reçu %q", c.want, st.Blocker)
			}
			if st.SelfUpdatable != c.selfOK {
				t.Fatalf("selfUpdatable attendu %v, reçu %v", c.selfOK, st.SelfUpdatable)
			}
			// La commande à la main reste toujours disponible, y compris quand
			// le clic est impossible : c'est le repli universel.
			if st.Command == "" {
				t.Fatalf("command vide : l'utilisateur n'a aucun recours")
			}
		})
	}
}

// TestUpdateApplyRefusesWhileAgentWorks : un agent au travail interdit le
// remplacement du binaire. Le blocage est temporaire, donc le bouton reste
// proposé (selfUpdatable), mais l'application échoue.
func TestUpdateApplyRefusesWhileAgentWorks(t *testing.T) {
	srv := newTestServer(t)
	withVersion(t, "0.8.0")
	withInstall(t, installInfo{Method: "binary", Path: filepath.Join(t.TempDir(), "sillage"), Writable: true})
	withLatestRelease(t, "v0.9.0", nil)
	srv.checkForUpdate(context.Background())

	srv.runner.mu.Lock()
	srv.runner.procs["task-1"] = &procHandle{}
	srv.runner.mu.Unlock()

	st := srv.UpdateStatus()
	if st.Blocker != "tasksRunning" || !st.SelfUpdatable {
		t.Fatalf("attendu un blocage temporaire, reçu %#v", st)
	}
	if _, _, err := srv.applyUpdate(); err == nil || !strings.Contains(err.Error(), "agent") {
		t.Fatalf("applyUpdate devait refuser tant qu'un agent travaille, reçu %v", err)
	}

	srv.runner.mu.Lock()
	delete(srv.runner.procs, "task-1")
	srv.runner.mu.Unlock()

	called := ""
	prev := downloadReleaseFn
	downloadReleaseFn = func(tag, dest string) error { called = tag; return nil }
	t.Cleanup(func() { downloadReleaseFn = prev })

	out, execPath, err := srv.applyUpdate()
	if err != nil {
		t.Fatalf("applyUpdate: %v", err)
	}
	if called != "v0.9.0" {
		t.Fatalf("tag téléchargé %q, attendu \"v0.9.0\"", called)
	}
	if execPath == "" || !strings.Contains(out, "0.9.0") {
		t.Fatalf("réponse inattendue : out=%q execPath=%q", out, execPath)
	}
}

// TestUpdateApplyRequiresConfirm : comme toute action sortante du produit
// (invariant 2 de CONTRIBUTING.md).
func TestUpdateApplyRequiresConfirm(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/update/apply", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	srv.handleUpdateApply(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("attendu 400 sans confirmation, reçu %d", w.Code)
	}

	// Avec confirmation mais sans mise à jour connue : refus explicite, et
	// surtout aucun téléchargement.
	prev := downloadReleaseFn
	downloadReleaseFn = func(tag, dest string) error {
		t.Fatalf("téléchargement déclenché sans mise à jour disponible")
		return nil
	}
	t.Cleanup(func() { downloadReleaseFn = prev })

	req = httptest.NewRequest(http.MethodPost, "/api/update/apply", strings.NewReader(`{"confirm":true}`))
	w = httptest.NewRecorder()
	srv.handleUpdateApply(w, req)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("attendu 502 sans mise à jour disponible, reçu %d", w.Code)
	}
}

// TestDownloadReleaseVerifiesChecksum est le test de sécurité de la mise à
// jour : un binaire dont le sha256 ne correspond pas au checksums.txt de la
// release n'est jamais posé, et ne laisse aucun résidu.
func TestDownloadReleaseVerifiesChecksum(t *testing.T) {
	payload := []byte("#!/bin/sh\necho nouvelle version\n")
	asset := fmt.Sprintf("sillage_%s_%s", runtime.GOOS, runtime.GOARCH)

	var sum string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "checksums.txt"):
			fmt.Fprintf(w, "%s  %s\n", sum, asset)
		case strings.HasSuffix(r.URL.Path, asset):
			w.Write(payload)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	prev := releaseDownloadBase
	releaseDownloadBase = srv.URL + "/"
	t.Cleanup(func() { releaseDownloadBase = prev })

	dir := t.TempDir()
	dest := filepath.Join(dir, "sillage")
	if err := os.WriteFile(dest, []byte("ancienne version"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Empreinte fausse : refus, binaire courant intact, pas de temporaire laissé.
	sum = strings.Repeat("00", 32)
	if err := downloadRelease("v0.9.0", dest); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("attendu un refus sur empreinte fausse, reçu %v", err)
	}
	if content, _ := os.ReadFile(dest); string(content) != "ancienne version" {
		t.Fatalf("le binaire courant a été remplacé malgré l'empreinte fausse")
	}
	if leftovers := tempLeftovers(t, dir); len(leftovers) > 0 {
		t.Fatalf("temporaires laissés derrière : %v", leftovers)
	}

	// Bonne empreinte : le binaire est remplacé et exécutable.
	digest := sha256.Sum256(payload)
	sum = hex.EncodeToString(digest[:])
	if err := downloadRelease("v0.9.0", dest); err != nil {
		t.Fatalf("downloadRelease: %v", err)
	}
	content, err := os.ReadFile(dest)
	if err != nil || string(content) != string(payload) {
		t.Fatalf("binaire non remplacé (%v) : %q", err, content)
	}
	info, err := os.Stat(dest)
	if err != nil || info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("binaire non exécutable : %v %v", err, info.Mode())
	}
	if leftovers := tempLeftovers(t, dir); len(leftovers) > 0 {
		t.Fatalf("temporaires laissés derrière : %v", leftovers)
	}
}

// TestUpdateCheckSettingPersisted : le réglage coupe les vérifications, et son
// absence (state.json d'avant cette fonctionnalité) vaut « activé ».
func TestUpdateCheckSettingPersisted(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if !settingsUpdateCheck(store.GetSettings()) {
		t.Fatalf("par défaut (champ absent), la vérification doit être active")
	}

	off := false
	if _, err := store.UpdateSettings(nil, nil, &off); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}
	reloaded, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore (relecture): %v", err)
	}
	if settingsUpdateCheck(reloaded.GetSettings()) {
		t.Fatalf("le réglage coupé doit survivre à un redémarrage")
	}

	srv := NewServer(reloaded, "", dir, fstest.MapFS{})
	withVersion(t, "0.8.0")
	if srv.updateChecksEnabled() {
		t.Fatalf("réglage coupé : aucune vérification ne doit être autorisée")
	}
	if st := srv.UpdateStatus(); st.CheckEnabled {
		t.Fatalf("checkEnabled devrait être faux : %#v", st)
	}
}

// TestUpdateStatusInState : le frontend hydrate l'état des mises à jour avec le
// reste (GET /api/state), sans requête supplémentaire.
func TestUpdateStatusInState(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/state", nil)
	w := httptest.NewRecorder()
	srv.handleState(w, req)

	var state struct {
		Update UpdateStatus `json:"update"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatalf("réponse illisible: %v", err)
	}
	if state.Update.Current != buildVersion {
		t.Fatalf("version absente de l'état : %#v", state.Update)
	}
}

// --- Helpers ---

func tempLeftovers(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".sillage-update-") {
			out = append(out, e.Name())
		}
	}
	return out
}

func withBrewService(t *testing.T, report *brewServiceReport, err error) {
	t.Helper()
	prev := brewServiceInfoFn
	brewServiceInfoFn = func(string) (*brewServiceReport, error) { return report, err }
	t.Cleanup(func() { brewServiceInfoFn = prev })
}

func boolPtr(b bool) *bool { return &b }
func intPtr(n int) *int    { return &n }

// TestServiceStatusOnlyForBrew : les autres modes d'installation n'ont aucun
// registre à interroger. Ne rien dire est la seule réponse honnête, et surtout
// aucun process brew ne doit être lancé.
func TestServiceStatusOnlyForBrew(t *testing.T) {
	for _, method := range []string{"binary", "go", "unknown"} {
		t.Run(method, func(t *testing.T) {
			srv := newTestServer(t)
			withInstall(t, installInfo{Method: method, Command: "..."})
			called := false
			prev := brewServiceInfoFn
			brewServiceInfoFn = func(string) (*brewServiceReport, error) {
				called = true
				return &brewServiceReport{Registered: boolPtr(false)}, nil
			}
			t.Cleanup(func() { brewServiceInfoFn = prev })

			srv.refreshServiceStatus()
			if called {
				t.Fatalf("brew interrogé pour une installation %q", method)
			}
			if srv.UpdateStatus().Service != nil {
				t.Fatalf("aucun état de service ne doit être exposé pour %q", method)
			}
		})
	}
}

func TestServiceStatusBrew(t *testing.T) {
	t.Run("non enregistré", func(t *testing.T) {
		srv := newTestServer(t)
		withInstall(t, installInfo{Method: "brew", Path: "/opt/homebrew/bin/sillage", Writable: true, Command: "brew upgrade sillage"})
		withBrewService(t, &brewServiceReport{Registered: boolPtr(false), Status: "none"}, nil)

		srv.refreshServiceStatus()
		svc := srv.UpdateStatus().Service
		if svc == nil {
			t.Fatal("état de service attendu")
		}
		if svc.Registered || svc.IsThisOne {
			t.Fatalf("attendu ni enregistré ni service courant : %#v", svc)
		}
		if svc.Command != "brew services start sillage" {
			t.Fatalf("commande inattendue : %q", svc.Command)
		}
	})

	t.Run("enregistré, et c'est nous", func(t *testing.T) {
		srv := newTestServer(t)
		withInstall(t, installInfo{Method: "brew", Writable: true})
		withBrewService(t, &brewServiceReport{Registered: boolPtr(true), Status: "started", PID: intPtr(os.Getpid())}, nil)

		srv.refreshServiceStatus()
		svc := srv.UpdateStatus().Service
		if svc == nil || !svc.Registered || !svc.IsThisOne {
			t.Fatalf("attendu enregistré et instance de service : %#v", svc)
		}
	})

	t.Run("enregistré, mais un autre process", func(t *testing.T) {
		srv := newTestServer(t)
		withInstall(t, installInfo{Method: "brew", Writable: true})
		withBrewService(t, &brewServiceReport{Registered: boolPtr(true), PID: intPtr(os.Getpid() + 1)}, nil)

		srv.refreshServiceStatus()
		svc := srv.UpdateStatus().Service
		if svc == nil || !svc.Registered || svc.IsThisOne {
			t.Fatalf("le PID ne correspond pas : IsThisOne devait rester faux : %#v", svc)
		}
	})

	// brew injoignable, ou trop ancien pour répondre : silence complet. Une
	// suggestion fondée sur une non-réponse serait un mensonge.
	t.Run("brew muet", func(t *testing.T) {
		for _, c := range []struct {
			name   string
			report *brewServiceReport
			err    error
		}{
			{"erreur", nil, fmt.Errorf("brew: command not found")},
			{"champ absent, statut inconnu", &brewServiceReport{Status: "wat"}, nil},
			{"champ absent, statut vide", &brewServiceReport{}, nil},
		} {
			t.Run(c.name, func(t *testing.T) {
				srv := newTestServer(t)
				withInstall(t, installInfo{Method: "brew", Writable: true})
				withBrewService(t, c.report, c.err)
				srv.refreshServiceStatus()
				if svc := srv.UpdateStatus().Service; svc != nil {
					t.Fatalf("silence attendu, reçu %#v", svc)
				}
			})
		}
	})

	// Un brew qui n'expose pas `registered` : le statut suffit à conclure.
	t.Run("repli sur le statut", func(t *testing.T) {
		for status, want := range map[string]bool{"started": true, "scheduled": true, "none": false} {
			srv := newTestServer(t)
			withInstall(t, installInfo{Method: "brew", Writable: true})
			withBrewService(t, &brewServiceReport{Status: status}, nil)
			srv.refreshServiceStatus()
			svc := srv.UpdateStatus().Service
			if svc == nil || svc.Registered != want {
				t.Fatalf("statut %q : attendu registered=%v, reçu %#v", status, want, svc)
			}
		}
	})
}

// TestParseBrewServiceInfoFixture : la sortie réelle de
// `brew services info sillage --json` (Homebrew 4.x, Linux), capturée telle
// quelle. Les mocks des autres tests valident la logique ; celui-ci valide la
// forme, la seule chose qu'un mock ne peut pas garantir.
func TestParseBrewServiceInfoFixture(t *testing.T) {
	out := []byte(`[
  {
    "name": "sillage",
    "service_name": "homebrew.sillage",
    "running": false,
    "loaded": false,
    "schedulable": false,
    "pid": null,
    "exit_code": null,
    "user": null,
    "status": "none",
    "file": "/home/linuxbrew/.linuxbrew/opt/sillage/homebrew.sillage.service",
    "registered": false,
    "loaded_file": null,
    "command": "/home/linuxbrew/.linuxbrew/opt/sillage/bin/sillage",
    "working_dir": null,
    "root_dir": null,
    "log_path": "/home/linuxbrew/.linuxbrew/var/log/sillage.log",
    "error_log_path": "/home/linuxbrew/.linuxbrew/var/log/sillage.log",
    "interval": null,
    "cron": null
  }
]`)

	report, err := parseBrewServiceInfo(out, "sillage")
	if err != nil {
		t.Fatalf("parseBrewServiceInfo: %v", err)
	}
	if report.Registered == nil || *report.Registered {
		t.Fatalf("registered=false attendu, reçu %v", report.Registered)
	}
	if report.Status != "none" || report.PID != nil {
		t.Fatalf("statut/pid inattendus : %q %v", report.Status, report.PID)
	}
	registered, ok := serviceRegistered(report)
	if !ok || registered {
		t.Fatalf("serviceRegistered = (%v, %v), attendu (false, true)", registered, ok)
	}

	// Un autre nom que celui demandé n'est jamais confondu.
	if _, err := parseBrewServiceInfo(out, "autre"); err == nil {
		t.Fatalf("un service absent de la sortie doit produire une erreur")
	}
}
