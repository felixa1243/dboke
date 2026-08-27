package middleware

import (
	"context"
	"net/http"
	"dboke-api/internal/core/ports"
)

// AuthMiddleware enforces session and CSRF security
func AuthMiddleware(sessionStore ports.ISessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extract HTTP-Only Cookie
			cookie, err := r.Cookie("dboke_session")
			if err != nil {
				http.Error(w, "Unauthorized: No session cookie", http.StatusUnauthorized)
				return
			}

			// 2. Validate Session (Prevents forged/expired cookies)
			session, err := sessionStore.GetSession(r.Context(), cookie.Value)
			if err != nil || session == nil {
				http.Error(w, "Invalid or expired session", http.StatusUnauthorized)
				return
			}

			// 3. Strict CSRF Protection Check
			if r.Method != http.MethodGet && r.Method != http.MethodOptions {
				csrfHeader := r.Header.Get("X-CSRF-Token")
				if csrfHeader == "" || csrfHeader != session.CSRFToken { 
					http.Error(w, "Missing or invalid CSRF token", http.StatusForbidden)
					return
				}
			}

			// 4. Inject context (User ID, RBAC Role, Session ID)
			ctx := context.WithValue(r.Context(), "user_id", session.UserID)
			ctx = context.WithValue(ctx, "role", session.Role)
			ctx = context.WithValue(ctx, "session_id", session.ID)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
