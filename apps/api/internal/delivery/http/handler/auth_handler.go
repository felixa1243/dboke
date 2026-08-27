package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/services"
	"log/slog"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	DbType   string `json:"dbType"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}


func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Invalid login request payload", slog.String("error", err.Error()))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	portInt, _ := strconv.Atoi(req.Port)
	
	defaultDb := ""
	if req.DbType == "pgsql" {
		defaultDb = "postgres"
	} else if req.DbType == "mysql" {
		defaultDb = "mysql"
	}
	
	config := domain.DBConnectionConfig{
		Host:     "localhost",
		Port:     portInt,
		User:     req.Username,
		Password: req.Password,
		Database: defaultDb,
	}

	session, err := h.authService.Authenticate(r.Context(), req.DbType, config)
	if err != nil {
		slog.Error("Login failed", slog.String("error", err.Error()))
		http.Error(w, "Invalid credentials or connection failed", http.StatusUnauthorized)
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
