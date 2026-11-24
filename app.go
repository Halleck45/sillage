// Fenêtre d'application. Sillage reste un serveur : l'« app » est une fenêtre
// Chromium en mode application ouverte sur lui, pas un runtime embarqué. Les
// agents tournent dans le serveur, donc fermer la fenêtre ne les arrête pas.
//
// Pourquoi pas un webview natif (Wails, Tauri, webview_go) : tous demandent cgo,
// ce qui casse la compilation croisée CGO_ENABLED=0 de la release, et sous
// Linux ils reposent sur WebKitGTK, dont le rendu de backdrop-filter (tout le
// design en verre) est nettement plus coûteux que celui de Chromium. Une
// fenêtre --app partage le process d'un Chromium déjà lancé : quelques dizaines
// de Mo de plus, aucun octet de plus dans le binaire, aucune dépendance.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// appURL renvoie l'URL locale de l'interface pour une adresse d'écoute. Une
// écoute sur toutes les interfaces s'ouvre par la boucle locale.
func appURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://" + addr + "/"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port) + "/"
}

// sillageServing dit si une instance de Sillage répond déjà sur cette URL. La
// page d'accueil est statique et sans session, et son titre suffit à ne pas
// confondre Sillage avec un autre programme posé sur le même port. Un refus de
// connexion en boucle locale est immédiat : la sonde ne retarde pas le
// démarrage quand le serveur ne tourne pas.
func sillageServing(url string) bool {
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	head, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return resp.StatusCode == http.StatusOK && bytes.Contains(head, []byte("<title>Sillage</title>"))
}

// chromiumCandidates : navigateurs de la famille Chromium qui savent ouvrir une
// fenêtre d'application (--app). SILLAGE_BROWSER passe devant tous.
func chromiumCandidates() []string {
	var list []string
	if custom := os.Getenv("SILLAGE_BROWSER"); custom != "" {
		list = append(list, custom)
	}
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		for _, app := range []string{"Google Chrome", "Chromium", "Microsoft Edge", "Brave Browser", "Vivaldi"} {
			bin := app + ".app/Contents/MacOS/" + app
			list = append(list, "/Applications/"+bin, filepath.Join(home, "Applications", bin))
		}
	default:
		list = append(list, "google-chrome-stable", "google-chrome", "chromium", "chromium-browser",
			"microsoft-edge-stable", "microsoft-edge", "brave-browser", "vivaldi-stable")
	}
	return list
}

// openAppWindow ouvre l'interface dans une fenêtre d'application, sinon dans un
// onglet du navigateur par défaut. Le profil du navigateur est celui de
// l'utilisateur, volontairement : la fenêtre rejoint le process déjà lancé
// (pas un second Chromium en mémoire) et reprend la session de connexion.
func openAppWindow(url string) error {
	for _, candidate := range chromiumCandidates() {
		bin, err := exec.LookPath(candidate)
		if err != nil {
			continue
		}
		if err := startDetached(bin, "--app="+url); err == nil {
			return nil
		}
	}
	opener := "xdg-open"
	if runtime.GOOS == "darwin" {
		opener = "open"
	}
	if bin, err := exec.LookPath(opener); err == nil {
		return startDetached(bin, url)
	}
	return fmt.Errorf("no browser found: open %s yourself", url)
}

// startDetached lance un process dans son propre groupe : un Ctrl+C sur le
// serveur ne doit pas fermer le navigateur que la fenêtre a pu démarrer.
func startDetached(bin string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }() // pas de zombie quand le navigateur rend la main
	return nil
}

// withoutFlag retire un drapeau booléen des arguments. Le redémarrage après
// mise à jour relance le binaire avec os.Args : sans ce retrait, chaque mise à
// jour rouvrirait une fenêtre, et le statut du service signalerait à tort des
// arguments personnalisés.
func withoutFlag(args []string, name string) []string {
	out := args[:1:1]
	for _, a := range args[1:] {
		trimmed := strings.TrimLeft(a, "-")
		if (a != trimmed) && (trimmed == name || strings.HasPrefix(trimmed, name+"=")) {
			continue
		}
		out = append(out, a)
	}
	return out
}

var errAppInstallUnsupported = errors.New("unsupported platform")

// installDesktopEntry pose un lanceur Sillage dans le menu d'applications du
// bureau (Linux, spécification freedesktop) : un .desktop qui lance
// `sillage -app`, et l'icône SVG de l'interface. Sous macOS, la voie est
// l'installation de la page depuis Chrome, avec le service brew pour le serveur.
func installDesktopEntry(web fs.FS, extraArgs []string) (string, error) {
	if runtime.GOOS != "linux" {
		return "", errAppInstallUnsupported
	}
	exe, err := stableExecutable()
	if err != nil {
		return "", err
	}
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	icon, err := fs.ReadFile(web, "favicon.svg")
	if err != nil {
		return "", err
	}
	iconDir := filepath.Join(dataHome, "icons", "hicolor", "scalable", "apps")
	if err := os.MkdirAll(iconDir, 0o755); err != nil {
		return "", err
	}
	// Chemin absolu dans Icon= plutôt qu'un nom de thème : un icon-theme.cache
	// déjà présent dans ce dossier (posé par un autre installeur) fait foi pour
	// GTK, qui ne voit alors pas une icône ajoutée après lui.
	iconPath := filepath.Join(iconDir, "sillage.svg")
	if err := os.WriteFile(iconPath, icon, 0o644); err != nil {
		return "", err
	}

	execLine := []string{desktopQuote(exe), "-app"}
	for _, a := range extraArgs {
		execLine = append(execLine, desktopQuote(a))
	}
	entry := strings.Join([]string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=Sillage",
		"Comment=Pilot AI coding agents across your projects",
		"Exec=" + strings.Join(execLine, " "),
		"Icon=" + iconPath,
		"Terminal=false",
		"Categories=Development;",
		"StartupNotify=false",
		"",
	}, "\n")
	appDir := filepath.Join(dataHome, "applications")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(appDir, "sillage.desktop")
	return path, os.WriteFile(path, []byte(entry), 0o644)
}

// stableExecutable renvoie un chemin du binaire qui survit à une mise à jour.
// Sous Homebrew, os.Executable donne le Cellar versionné, effacé au prochain
// upgrade : on préfère le lien du PATH quand il désigne le même fichier.
func stableExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return exe, nil
	}
	if onPath, err := exec.LookPath("sillage"); err == nil {
		if abs, err := filepath.Abs(onPath); err == nil {
			if target, err := filepath.EvalSymlinks(abs); err == nil && target == real {
				return abs, nil
			}
		}
	}
	return real, nil
}

// desktopQuote cite un argument de la clé Exec selon la spécification
// freedesktop : guillemets dès qu'un caractère réservé apparaît, et échappement
// de " ` $ \ à l'intérieur, puis échappement de chaîne de la valeur. Le % se double dans tous les cas (codes de champ).
func desktopQuote(arg string) string {
	arg = strings.ReplaceAll(arg, "%", "%%")
	if !strings.ContainsAny(arg, " \t\n\"'\\><~|&;$*?#()`") {
		return arg
	}
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", `$`, `\$`)
	// Second niveau : la valeur d'une clé est elle-même une chaîne échappée,
	// chaque barre oblique inverse s'y écrit doublée.
	return strings.ReplaceAll(`"`+r.Replace(arg)+`"`, `\`, `\\`)
}
