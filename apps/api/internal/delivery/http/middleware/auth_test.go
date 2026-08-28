package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/pkg/contextkeys"
)

// ============================================================================
// Mock Session Store for Auth Middleware Tests
// ============================================================================

type mockSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

func newMockSessionStore() *mockSessionStore {
	return &mockSessionStore{sessions: make(map[string]*domain.Session)}
}

func (m *mockSessionStore) CreateSession(ctx context.Context, session *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

func (m *mockSessionStore) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.sessions[sessionID]; ok {
		if time.Now().After(s.ExpiresAt) {
			return nil, fmt.Errorf("expired")
		}
		return s, nil
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockSessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

func (m *mockSessionStore) ExtendSession(ctx context.Context, sessionID string, newExpiry time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[sessionID]; ok {
		s.ExpiresAt = newExpiry
	}
	return nil
}

// ============================================================================
// Auth Middleware Tests
// ============================================================================

func TestAuthMiddleware_NoCookie(t *testing.T) {
	store := newMockSessionStore()
	middleware := AuthMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called when no cookie is present")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/databases", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["code"] != "UNAUTHORIZED" {
		t.Errorf("Error code = %q, want UNAUTHORIZED", body["code"])
	}
}

func TestAuthMiddleware_InvalidSession(t *testing.T) {
	store := newMockSessionStore()
	middleware := AuthMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for invalid session")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/databases", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "non-existent-session"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["code"] != "INVALID_SESSION" {
		t.Errorf("Error code = %q, want INVALID_SESSION", body["code"])
	}
}

func TestAuthMiddleware_ExpiredSession(t *testing.T) {
	store := newMockSessionStore()
	ctx := context.Background()

	store.CreateSession(ctx, &domain.Session{
		ID:        "expired-sess",
		UserID:    "user",
		Role:      "admin",
		CSRFToken: "token",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // expired
	})

	middleware := AuthMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called for expired session")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/databases", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "expired-sess"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_GET_NoCSRFRequired(t *testing.T) {
	store := newMockSessionStore()
	ctx := context.Background()

	store.CreateSession(ctx, &domain.Session{
		ID:        "valid-sess",
		UserID:    "testuser",
		Role:      "admin",
		CSRFToken: "csrf-token-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	middleware := AuthMiddleware(store)

	var capturedUserID, capturedRole, capturedSessionID string

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = r.Context().Value(contextkeys.UserIDKey).(string)
		capturedRole = r.Context().Value(contextkeys.RoleKey).(string)
		capturedSessionID = r.Context().Value(contextkeys.SessionIDKey).(string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/databases", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "valid-sess"})
	// No CSRF header — GET should not require it
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
	if capturedUserID != "testuser" {
		t.Errorf("UserID = %q, want %q", capturedUserID, "testuser")
	}
	if capturedRole != "admin" {
		t.Errorf("Role = %q, want %q", capturedRole, "admin")
	}
	if capturedSessionID != "valid-sess" {
		t.Errorf("SessionID = %q, want %q", capturedSessionID, "valid-sess")
	}
}

func TestAuthMiddleware_OPTIONS_NoCSRFRequired(t *testing.T) {
	store := newMockSessionStore()
	ctx := context.Background()

	store.CreateSession(ctx, &domain.Session{
		ID:        "options-sess",
		UserID:    "testuser",
		Role:      "admin",
		CSRFToken: "csrf-token-123",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	middleware := AuthMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/databases", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "options-sess"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d for OPTIONS request", w.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_POST_MissingCSRF(t *testing.T) {
	store := newMockSessionStore()
	ctx := context.Background()

	store.CreateSession(ctx, &domain.Session{
		ID:        "csrf-test-sess",
		UserID:    "testuser",
		Role:      "admin",
		CSRFToken: "correct-csrf-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	middleware := AuthMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without CSRF token on POST")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/databases/test/query", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "csrf-test-sess"})
	// No X-CSRF-Token header
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["code"] != "CSRF_FAILED" {
		t.Errorf("Error code = %q, want CSRF_FAILED", body["code"])
	}
}

func TestAuthMiddleware_POST_WrongCSRF(t *testing.T) {
	store := newMockSessionStore()
	ctx := context.Background()

	store.CreateSession(ctx, &domain.Session{
		ID:        "csrf-wrong-sess",
		UserID:    "testuser",
		Role:      "admin",
		CSRFToken: "correct-csrf-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	middleware := AuthMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with wrong CSRF token")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/databases/test/query", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "csrf-wrong-sess"})
	req.Header.Set("X-CSRF-Token", "wrong-csrf-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestAuthMiddleware_POST_CorrectCSRF(t *testing.T) {
	store := newMockSessionStore()
	ctx := context.Background()

	store.CreateSession(ctx, &domain.Session{
		ID:        "csrf-ok-sess",
		UserID:    "testuser",
		Role:      "admin",
		CSRFToken: "valid-csrf-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	middleware := AuthMiddleware(store)
	handlerCalled := false

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/databases/test/query", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "csrf-ok-sess"})
	req.Header.Set("X-CSRF-Token", "valid-csrf-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
	if !handlerCalled {
		t.Error("Handler was not called with valid CSRF token")
	}
}

func TestAuthMiddleware_PUT_RequiresCSRF(t *testing.T) {
	store := newMockSessionStore()
	ctx := context.Background()

	store.CreateSession(ctx, &domain.Session{
		ID:        "put-csrf-sess",
		UserID:    "testuser",
		Role:      "admin",
		CSRFToken: "put-csrf-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	middleware := AuthMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without CSRF on PUT")
	}))

	req := httptest.NewRequest(http.MethodPut, "/api/v1/something", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "put-csrf-sess"})
	// No CSRF header
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestAuthMiddleware_DELETE_RequiresCSRF(t *testing.T) {
	store := newMockSessionStore()
	ctx := context.Background()

	store.CreateSession(ctx, &domain.Session{
		ID:        "delete-csrf-sess",
		UserID:    "testuser",
		Role:      "admin",
		CSRFToken: "delete-csrf-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	middleware := AuthMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called without CSRF on DELETE")
	}))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/plugins/test", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "delete-csrf-sess"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestAuthMiddleware_ContextInjection(t *testing.T) {
	store := newMockSessionStore()
	ctx := context.Background()

	store.CreateSession(ctx, &domain.Session{
		ID:        "context-test-sess",
		UserID:    "db_admin",
		Role:      "viewer",
		CSRFToken: "csrf",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	middleware := AuthMiddleware(store)

	var gotUserID, gotRole, gotSessionID string

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = r.Context().Value(contextkeys.UserIDKey).(string)
		gotRole, _ = r.Context().Value(contextkeys.RoleKey).(string)
		gotSessionID, _ = r.Context().Value(contextkeys.SessionIDKey).(string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: "context-test-sess"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if gotUserID != "db_admin" {
		t.Errorf("Context UserID = %q, want %q", gotUserID, "db_admin")
	}
	if gotRole != "viewer" {
		t.Errorf("Context Role = %q, want %q", gotRole, "viewer")
	}
	if gotSessionID != "context-test-sess" {
		t.Errorf("Context SessionID = %q, want %q", gotSessionID, "context-test-sess")
	}
}

func TestAuthMiddleware_EmptyCookieValue(t *testing.T) {
	store := newMockSessionStore()
	middleware := AuthMiddleware(store)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called with empty cookie value")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/databases", nil)
	req.AddCookie(&http.Cookie{Name: "dboke_session", Value: ""})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
