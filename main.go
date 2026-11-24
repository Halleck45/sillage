// Sillage : serveur backend unique (API + frontend statique embarqué).
package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Halleck45/sillage/internal/server"
)

//go:embed web
var webFS embed.FS

// version est injectée au build de release (-ldflags "-X main.version=vX.Y.Z").
// Une compilation locale garde "dev", ce qui désactive toute la mécanique de
// mise à jour : rien à comparer, aucun appel réseau (voir internal/server/update.go).
var version = "dev"

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "listen address")
	dataDir := flag.String("data", defaultDataDir(), "data directory (state, worktrees, config)")
	showVersion := flag.Bool("version", false, "print version and exit")
	appMode := flag.Bool("app", false, "open Sillage in its own app window, starting the server if it is not running")
	installApp := flag.Bool("install-app", false, "add Sillage to the desktop application menu (Linux) and exit")
	flag.Parse()

	server.SetVersion(version)
	if *showVersion {
		fmt.Println(server.CurrentVersion())
		return
	}

	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("embedded frontend not found: %v", err)
	}

	if *installApp {
		runInstallApp(webContent, *addr, *dataDir)
		return
	}

	url := appURL(*addr)
	if *appMode {
		os.Args = withoutFlag(os.Args, "app")
	}

	if err := os.MkdirAll(*dataDir, 0o700); err != nil {
		log.Fatalf("cannot create data directory %s: %v", *dataDir, err)
	}

	// Le verrou avant l'état : NewStore réécrit state.json dès le chargement.
	// Une instance tient déjà ce dossier : le mode app ouvre sa fenêtre sur
	// elle, quel que soit son port, sans rien charger ; sinon on refuse de
	// démarrer un second serveur sur les mêmes données.
	lock, err := acquireInstanceLock(*dataDir)
	if errors.Is(err, errInstanceRunning) {
		joinRunningInstance(*dataDir, *appMode)
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	defer lock.Close()

	// Le port avant l'état, lui aussi : un bind refusé ne doit pas avoir déjà
	// réécrit state.json. Les connexions attendent dans la file d'écoute le
	// temps du chargement.
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		// Un Sillage antérieur au verrou d'instance occupe le port : il ne
		// s'annonce pas dans sillage.lock, mais il sert déjà l'interface.
		if *appMode && sillageServing(url) {
			if err := openAppWindow(url); err != nil {
				log.Fatal(err)
			}
			return
		}
		log.Fatalf("cannot listen on %s: %v", *addr, err)
	}

	// L'état avant tout log : son refus de démarrer ne doit pas être précédé
	// de lignes sans rapport.
	store, err := server.NewStore(*dataDir)
	if err != nil {
		if errors.Is(err, server.ErrStateTooNew) {
			log.Fatal(err) // le message porte déjà la raison et la sortie de secours
		}
		log.Fatalf("cannot load state: %v", err)
	}

	// Le format ne voit pas tout : deux versions publiées du même format, dont
	// celle qui tourne est la plus ancienne, chargent sans erreur mais perdront
	// des champs à la première sauvegarde.
	if warning := store.DowngradeWarning(); warning != "" {
		log.Print(warning)
	}

	hash, err := server.LoadPasswordHash(*dataDir)
	if err != nil {
		log.Fatalf("cannot initialize authentication: %v", err)
	}
	if hash == "" {
		log.Print("No SILLAGE_PASSWORD set: running without a login password.")
	}

	srv := server.NewServer(store, hash, *dataDir, webContent)
	srv.SetListenAddr(*addr)
	httpSrv := &http.Server{Addr: *addr, Handler: srv.Handler()}

	// Arrêt propre : les recettes manuelles lancées par Sillage (serveurs de
	// dev, scripts) sont tuées avec lui. Sans ce passage, un Ctrl+C laisserait
	// des process en vie sur leurs ports, invisibles depuis l'interface.
	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	go func() {
		<-ctx.Done()
		stopSignals() // un second Ctrl+C reprend son comportement par défaut
		log.Print("Sillage shutting down: stopping previews")
		srv.Shutdown()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shutdownCtx)
	}()

	if err := publishInstance(lock, instanceInfo{PID: os.Getpid(), Addr: ln.Addr().String(), Version: server.CurrentVersion()}); err != nil {
		log.Printf("cannot publish the instance address: %v", err)
	}
	log.Printf("Sillage %s listening on %s (data: %s)", server.CurrentVersion(), *addr, *dataDir)
	if *appMode {
		if err := openAppWindow(url); err != nil {
			log.Print(err)
		}
	}
	if err := httpSrv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server stopped: %v", err)
	}
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sillage-data"
	}
	return filepath.Join(home, ".local", "share", "sillage")
}

// runInstallApp pose le lanceur de bureau, en reportant -addr et -data quand
// ils diffèrent des valeurs par défaut.
func runInstallApp(web fs.FS, addr, dataDir string) {
	var extra []string
	if addr != "127.0.0.1:8787" {
		extra = append(extra, "-addr", addr)
	}
	if dataDir != defaultDataDir() {
		if abs, err := filepath.Abs(dataDir); err == nil {
			dataDir = abs
		}
		extra = append(extra, "-data", dataDir)
	}
	path, err := installDesktopEntry(web, extra)
	if errors.Is(err, errAppInstallUnsupported) {
		fmt.Println("No desktop launcher to install on this system.")
		fmt.Println("Run `sillage -app`, or open Sillage in Chrome or Edge and choose \"Install Sillage\" in the address bar.")
		fmt.Println("With Homebrew, `brew services start sillage` keeps the server running for the installed app.")
		return
	}
	if err != nil {
		log.Fatalf("cannot install the desktop launcher: %v", err)
	}
	fmt.Printf("Sillage added to the application menu (%s).\n", path)
}

// joinRunningInstance : une autre instance tient le dossier de données. Le mode
// app ouvre sa fenêtre sur elle ; un lancement ordinaire s'arrête en disant où
// la trouver.
func joinRunningInstance(dataDir string, appMode bool) {
	wait := time.Second
	if appMode {
		wait = 15 * time.Second // elle peut être en train de charger son état
	}
	info, err := readInstance(dataDir, wait)
	if !appMode {
		where := ""
		if err == nil {
			where = fmt.Sprintf(" (pid %d, %s)", info.PID, appURL(info.Addr))
		}
		log.Fatalf("Sillage is already running for %s%s. Open it there, or pass -data to start a separate workspace.", dataDir, where)
	}
	if err != nil {
		log.Fatal(err)
	}
	url := appURL(info.Addr)
	if !sillageServing(url) {
		log.Fatalf("Sillage (pid %d) holds %s but does not answer at %s", info.PID, dataDir, url)
	}
	if err := openAppWindow(url); err != nil {
		log.Fatal(err)
	}
}
