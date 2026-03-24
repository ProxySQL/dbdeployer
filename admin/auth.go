package admin

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

type AuthManager struct {
	otp      string
	otpUsed  bool
	sessions map[string]time.Time
	mu       sync.RWMutex
}

func NewAuthManager() *AuthManager {
	otp := generateOTP()
	return &AuthManager{
		otp:      otp,
		sessions: make(map[string]time.Time),
	}
}

func (a *AuthManager) OTP() string {
	return a.otp
}

func (a *AuthManager) ValidateOTP(token string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.otpUsed || token != a.otp {
		return false
	}
	a.otpUsed = true
	return true
}

func (a *AuthManager) CreateSession() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	sessionID := generateOTP()
	a.sessions[sessionID] = time.Now().Add(1 * time.Hour)
	return sessionID
}

func (a *AuthManager) ValidateSession(r *http.Request) bool {
	cookie, err := r.Cookie("dbdeployer_session")
	if err != nil {
		return false
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	expiry, ok := a.sessions[cookie.Value]
	return ok && time.Now().Before(expiry)
}

func (a *AuthManager) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.ValidateSession(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func generateOTP() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand.Read failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}
