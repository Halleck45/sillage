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
	addr := flag.String("addr", "127.0.0.1:8787", "adresse d'écoute du serveur")
	dataDir := flag.String("data", defaultDataDir(), "répertoire de données (état, worktrees, config)")
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0o700); err != nil {
		log.Fatalf("impossible de créer le répertoire de données %s : %v", *dataDir, err)
	}

	hash, generated, err := server.LoadOrInitPasswordHash(*dataDir)
	if err != nil {
		log.Fatalf("initialisation de l'authentification impossible : %v", err)
	}
	if generated != "" {
		fmt.Printf("Mot de passe initial : %s\n", generated)
	}

	store, err := server.NewStore(*dataDir)
	if err != nil {
		log.Fatalf("chargement de l'état impossible : %v", err)
	}

	webContent, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("frontend embarqué introuvable : %v", err)
	}

	srv := server.NewServer(store, hash, *dataDir, webContent)

	log.Printf("Sillage écoute sur %s (données : %s)", *addr, *dataDir)
	if err := http.ListenAndServe(*addr, srv.Handler()); err != nil {
		log.Fatalf("serveur arrêté : %v", err)
	}
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sillage-data"
	}
	return filepath.Join(home, ".local", "share", "sillage")
}
