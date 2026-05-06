package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"

	"proxy-lite-wg/internal/config"
	"proxy-lite-wg/internal/domain"
)

type Service interface {
	CreateUser(ctx context.Context, name, email string) (domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	DeleteUser(ctx context.Context, id string) error
	CreateClient(ctx context.Context, userID, label string) (domain.Client, error)
	ListClients(ctx context.Context) ([]domain.Client, error)
	GetClient(ctx context.Context, id string) (domain.Client, error)
	DisableClient(ctx context.Context, id string) error
	EnableClient(ctx context.Context, id string) error
	DeleteClient(ctx context.Context, id string) error
	RenderShareLink(ctx context.Context, id string) (string, error)
}

type Server struct {
	cfg      config.Config
	svc      Service
	sessions *SessionManager
	mux      *http.ServeMux
}

func NewServer(cfg config.Config, svc Service) (*Server, error) {
	sessions, err := NewSessionManager(cfg)
	if err != nil {
		return nil, err
	}

	server := &Server{
		cfg:      cfg,
		svc:      svc,
		sessions: sessions,
		mux:      http.NewServeMux(),
	}
	server.routes()
	return server, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !s.isPublicPath(r.URL.Path) && !s.authorized(r) {
		writeError(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	s.mux.HandleFunc("/manage", s.handleManagePage)
	s.mux.HandleFunc("/api/v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("/api/v1/auth/logout", s.handleLogout)
	s.mux.HandleFunc("/api/v1/auth/me", s.handleMe)
	s.mux.HandleFunc("/api/v1/users", s.handleUsers)
	s.mux.HandleFunc("/api/v1/users/", s.handleUserByID)
	s.mux.HandleFunc("/api/v1/clients", s.handleClients)
	s.mux.HandleFunc("/api/v1/clients/", s.handleClientByID)
}

func (s *Server) isPublicPath(path string) bool {
	return slices.Contains([]string{
		"/healthz",
		"/manage",
		"/api/v1/auth/login",
		"/api/v1/auth/me",
	}, path)
}

func (s *Server) authorized(r *http.Request) bool {
	_, ok := s.sessions.CurrentUser(r)
	return ok
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC(),
		"mode":   "vless-ws",
	})
}

func (s *Server) handleManagePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(managePageHTML))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
		return
	}

	if !s.sessions.Authenticate(request.Username, request.Password) {
		writeError(w, http.StatusUnauthorized, errors.New("invalid username or password"))
		return
	}
	if err := s.sessions.SetSession(w, r); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"username": s.cfg.AdminUsername,
		"status":   "authenticated",
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	s.sessions.ClearSession(w, r)
	writeJSON(w, http.StatusOK, map[string]any{"status": "logged_out"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	username, ok := s.sessions.CurrentUser(r)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated": false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"username":    username,
		"public_host": s.cfg.PublicHost,
		"public_port": s.cfg.PublicPort,
		"ws_path":     s.cfg.VLESSWSPath,
	})
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		users, err := s.svc.ListUsers(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": users})
	case http.MethodPost:
		var request struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
			return
		}

		user, err := s.svc.CreateUser(r.Context(), request.Name, request.Email)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, user)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) handleUserByID(w http.ResponseWriter, r *http.Request) {
	userID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/users/"), "/")
	if userID == "" {
		writeError(w, http.StatusNotFound, errors.New("not found"))
		return
	}

	if r.Method != http.MethodDelete {
		writeMethodNotAllowed(w, http.MethodDelete)
		return
	}

	if err := s.svc.DeleteUser(r.Context(), userID); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":      userID,
		"status":  "deleted",
		"message": "user deleted",
	})
}

func (s *Server) handleClients(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		clients, err := s.svc.ListClients(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": clients})
	case http.MethodPost:
		var request struct {
			UserID string `json:"user_id"`
			Label  string `json:"label"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid JSON body"))
			return
		}

		clientRecord, err := s.svc.CreateClient(r.Context(), request.UserID, request.Label)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, clientRecord)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (s *Server) handleClientByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/clients/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, errors.New("not found"))
		return
	}

	clientID := parts[0]
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			clientRecord, err := s.svc.GetClient(r.Context(), clientID)
			if err != nil {
				writeError(w, http.StatusNotFound, err)
				return
			}
			writeJSON(w, http.StatusOK, clientRecord)
		case http.MethodDelete:
			if err := s.svc.DeleteClient(r.Context(), clientID); err != nil {
				writeError(w, http.StatusNotFound, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"id":      clientID,
				"status":  "deleted",
				"message": "client deleted",
			})
		default:
			writeMethodNotAllowed(w, http.MethodGet, http.MethodDelete)
		}
		return
	}

	switch parts[1] {
	case "disable":
		if r.Method != http.MethodPost {
			writeMethodNotAllowed(w, http.MethodPost)
			return
		}
		if err := s.svc.DisableClient(r.Context(), clientID); err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":      clientID,
			"status":  "disabled",
			"message": "client disabled",
		})
	case "enable":
		if r.Method != http.MethodPost {
			writeMethodNotAllowed(w, http.MethodPost)
			return
		}
		if err := s.svc.EnableClient(r.Context(), clientID); err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id":      clientID,
			"status":  "active",
			"message": "client enabled",
		})
	case "share-link":
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, http.MethodGet)
			return
		}
		shareLink, err := s.svc.RenderShareLink(r.Context(), clientID)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		if r.URL.Query().Get("download") == "1" {
			clientRecord, getErr := s.svc.GetClient(r.Context(), clientID)
			if getErr != nil {
				writeError(w, http.StatusNotFound, getErr)
				return
			}
			filename := sanitizeFilename(fmt.Sprintf("%s-%s.txt", clientRecord.Label, clientRecord.ID))
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(shareLink))
	case "share-link-qr.png":
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, http.MethodGet)
			return
		}
		shareLink, err := s.svc.RenderShareLink(r.Context(), clientID)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		png, err := qrcode.Encode(shareLink, qrcode.Medium, 512)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(png)
	default:
		writeError(w, http.StatusNotFound, errors.New("not found"))
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{
		"error": err.Error(),
	})
}

func writeMethodNotAllowed(w http.ResponseWriter, methods ...string) {
	w.Header().Set("Allow", strings.Join(methods, ", "))
	writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
}

func sanitizeFilename(value string) string {
	replacer := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		";", "-",
		"\"", "",
		"'", "",
	)
	return replacer.Replace(value)
}
