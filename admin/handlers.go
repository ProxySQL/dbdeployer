package admin

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		if err := s.templates.ExecuteTemplate(w, "login.html", nil); err != nil {
			http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if !s.auth.ValidateOTP(token) {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	sessionID := s.auth.CreateSession()
	http.SetCookie(w, &http.Cookie{
		Name:     "dbdeployer_session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	sandboxes, err := GetAllSandboxes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Sandboxes": sandboxes,
		"Count":     len(sandboxes),
	}
	if err := s.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleListSandboxes(w http.ResponseWriter, r *http.Request) {
	sandboxes, err := GetAllSandboxes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sandboxes)
}

func (s *Server) handleRefreshSandboxList(w http.ResponseWriter, r *http.Request) {
	sandboxes, err := GetAllSandboxes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Sandboxes": sandboxes,
		"Count":     len(sandboxes),
	}
	if err := s.templates.ExecuteTemplate(w, "sandbox-list.html", data); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleSandboxAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse: /api/sandboxes/<name>/<action>
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/sandboxes/"), "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid path format. Expected /api/sandboxes/<name>/<action>", http.StatusBadRequest)
		return
	}
	rawName, action := parts[0], parts[1]

	// URL-decode the sandbox name (it was urlquery-encoded in the template).
	name, err := url.PathUnescape(rawName)
	if err != nil {
		http.Error(w, "Invalid sandbox name encoding", http.StatusBadRequest)
		return
	}

	// Reject names that could escape the sandbox directory.
	if name == "" || strings.ContainsAny(name, "/\\") || name == "." || name == ".." {
		http.Error(w, "Invalid sandbox name", http.StatusBadRequest)
		return
	}

	var actionErr error
	switch action {
	case "start":
		actionErr = ExecuteSandboxScript(name, "start")
	case "stop":
		actionErr = ExecuteSandboxScript(name, "stop")
	case "destroy":
		actionErr = DestroySandbox(name)
	default:
		http.Error(w, "Unknown action", http.StatusBadRequest)
		return
	}

	if actionErr != nil {
		http.Error(w, actionErr.Error(), http.StatusInternalServerError)
		return
	}

	// Wait briefly for start/stop to take effect before reading status.
	// Start scripts launch mysqld in the background; the PID file
	// and status script need a moment to reflect the new state.
	if action == "start" || action == "stop" {
		time.Sleep(2 * time.Second)
	}

	// Return updated sandbox list fragment for HTMX.
	sandboxes, err := GetAllSandboxes()
	if err != nil {
		http.Error(w, "Action succeeded but failed to refresh: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Sandboxes": sandboxes,
		"Count":     len(sandboxes),
	}
	if err := s.templates.ExecuteTemplate(w, "sandbox-list.html", data); err != nil {
		http.Error(w, "template error: "+err.Error(), http.StatusInternalServerError)
	}
}
