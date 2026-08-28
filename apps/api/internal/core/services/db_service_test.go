package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"
	"dboke-api/internal/infrastructure/security"
)

// ============================================================================
// DBService Tests
// ============================================================================

func TestDBService_GetAdapter_SessionNotFound(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{}

	svc := NewDBService(store, factories)
	ctx := context.Background()

	_, err := svc.GetAdapter(ctx, "non-existent-session", "")
	if err == nil {
		t.Error("Expected error for non-existent session, got nil")
	}
}

func TestDBService_GetAdapter_DecryptionFailure(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			return &MockDBAdapter{}
		},
	}

	ctx := context.Background()

	// Create session with bad encrypted config
	session := &domain.Session{
		ID:                "test-session-bad",
		UserID:            "user",
		Role:              "admin",
		CSRFToken:         "csrf",
		DBType:            "pgsql",
		EncryptedDBConfig: "invalid-not-encrypted-data",
		ExpiresAt:         time.Now().Add(24 * time.Hour),
	}
	store.CreateSession(ctx, session)

	t.Setenv("DBOKE_MASTER_KEY", "01234567890123456789012345678901")

	svc := NewDBService(store, factories)

	_, err := svc.GetAdapter(ctx, "test-session-bad", "")
	if err == nil {
		t.Error("Expected error for invalid encrypted config, got nil")
	}
}

func TestDBService_GetAdapter_UnsupportedDBType(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{} // no factories

	ctx := context.Background()
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)

	config := domain.DBConnectionConfig{Host: "localhost", Port: 5432, User: "u", Password: "p"}
	configBytes, _ := json.Marshal(config)
	encrypted, _ := security.Encrypt(string(configBytes), masterKey)

	session := &domain.Session{
		ID:                "test-session-unsupported",
		UserID:            "user",
		Role:              "admin",
		CSRFToken:         "csrf",
		DBType:            "oracle", // not registered
		EncryptedDBConfig: encrypted,
		ExpiresAt:         time.Now().Add(24 * time.Hour),
	}
	store.CreateSession(ctx, session)

	svc := NewDBService(store, factories)
	_, err := svc.GetAdapter(ctx, "test-session-unsupported", "")
	if err == nil {
		t.Error("Expected error for unsupported db type, got nil")
	}
}

func TestDBService_GetAdapter_Success(t *testing.T) {
	store := NewMockSessionStore()

	mockAdapter := &MockDBAdapter{}
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			return mockAdapter
		},
	}

	ctx := context.Background()
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)

	config := domain.DBConnectionConfig{Host: "localhost", Port: 5432, User: "u", Password: "p"}
	configBytes, _ := json.Marshal(config)
	encrypted, _ := security.Encrypt(string(configBytes), masterKey)

	session := &domain.Session{
		ID:                "test-session-ok",
		UserID:            "user",
		Role:              "admin",
		CSRFToken:         "csrf",
		DBType:            "pgsql",
		EncryptedDBConfig: encrypted,
		ExpiresAt:         time.Now().Add(24 * time.Hour),
	}
	store.CreateSession(ctx, session)

	svc := NewDBService(store, factories)
	adapter, err := svc.GetAdapter(ctx, "test-session-ok", "")
	if err != nil {
		t.Fatalf("GetAdapter failed: %v", err)
	}
	if adapter == nil {
		t.Fatal("Expected non-nil adapter")
	}
}

func TestDBService_GetAdapter_CachesAdapter(t *testing.T) {
	store := NewMockSessionStore()
	callCount := 0

	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			callCount++
			return &MockDBAdapter{}
		},
	}

	ctx := context.Background()
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)

	config := domain.DBConnectionConfig{Host: "localhost", Port: 5432, User: "u", Password: "p"}
	configBytes, _ := json.Marshal(config)
	encrypted, _ := security.Encrypt(string(configBytes), masterKey)

	session := &domain.Session{
		ID:                "cached-session",
		UserID:            "user",
		Role:              "admin",
		CSRFToken:         "csrf",
		DBType:            "pgsql",
		EncryptedDBConfig: encrypted,
		ExpiresAt:         time.Now().Add(24 * time.Hour),
	}
	store.CreateSession(ctx, session)

	svc := NewDBService(store, factories)

	// First call creates adapter
	_, err := svc.GetAdapter(ctx, "cached-session", "")
	if err != nil {
		t.Fatalf("First GetAdapter failed: %v", err)
	}

	// Second call should use cached adapter
	_, err = svc.GetAdapter(ctx, "cached-session", "")
	if err != nil {
		t.Fatalf("Second GetAdapter failed: %v", err)
	}

	if callCount != 1 {
		t.Errorf("Factory was called %d times, want 1 (should cache)", callCount)
	}
}

func TestDBService_GetAdapter_ReconnectsOnPingFailure(t *testing.T) {
	store := NewMockSessionStore()
	callCount := 0

	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			callCount++
			adapter := &MockDBAdapter{}
			if callCount == 1 {
				// First adapter will fail on ping (simulating dead connection)
				adapter.PingError = fmt.Errorf("connection lost")
			}
			return adapter
		},
	}

	ctx := context.Background()
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)

	config := domain.DBConnectionConfig{Host: "localhost", Port: 5432, User: "u", Password: "p"}
	configBytes, _ := json.Marshal(config)
	encrypted, _ := security.Encrypt(string(configBytes), masterKey)

	session := &domain.Session{
		ID:                "reconnect-session",
		UserID:            "user",
		Role:              "admin",
		CSRFToken:         "csrf",
		DBType:            "pgsql",
		EncryptedDBConfig: encrypted,
		ExpiresAt:         time.Now().Add(24 * time.Hour),
	}
	store.CreateSession(ctx, session)

	svc := NewDBService(store, factories)

	// First call — creates adapter that has PingError
	_, _ = svc.GetAdapter(ctx, "reconnect-session", "")

	// Second call — should detect dead connection and reconnect
	_, err := svc.GetAdapter(ctx, "reconnect-session", "")
	if err != nil {
		t.Fatalf("GetAdapter after reconnect failed: %v", err)
	}

	if callCount < 2 {
		t.Errorf("Factory was called %d times, expected at least 2 (should reconnect)", callCount)
	}
}

func TestDBService_GetAdapter_WithTargetDatabase(t *testing.T) {
	store := NewMockSessionStore()
	callCount := 0

	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			callCount++
			return &MockDBAdapter{}
		},
	}

	ctx := context.Background()
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)

	config := domain.DBConnectionConfig{Host: "localhost", Port: 5432, User: "u", Password: "p"}
	configBytes, _ := json.Marshal(config)
	encrypted, _ := security.Encrypt(string(configBytes), masterKey)

	session := &domain.Session{
		ID:                "multi-db-session",
		UserID:            "user",
		Role:              "admin",
		CSRFToken:         "csrf",
		DBType:            "pgsql",
		EncryptedDBConfig: encrypted,
		ExpiresAt:         time.Now().Add(24 * time.Hour),
	}
	store.CreateSession(ctx, session)

	svc := NewDBService(store, factories)

	// Connect to different databases — each should create a separate adapter
	_, _ = svc.GetAdapter(ctx, "multi-db-session", "db1")
	_, _ = svc.GetAdapter(ctx, "multi-db-session", "db2")

	if callCount != 2 {
		t.Errorf("Factory was called %d times, want 2 (different databases should create separate adapters)", callCount)
	}
}

func TestDBService_CloseSessionConnection(t *testing.T) {
	store := NewMockSessionStore()
	adapter1 := &MockDBAdapter{}
	adapter2 := &MockDBAdapter{}
	callIndex := 0

	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			callIndex++
			if callIndex == 1 {
				return adapter1
			}
			return adapter2
		},
	}

	ctx := context.Background()
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)

	config := domain.DBConnectionConfig{Host: "localhost", Port: 5432, User: "u", Password: "p"}
	configBytes, _ := json.Marshal(config)
	encrypted, _ := security.Encrypt(string(configBytes), masterKey)

	session := &domain.Session{
		ID:                "close-session",
		UserID:            "user",
		Role:              "admin",
		CSRFToken:         "csrf",
		DBType:            "pgsql",
		EncryptedDBConfig: encrypted,
		ExpiresAt:         time.Now().Add(24 * time.Hour),
	}
	store.CreateSession(ctx, session)

	svc := NewDBService(store, factories)

	// Create adapters for the session
	_, _ = svc.GetAdapter(ctx, "close-session", "")
	_, _ = svc.GetAdapter(ctx, "close-session", "mydb")

	// Close all session connections
	svc.CloseSessionConnection("close-session")

	// Both adapters should have been disconnected
	if adapter1.Connected {
		t.Error("Adapter1 should have been disconnected after CloseSessionConnection")
	}
	if adapter2.Connected {
		t.Error("Adapter2 should have been disconnected after CloseSessionConnection")
	}
}

func TestDBService_ConcurrentGetAdapter(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			return &MockDBAdapter{}
		},
	}

	ctx := context.Background()
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)

	config := domain.DBConnectionConfig{Host: "localhost", Port: 5432, User: "u", Password: "p"}
	configBytes, _ := json.Marshal(config)
	encrypted, _ := security.Encrypt(string(configBytes), masterKey)

	session := &domain.Session{
		ID:                "concurrent-session",
		UserID:            "user",
		Role:              "admin",
		CSRFToken:         "csrf",
		DBType:            "pgsql",
		EncryptedDBConfig: encrypted,
		ExpiresAt:         time.Now().Add(24 * time.Hour),
	}
	store.CreateSession(ctx, session)

	svc := NewDBService(store, factories)

	// Concurrent access should not panic or race
	var wg sync.WaitGroup
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			db := fmt.Sprintf("db%d", idx%5) // 5 different databases
			_, err := svc.GetAdapter(ctx, "concurrent-session", db)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent GetAdapter failed: %v", err)
	}
}

func TestDBService_ExpiredSession(t *testing.T) {
	store := NewMockSessionStore()
	factories := map[string]func() ports.IDBAdapter{}

	ctx := context.Background()
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)

	// Create an already-expired session
	session := &domain.Session{
		ID:                "expired-session",
		UserID:            "user",
		Role:              "admin",
		CSRFToken:         "csrf",
		DBType:            "pgsql",
		EncryptedDBConfig: "doesntmatter",
		ExpiresAt:         time.Now().Add(-1 * time.Hour), // expired
	}
	store.CreateSession(ctx, session)

	svc := NewDBService(store, factories)

	_, err := svc.GetAdapter(ctx, "expired-session", "")
	if err == nil {
		t.Error("Expected error for expired session, got nil")
	}
}
