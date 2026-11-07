package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"sync"
	"time"
)

// Session represents an active user session container.
type Session struct {
	ID        string
	values    map[string]any
	flashes   map[string]string
	createdAt time.Time
	mu        sync.RWMutex
}

// NewSession creates an empty session instance.
func NewSession(id string) *Session {
	return &Session{
		ID:        id,
		values:    make(map[string]any),
		flashes:   make(map[string]string),
		createdAt: time.Now(),
	}
}

// Get retrieves a value from the session.
func (s *Session) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.values[key]
	return val, ok
}

// GetString retrieves a string value from the session with fallback.
func (s *Session) GetString(key, fallback string) string {
	val, ok := s.Get(key)
	if !ok {
		return fallback
	}
	if str, ok := val.(string); ok {
		return str
	}
	return fallback
}

// Set stores a value in the session.
func (s *Session) Set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = val
}

// Delete removes a key from the session.
func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
}

// All retrieves a copy of all stored session values.
func (s *Session) All() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyValues := make(map[string]any, len(s.values))
	for k, v := range s.values {
		copyValues[k] = v
	}
	return copyValues
}

// Flash sets a one-time flash message for the next request.
func (s *Session) Flash(key, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flashes[key] = message
}

// GetFlash retrieves and consumes a one-time flash message.
func (s *Session) GetFlash(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg := s.flashes[key]
	delete(s.flashes, key)
	return msg
}

// AllFlashes retrieves all flash messages and clears them from session.
func (s *Session) AllFlashes() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	copyFlashes := make(map[string]string, len(s.flashes))
	for k, v := range s.flashes {
		copyFlashes[k] = v
	}
	s.flashes = make(map[string]string)
	return copyFlashes
}

var (
	defaultManager     *Manager
	defaultManagerOnce sync.Once
)

// DefaultManager returns a shared singleton session manager.
func DefaultManager() *Manager {
	defaultManagerOnce.Do(func() {
		secret := os.Getenv("APP_KEY")
		if secret == "" {
			secret = "larapak_default_secret_key_ngapak_session"
		}
		defaultManager = NewManager(secret, 24*time.Hour)
	})
	return defaultManager
}

// Manager handles session lifecycles, cookies, and signing.
type Manager struct {
	CookieName string
	SecretKey  []byte
	MaxAge     time.Duration
	sessions   map[string]*Session
	mu         sync.RWMutex
}

// NewManager creates a thread-safe Session Manager.
func NewManager(secretKey string, maxAge time.Duration) *Manager {
	if secretKey == "" {
		secretKey = "larapak_default_secret_key_ngapak"
	}
	return &Manager{
		CookieName: "larapak_session",
		SecretKey:  []byte(secretKey),
		MaxAge:     maxAge,
		sessions:   make(map[string]*Session),
	}
}

// Start retrieves an existing session or creates a new one for the incoming HTTP request.
func (m *Manager) Start(w http.ResponseWriter, r *http.Request) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()

	var sessionID string
	if cookie, err := r.Cookie(m.CookieName); err == nil {
		if validID, ok := m.verifySignedID(cookie.Value); ok {
			sessionID = validID
		}
	}

	if sessionID != "" {
		if sess, exists := m.sessions[sessionID]; exists {
			return sess
		}
	}

	// Create new session ID
	newID := generateRandomHex(24)
	sess := NewSession(newID)
	m.sessions[newID] = sess

	// Set signed HTTP-only cookie
	signed := m.signID(newID)
	http.SetCookie(w, &http.Cookie{
		Name:     m.CookieName,
		Value:    signed,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(m.MaxAge.Seconds()),
	})

	return sess
}

func (m *Manager) signID(id string) string {
	mac := hmac.New(sha256.New, m.SecretKey)
	mac.Write([]byte(id))
	sig := hex.EncodeToString(mac.Sum(nil))
	return id + "." + sig
}

func (m *Manager) verifySignedID(signed string) (string, bool) {
	for i := len(signed) - 1; i >= 0; i-- {
		if signed[i] == '.' {
			id := signed[:i]
			sig := signed[i+1:]

			mac := hmac.New(sha256.New, m.SecretKey)
			mac.Write([]byte(id))
			expectedSig := hex.EncodeToString(mac.Sum(nil))

			if hmac.Equal([]byte(sig), []byte(expectedSig)) {
				return id, true
			}
			return "", false
		}
	}
	return "", false
}

func generateRandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

