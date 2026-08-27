package ports

import (
	"context"
	"dboke-api/internal/core/domain"
	"time"
)

// ISessionStore defines the contract for session management (e.g., Redis)
type ISessionStore interface {
	CreateSession(ctx context.Context, session *domain.Session) error
	GetSession(ctx context.Context, sessionID string) (*domain.Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	ExtendSession(ctx context.Context, sessionID string, newExpiry time.Time) error
}
