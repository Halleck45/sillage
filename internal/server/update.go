// Mises à jour de Sillage : version courante, dernière version publiée, et
// application en un clic selon le mode d'installation.
//
// Trois garde-fous, dans l'esprit des invariants du produit (CONTRIBUTING.md) :
// la vérification est une lecture seule (un GET vers l'API GitHub, rien de la
// machine ne sort) et se coupe d'un réglage ; l'application exige
// {"confirm": true} comme toute action sortante et refuse de tourner tant qu'un
// agent ou une recette travaille ; un binaire téléchargé n'est posé qu'après
// vérification de son sha256 contre le checksums.txt de la release.
package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	updateRepo          = "Halleck45/sillage"
	updateCheckInterval = 24 * time.Hour
	updateHTTPTimeout   = 15 * time.Second
	// Délai avant la vérification de démarrage : le serveur écoute déjà, la
	// première page est servie sans attendre le réseau.
	updateFirstCheckDelay = 3 * time.Second
	// Délai entre la réponse HTTP et le remplacement du process : laisse la
	// réponse arriver au navigateur avant que la socket ne se ferme.
	updateRestartGrace = 500 * time.Millisecond

	installScriptURL = "https://raw.githubusercontent.com/" + updateRepo + "/main/install.sh"
)

// version courante, renseignée par main au démarrage (ldflags du build de
// release). "dev" désactive tout le mécanisme : pas de vérification, pas de
// bouton, rien à comparer.
var buildVersion = "dev"

// SetVersion fixe la version du binaire courant. Appelée une fois par main
// avant le démarrage du serveur. La valeur passée gagne ; si elle ne ressemble
// pas à une version publiée (compilation locale sans ldflags), on retombe sur
// les informations de build du module, renseignées par `go install`.
func SetVersion(v string) {
	if isReleaseVersion(v) {
		buildVersion = normalizeVersion(v)
		return
	}
	if info, ok := debug.ReadBuildInfo(); ok && isReleaseVersion(info.Main.Version) {
		buildVersion = normalizeVersion(info.Main.Version)
	}
}

// CurrentVersion retourne la version affichée par l'interface ("dev" pour une
// compilation locale).
func CurrentVersion() string { return buildVersion }

// updateTracker garde ce que la dernière vérification a appris. En mémoire
// uniquement : une version publiée n'est pas un état du produit, la relire au
// démarrage n'apporte rien et la persister ferait mentir un state.json
// rapatrié d'une autre machine.
type updateTracker struct {
	mu         sync.Mutex
	latest     string
	releaseURL string
	checkedAt  *time.Time
	lastErr    string
	applying   bool
	stop       chan struct{}
}

// installInfo décrit comment ce binaire a été installé, donc comment il peut
// se remplacer. Calculé une seule fois : un binaire ne bouge pas pendant qu'il
// tourne.
type installInfo struct {
	Method   string // "brew" | "binary" | "go" | "unknown"
	Path     string // chemin du binaire courant, cible du remplacement
	Writable bool   // le dossier du binaire accepte une écriture
	Command  string // la commande équivalente, à jouer à la main
}

var (
	installOnce   sync.Once
	installCached installInfo
)

// Indirections pour les tests : aucun test ne doit sortir sur le réseau, ni
// remplacer un binaire, ni dépendre de la façon dont le binaire de test a été
// installé.
var (
	fetchLatestReleaseFn = fetchLatestRelease
	downloadReleaseFn    = downloadRelease
	brewUpgradeFn        = brewUpgrade
	detectInstallFn      = detectInstall
	releaseDownloadBase  = "https://github.com/" + updateRepo + "/releases/download/"
)

// --- Version ---

// isReleaseVersion reconnaît une version publiée (vX.Y.Z, suffixe de
// pré-release toléré). Tout le reste ("dev", "(devel)", "") est une
// compilation locale.
func isReleaseVersion(v string) bool {
	v = normalizeVersion(v)
	if v == "" {
		return false
	}
	core, _, _ := strings.Cut(v, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return false
	}
	for _, p := range parts {
		if p == "" {
			return false
		}
		if _, err := strconv.Atoi(p); err != nil {
			return false
		}
	}
	return true
}

func normalizeVersion(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

// compareVersions retourne -1 si a < b, 0 si égales, 1 si a > b. Les suffixes
// de pré-release sont ignorés : une v1.0.0-rc1 et une v1.0.0 se valent, ce qui
// évite de proposer une « mise à jour » vers la release qu'on a déjà en rc.
func compareVersions(a, b string) int {
	pa := versionParts(a)
	pb := versionParts(b)
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			if pa[i] < pb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

func versionParts(v string) [3]int {
	out := [3]int{}
	core, _, _ := strings.Cut(normalizeVersion(v), "-")
	for i, p := range strings.Split(core, ".") {
		if i > 2 {
			break
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return out
		}
		out[i] = n
	}
	return out
}

// --- Mode d'installation ---

func detectInstall() installInfo {
	installOnce.Do(func() { installCached = detectInstallUncached() })
	return installCached
}

func detectInstallUncached() installInfo {
	info := installInfo{Method: "unknown", Command: "curl -fsSL " + installScriptURL + " | sh"}

	exe, err := os.Executable()
	if err != nil {
		return info
	}
	info.Path = exe
	// Le chemin réel (symlinks résolus) est le seul qui dise la vérité sur
	// Homebrew : /opt/homebrew/bin/sillage n'est qu'un lien vers le Cellar.
	real := exe
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		real = resolved
	}

	switch {
	case strings.Contains(real, string(filepath.Separator)+"Cellar"+string(filepath.Separator)):
		info.Method = "brew"
		info.Command = "brew update && brew upgrade sillage"
		// Sans le binaire brew, un clic ne peut rien faire : on garde la
		// commande affichée et le bouton reste éteint (voir selfUpdatable).
		info.Writable = lookPathFn("brew") != ""
		return info
	case underGoBin(real):
		info.Method = "go"
		info.Command = "go install github.com/" + updateRepo + "@latest"
		return info
	}

	info.Method = "binary"
	info.Writable = dirWritable(filepath.Dir(exe))
	return info
}

// lookPathFn est indirect pour les tests (une machine de CI n'a pas brew).
var lookPathFn = func(name string) string {
	p, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return p
}

func underGoBin(path string) bool {
	dir := filepath.Dir(path)
	if gp := os.Getenv("GOPATH"); gp != "" {
		for _, entry := range filepath.SplitList(gp) {
			if dir == filepath.Join(entry, "bin") {
				return true
			}
		}
	}
	if gb := os.Getenv("GOBIN"); gb != "" && dir == gb {
		return true
	}
	if home, err := os.UserHomeDir(); err == nil && dir == filepath.Join(home, "go", "bin") {
		return true
	}
	return false
}

// dirWritable teste réellement l'écriture : les permissions seules ne disent
// rien d'un montage en lecture seule.
func dirWritable(dir string) bool {
	f, err := os.CreateTemp(dir, ".sillage-write-test-*")
	if err != nil {
		return false
	}
	name := f.Name()
	_ = f.Close()
	_ = os.Remove(name)
	return true
}

// --- État exposé ---

// updateChecksEnabled dit si la vérification automatique est autorisée : une
// compilation locale n'a rien à comparer, et le réglage coupe tout.
func (s *Server) updateChecksEnabled() bool {
	if !isReleaseVersion(buildVersion) {
		return false
	}
	return settingsUpdateCheck(s.store.GetSettings())
}

// settingsUpdateCheck : absent = activé. Le champ est un pointeur pour qu'un
// state.json écrit avant cette fonctionnalité garde le comportement par
// défaut sans migration.
func settingsUpdateCheck(st Settings) bool {
	return st.UpdateCheck == nil || *st.UpdateCheck
}

// UpdateStatus construit l'état des mises à jour pour GET /api/state,
// GET /api/update et l'événement SSE "update".
func (s *Server) UpdateStatus() UpdateStatus {
	install := detectInstallFn()

	s.updates.mu.Lock()
	latest := s.updates.latest
	releaseURL := s.updates.releaseURL
	checkedAt := s.updates.checkedAt
	lastErr := s.updates.lastErr
	applying := s.updates.applying
	s.updates.mu.Unlock()

	st := UpdateStatus{
		Current:      buildVersion,
		Latest:       latest,
		Method:       install.Method,
		Path:         install.Path,
		Command:      install.Command,
		ReleaseURL:   releaseURL,
		CheckEnabled: settingsUpdateCheck(s.store.GetSettings()),
		CheckedAt:    checkedAt,
		Error:        lastErr,
		Applying:     applying,
	}
	if !isReleaseVersion(buildVersion) {
		st.Method = "dev"
		st.Blocker = "dev"
		return st
	}
	st.Available = latest != "" && compareVersions(latest, buildVersion) > 0
	if !st.Available {
		return st
	}

	switch {
	case install.Method == "go":
		st.Blocker = "goInstall"
	case install.Method == "unknown":
		st.Blocker = "unknownMethod"
	case install.Method == "brew" && !install.Writable:
		st.Blocker = "brewMissing"
	case install.Method == "binary" && !install.Writable:
		st.Blocker = "notWritable"
	case s.runner.RunningCount() > 0:
		st.Blocker = "tasksRunning"
	case s.previews.RunningCount() > 0:
		st.Blocker = "previewsRunning"
	}
	// Les deux derniers blocages sont temporaires : le bouton existe, il
	// attend juste que la machine soit au repos. Les autres sont structurels,
	// seule la commande reste.
	st.SelfUpdatable = st.Blocker == "" || st.Blocker == "tasksRunning" || st.Blocker == "previewsRunning"
	return st
}

// --- Vérification ---

// checkForUpdate interroge GitHub et met le tracker à jour. Publie toujours
// l'événement SSE : une vérification qui échoue est une information (l'UI
// affiche la version courante sans promesse).
func (s *Server) checkForUpdate(ctx context.Context) UpdateStatus {
	tag, releaseURL, err := fetchLatestReleaseFn(ctx)
	now := time.Now().UTC()

	s.updates.mu.Lock()
	s.updates.checkedAt = &now
	if err != nil {
		s.updates.lastErr = err.Error()
	} else {
		s.updates.lastErr = ""
		s.updates.latest = normalizeVersion(tag)
		s.updates.releaseURL = releaseURL
	}
	s.updates.mu.Unlock()

	st := s.UpdateStatus()
	s.runner.publishUpdate(st)
	return st
}

func fetchLatestRelease(ctx context.Context) (tag, htmlURL string, err error) {
	url := "https://api.github.com/repos/" + updateRepo + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "sillage/"+buildVersion)

	resp, err := (&http.Client{Timeout: updateHTTPTimeout}).Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("github returned %s", resp.Status)
	}

	var payload struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return "", "", err
	}
	if !isReleaseVersion(payload.TagName) {
		return "", "", fmt.Errorf("unexpected release tag %q", payload.TagName)
	}
	return payload.TagName, payload.HTMLURL, nil
}

// startUpdateChecker démarre la vérification périodique si elle ne tourne pas
// déjà. Appelée au boot et à chaque activation du réglage. Même patron que
// startAutoSync (workspace.go).
func (s *Server) startUpdateChecker() {
	if !s.updateChecksEnabled() {
		return
	}
	s.updates.mu.Lock()
	if s.updates.stop != nil {
		s.updates.mu.Unlock()
		return
	}
	stop := make(chan struct{})
	s.updates.stop = stop
	s.updates.mu.Unlock()

	go func() {
		timer := time.NewTimer(updateFirstCheckDelay)
		defer timer.Stop()
		ticker := time.NewTicker(updateCheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-timer.C:
			case <-ticker.C:
			}
			ctx, cancel := context.WithTimeout(context.Background(), updateHTTPTimeout)
			s.checkForUpdate(ctx)
			cancel()
		}
	}()
}

func (s *Server) stopUpdateChecker() {
	s.updates.mu.Lock()
	defer s.updates.mu.Unlock()
	if s.updates.stop == nil {
		return
	}
	close(s.updates.stop)
	s.updates.stop = nil
}

// --- Application ---

// applyUpdate installe la dernière version et retourne la sortie de
// l'opération, plus le chemin du binaire à ré-exécuter. Ne redémarre rien :
// c'est le handler qui répond d'abord, puis remplace le process.
func (s *Server) applyUpdate() (output string, execPath string, err error) {
	st := s.UpdateStatus()
	if !st.Available {
		return "", "", fmt.Errorf("no update available")
	}
	if !st.SelfUpdatable {
		return "", "", fmt.Errorf("this installation must be updated by hand: %s", st.Command)
	}
	switch st.Blocker {
	case "tasksRunning":
		return "", "", fmt.Errorf("an agent is still working: interrupt it before updating")
	case "previewsRunning":
		return "", "", fmt.Errorf("a preview is still running: stop it before updating")
	}

	s.updates.mu.Lock()
	if s.updates.applying {
		s.updates.mu.Unlock()
		return "", "", fmt.Errorf("an update is already in progress")
	}
	s.updates.applying = true
	s.updates.mu.Unlock()

	defer func() {
		s.updates.mu.Lock()
		s.updates.applying = false
		s.updates.mu.Unlock()
		s.runner.publishUpdate(s.UpdateStatus())
	}()
	s.runner.publishUpdate(s.UpdateStatus())

	install := detectInstallFn()
	switch install.Method {
	case "brew":
		out, err := brewUpgradeFn()
		if err != nil {
			return out, "", err
		}
		return out, brewExecPath(install.Path), nil
	case "binary":
		if err := downloadReleaseFn("v"+st.Latest, install.Path); err != nil {
			return "", "", err
		}
		return "installed " + st.Latest + " in " + install.Path, install.Path, nil
	}
	return "", "", fmt.Errorf("this installation must be updated by hand: %s", install.Command)
}

// brewUpgrade joue la mise à jour Homebrew. `brew update` d'abord : sans lui,
// le tap ne connaît pas encore la nouvelle formule et `brew upgrade` ne voit
// rien à faire.
func brewUpgrade() (string, error) {
	var out strings.Builder
	for _, args := range [][]string{{"update"}, {"upgrade", "sillage"}} {
		cmd := exec.Command("brew", args...)
		combined, err := cmd.CombinedOutput()
		out.WriteString("$ brew " + strings.Join(args, " ") + "\n")
		out.Write(combined)
		if err != nil {
			return out.String(), fmt.Errorf("brew %s failed: %v", strings.Join(args, " "), err)
		}
	}
	return out.String(), nil
}

// brewExecPath : après un upgrade, le Cellar de l'ancienne version peut avoir
// disparu. On repart du binaire visible dans le PATH (un lien vers la nouvelle
// version), et seulement à défaut du chemin de départ.
func brewExecPath(fallback string) string {
	if p := lookPathFn("sillage"); p != "" {
		return p
	}
	if prefix := os.Getenv("HOMEBREW_PREFIX"); prefix != "" {
		candidate := filepath.Join(prefix, "bin", "sillage")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if _, err := os.Stat(fallback); err == nil {
		return fallback
	}
	return ""
}

// downloadRelease télécharge le binaire de la release `tag` et le pose à la
// place de `dest`. Le sha256 est vérifié contre le checksums.txt de la même
// release : un binaire qui ne correspond pas n'est jamais posé. L'écriture se
// fait dans un temporaire du même dossier puis un os.Rename (atomique, et sûr
// sur un binaire en cours d'exécution : l'inode courant survit).
func downloadRelease(tag, dest string) error {
	asset := fmt.Sprintf("sillage_%s_%s", runtime.GOOS, runtime.GOARCH)
	base := releaseDownloadBase + tag + "/"

	sums, err := fetchChecksums(base + "checksums.txt")
	if err != nil {
		return err
	}
	want, ok := sums[asset]
	if !ok {
		return fmt.Errorf("no checksum published for %s", asset)
	}

	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, ".sillage-update-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}

	resp, err := httpGet(base + asset)
	if err != nil {
		cleanup()
		return err
	}
	defer resp.Body.Close()

	sum := sha256.New()
	if _, err := io.Copy(io.MultiWriter(tmp, sum), resp.Body); err != nil {
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if got := hex.EncodeToString(sum.Sum(nil)); got != want {
		_ = os.Remove(tmpName)
		return fmt.Errorf("checksum mismatch for %s: refusing to install", asset)
	}
	if err := os.Chmod(tmpName, 0o755); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, dest); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

func fetchChecksums(url string) (map[string]string, error) {
	resp, err := httpGet(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return nil, err
	}
	return parseChecksums(string(body)), nil
}

// parseChecksums lit un fichier au format sha256sum ("<hex>  <nom>"). Seules
// les lignes dont le premier champ est bien une empreinte sha256 sont retenues :
// une ligne de prose ne doit jamais devenir une empreinte de référence.
func parseChecksums(body string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || !isSHA256(fields[0]) {
			continue
		}
		out[strings.TrimPrefix(fields[1], "*")] = fields[0]
	}
	return out
}

func isSHA256(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func httpGet(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "sillage/"+buildVersion)
	resp, err := (&http.Client{Timeout: 5 * time.Minute}).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("GET %s returned %s", url, resp.Status)
	}
	return resp, nil
}

// restartInPlace remplace le process courant par le binaire fraîchement
// installé : même PID, même terminal, mêmes arguments. Les navigateurs ouverts
// se reconnectent tout seuls (l'EventSource retente, et le frontend re-fetch
// l'état complet à la reconnexion).
func restartInPlace(execPath string) error {
	if execPath == "" {
		return fmt.Errorf("cannot locate the new binary")
	}
	argv := append([]string{execPath}, os.Args[1:]...)
	return syscall.Exec(execPath, argv, os.Environ())
}
