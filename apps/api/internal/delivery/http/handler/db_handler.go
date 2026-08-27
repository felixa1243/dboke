package handler

import (
	"encoding/json"
	"net/http"
	"dboke-api/internal/core/services"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

type DBHandler struct {
	dbService *services.DBService
}

func NewDBHandler(dbService *services.DBService) *DBHandler {
	return &DBHandler{dbService: dbService}
}

func (h *DBHandler) GetDatabases(w http.ResponseWriter, r *http.Request) {
	sessionID, ok := r.Context().Value("session_id").(string)
	if !ok || sessionID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	adapter, err := h.dbService.GetAdapter(r.Context(), sessionID, "")
	if err != nil {
		slog.Error("Failed to get adapter", slog.String("error", err.Error()))
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	databases, err := adapter.GetDatabases(r.Context())
	if err != nil {
		slog.Error("Failed to list databases", slog.String("error", err.Error()))
		http.Error(w, "Failed to list databases", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"databases": databases,
	})
}

func (h *DBHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	database := chi.URLParam(r, "database")
	sessionID, ok := r.Context().Value("session_id").(string)
	if !ok || sessionID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	adapter, err := h.dbService.GetAdapter(r.Context(), sessionID, database)
	if err != nil {
		slog.Error("Failed to get adapter", slog.String("error", err.Error()))
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}

	tables, err := adapter.GetTables(r.Context(), database)
	if err != nil {
		slog.Error("Failed to list tables", slog.String("error", err.Error()))
		http.Error(w, "Failed to list tables", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tables": tables,
	})
}
