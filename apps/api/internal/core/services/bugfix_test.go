package services

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"
	"dboke-api/internal/infrastructure/security"
)

func TestIsValidPluginID(t *testing.T) {
	valid := []string{"data-seeder", "plugin_1", "A"}
	invalid := []string{"", ".", "..", "../x", "a/b", `a\b`, "_hidden", "-x", "a b", "a.b"}
	for _, id := range valid {
		if !IsValidPluginID(id) {
			t.Errorf("IsValidPluginID(%q) = false, want true", id)
		}
	}
	for _, id := range invalid {
		if IsValidPluginID(id) {
			t.Errorf("IsValidPluginID(%q) = true, want false", id)
		}
	}
}

func TestPluginManager_LoadPlugin_RejectsTraversal(t *testing.T) {
	m := NewPluginManager()
	if err := m.LoadPlugin("../../bin"); err == nil {
		t.Fatal("LoadPlugin accepted a path-traversal plugin id")
	}
}

func createTestSession(t *testing.T, store ports.ISessionStore, id string, expiresAt time.Time) {
	t.Helper()
	masterKey := "01234567890123456789012345678901"
	t.Setenv("DBOKE_MASTER_KEY", masterKey)
	configBytes, _ := json.Marshal(domain.DBConnectionConfig{Host: "localhost", Port: 5432, User: "u", Password: "p"})
	encrypted, _ := security.Encrypt(string(configBytes), masterKey)
	store.CreateSession(context.Background(), &domain.Session{
		ID:                id,
		DBType:            "pgsql",
		EncryptedDBConfig: encrypted,
		ExpiresAt:         expiresAt,
	})
}

func TestDBService_CloseExpiredConnections(t *testing.T) {
	store := NewMockSessionStore()
	ctx := context.Background()
	var adapters []*MockDBAdapter
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			a := &MockDBAdapter{}
			adapters = append(adapters, a)
			return a
		},
	}
	svc := NewDBService(store, factories)

	createTestSession(t, store, "aaaa", time.Now().Add(time.Hour))
	createTestSession(t, store, "bbbb", time.Now().Add(time.Hour))
	if _, err := svc.GetAdapter(ctx, "aaaa", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetAdapter(ctx, "bbbb", "my_db"); err != nil {
		t.Fatal(err)
	}

	// Session bbbb goes away; its connection must be closed, aaaa's kept.
	store.DeleteSession(ctx, "bbbb")
	svc.CloseExpiredConnections(ctx)

	if !adapters[0].Connected {
		t.Error("connection for live session was closed")
	}
	if adapters[1].Connected {
		t.Error("connection for deleted session was not closed")
	}
	svc.mu.RLock()
	defer svc.mu.RUnlock()
	if _, ok := svc.activeAdapters["bbbb_my_db"]; ok {
		t.Error("expired session's adapter still cached")
	}
}

// blockingAdapter holds Connect until released so two requests both miss the cache.
type blockingAdapter struct {
	MockDBAdapter
	release <-chan struct{}
	mu      sync.Mutex
}

func (b *blockingAdapter) Connect(ctx context.Context, c domain.DBConnectionConfig) error {
	<-b.release
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Connected = true
	return nil
}

func (b *blockingAdapter) Disconnect() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Connected = false
	return nil
}

func (b *blockingAdapter) Ping(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.MockDBAdapter.Ping(ctx)
}

func TestDBService_GetAdapter_ConcurrentMissDoesNotLeak(t *testing.T) {
	store := NewMockSessionStore()
	createTestSession(t, store, "cccc", time.Now().Add(time.Hour))

	release := make(chan struct{})
	var mu sync.Mutex
	var created []*blockingAdapter
	factories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter {
			a := &blockingAdapter{release: release}
			mu.Lock()
			created = append(created, a)
			mu.Unlock()
			return a
		},
	}
	svc := NewDBService(store, factories)

	var wg sync.WaitGroup
	results := make([]ports.IDBAdapter, 2)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			a, err := svc.GetAdapter(context.Background(), "cccc", "")
			if err != nil {
				t.Error(err)
			}
			results[i] = a
		}(i)
	}
	// Wait until both requests are inside Connect, then let them finish.
	for {
		mu.Lock()
		n := len(created)
		mu.Unlock()
		if n == 2 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	close(release)
	wg.Wait()

	if results[0] != results[1] {
		t.Error("concurrent callers received different adapters")
	}
	open := 0
	for _, a := range created {
		if a.Ping(context.Background()) == nil {
			open++
		}
	}
	if open != 1 {
		t.Errorf("%d adapters left connected, want 1 (the loser must be disconnected)", open)
	}
}
