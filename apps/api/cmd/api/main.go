package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"dboke-api/internal/core/domain"
	"dboke-api/internal/core/ports"
	"dboke-api/internal/core/services"
	"dboke-api/internal/delivery/http/router"
	"dboke-api/internal/infrastructure/database"
	"dboke-api/internal/pkg/logger"

	"github.com/joho/godotenv"
)

// InMemorySessionStore temporarily stores sessions in memory for development
type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*domain.Session
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{sessions: make(map[string]*domain.Session)}
}

func (m *InMemorySessionStore) CreateSession(ctx context.Context, session *domain.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.ID] = session
	return nil
}

func (m *InMemorySessionStore) GetSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if session, exists := m.sessions[sessionID]; exists {
		if time.Now().After(session.ExpiresAt) {
			return nil, fmt.Errorf("session expired")
		}
		return session, nil
	}
	return nil, fmt.Errorf("session not found")
}

func (m *InMemorySessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
	return nil
}

func (m *InMemorySessionStore) ExtendSession(ctx context.Context, sessionID string, newExpiry time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if session, exists := m.sessions[sessionID]; exists {
		session.ExpiresAt = newExpiry
	}
	return nil
}


func main() {
	// Attempt to load .env from common locations
	err1 := godotenv.Load("../../.env")
	err2 := godotenv.Load(".env")
	
	if err1 != nil && err2 != nil {
		slog.Warn("No .env file found in common locations, falling back to system environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	env := os.Getenv("ENV")
	logger.Init(env)

	// 1. Initialize Infrastructure (e.g. Session Store, Internal DB)
	sessionStore := NewInMemorySessionStore()

	// Initialize Database Adapters
	adapterFactories := map[string]func() ports.IDBAdapter{
		"pgsql": func() ports.IDBAdapter { return database.NewPostgresAdapter() },
		// Future adapters go here
	}

	// 2. Initialize Core Services
	authService := services.NewAuthService(sessionStore, adapterFactories)
	dbService := services.NewDBService(sessionStore, adapterFactories)

	// 3. Initialize the API Gateway Router
	r := router.NewRouter(sessionStore, authService, dbService)

	// 3. Start the HTTP Server
	slog.Info("Starting Dboke API server securely", slog.String("port", port))
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: r,
	}

	// Setup graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server startup failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()
	
	<-stop
	slog.Info("Shutting down API server gracefully...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("error", err.Error()))
	}
	
	slog.Info("Server stopped")
}
