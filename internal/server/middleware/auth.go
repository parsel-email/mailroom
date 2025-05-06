package middleware

import (
	"net/http"
	"strings"

	"github.com/parsel-email/lib-go/logger"
	"github.com/parsel-email/mailroom/internal/auth"
)

var UnprotectedAPIRoutes = map[string]bool{
	"/api/v1/renew":  true,
	"/api/v1/health": true,
	"/api/v1/logout": true, // Allow logout without a valid token
}

func AuthenticatedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if the route is an API route and is not in the unprotected list, check for JWT
		if (strings.Contains(r.URL.Path, "/api/") && !UnprotectedAPIRoutes[r.URL.Path]) && !auth.ValidateJWT(r.Header.Get("Authorization")) {
			logger.Warn(r.Context(), "Unauthorized access attempt",
				"path", r.URL.Path,
				"method", r.Method,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
