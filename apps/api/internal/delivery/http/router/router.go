package router

import (
	"net/http"
	"os"
	"dboke-api/internal/core/ports"
	customMiddleware "dboke-api/internal/delivery/http/middleware"
	"dboke-api/internal/delivery/http/handler"
	"dboke-api/internal/pkg/logger"
	"dboke-api/internal/core/services"
	"dboke-api/internal/pkg/contextkeys"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter initializes the HTTP router with middleware and route definitions
func NewRouter(sessionStore ports.ISessionStore, authService *services.AuthService, dbService *services.DBService, pluginManager *services.PluginManager) *chi.Mux {
	r := chi.NewRouter()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	dbHandler := handler.NewDBHandler(dbService)
	pluginHandler := handler.NewPluginHandler(pluginManager, dbService)

	// 1. Core structural middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(logger.RequestLogger)
	r.Use(middleware.Recoverer)

	// 2. Network Security: Strict CORS configuration for Next.js frontend
	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{allowedOrigin}, // Comes from ENV
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true, // Crucial for accepting HTTP-Only cookies
		MaxAge:           300,
	}))

	// 3. Public Routes Group (Health checks, Login)
	r.Group(func(public chi.Router) {
		public.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Dboke API Server is operational"))
		})
		
		// Setup auth routes with rate limiting
		public.With(customMiddleware.RateLimit).Post("/api/v1/auth/login", authHandler.Login)
		public.Post("/api/v1/auth/logout", authHandler.Logout)
	})

	// 4. Protected Routes Group (Database Management, Raw SQL)
	r.Group(func(protected chi.Router) {
		// Enforce Session & CSRF Security via our custom middleware
		protected.Use(customMiddleware.AuthMiddleware(sessionStore))

		protected.Get("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(contextkeys.UserIDKey).(string)
			role := r.Context().Value(contextkeys.RoleKey).(string)
			
			response := "Hello Secured User! ID: " + userID + ", Role: " + role
			w.Write([]byte(response))
		})

		protected.Get("/api/v1/databases", dbHandler.GetDatabases)
		protected.Get("/api/v1/databases/{database}/tables", dbHandler.GetTables)
		protected.Get("/api/v1/databases/{database}/tables/{table}/columns", dbHandler.GetTableSchema)
		protected.Post("/api/v1/databases/{database}/query", dbHandler.ExecuteQuery)

		// Plugin Management
		protected.Post("/api/v1/plugins/upload", pluginHandler.UploadPlugin)
		protected.Get("/api/v1/plugins", pluginHandler.ListPlugins)
		protected.Post("/api/v1/plugins/{id}/toggle", pluginHandler.TogglePlugin)
		protected.Delete("/api/v1/plugins/{id}", pluginHandler.DeletePlugin)
		
		// Tasks
		protected.Get("/api/v1/tasks/{id}", pluginHandler.GetTaskStatus)
		
		// Plugin Execution
		protected.Post("/api/v1/databases/{database}/plugins/{id}/execute", pluginHandler.ExecutePluginQuery)

		// Example hooks for future handlers:
		// protected.Post("/api/v1/db/connect", connectDBHandler)
		// protected.Post("/api/v1/db/query", executeRawSQLHandler)
	})

	return r
}
