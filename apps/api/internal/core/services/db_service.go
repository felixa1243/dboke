package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"
	"dboke-api/internal/infrastructure/security"
)

type DBService struct {
	sessionStore     ports.ISessionStore
	adapterFactories map[string]func() ports.IDBAdapter
	activeAdapters   map[string]ports.IDBAdapter // sessionID -> adapter
	mu               sync.RWMutex
}

func NewDBService(sessionStore ports.ISessionStore, adapterFactories map[string]func() ports.IDBAdapter) *DBService {
	return &DBService{
		sessionStore:     sessionStore,
		adapterFactories: adapterFactories,
		activeAdapters:   make(map[string]ports.IDBAdapter),
	}
}

func (s *DBService) GetAdapter(ctx context.Context, sessionID, targetDatabase string) (ports.IDBAdapter, error) {
	cacheKey := sessionID
	if targetDatabase != "" {
		cacheKey = fmt.Sprintf("%s_%s", sessionID, targetDatabase)
	}

	s.mu.RLock()
	adapter, exists := s.activeAdapters[cacheKey]
	s.mu.RUnlock()
	if exists {
		// Ping to ensure connection is still alive
		if err := adapter.Ping(ctx); err == nil {
			return adapter, nil
		}
		// If ping fails, force reconnect
		s.mu.Lock()
		adapter.Disconnect()
		delete(s.activeAdapters, cacheKey)
		s.mu.Unlock()
	}

	session, err := s.sessionStore.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	masterKey := os.Getenv("DBOKE_MASTER_KEY")
	decryptedConfigStr, err := security.Decrypt(session.EncryptedDBConfig, masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt credentials: %w", err)
	}

	var config domain.DBConnectionConfig
	if err := json.Unmarshal([]byte(decryptedConfigStr), &config); err != nil {
		return nil, fmt.Errorf("corrupted session db config")
	}

	if targetDatabase != "" {
		config.Database = targetDatabase
	}

	factory, exists := s.adapterFactories[session.DBType]
	if !exists {
		return nil, fmt.Errorf("unsupported db type in session: %s", session.DBType)
	}

	newAdapter := factory()
	if err := newAdapter.Connect(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	s.mu.Lock()
	s.activeAdapters[cacheKey] = newAdapter
	s.mu.Unlock()

	return newAdapter, nil
}

func (s *DBService) CloseSessionConnection(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Close all connections for this session
	for key, adapter := range s.activeAdapters {
		if key == sessionID || (len(key) > len(sessionID) && key[:len(sessionID)+1] == sessionID+"_") {
			adapter.Disconnect()
			delete(s.activeAdapters, key)
		}
	}
}
