package middleware

import (
	"net/http"
	"strings"

	"github.com/parsel-email/lib-go/logger"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldSkipLogging(r.URL.Path) {
			logger.Info(r.Context(), "Request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				// "user_agent", r.UserAgent(),
			)
		}
		next.ServeHTTP(w, r)
	})
}

func shouldSkipLogging(path string) bool {
	if path == "/favicon.png" || path == "/metrics" {
		return true
	}
	return strings.HasPrefix(path, "/_app/")
}
