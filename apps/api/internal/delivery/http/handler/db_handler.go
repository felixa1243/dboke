package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/services"
	"dboke-api/internal/pkg/contextkeys"
	"dboke-api/internal/pkg/response"

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

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	databases, err := adapter.GetDatabases(ctx)
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

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	tables, err := adapter.GetTables(ctx, database)
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

	qStr := strings.TrimRight(req.Query, "; \n\r\t")

	// A simplistic but effective pagination wrapper for SELECT queries
	isSelect := isSelectQuery(qStr)

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

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	rs, err := adapter.ExecuteRaw(ctx, finalQuery, req.Params...)
	if err != nil {
		slog.Error("Failed to execute query", slog.String("error", err.Error()))
		response.WriteError(w, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}

	// Paginated tells the client whether LIMIT/OFFSET was applied, so it only
	// offers "fetch next" for queries that can safely be re-run page by page.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		*domain.ResultSet
		Paginated bool `json:"paginated"`
	}{rs, isSelect})
}

// isSelectQuery reports whether q is a plain SELECT statement, ignoring case,
// leading whitespace and leading SQL comments.
func isSelectQuery(q string) bool {
	for {
		q = strings.TrimLeft(q, " \n\r\t")
		if strings.HasPrefix(q, "--") {
			if i := strings.IndexByte(q, '\n'); i >= 0 {
				q = q[i+1:]
				continue
			}
			return false
		}
		if strings.HasPrefix(q, "/*") {
			if i := strings.Index(q, "*/"); i >= 0 {
				q = q[i+2:]
				continue
			}
			return false
		}
		break
	}
	if len(q) < 6 || !strings.EqualFold(q[:6], "SELECT") {
		return false
	}
	// Require a word boundary so identifiers like "selection" don't match
	if len(q) > 6 {
		c := q[6]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			return false
		}
	}
	return true
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

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	columns, err := adapter.GetTableSchema(ctx, database, table)
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
