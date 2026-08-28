package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"
	"dboke-api/internal/core/services"
)

// ============================================================================
// Mock Infrastructure for E2E Tests
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

type mockDBAdapter struct {
	Connected bool
}

func (m *mockDBAdapter) Connect(ctx context.Context, config domain.DBConnectionConfig) error {
	m.Connected = true
	return nil
}
func (m *mockDBAdapter) Disconnect() error {
	m.Connected = false
	return nil
}
func (m *mockDBAdapter) Ping(ctx context.Context) error {
	if !m.Connected {
		return fmt.Errorf("not connected")
	}
	return nil
}
func (m *mockDBAdapter) GetDatabases(ctx context.Context) ([]string, error) {
	return []string{"testdb1", "testdb2"}, nil
}
func (m *mockDBAdapter) GetTables(ctx context.Context, database string) ([]domain.TableMetadata, error) {
	return []domain.TableMetadata{{Name: "users", Type: "BASE TABLE", Rows: "100", Size: "10KB"}}, nil
}
func (m *mockDBAdapter) GetTableSchema(ctx context.Context, database, table string) ([]domain.ColumnMetadata, error) {
	return []domain.ColumnMetadata{{Name: "id", Type: "int", IsPrimaryKey: true, IsNullable: false}}, nil
}
func (m *mockDBAdapter) FetchRows(ctx context.Context, query domain.SelectQuery) (*domain.ResultSet, error) {
	return &domain.ResultSet{Columns: []string{"id"}, Rows: []map[string]interface{}{{"id": 1}}}, nil
}
func (m *mockDBAdapter) UpdateRow(ctx context.Context, update domain.UpdateQuery) (int64, error) {
	return 1, nil
}
func (m *mockDBAdapter) ExecuteRaw(ctx context.Context, query string, params ...interface{}) (*domain.ResultSet, error) {
	return &domain.ResultSet{Columns: []string{"result"}, Rows: []map[string]interface{}{{"result": "ok"}}}, nil
}

// ============================================================================
// E2E Router Tests
// ============================================================================

func TestAPI_E2E_Flow(t *testing.T) {
	os.Setenv("DBOKE_MASTER_KEY", "01234567890123456789012345678901")
	defer os.Unsetenv("DBOKE_MASTER_KEY")

	store := newMockSessionStore()
	factories := map[string]func() ports.IDBAdapter {
		"pgsql": func() ports.IDBAdapter { return &mockDBAdapter{} },
	}

	authSvc := services.NewAuthService(store, factories)
	dbSvc := services.NewDBService(store, factories)
	pluginManager := services.NewPluginManager()

	r := NewRouter(store, authSvc, dbSvc, pluginManager)
	server := httptest.NewServer(r)
	defer server.Close()

	client := &http.Client{
		// Don't follow redirects automatically, handle cookies manually
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 1. Health check
	resp, err := client.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Health check status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// 2. Unauthenticated access to protected route
	resp, err = client.Get(server.URL + "/api/v1/ping")
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Ping status = %d, want %d for unauthenticated", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()

	// 3. Login
	loginBody := []byte(`{"dbType":"pgsql","host":"localhost","port":"5432","username":"admin","password":"password"}`)
	resp, err = client.Post(server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Login status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Extract CSRF token
	var loginResp map[string]string
	json.NewDecoder(resp.Body).Decode(&loginResp)
	csrfToken := loginResp["csrfToken"]
	if csrfToken == "" {
		t.Fatal("Expected csrfToken in login response")
	}
	resp.Body.Close()

	// Extract session cookie
	var sessionCookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == "dboke_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("Expected dboke_session cookie in login response")
	}

	// 4. Authenticated access to protected GET route
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/ping", nil)
	req.AddCookie(sessionCookie)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Authenticated ping failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Authenticated ping status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// 5. Authenticated access to protected POST route WITHOUT CSRF
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/databases/testdb1/query", bytes.NewBufferString(`{"query":"SELECT 1"}`))
	req.AddCookie(sessionCookie)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Query POST failed: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("Query POST without CSRF status = %d, want %d (Forbidden)", resp.StatusCode, http.StatusForbidden)
	}
	resp.Body.Close()

	// 6. Authenticated access to protected POST route WITH CSRF
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/databases/testdb1/query", bytes.NewBufferString(`{"query":"SELECT 1"}`))
	req.AddCookie(sessionCookie)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfToken)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Query POST failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Query POST with CSRF status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// 7. Test DB endpoints (GetDatabases)
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/databases", nil)
	req.AddCookie(sessionCookie)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("GetDatabases failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetDatabases status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	
	var dbsResp struct {
		Databases []string `json:"databases"`
	}
	json.NewDecoder(resp.Body).Decode(&dbsResp)
	if len(dbsResp.Databases) == 0 {
		t.Error("Expected databases to be returned")
	}
	resp.Body.Close()

	// 8. Logout
	req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/v1/auth/logout", nil)
	req.AddCookie(sessionCookie)
	// We might need CSRF for logout depending on implementation, let's include it
	req.Header.Set("X-CSRF-Token", csrfToken)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Logout failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Logout status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()

	// 9. Verify session is destroyed
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/api/v1/ping", nil)
	req.AddCookie(sessionCookie) // Try with old cookie
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Post-logout ping failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Post-logout ping status = %d, want %d (Unauthorized)", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()
}
