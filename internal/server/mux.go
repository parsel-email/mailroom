package server

import (
	"encoding/json"
	"net/http"

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
