package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "atelier_session"
const sessionTTL = 30 * 24 * time.Hour

// Config est le contenu persisté de config.json.
type Config struct {
	PasswordHash string `json:"passwordHash"`
}

func configPath(dataDir string) string {
	return filepath.Join(dataDir, "config.json")
}

// GenerateRandomPassword génère un mot de passe alphanumérique de longueur n
// à l'aide de crypto/rand.
func GenerateRandomPassword(n int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	out := make([]byte, n)
	max := big.NewInt(int64(len(charset)))
	for i := range out {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = charset[idx.Int64()]
	}
	return string(out), nil
}

// LoadOrInitPasswordHash détermine le hash bcrypt du mot de passe à utiliser.
//
// Priorité :
//  1. ATELIER_PASSWORD (si présent au démarrage), pour les tests ; non persisté.
//  2. config.json existant dans dataDir.
//  3. Génération d'un mot de passe aléatoire, affiché une seule fois
//     (valeur non vide de generated), et persistance du hash dans config.json.
func LoadOrInitPasswordHash(dataDir string) (hash string, generated string, err error) {
	if envPass := os.Getenv("ATELIER_PASSWORD"); envPass != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(envPass), bcrypt.DefaultCost)
		if err != nil {
			return "", "", err
		}
		return string(h), "", nil
	}

	path := configPath(dataDir)
	data, err := os.ReadFile(path)
	if err == nil {
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return "", "", fmt.Errorf("lecture de config.json impossible : %w", err)
		}
		if cfg.PasswordHash != "" {
			return cfg.PasswordHash, "", nil
		}
	} else if !os.IsNotExist(err) {
		return "", "", err
	}

	pass, err := GenerateRandomPassword(16)
	if err != nil {
		return "", "", err
	}
	h, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", "", err
	}
	cfg := Config{PasswordHash: string(h)}
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", "", err
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return "", "", err
	}
	return string(h), pass, nil
}

// SessionManager gère les sessions en mémoire (token -> expiration).
type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]time.Time
}

// NewSessionManager crée un gestionnaire de sessions vide.
func NewSessionManager() *SessionManager {
	return &SessionManager{sessions: map[string]time.Time{}}
}

// Create génère un nouveau token de session (32 octets aléatoires, hex) valable 30 jours.
func (m *SessionManager) Create() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := hex.EncodeToString(buf)
	m.mu.Lock()
	m.sessions[token] = time.Now().Add(sessionTTL)
	m.mu.Unlock()
	return token, nil
}

// Validate indique si un token de session est connu et non expiré.
func (m *SessionManager) Validate(token string) bool {
	if token == "" {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	expiry, ok := m.sessions[token]
	if !ok {
		return false
	}
	if time.Now().After(expiry) {
		delete(m.sessions, token)
		return false
	}
	return true
}

// Delete invalide un token de session (déconnexion).
func (m *SessionManager) Delete(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
}

// LoginLimiter limite les tentatives de connexion échouées par adresse IP
// (max 5 échecs par fenêtre glissante d'une minute).
type LoginLimiter struct {
	mu       sync.Mutex
	failures map[string][]time.Time
}

// NewLoginLimiter crée un limiteur de tentatives de connexion.
func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{failures: map[string][]time.Time{}}
}

// Blocked indique si l'IP a atteint la limite d'échecs sur la dernière minute.
func (l *LoginLimiter) Blocked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := time.Now().Add(-time.Minute)
	kept := l.failures[ip][:0]
	for _, t := range l.failures[ip] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	l.failures[ip] = kept
	return len(kept) >= 5
}

// RecordFailure enregistre un échec de connexion pour l'IP donnée.
func (l *LoginLimiter) RecordFailure(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failures[ip] = append(l.failures[ip], time.Now())
}

// isSecureRequest indique si la connexion doit être considérée comme HTTPS
// (TLS direct ou en-tête X-Forwarded-Proto positionné par un proxy).
func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
		Expires:  time.Now().Add(sessionTTL),
	})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isSecureRequest(r),
		MaxAge:   -1,
	})
}
