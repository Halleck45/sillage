package server

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectRuntimeProducts(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("composer.json", `{"require":{"php":">=8.1","symfony/console":"^7.0"}}`)
	write("go.mod", "module x\n\ngo 1.22\n")
	got := strings.Join(detectRuntimeProducts(dir), ",")
	if got != "go,php,symfony" {
		t.Fatalf("attendu go,php,symfony, reçu %s", got)
	}
}

func TestEOLSummary(t *testing.T) {
	var cycles []eolCycle
	_ = json.Unmarshal([]byte(`[{"cycle":"8.6","releaseDate":"2026-11-19","latest":"8.6.0","support":"2028-12-31","eol":"2030-12-31"},
		{"cycle":8.0,"releaseDate":"2020-11-26","latest":"8.0.30","support":true,"eol":true}]`), &cycles)
	got := eolSummary("php", cycles)
	if !strings.Contains(got, "php 8.6: released 2026-11-19, latest 8.6.0, active support until 2028-12-31, end of life 2030-12-31") ||
		!strings.Contains(got, "php 8.0:") || !strings.Contains(got, "already end of life") {
		t.Fatalf("résumé inattendu :\n%s", got)
	}
}

func TestExtractUpkeepFindings(t *testing.T) {
	f, ok := extractUpkeepFindings(`Example {"findings":[{"title":"...","verdict":"..."}]} then {"findings":[{"title":"Support PHP 8.6","priority":"urgent","verdict":"PHP 8.6 is out.","details":"x"}]}`)
	if !ok || len(f) != 1 || f[0].Title != "Support PHP 8.6" || f[0].Priority != "normal" {
		t.Fatalf("constat attendu, reçu %+v", f)
	}
	if f, ok := extractUpkeepFindings(`{"findings":[]}`); !ok || len(f) != 0 {
		t.Fatal("une liste vide est une réponse valide : le projet est à jour")
	}
}

// La veille passe une fois, rend un signal d'entretien prêt, ne le duplique
// pas à la passe suivante, et transmet à l'agent les versions publiées.
func TestUpkeepCreatesSignalOnce(t *testing.T) {
	srv, p := newWatchFixture(t, nil)
	if err := os.WriteFile(filepath.Join(p.Repos[0].Path, "composer.json"), []byte(`{"require":{"php":">=8.1"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var asked []string
	srv.watch.mu.Lock()
	srv.watch.eol = func(ctx context.Context, product string) ([]eolCycle, error) {
		asked = append(asked, product)
		return nil, nil
	}
	srv.watch.mu.Unlock()
	if err := srv.store.ResetUpkeepAt(p.ID); err != nil {
		t.Fatal(err)
	}
	srv.watchOnce(context.Background())
	var upkeep []Signal
	for _, s := range srv.store.Snapshot().Signals {
		if s.Kind == "upkeep" {
			upkeep = append(upkeep, s)
		}
	}
	if len(upkeep) != 1 || upkeep[0].Status != "ready" || upkeep[0].Review == "" {
		t.Fatalf("un signal d'entretien prêt attendu, reçu %+v", upkeep)
	}
	if strings.Join(asked, ",") != "php" {
		t.Fatalf("les versions de php doivent être demandées, reçu %v", asked)
	}
	got, _ := srv.store.GetProject(p.ID)
	if got.Watch.UpkeepAt == nil {
		t.Fatal("la date de veille doit être notée")
	}
	// Redemandée : même constat, pas de doublon.
	_ = srv.store.ResetUpkeepAt(p.ID)
	srv.watchOnce(context.Background())
	n := 0
	for _, s := range srv.store.Snapshot().Signals {
		if s.Kind == "upkeep" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("pas de doublon attendu, reçu %d", n)
	}
}
