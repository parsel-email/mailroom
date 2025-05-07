package server

import (
	"encoding/json"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/parsel-email/lib-go/logger"
	"github.com/parsel-email/lib-go/metrics"
	"github.com/parsel-email/mailroom/internal/server/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Prometheus metrics
	mux.Handle("/metrics", promhttp.Handler())

	// API routes
	mux.HandleFunc("/api/v1/health", s.healthHandler)
	mux.HandleFunc("/api/v1/auth/health", s.authHealthHandler) // Dedicated auth health check

	// Wrap with middleware in the following order
	handler := middleware.TracingMiddleware(mux)          // Add tracing (first to capture all other middleware)
	handler = middleware.LoggingMiddleware(handler)       // Add logging
	handler = metrics.InstrumentHandler(handler)          // Add metrics instrumentation
	handler = middleware.AuditLogMiddleware(handler)      // Add audit logging
	handler = middleware.CorsMiddleware(handler)          // Add CORS
	handler = middleware.AuthenticatedMiddleware(handler) // Add authentication

	return handler
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		logger.Error(r.Context(), "Failed to marshal health check response", "error", err)
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		logger.Error(r.Context(), "Failed to write response", "error", err)
	}
}

// authHealthHandler provides a detailed health check for the authentication service
func (s *Server) authHealthHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "up",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
		"components": map[string]interface{}{
			"database": s.db.Health(),
			"google_oauth": map[string]string{
				"status": "up", // In a real implementation, you'd check the Google OAuth API
			},
			"microsoft_oauth": map[string]string{
				"status": "up", // In a real implementation, you'd check the Microsoft OAuth API
			},
		},
	}

	// Check database connectivity
	dbHealth := s.db.Health()
	if dbHealth["status"] != "up" {
		status["status"] = "degraded"
		logger.Warn(r.Context(), "Database health check indicates degraded service", "db_status", dbHealth["status"])
	}

	// Return the health check response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		logger.Error(r.Context(), "Failed to encode health status response", "error", err)
	}
}
