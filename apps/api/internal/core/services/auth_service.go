package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"
	"dboke-api/internal/infrastructure/security"
)

type AuthService struct {
	sessionStore    ports.ISessionStore
	adapterFactories map[string]func() ports.IDBAdapter
}

func NewAuthService(sessionStore ports.ISessionStore, adapterFactories map[string]func() ports.IDBAdapter) *AuthService {
	return &AuthService{
		sessionStore:     sessionStore,
		adapterFactories: adapterFactories,
	}
}

func generateRandomString(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *AuthService) Authenticate(ctx context.Context, dbType string, config domain.DBConnectionConfig) (*domain.Session, error) {
	factory, exists := s.adapterFactories[dbType]
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	adapter := factory()

	// Actually validate the DB connection (Task 5)
	if err := adapter.Connect(ctx, config); err != nil {
		return nil, fmt.Errorf("invalid credentials or connection error: %w", err)
	}
	defer adapter.Disconnect()

	sessionID, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session id: %w", err)
	}

	csrfToken, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate csrf token: %w", err)
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal db config: %w", err)
	}

	masterKey := os.Getenv("DBOKE_MASTER_KEY")
	encryptedConfig, err := security.Encrypt(string(configBytes), masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt db config: %w", err)
	}

	session := &domain.Session{
		ID:                sessionID,
		UserID:            config.User, // Using DB user as Dboke user
		Role:              "admin",
		CSRFToken:         csrfToken,
		DBType:            dbType,
		EncryptedDBConfig: encryptedConfig,
		ExpiresAt:         time.Now().Add(24 * time.Hour),
	}

	if err := s.sessionStore.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return session, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.sessionStore.DeleteSession(ctx, sessionID)
}
