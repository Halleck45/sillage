package server

// Accès depuis le réseau local (réglage `lan`) : un second listener, à côté de
// celui de -addr qui reste sur localhost. Il est en HTTPS avec un certificat
// auto-signé généré dans le dossier de données, parce qu'un wifi de café en
// HTTP clair laisserait passer le mot de passe et le cookie de session, et
// qu'une session volée ici vaut une exécution de code sur la machine (les
// agents écrivent et lancent du code). Refusé tant qu'aucun mot de passe n'est
// défini, pour la même raison.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LanStatus est exposé dans GET /api/state (champ `lan`) et l'événement SSE
// "lan". Jamais persisté : seul le réglage Settings.LAN l'est.
type LanStatus struct {
	Enabled     bool     `json:"enabled"`
	URLs        []string `json:"urls"`
	Fingerprint string   `json:"fingerprint"` // SHA-256 du certificat, à comparer dans le navigateur
	Error       string   `json:"error"`
	NoPassword  bool     `json:"noPassword"` // le réglage ne peut pas s'activer
}

type lanListener struct {
	mu          sync.Mutex
	port        int
	srv         *http.Server
	fingerprint string
	err         string
}

// lanPortFor : le port de -addr plus un, pour ne pas entrer en conflit
// avec le listener localhost sur le même port.
func lanPortFor(addr string) int {
	_, p, err := net.SplitHostPort(addr)
	if err != nil {
		return 8788
	}
	n, err := strconv.Atoi(p)
	if err != nil || n <= 0 || n >= 65535 {
		return 8788
	}
	return n + 1
}

// SetListenAddr dit au serveur sur quelle adresse tourne le listener principal
// (pour choisir le port LAN), puis démarre l'accès réseau si le réglage le veut.
func (s *Server) SetListenAddr(addr string) {
	s.lan.mu.Lock()
	s.lan.port = lanPortFor(addr)
	s.lan.mu.Unlock()
	if s.store.Snapshot().Settings.LAN {
		if err := s.startLAN(); err != nil {
			log.Printf("LAN access not started: %v", err)
		}
	}
}

func (s *Server) publishLan() {
	s.hub.Publish(Event{Name: "lan", Data: s.LanStatus()})
}

func (s *Server) LanStatus() LanStatus {
	s.lan.mu.Lock()
	defer s.lan.mu.Unlock()
	st := LanStatus{Error: s.lan.err, NoPassword: s.getPasswordHash() == ""}
	if s.lan.srv != nil {
		st.Enabled = true
		st.Fingerprint = s.lan.fingerprint
		for _, ip := range lanIPs() {
			st.URLs = append(st.URLs, "https://"+net.JoinHostPort(ip, strconv.Itoa(s.lan.port)))
		}
		if h, err := os.Hostname(); err == nil && h != "" && !strings.Contains(h, ".") {
			st.URLs = append(st.URLs, "https://"+net.JoinHostPort(h+".local", strconv.Itoa(s.lan.port)))
		}
	}
	return st
}

func (s *Server) startLAN() error {
	if s.getPasswordHash() == "" {
		return errors.New("a password is required to open Sillage to the network")
	}
	s.lan.mu.Lock()
	defer s.lan.mu.Unlock()
	if s.lan.srv != nil {
		return nil
	}
	if s.lan.port == 0 {
		s.lan.port = 8788
	}
	cert, fp, err := loadOrCreateLanCert(s.dataDir)
	if err != nil {
		s.lan.err = err.Error()
		return err
	}
	ln, err := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(s.lan.port)))
	if err != nil {
		s.lan.err = err.Error()
		return err
	}
	srv := &http.Server{
		Handler:           s.Handler(),
		TLSConfig:         &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
		ReadHeaderTimeout: 10 * time.Second,
	}
	s.lan.srv, s.lan.fingerprint, s.lan.err = srv, fp, ""
	go func() {
		if err := srv.ServeTLS(ln, "", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("LAN listener stopped: %v", err)
			s.lan.mu.Lock()
			if s.lan.srv == srv {
				s.lan.srv, s.lan.err = nil, err.Error()
			}
			s.lan.mu.Unlock()
		}
	}()
	log.Printf("LAN access on https://0.0.0.0:%d", s.lan.port)
	return nil
}

func (s *Server) stopLAN() {
	s.lan.mu.Lock()
	srv := s.lan.srv
	s.lan.srv, s.lan.err = nil, ""
	s.lan.mu.Unlock()
	if srv != nil {
		_ = srv.Close()
	}
}

// lanIPs : les adresses IPv4 privées des interfaces actives, triées. Les
// interfaces de conteneurs (docker, bridges virtuels) sont écartées : leurs
// adresses ne mènent nulle part depuis un téléphone.
func lanIPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		name := iface.Name
		if strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "br-") || strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "virbr") {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := ipn.IP.To4()
			if ip == nil || !(ip.IsPrivate() || isCGNAT(ip)) {
				continue
			}
			out = append(out, ip.String())
		}
	}
	sort.Strings(out)
	return out
}

// isCGNAT : 100.64.0.0/10, la plage de Tailscale.
func isCGNAT(ip net.IP) bool { return ip[0] == 100 && ip[1]&0xc0 == 64 }

// loadOrCreateLanCert relit le certificat du dossier de données, ou en génère
// un (ECDSA P-256, dix ans) : l'empreinte reste stable d'un démarrage à
// l'autre, donc le navigateur ne redemande l'exception qu'une fois. Les
// adresses IP n'y figurent pas : elles changent avec le DHCP, et un certificat
// auto-signé déclenche l'avertissement du navigateur de toute façon.
func loadOrCreateLanCert(dataDir string) (tls.Certificate, string, error) {
	dir := filepath.Join(dataDir, "tls")
	certPath, keyPath := filepath.Join(dir, "lan-cert.pem"), filepath.Join(dir, "lan-key.pem")
	if cert, err := tls.LoadX509KeyPair(certPath, keyPath); err == nil {
		return cert, certFingerprint(cert.Certificate[0]), nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return tls.Certificate{}, "", err
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, "", err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 120))
	if err != nil {
		return tls.Certificate{}, "", err
	}
	host, _ := os.Hostname()
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "Sillage " + host},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	if host != "" {
		tmpl.DNSNames = []string{host, host + ".local"}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, "", err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return tls.Certificate{}, "", err
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600); err != nil {
		return tls.Certificate{}, "", err
	}
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		return tls.Certificate{}, "", err
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return tls.Certificate{}, "", fmt.Errorf("reload generated certificate: %w", err)
	}
	return cert, certFingerprint(der), nil
}

// certFingerprint : SHA-256 en hexadécimal par paires, le format qu'affichent
// les navigateurs dans le détail du certificat.
func certFingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	h := strings.ToUpper(hex.EncodeToString(sum[:]))
	var parts []string
	for i := 0; i < len(h); i += 2 {
		parts = append(parts, h[i:i+2])
	}
	return strings.Join(parts, ":")
}
