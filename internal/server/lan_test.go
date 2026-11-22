package server

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"golang.org/x/crypto/bcrypt"
)

func freePort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}

func TestLanRefusedWithoutPassword(t *testing.T) {
	srv := newTestServer(t)
	req := httptest.NewRequest(http.MethodPatch, "/api/settings", strings.NewReader(`{"lan":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("sans mot de passe, attendu 400, reçu %d", w.Code)
	}
	if srv.store.Snapshot().Settings.LAN {
		t.Fatal("le réglage ne doit pas rester allumé après un refus")
	}
}

func TestLanServesHTTPSAndRequiresLogin(t *testing.T) {
	store, err := NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.MinCost)
	dataDir := t.TempDir()
	srv := NewServer(store, string(hash), dataDir, fstest.MapFS{})
	defer srv.Shutdown()
	port := freePort(t)
	srv.lan.port = port
	if err := srv.startLAN(); err != nil {
		t.Fatalf("startLAN: %v", err)
	}
	st := srv.LanStatus()
	if !st.Enabled || st.Fingerprint == "" {
		t.Fatalf("statut inattendu : %+v", st)
	}

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Get("https://127.0.0.1:" + strconv.Itoa(port) + "/api/state")
	if err != nil {
		t.Fatalf("GET en HTTPS : %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("l'API doit exiger une session, reçu %d", resp.StatusCode)
	}

	// Le certificat est réutilisé : même empreinte après un redémarrage.
	srv.stopLAN()
	if err := srv.startLAN(); err != nil {
		t.Fatal(err)
	}
	if got := srv.LanStatus().Fingerprint; got != st.Fingerprint {
		t.Fatalf("empreinte changée au redémarrage : %s puis %s", st.Fingerprint, got)
	}
}

func TestLanPortFor(t *testing.T) {
	if p := lanPortFor("127.0.0.1:8787"); p != 8788 {
		t.Fatalf("attendu 8788, reçu %d", p)
	}
	if p := lanPortFor("garbage"); p != 8788 {
		t.Fatalf("repli attendu 8788, reçu %d", p)
	}
}
