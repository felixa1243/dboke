package services

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"
)

// ============================================================================
// Mock Implementations
// ============================================================================

// MockSessionStore is a test-only in-memory session store
type MockSessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

func NewMockSessionStore() *MockSessionStore {
	return &MockSessionStore{sessions: make(map[string]*domain.Session)}
}

func (m *MockSessionStore) CreateSession(ctx context.Context, session *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

func (m *MockSessionStore) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.sessions[sessionID]; ok {
		if time.Now().After(s.ExpiresAt) {
			return nil, fmt.Errorf("session expired")
		}
		return s, nil
	}
	return nil, fmt.Errorf("session not found")
}

func (m *MockSessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

func (m *MockSessionStore) ExtendSession(ctx context.Context, sessionID string, newExpiry time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[sessionID]; ok {
		s.ExpiresAt = newExpiry
	}
	return nil
}

// MockDBAdapter simulates a database adapter for testing
type MockDBAdapter struct {
	Connected    bool
	PingError    error
	ConnectError error
	Databases    []string
	Tables       []domain.TableMetadata
	Columns      []domain.ColumnMetadata
}

func (m *MockDBAdapter) Connect(ctx context.Context, config domain.DBConnectionConfig) error {
	if m.ConnectError != nil {
		return m.ConnectError
	}
	m.Connected = true
	return nil
}

func (m *MockDBAdapter) Disconnect() error {
	m.Connected = false
	return nil
}

func (m *MockDBAdapter) Ping(ctx context.Context) error {
	if m.PingError != nil {
		return m.PingError
	}
	if !m.Connected {
		return fmt.Errorf("not connected")
	}
	return nil
}

func (m *MockDBAdapter) GetDatabases(ctx context.Context) ([]string, error) {
	return m.Databases, nil
}

func (m *MockDBAdapter) GetTables(ctx context.Context, database string) ([]domain.TableMetadata, error) {
	return m.Tables, nil
}

func (m *MockDBAdapter) GetTableSchema(ctx context.Context, database, table string) ([]domain.ColumnMetadata, error) {
	return m.Columns, nil
}

func (m *MockDBAdapter) FetchRows(ctx context.Context, query domain.SelectQuery) (*domain.ResultSet, error) {
	return &domain.ResultSet{Columns: []string{"id"}, Rows: []map[string]interface{}{{"id": 1}}}, nil
}

func (m *MockDBAdapter) UpdateRow(ctx context.Context, update domain.UpdateQuery) (int64, error) {
	return 1, nil
}

func (m *MockDBAdapter) ExecuteRaw(ctx context.Context, query string, params ...interface{}) (*domain.ResultSet, error) {
	return &domain.ResultSet{Columns: []string{"result"}, Rows: []map[string]interface{}{{"result": "ok"}}}, nil
}

// ============================================================================
// AuthService Tests
// ============================================================================

func TestAuthService_Authenticate_UnsupportedDBType(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{}

	svc := NewAuthService(store, factories)
	ctx := context.Background()

	_, err := svc.Authenticate(ctx, "oracle", domain.DBConnectionConfig{
		Host: "localhost", Port: 1521, User: "user", Password: "pass",
	})

	if err == nil {
		t.Error("Expected error for unsupported db type, got nil")
	}
}

func TestAuthService_Authenticate_ConnectionFailure(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			return &MockDBAdapter{ConnectError: fmt.Errorf("connection refused")}
		},
	}

	svc := NewAuthService(store, factories)
	ctx := context.Background()

	_, err := svc.Authenticate(ctx, "pgsql", domain.DBConnectionConfig{
		Host: "badhost", Port: 5432, User: "user", Password: "wrong",
	})

	if err == nil {
		t.Error("Expected error for connection failure, got nil")
	}
}

func TestAuthService_Authenticate_Success(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			return &MockDBAdapter{}
		},
	}

	// Set required env var
	t.Setenv("DBOKE_MASTER_KEY", "01234567890123456789012345678901")

	svc := NewAuthService(store, factories)
	ctx := context.Background()

	session, err := svc.Authenticate(ctx, "pgsql", domain.DBConnectionConfig{
		Host: "localhost", Port: 5432, User: "testuser", Password: "testpass",
	})

	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}
	if session.CSRFToken == "" {
		t.Error("CSRF token should not be empty")
	}
	if session.UserID != "testuser" {
		t.Errorf("UserID = %q, want %q", session.UserID, "testuser")
	}
	if session.Role != "admin" {
		t.Errorf("Role = %q, want %q", session.Role, "admin")
	}
	if session.DBType != "pgsql" {
		t.Errorf("DBType = %q, want %q", session.DBType, "pgsql")
	}
	if session.EncryptedDBConfig == "" {
		t.Error("EncryptedDBConfig should not be empty")
	}
	if time.Until(session.ExpiresAt) < 23*time.Hour {
		t.Error("Session should expire in approximately 24 hours")
	}
}

func TestAuthService_Authenticate_SessionIsStored(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			return &MockDBAdapter{}
		},
	}

	t.Setenv("DBOKE_MASTER_KEY", "01234567890123456789012345678901")

	svc := NewAuthService(store, factories)
	ctx := context.Background()

	session, err := svc.Authenticate(ctx, "pgsql", domain.DBConnectionConfig{
		Host: "localhost", Port: 5432, User: "user", Password: "pass",
	})
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	// Verify the session was stored
	retrieved, err := store.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved.ID != session.ID {
		t.Errorf("Stored session ID mismatch: got %q, want %q", retrieved.ID, session.ID)
	}
}

func TestAuthService_Authenticate_UniqueSessionIDs(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			return &MockDBAdapter{}
		},
	}

	t.Setenv("DBOKE_MASTER_KEY", "01234567890123456789012345678901")

	svc := NewAuthService(store, factories)
	ctx := context.Background()
	config := domain.DBConnectionConfig{
		Host: "localhost", Port: 5432, User: "user", Password: "pass",
	}

	s1, _ := svc.Authenticate(ctx, "pgsql", config)
	s2, _ := svc.Authenticate(ctx, "pgsql", config)

	if s1.ID == s2.ID {
		t.Error("Two authentications produced the same session ID")
	}
	if s1.CSRFToken == s2.CSRFToken {
		t.Error("Two authentications produced the same CSRF token")
	}
}

func TestAuthService_Authenticate_MissingMasterKey(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			return &MockDBAdapter{}
		},
	}

	t.Setenv("DBOKE_MASTER_KEY", "") // empty key

	svc := NewAuthService(store, factories)
	ctx := context.Background()

	_, err := svc.Authenticate(ctx, "pgsql", domain.DBConnectionConfig{
		Host: "localhost", Port: 5432, User: "user", Password: "pass",
	})

	if err == nil {
		t.Error("Expected error when master key is empty, got nil")
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			return &MockDBAdapter{}
		},
	}

	t.Setenv("DBOKE_MASTER_KEY", "01234567890123456789012345678901")

	svc := NewAuthService(store, factories)
	ctx := context.Background()

	session, err := svc.Authenticate(ctx, "pgsql", domain.DBConnectionConfig{
		Host: "localhost", Port: 5432, User: "user", Password: "pass",
	})
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	// Logout
	err = svc.Logout(ctx, session.ID)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}

	// Verify session is gone
	_, err = store.GetSession(ctx, session.ID)
	if err == nil {
		t.Error("Session should be deleted after logout")
	}
}

func TestAuthService_Logout_NonExistentSession(t *testing.T) {
	store := NewMockSessionStore()
	svc := NewAuthService(store, nil)
	ctx := context.Background()

	// Should not error on deleting non-existent session
	err := svc.Logout(ctx, "non-existent-session-id")
	if err != nil {
		t.Errorf("Logout of non-existent session should not error, got: %v", err)
	}
}
