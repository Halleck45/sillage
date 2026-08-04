// Sillage : serveur backend unique (API + frontend statique embarqué).
package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"io/fs"
	"log"
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

func main() {
	addr := flag.String("addr", "127.0.0.1:8787", "listen address")
	dataDir := flag.String("data", defaultDataDir(), "data directory (state, worktrees, config)")
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0o700); err != nil {
		log.Fatalf("cannot create data directory %s: %v", *dataDir, err)
	}

	hash, err := server.LoadPasswordHash(*dataDir)
	if err != nil {
		log.Fatalf("cannot initialize authentication: %v", err)
	}
	if hash == "" {
		log.Print("No SILLAGE_PASSWORD set: running without a login password.")
	}

	store, err := server.NewStore(*dataDir)
	if err != nil {
		log.Fatalf("cannot load state: %v", err)
	}

	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("embedded frontend not found: %v", err)
	}

	srv := server.NewServer(store, hash, *dataDir, webContent)
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

	log.Printf("Sillage listening on %s (data: %s)", *addr, *dataDir)
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
