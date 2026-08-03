package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "sillage_session"
const sessionTTL = 30 * 24 * time.Hour

// Config est le contenu persisté de config.json.
type Config struct {
	PasswordHash string `json:"passwordHash"`
}

func configPath(dataDir string) string {
	return filepath.Join(dataDir, "config.json")
}

// LoadPasswordHash détermine le hash bcrypt du mot de passe à utiliser, ou
// une chaîne vide si aucun mot de passe n'est configuré. Dans ce dernier cas,
// le serveur tourne sans authentification (usage local, réseau non exposé).
//
// Priorité :
//  1. SILLAGE_PASSWORD (si présent au démarrage) ; non persisté.
//  2. config.json existant dans dataDir (mot de passe déjà configuré lors
//     d'un précédent lancement avec SILLAGE_PASSWORD).
//  3. Aucun mot de passe.
func LoadPasswordHash(dataDir string) (hash string, err error) {
	if envPass := os.Getenv("SILLAGE_PASSWORD"); envPass != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(envPass), bcrypt.DefaultCost)
		if err != nil {
			return "", err
		}
		return string(h), nil
	}

	hash, ok, err := ReadPasswordHash(dataDir)
	if err != nil {
		return "", err
	}
	if ok {
		return hash, nil
	}
	return "", nil
}

// ReadPasswordHash lit le hash de mot de passe depuis config.json dans
// dataDir, sans jamais en générer un nouveau si absent. Utilisé après le
// rapatriement (clone) d'un espace de travail pour recharger immédiatement le
// mot de passe en mémoire, et par LoadPasswordHash au démarrage.
func ReadPasswordHash(dataDir string) (hash string, ok bool, err error) {
	data, err := os.ReadFile(configPath(dataDir))
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", false, err
	}
	if cfg.PasswordHash == "" {
		return "", false, nil
	}
	return cfg.PasswordHash, true, nil
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
