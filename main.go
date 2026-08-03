// Sillage : serveur backend unique (API + frontend statique embarqué).
package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

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

	hash, generated, err := server.LoadOrInitPasswordHash(*dataDir)
	if err != nil {
		log.Fatalf("cannot initialize authentication: %v", err)
	}
	if generated != "" {
		fmt.Printf("Initial password: %s\n", generated)
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

	log.Printf("Sillage listening on %s (data: %s)", *addr, *dataDir)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
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
