package admin

import (
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

//go:embed templates static
var content embed.FS

// Server holds the HTTP server state.
type Server struct {
	auth      *AuthManager
	port      int
	templates *template.Template
}

// NewServer creates a new admin Server.
func NewServer(port int) (*Server, error) {
	tmpl, err := template.New("").ParseFS(content,
		"templates/*.html",
		"templates/components/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}
	return &Server{
		auth:      NewAuthManager(),
		port:      port,
		templates: tmpl,
	}, nil
}

// Start binds the listener and begins serving requests. It blocks until the server exits.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Static files.
	mux.Handle("/static/", http.FileServer(http.FS(content)))

	// Auth routes.
	mux.HandleFunc("/login", s.handleLogin)

	// Dashboard (auth required).
	mux.HandleFunc("/", s.auth.AuthMiddleware(s.handleDashboard))

	// API (auth required).
	mux.HandleFunc("/api/sandboxes", s.auth.AuthMiddleware(s.handleListSandboxes))
	mux.HandleFunc("/api/sandboxes/", s.auth.AuthMiddleware(s.handleSandboxAction))

	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	url := fmt.Sprintf("http://%s/login?token=%s", addr, s.auth.OTP())

	fmt.Printf("\n  dbdeployer admin\n")
	fmt.Printf("  ────────────────────────────────\n")
	fmt.Printf("  URL: %s\n", url)
	fmt.Printf("  Press Ctrl+C to stop\n\n")

	// Open browser AFTER listener is ready
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	// Now that listener is ready, open browser
	if err := openBrowser(url); err != nil {
		fmt.Printf("  Warning: could not open browser: %v\n", err)
		fmt.Printf("  Open this URL manually: %s\n\n", url)
	}

	return server.Serve(listener)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return cmd.Start()
}
