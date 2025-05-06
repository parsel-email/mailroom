// Package middleware provides HTTP middleware components
package middleware

import (
	"context"
	"net/http"

	"github.com/parsel-email/lib-go/logger"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware wraps handlers with OpenTelemetry tracing
func TracingMiddleware(next http.Handler) http.Handler {
	// Create a custom handler that adds additional attributes to spans
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Standard attributes collected by otelhttp
		handler := otelhttp.NewHandler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Extract request ID from logger context if available and add as attribute
				if requestID := logger.GetRequestID(r.Context()); requestID != "" {
					span := trace.SpanFromContext(r.Context())
					span.SetAttributes(attribute.String("request_id", requestID))
				}

				// Continue processing the request
				next.ServeHTTP(w, r)
			}),
			r.URL.Path,
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return r.Method + " " + r.URL.Path
			}),
		)

		handler.ServeHTTP(w, r)
	})
}

// AddTraceIDToContext adds the trace ID to the request context
func AddTraceIDToContext(ctx context.Context) context.Context {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		ctx = logger.WithTraceID(ctx, span.SpanContext().TraceID().String())
	}
	return ctx
}

// ExtractBaggageItems extracts baggage items from the context and returns them as a map
func ExtractBaggageItems(ctx context.Context) map[string]string {
	items := make(map[string]string)
	bags := baggage.FromContext(ctx)
	for _, member := range bags.Members() {
		items[member.Key()] = member.Value()
	}
	return items
}
