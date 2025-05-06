// This file implements rate limiting for the auth service

package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/parsel-email/lib-go/logger"
	"github.com/parsel-email/lib-go/metrics"
	"github.com/parsel-email/mailroom/internal/auth"
)

// We'll add a counter for rate limit hits
type RateLimiter struct {
	mu             sync.Mutex
	ipLimits       map[string]*ClientLimit
	serviceLimits  map[string]*ClientLimit
	ipMaxRequests  int
	ipWindow       time.Duration
	svcMaxRequests int
	svcWindow      time.Duration
}

type ClientLimit struct {
	requests    int
	windowStart time.Time
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(ipMaxRequests int, ipWindow time.Duration, svcMaxRequests int, svcWindow time.Duration) *RateLimiter {
	return &RateLimiter{
		ipLimits:       make(map[string]*ClientLimit),
		serviceLimits:  make(map[string]*ClientLimit),
		ipMaxRequests:  ipMaxRequests,
		ipWindow:       ipWindow,
		svcMaxRequests: svcMaxRequests,
		svcWindow:      svcWindow,
	}
}

// Allow checks if a request should be allowed based on rate limits
func (rl *RateLimiter) Allow(identifier string, isService bool) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Determine which limits map to use
	limitsMap := rl.ipLimits
	maxRequests := rl.ipMaxRequests
	window := rl.ipWindow
	limiterType := "ip"

	if isService {
		limitsMap = rl.serviceLimits
		maxRequests = rl.svcMaxRequests
		window = rl.svcWindow
		limiterType = "service"
	}

	// Get or create limit for this identifier
	limit, exists := limitsMap[identifier]
	if !exists || now.Sub(limit.windowStart) > window {
		// First request or window expired, reset counter
		limitsMap[identifier] = &ClientLimit{
			requests:    1,
			windowStart: now,
		}
		return true
	}

	// Check if limit exceeded
	if limit.requests >= maxRequests {
		// Track rate limit exceeded metric
		metrics.Errors.WithLabelValues("rate_limit_exceeded_" + limiterType).Inc()
		return false
	}

	// Increment counter
	limit.requests++
	return true
}

// Cleanup removes old entries from the rate limiter
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Clean up IP limits
	for ip, limit := range rl.ipLimits {
		if now.Sub(limit.windowStart) > rl.ipWindow {
			delete(rl.ipLimits, ip)
		}
	}

	// Clean up service limits
	for service, limit := range rl.serviceLimits {
		if now.Sub(limit.windowStart) > rl.svcWindow {
			delete(rl.serviceLimits, service)
		}
	}
}

// Create global rate limiter with different limits for IPs and services
var rateLimiter = NewRateLimiter(
	100,         // IP: 100 requests per minute
	time.Minute, // IP: 1 minute window
	1000,        // Service: 1000 requests per minute
	time.Minute, // Service: 1 minute window
)

// Start a background goroutine to clean up the rate limiter
func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			rateLimiter.Cleanup()
		}
	}()
}

// RateLimitMiddleware implements rate limiting for API endpoints
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limiting for specified paths
		if shouldSkipRateLimiting(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Default client identifier (IP address)
		identifier := r.RemoteAddr
		isService := false
		clientType := "user"

		// Check if this is a service-to-service request based on Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Determine the type of token (JWT or API key)
			if strings.HasPrefix(token, "pk_") {
				// Handle API key
				isService = true
				clientType = "api_key"
				// Use a temporary identifier until we can validate the key
				// The actual service name would be determined during full validation
				identifier = "api-key-" + token[:16]
			} else {
				// Try to parse as JWT
				parsedToken, err := auth.ParseToken(token)
				if err == nil {
					// Extract claims from the token
					if mapClaims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
						// Check if it's a service token
						if isServiceToken, ok := mapClaims["isService"].(bool); ok && isServiceToken {
							if serviceName, ok := mapClaims["sub"].(string); ok && serviceName != "" {
								// Use the service name as identifier
								identifier = serviceName
								isService = true
								clientType = "service"
							}
						}
					}
				}
			}
		}

		// Apply rate limiting based on identifier and service type
		if !rateLimiter.Allow(identifier, isService) {
			logger.Warn(r.Context(), "Rate limit exceeded",
				"identifier", identifier,
				"isService", isService,
				"clientType", clientType,
				"path", r.URL.Path,
				"method", r.Method)

			// Track rate limit exceeded in metrics with more details
			metrics.Errors.WithLabelValues("rate_limit_exceeded").Inc()

			// Set headers per RFC 6585 for rate limiting
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// Continue with the request
		next.ServeHTTP(w, r)
	})
}

// shouldSkipRateLimiting determines if a path should skip rate limiting
func shouldSkipRateLimiting(path string) bool {
	// Skip rate limiting for health checks
	return path == "/api/v1/health"
}
