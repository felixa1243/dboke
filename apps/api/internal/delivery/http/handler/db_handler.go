package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"dboke-api/internal/core/services"
	"dboke-api/internal/pkg/contextkeys"
	"dboke-api/internal/pkg/response"
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
	sessionID, ok := r.Context().Value(contextkeys.SessionIDKey).(string)
	if !ok || sessionID == "" {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	adapter, err := h.dbService.GetAdapter(r.Context(), sessionID, "")
	if err != nil {
		slog.Error("Failed to get adapter", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", "Database connection failed")
		return
	}

	databases, err := adapter.GetDatabases(r.Context())
	if err != nil {
		slog.Error("Failed to list databases", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to list databases")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"databases": databases,
	})
}

func (h *DBHandler) GetTables(w http.ResponseWriter, r *http.Request) {
	database := chi.URLParam(r, "database")
	sessionID, ok := r.Context().Value(contextkeys.SessionIDKey).(string)
	if !ok || sessionID == "" {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	adapter, err := h.dbService.GetAdapter(r.Context(), sessionID, database)
	if err != nil {
		slog.Error("Failed to get adapter", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", "Database connection failed")
		return
	}

	tables, err := adapter.GetTables(r.Context(), database)
	if err != nil {
		slog.Error("Failed to list tables", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to list tables")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tables": tables,
	})
}

func (h *DBHandler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	database := chi.URLParam(r, "database")
	sessionID, ok := r.Context().Value(contextkeys.SessionIDKey).(string)
	if !ok || sessionID == "" {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req struct {
		Query  string        `json:"query"`
		Params []interface{} `json:"params"`
		Limit  int           `json:"limit"`
		Offset int           `json:"offset"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Limit <= 0 {
		req.Limit = 200
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Trim trailing semicolons and whitespace
	importStrings := false
	_ = importStrings // to avoid unused error if strings isn't imported, wait I should use strings.HasPrefix

	// Import strings if necessary, but we can't do it inline easily without breaking format. 
	// Actually we should just trim it manually if we don't have strings package guaranteed
	qStr := req.Query
	for len(qStr) > 0 && (qStr[len(qStr)-1] == ';' || qStr[len(qStr)-1] == ' ' || qStr[len(qStr)-1] == '\n' || qStr[len(qStr)-1] == '\r' || qStr[len(qStr)-1] == '\t') {
		qStr = qStr[:len(qStr)-1]
	}

	// A simplistic but effective pagination wrapper for SELECT queries
	isSelect := false
	if len(qStr) >= 6 {
		prefix := qStr[:6]
		if prefix == "SELECT" || prefix == "select" || prefix == "Select" {
			isSelect = true
		}
	}

	finalQuery := qStr
	if isSelect {
		// Both Postgres and MySQL support LIMIT and OFFSET
		finalQuery = "SELECT * FROM (" + qStr + ") AS _dboke_subq LIMIT " + strconv.Itoa(req.Limit) + " OFFSET " + strconv.Itoa(req.Offset)
	}

	adapter, err := h.dbService.GetAdapter(r.Context(), sessionID, database)
	if err != nil {
		slog.Error("Failed to get adapter", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", "Database connection failed")
		return
	}

	rs, err := adapter.ExecuteRaw(r.Context(), finalQuery, req.Params...)
	if err != nil {
		slog.Error("Failed to execute query", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rs)
}

func (h *DBHandler) GetTableSchema(w http.ResponseWriter, r *http.Request) {
	database := chi.URLParam(r, "database")
	table := chi.URLParam(r, "table")
	
	sessionID, ok := r.Context().Value(contextkeys.SessionIDKey).(string)
	if !ok || sessionID == "" {
		response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	adapter, err := h.dbService.GetAdapter(r.Context(), sessionID, database)
	if err != nil {
		slog.Error("Failed to get adapter", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", "Database connection failed")
		return
	}

	columns, err := adapter.GetTableSchema(r.Context(), database, table)
	if err != nil {
		slog.Error("Failed to get table schema", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", "Failed to list columns")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"columns": columns,
	})
}
