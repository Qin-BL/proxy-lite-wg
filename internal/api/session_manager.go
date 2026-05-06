package api

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"proxy-lite-wg/internal/config"
)

const sessionCookieName = "proxy_lite_admin_session"

type SessionManager struct {
	username     string
	password     string
	passwordHash string
	secret       []byte
	ttl          time.Duration
}

func NewSessionManager(cfg config.Config) (*SessionManager, error) {
	secret, err := loadSessionSecret(cfg.AdminSessionSecret, cfg.AdminSessionSecretPath)
	if err != nil {
		return nil, err
	}

	return &SessionManager{
		username:     cfg.AdminUsername,
		password:     cfg.AdminPassword,
		passwordHash: cfg.AdminPasswordHash,
		secret:       secret,
		ttl:          cfg.AdminSessionTTL,
	}, nil
}

func loadSessionSecret(seed, path string) ([]byte, error) {
	if token := strings.TrimSpace(seed); token != "" {
		return []byte(token), nil
	}

	if path != "" {
		if stored, err := os.ReadFile(path); err == nil {
			if value := strings.TrimSpace(string(stored)); value != "" {
				return []byte(value), nil
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	secret := base64.RawURLEncoding.EncodeToString(buf)

	if path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(secret+"\n"), 0o600); err != nil {
			return nil, err
		}
	}

	return []byte(secret), nil
}

func (m *SessionManager) Authenticate(username, password string) bool {
	if strings.TrimSpace(username) != m.username {
		return false
	}
	if m.passwordHash != "" {
		return bcrypt.CompareHashAndPassword([]byte(m.passwordHash), []byte(password)) == nil
	}
	return password == m.password
}

func (m *SessionManager) SetSession(w http.ResponseWriter, r *http.Request) error {
	expiresAt := time.Now().UTC().Add(m.ttl)
	payload := fmt.Sprintf("%s|%d", m.username, expiresAt.Unix())
	signature := m.sign(payload)
	value := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(signature)

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   requestIsSecure(r),
		Expires:  expiresAt,
	})
	return nil
}

func (m *SessionManager) ClearSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   requestIsSecure(r),
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func (m *SessionManager) CurrentUser(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return "", false
	}

	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return "", false
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", false
	}

	payload := string(payloadBytes)
	if !hmac.Equal(signature, m.sign(payload)) {
		return "", false
	}

	payloadParts := strings.Split(payload, "|")
	if len(payloadParts) != 2 {
		return "", false
	}
	if payloadParts[0] != m.username {
		return "", false
	}

	expiresUnix, err := strconv.ParseInt(payloadParts[1], 10, 64)
	if err != nil {
		return "", false
	}
	if time.Now().UTC().After(time.Unix(expiresUnix, 0).UTC()) {
		return "", false
	}

	return m.username, true
}

func (m *SessionManager) sign(payload string) []byte {
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(payload))
	return mac.Sum(nil)
}

func requestIsSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https")
}
