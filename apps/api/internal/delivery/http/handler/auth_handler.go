package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/services"
	"dboke-api/internal/pkg/response"
	"log/slog"
)

type AuthHandler struct {
	authService *services.AuthService
	dbService   *services.DBService
}

// NewAuthHandler creates the auth handler. dbService may be nil; when set, a
// session's cached database connections are closed on logout.
func NewAuthHandler(authService *services.AuthService, dbService *services.DBService) *AuthHandler {
	return &AuthHandler{authService: authService, dbService: dbService}
}

// isAllowedDBHost checks host against the comma-separated DBOKE_ALLOWED_DB_HOSTS
// allowlist. An unset allowlist permits any host.
func isAllowedDBHost(host string) bool {
	allowed := os.Getenv("DBOKE_ALLOWED_DB_HOSTS")
	if strings.TrimSpace(allowed) == "" {
		return true
	}
	for _, h := range strings.Split(allowed, ",") {
		if strings.EqualFold(strings.TrimSpace(h), host) {
			return true
		}
	}
	return false
}

// resolveDBHost maps loopback hosts to DBOKE_LOCALHOST_ALIAS when set. In a
// container, "localhost" is the container itself, so compose sets the alias to
// host.docker.internal to make "localhost" mean the user's machine.
func resolveDBHost(host string) string {
	alias := strings.TrimSpace(os.Getenv("DBOKE_LOCALHOST_ALIAS"))
	if alias == "" {
		return host
	}
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1", "[::1]":
		return alias
	}
	return host
}

type LoginRequest struct {
	DbType   string `json:"dbType"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}


func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Invalid login request payload", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Invalid request payload")
		return
	}

	portInt, _ := strconv.Atoi(req.Port)
	
	defaultDb := ""
	if req.DbType == "pgsql" {
		defaultDb = "postgres"
	} else if req.DbType == "mysql" {
		defaultDb = "mysql"
	}
	
	host := req.Host
	if host == "" {
		host = "localhost"
	}

	if !isAllowedDBHost(host) {
		slog.Warn("Login rejected for disallowed database host", slog.String("host", host))
		response.WriteError(w, http.StatusForbidden, "HOST_NOT_ALLOWED", "Connections to this database host are not allowed")
		return
	}
	
	config := domain.DBConnectionConfig{
		Host:     resolveDBHost(host),
		Port:     portInt,
		User:     req.Username,
		Password: req.Password,
		Database: defaultDb,
	}

	session, err := h.authService.Authenticate(r.Context(), req.DbType, config)
	if err != nil {
		slog.Error("Login failed", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusUnauthorized, "AUTH_FAILED", "Invalid credentials or connection failed")
		return
	}

	sessionID := session.ID
	csrfToken := session.CSRFToken


	env := os.Getenv("ENV")
	isProduction := env == "production"

	// Set a secure HTTP-Only cookie for the session
	http.SetCookie(w, &http.Cookie{
		Name:     "dboke_session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Connection established successfully",
		"dbType":    req.DbType,
		"csrfToken": csrfToken,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("dboke_session")
	if err == nil {
		// Remove from session store via service
		_ = h.authService.Logout(r.Context(), cookie.Value)
		if h.dbService != nil {
			h.dbService.CloseSessionConnection(cookie.Value)
		}
	}

	env := os.Getenv("ENV")
	isProduction := env == "production"

	// Clear the cookie by setting MaxAge to -1
	http.SetCookie(w, &http.Cookie{
		Name:     "dboke_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}
