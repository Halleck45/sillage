package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- Validation des liens épinglés (v0.3.5 section 2) ---

func TestNormalizeLinksValidation(t *testing.T) {
	if _, err := NormalizeLinks([]Link{{URL: "ftp://example.com/file"}}); err == nil {
		t.Fatalf("un schéma non http(s) devrait être refusé")
	}
	if _, err := NormalizeLinks([]Link{{URL: "file:///etc/passwd"}}); err == nil {
		t.Fatalf("file:// devrait être refusé")
	}
	if _, err := NormalizeLinks([]Link{{URL: ""}}); err == nil {
		t.Fatalf("une url vide devrait être refusée")
	}
	if _, err := NormalizeLinks([]Link{{URL: "not a url"}}); err == nil {
		t.Fatalf("une url invalide devrait être refusée")
	}

	// Au-delà de 12 liens : refusé.
	var tooMany []Link
	for i := 0; i < 13; i++ {
		tooMany = append(tooMany, Link{URL: "https://example.com/"})
	}
	if _, err := NormalizeLinks(tooMany); err == nil {
		t.Fatalf("plus de 12 liens devrait être refusé")
	}

	// 12 liens exactement : accepté.
	twelve := tooMany[:12]
	if _, err := NormalizeLinks(twelve); err != nil {
		t.Fatalf("12 liens devraient être acceptés : %v", err)
	}

	out, err := NormalizeLinks([]Link{
		{URL: "https://example.com", Title: "Example"},
		{URL: "http://example.org"},
	})
	if err != nil {
		t.Fatalf("NormalizeLinks: %v", err)
	}
	if len(out) != 2 || out[0].Title != "Example" || out[1].Title != "" {
		t.Fatalf("résultat inattendu : %+v", out)
	}
}

func TestAddProjectWithLinks(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	links := []Link{{URL: "https://example.com", Title: "Example"}}
	p, err := s.AddProject("demo", "", "", []Repo{{Path: "/tmp/demo"}}, links, nil)
	if err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if len(p.Links) != 1 || p.Links[0].URL != "https://example.com" || p.Links[0].Title != "Example" {
		t.Fatalf("links inattendus : %+v", p.Links)
	}

	if _, err := s.AddProject("bad", "", "", []Repo{{Path: "/tmp/bad"}}, []Link{{URL: "ftp://nope"}}, nil); err == nil {
		t.Fatalf("un lien invalide devrait refuser la création du projet")
	}

	newLinks := []Link{{URL: "https://other.example"}}
	updated, err := s.UpdateProject(p.ID, nil, nil, nil, nil, nil, nil, &newLinks, nil, nil)
	if err != nil {
		t.Fatalf("UpdateProject: %v", err)
	}
	if len(updated.Links) != 1 || updated.Links[0].URL != "https://other.example" {
		t.Fatalf("links après mise à jour inattendus : %+v", updated.Links)
	}

	// links=nil (non fourni) ne modifie pas la liste existante.
	name := "demo2"
	updated, err = s.UpdateProject(p.ID, &name, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("UpdateProject (sans links): %v", err)
	}
	if len(updated.Links) != 1 || updated.Links[0].URL != "https://other.example" {
		t.Fatalf("links ne devraient pas changer quand links=nil : %+v", updated.Links)
	}
}

// --- Extraction best-effort du <title> ---

func TestFetchLinkTitleExtractsTitle(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><head><title>  Ma Page \n de Test  </title></head><body>hi</body></html>"))
	}))
	defer srv.Close()

	got := fetchLinkTitle(srv.URL)
	if got != "Ma Page de Test" {
		t.Fatalf("titre attendu 'Ma Page de Test', reçu %q", got)
	}
}

func TestFetchLinkTitleEscapedEntities(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<title>Tom &amp; Jerry</title>"))
	}))
	defer srv.Close()

	if got := fetchLinkTitle(srv.URL); got != "Tom & Jerry" {
		t.Fatalf("titre attendu 'Tom & Jerry', reçu %q", got)
	}
}

// TestFetchLinkTitleFallsBackToHostname couvre le repli sur le nom d'hôte :
// page sans <title>, page en erreur HTTP, et hôte injoignable.
func TestFetchLinkTitleFallsBackToHostname(t *testing.T) {
	noTitle := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body>pas de titre ici</body></html>"))
	}))
	defer noTitle.Close()
	if got, want := fetchLinkTitle(noTitle.URL), hostnameOrURL(noTitle.URL); got != want {
		t.Fatalf("repli attendu %q (hostname), reçu %q", want, got)
	}

	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errServer.Close()
	if got, want := fetchLinkTitle(errServer.URL), hostnameOrURL(errServer.URL); got != want {
		t.Fatalf("repli attendu %q sur erreur HTTP, reçu %q", want, got)
	}

	// Hôte injoignable : repli sur le nom d'hôte, jamais de blocage ni d'erreur.
	unreachable := "http://127.0.0.1:1" // port fermé
	if got, want := fetchLinkTitle(unreachable), "127.0.0.1"; got != want {
		t.Fatalf("repli attendu %q pour un hôte injoignable, reçu %q", want, got)
	}
}

// TestFetchLinkTitleRejectsNonHTTPRedirect vérifie qu'une redirection vers un
// schéma non http(s) est refusée : repli sur le nom d'hôte, jamais suivie.
func TestFetchLinkTitleRejectsNonHTTPRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "file:///etc/passwd")
		w.WriteHeader(http.StatusFound)
	}))
	defer srv.Close()

	got := fetchLinkTitle(srv.URL)
	want := hostnameOrURL(srv.URL)
	if got != want {
		t.Fatalf("repli attendu %q (redirection non http(s) refusée), reçu %q", want, got)
	}
}

// TestFetchLinkTitleCapsBodySize vérifie que la lecture est plafonnée : une
// page dont le <title> arrive après 64 Ko de remplissage n'est pas trouvée
// (repli sur le nom d'hôte), preuve que le corps n'est pas lu intégralement.
func TestFetchLinkTitleCapsBodySize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><head>"))
		w.Write([]byte(strings.Repeat("x", 100*1024)))
		w.Write([]byte("<title>Trop tard</title></head></html>"))
	}))
	defer srv.Close()

	got := fetchLinkTitle(srv.URL)
	want := hostnameOrURL(srv.URL)
	if got != want {
		t.Fatalf("le titre au-delà de 64 Ko ne devrait pas être trouvé, reçu %q (attendu repli %q)", got, want)
	}
}

func TestExtractTitleEmpty(t *testing.T) {
	if got := extractTitle("<html><body>no title here</body></html>"); got != "" {
		t.Fatalf("titre vide attendu, reçu %q", got)
	}
}
