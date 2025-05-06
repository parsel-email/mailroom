// This file implements audit logging for authentication events

package middleware

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/parsel-email/lib-go/logger"
	"github.com/parsel-email/lib-go/metrics"
	"github.com/parsel-email/mailroom/internal/auth"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AuditLogMiddleware captures and logs authentication events
func AuditLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := uuid.New().String()

		// Add request ID to context for consistent logging
		ctx := logger.WithRequestID(r.Context(), requestID)

		// Get current span from context if it exists
		span := trace.SpanFromContext(ctx)

		// Add request ID as span attribute if we have an active span
		if span.SpanContext().IsValid() {
			span.SetAttributes(attribute.String("request_id", requestID))
			// Add trace ID to logger context
			ctx = logger.WithTraceID(ctx, span.SpanContext().TraceID().String())
		}

		r = r.WithContext(ctx)

		// Create a response wrapper to capture the status code
		rw := newResponseWriter(w)

		// Extract authentication info before the request is processed
		authInfo := extractAuthInfo(r)

		// Track auth type in metrics
		if authInfo.authType != "none" && isAuthEndpoint(r.URL.Path) {
			// Track active sessions for user auth
			if authInfo.authType == "jwt" && !authInfo.isService {
				metrics.ActiveSessions.Inc()
			}
		}

		// Add auth info to span if available
		if span.SpanContext().IsValid() && authInfo.authType != "none" {
			span.SetAttributes(
				attribute.String("auth.type", authInfo.authType),
				attribute.Bool("auth.is_service", authInfo.isService),
			)

			if authInfo.userID != "" {
				span.SetAttributes(attribute.String("auth.user_id", authInfo.userID))
			}

			if authInfo.serviceName != "" {
				span.SetAttributes(attribute.String("auth.service_name", authInfo.serviceName))
			}
		}

		// Process the request
		next.ServeHTTP(rw, r)

		// Determine if this was an auth-related endpoint
		if isAuthEndpoint(r.URL.Path) {
			duration := time.Since(start)
			success := rw.statusCode >= 200 && rw.statusCode < 300

			// Track auth metrics
			if authInfo.authType != "none" {
				// On logout, decrement active sessions
				if r.URL.Path == "/api/v1/logout" && authInfo.authType == "jwt" && !authInfo.isService && success {
					metrics.ActiveSessions.Dec()
				}

				// // Track auth events by path, status and auth type
				// statusLabel := "success"
				// if !success {
				// 	statusLabel = "failure"
				// }

				// // Use a more specific label for auth type
				// authTypeLabel := authInfo.authType
				// if authInfo.isService {
				// 	authTypeLabel = authInfo.authType + "_service"
				// }

				// // Track as an auth request
				// metrics.AuthRequests.WithLabelValues(authTypeLabel, statusLabel).Inc()
			}

			// If we have an active span, add more attributes
			if span.SpanContext().IsValid() {
				span.SetAttributes(
					attribute.Bool("auth.success", success),
					attribute.Int("http.status_code", rw.statusCode),
				)

				// Add event for authentication
				span.AddEvent("auth_event", trace.WithAttributes(
					attribute.String("auth.endpoint", r.URL.Path),
					attribute.Bool("auth.success", success),
				))
			}

			// Log the authentication event
			logger.Info(ctx, "Auth event",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
				"success", success,
				"client_ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"auth_type", authInfo.authType,
				"user_id", authInfo.userID,
				"service_name", authInfo.serviceName,
				"is_service", authInfo.isService,
			)

			// If enabled, create an audit log in the database
			// This would be a good place to add tracing to your database calls
			// by passing the context with trace span to your database layer
		}
	})
}

// Custom response writer to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK} // Default to 200 OK
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// authInfo holds information about the authentication
type authInfo struct {
	authType    string // "jwt", "api_key", "none"
	userID      string
	serviceName string
	isService   bool
}

// extractAuthInfo extracts authentication information from the request
func extractAuthInfo(r *http.Request) authInfo {
	info := authInfo{
		authType: "none",
	}

	// Check for Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return info
	}

	// Check if it's a Bearer token
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token := authHeader[7:]

		// Check if it's an API key
		if len(token) > 3 && token[:3] == "pk_" {
			info.authType = "api_key"
			info.isService = true
			// In a real implementation, you'd extract the service name
			info.serviceName = "unknown-service"
			return info
		}

		// Try to parse as JWT
		parsedToken, err := auth.ParseToken(token)
		if err != nil {
			return info
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			return info
		}

		info.authType = "jwt"

		// Check if it's a service token
		if isService, ok := claims["isService"].(bool); ok && isService {
			info.isService = true
			if serviceName, ok := claims["sub"].(string); ok {
				info.serviceName = serviceName
			}
		} else {
			// Regular user token
			if id, ok := claims["ID"].(string); ok {
				info.userID = id
			}
		}
	}

	return info
}

// isAuthEndpoint determines if an endpoint is authentication-related
func isAuthEndpoint(path string) bool {
	authEndpoints := map[string]bool{
		"/auth/google":           true,
		"/auth/google/callback":  true,
		"/api/v1/token/refresh":  true,
		"/api/v1/logout":         true,
		"/api/v1/renew":          true,
		"/api/v1/apikeys":        true,
		"/api/v1/apikeys/create": true,
		"/api/v1/apikeys/revoke": true,
	}

	return authEndpoints[path]
}
