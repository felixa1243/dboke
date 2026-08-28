package middleware

import (
	"context"
	"net/http"
	"log/slog"
	"dboke-api/internal/core/ports"
	"dboke-api/internal/pkg/contextkeys"
	"dboke-api/internal/pkg/response"
)

// AuthMiddleware enforces session and CSRF security
func AuthMiddleware(sessionStore ports.ISessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extract HTTP-Only Cookie
			cookie, err := r.Cookie("dboke_session")
			if err != nil {
				response.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized: No session cookie")
				return
			}

			// 2. Validate Session (Prevents forged/expired cookies)
			session, err := sessionStore.GetSession(r.Context(), cookie.Value)
			if err != nil || session == nil {
				response.WriteError(w, http.StatusUnauthorized, "INVALID_SESSION", "Invalid or expired session")
				return
			}

			// 3. Strict CSRF Protection Check
			if r.Method != http.MethodGet && r.Method != http.MethodOptions {
				csrfHeader := r.Header.Get("X-CSRF-Token")
				if csrfHeader == "" || csrfHeader != session.CSRFToken { 
					slog.Warn("CSRF Validation Failed", 
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
						slog.String("received_csrf", csrfHeader),
						slog.String("expected_csrf", session.CSRFToken),
					)
					response.WriteError(w, http.StatusForbidden, "CSRF_FAILED", "Missing or invalid CSRF token")
					return
				}
			}

			// 4. Inject context (User ID, RBAC Role, Session ID)
			ctx := context.WithValue(r.Context(), contextkeys.UserIDKey, session.UserID)
			ctx = context.WithValue(ctx, contextkeys.RoleKey, session.Role)
			ctx = context.WithValue(ctx, contextkeys.SessionIDKey, session.ID)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
