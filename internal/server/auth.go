package server

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"time"

// 	"github.com/parsel-email/auth/internal/auth"
// 	"github.com/parsel-email/auth/internal/google"
// 	"github.com/parsel-email/auth/internal/microsoft"
// 	"github.com/parsel-email/auth/lib-go/logger"
// 	"github.com/parsel-email/auth/lib-go/metrics"
// )

// func (s *Server) handleRenewToken(w http.ResponseWriter, r *http.Request) {
// 	// Check if user has JWT token in the header
// 	if !auth.ValidateJWT(r.Header.Get("Authorization")) {
// 		metrics.TokenOperations.WithLabelValues("renew", "unauthorized").Inc()
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	id, err := auth.GetIDFromJWT(r.Header.Get("Authorization"))
// 	if err != nil {
// 		metrics.TokenOperations.WithLabelValues("renew", "error").Inc()
// 		metrics.Errors.WithLabelValues("jwt_decode").Inc()
// 		logger.Error(r.Context(), "Failed to get ID from JWT", "error", err)
// 		http.Error(w, "Failed to get ID from JWT", http.StatusInternalServerError)
// 		return
// 	}

// 	user, err := s.db.GetUserByID(r.Context(), id)
// 	if err != nil {
// 		metrics.TokenOperations.WithLabelValues("renew", "error").Inc()
// 		metrics.Errors.WithLabelValues("database_get_user").Inc()
// 		logger.Error(r.Context(), "Failed to get user from database", "user_id", id, "error", err)
// 		http.Error(w, "Failed to get user from the database", http.StatusInternalServerError)
// 		return
// 	}

// 	// Generate JWT token
// 	jwtToken, err := auth.GenerateJWT(*user) // Implement this based on your JWT package
// 	if err != nil {
// 		metrics.TokenOperations.WithLabelValues("renew", "error").Inc()
// 		metrics.Errors.WithLabelValues("jwt_generation").Inc()
// 		logger.Error(r.Context(), "Failed to generate JWT", "user_id", user.ParselID, "error", err)
// 		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
// 		return
// 	}

// 	// Set JWT as a secure cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "jwt",
// 		Value:    jwtToken,
// 		Path:     "/",
// 		Secure:   true,
// 		HttpOnly: false, // Allow JavaScript access
// 		SameSite: http.SameSiteLaxMode,
// 		MaxAge:   3600 * 24, // 24 hours
// 	})

// 	w.Header().Set("Content-Type", "application/json")
// 	resp := map[string]string{"message": "renewed"}
// 	if err := json.NewEncoder(w).Encode(resp); err != nil {
// 		metrics.Errors.WithLabelValues("response_encode").Inc()
// 		logger.Error(r.Context(), "Failed to encode response", "error", err)
// 		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
// 		return
// 	}

// 	metrics.TokenOperations.WithLabelValues("renew", "success").Inc()
// }

// // handleGoogleCallback handles the callback from Google after user authentication
// func (s *Server) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
// 	// Get the callback code from the URL
// 	code := r.URL.Query().Get("code")
// 	if code == "" {
// 		metrics.AuthRequests.WithLabelValues("google", "error").Inc()
// 		metrics.Errors.WithLabelValues("missing_auth_code").Inc()
// 		logger.Warn(r.Context(), "Failed to get auth code from Google callback")
// 		http.Error(w, "Failed to get auth code", http.StatusBadRequest)
// 		return
// 	}

// 	// Convert the auth code to a token
// 	token, err := auth.ConvertAuthCodeToToken(r.Context(), auth.GoogleProvider, code)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("google", "error").Inc()
// 		metrics.Errors.WithLabelValues("code_to_token").Inc()
// 		logger.Error(r.Context(), "Failed to convert auth code to token", "error", err)
// 		http.Error(w, "Failed to convert auth code to token", http.StatusInternalServerError)
// 		return
// 	}

// 	client, err := auth.GetClientFromToken(r.Context(), auth.GoogleProvider, token)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("google", "error").Inc()
// 		metrics.Errors.WithLabelValues("get_client").Inc()
// 		logger.Error(r.Context(), "Failed to get HTTP client from token", "error", err)
// 		http.Error(w, "Failed to get client from token", http.StatusInternalServerError)
// 		return
// 	}

// 	// Get user details from the google API using the token
// 	googleUser, err := google.GetUserInfo(r.Context(), client)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("google", "error").Inc()
// 		metrics.Errors.WithLabelValues("get_user_info").Inc()
// 		logger.Error(r.Context(), "Failed to get user info from Google", "error", err)
// 		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
// 		return
// 	}

// 	// Check if the token has a refresh token
// 	if token.RefreshToken != "" {
// 		logger.Info(r.Context(), "Setting refresh token", "provider_id", googleUser.ProviderID)
// 		googleUser.RefreshToken = token.RefreshToken
// 	}

// 	// Track database operations
// 	metrics.DatabaseOperations.WithLabelValues("get_user", "attempt").Inc()
// 	parselUser, err := s.db.GetUserByProviderID(r.Context(), "google", googleUser.ProviderID)
// 	if err != nil {
// 		if err != sql.ErrNoRows {
// 			metrics.DatabaseOperations.WithLabelValues("get_user", "error").Inc()
// 			metrics.Errors.WithLabelValues("database_get_user").Inc()
// 			logger.Error(r.Context(), "Failed to get user from database", "provider_id", googleUser.ProviderID, "error", err)
// 			http.Error(w, "Failed to get user from the database", http.StatusInternalServerError)
// 			return
// 		} else {
// 			metrics.DatabaseOperations.WithLabelValues("get_user", "not_found").Inc()
// 			logger.Info(r.Context(), "User not found in the database, creating new user", "provider_id", googleUser.ProviderID)
// 		}
// 	} else {
// 		metrics.DatabaseOperations.WithLabelValues("get_user", "success").Inc()
// 	}

// 	if parselUser == nil {
// 		logger.Info(r.Context(), "Creating new user in the database", "provider_id", googleUser.ProviderID, "email", googleUser.Email)
// 		metrics.DatabaseOperations.WithLabelValues("create_user", "attempt").Inc()
// 		createdUser, err := s.db.CreateUser(r.Context(), googleUser)
// 		if err != nil {
// 			metrics.DatabaseOperations.WithLabelValues("create_user", "error").Inc()
// 			metrics.Errors.WithLabelValues("database_create_user").Inc()
// 			metrics.AuthRequests.WithLabelValues("google", "error").Inc()
// 			logger.Error(r.Context(), "Failed to save user to the database", "provider_id", googleUser.ProviderID, "error", err)
// 			http.Error(w, "Failed to save user to the database", http.StatusInternalServerError)
// 			return
// 		}
// 		metrics.DatabaseOperations.WithLabelValues("create_user", "success").Inc()

// 		parselUser = createdUser
// 		parselUser.RefreshToken = "" // Do not return the refresh token
// 	} else if token.RefreshToken != "" {
// 		logger.Info(r.Context(), "Updating refresh token on user callback", "user_id", parselUser.ParselID)
// 		parselUser.RefreshToken = token.RefreshToken
// 		metrics.DatabaseOperations.WithLabelValues("update_refresh_token", "attempt").Inc()
// 		err = s.db.UpdateUserRefreshToken(r.Context(), parselUser.ParselID.String(), token.RefreshToken)
// 		if err != nil {
// 			metrics.DatabaseOperations.WithLabelValues("update_refresh_token", "error").Inc()
// 			metrics.Errors.WithLabelValues("database_update_token").Inc()
// 			logger.Error(r.Context(), "Failed to update refresh token on user callback", "user_id", parselUser.ParselID, "error", err)
// 			http.Error(w, "Failed to update refresh token", http.StatusInternalServerError)
// 			return
// 		}
// 		metrics.DatabaseOperations.WithLabelValues("update_refresh_token", "success").Inc()
// 	}

// 	// Generate JWT token
// 	jwtToken, err := auth.GenerateJWT(*parselUser)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("google", "error").Inc()
// 		metrics.Errors.WithLabelValues("jwt_generation").Inc()
// 		logger.Error(r.Context(), "Failed to generate JWT", "user_id", parselUser.ParselID, "error", err)
// 		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
// 		return
// 	}

// 	// Set JWT as a secure cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "jwt",
// 		Value:    jwtToken,
// 		Path:     "/",
// 		Secure:   true,
// 		HttpOnly: true,
// 		SameSite: http.SameSiteStrictMode,
// 		MaxAge:   3600, // 1 hour
// 	})

// 	// Increment active sessions gauge
// 	metrics.ActiveSessions.Inc()
// 	metrics.AuthRequests.WithLabelValues("google", "success").Inc()

// 	// Redirect with JWT in URL fragment
// 	http.Redirect(w, r, "/home", http.StatusFound)
// }

// func (s *Server) handleGoogleAuth(w http.ResponseWriter, r *http.Request) {
// 	// Check if user has JWT token in the header
// 	if auth.ValidateJWT(r.Header.Get("Authorization")) {
// 		metrics.AuthRequests.WithLabelValues("google", "already_authenticated").Inc()
// 		logger.Info(r.Context(), "User already authenticated, redirecting to home")
// 		http.Redirect(w, r, "/home", http.StatusFound)
// 		return
// 	}

// 	// Get user ID from URL query to check if we should force consent
// 	// This ID would be present if redirected from a refresh token failure
// 	userID := r.URL.Query().Get("user_id")
// 	forceConsent := userID != ""

// 	if forceConsent {
// 		logger.Info(r.Context(), "Forcing consent screen for user", "user_id", userID)
// 	}

// 	url, err := auth.GetAuthCodeURL(auth.GoogleProvider, forceConsent)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("google", "error").Inc()
// 		metrics.Errors.WithLabelValues("auth_code_url").Inc()
// 		logger.Error(r.Context(), "Failed to get auth code URL", "error", err)
// 		http.Error(w, "Failed to initiate authentication", http.StatusInternalServerError)
// 		return
// 	}

// 	metrics.AuthRequests.WithLabelValues("google", "initiated").Inc()
// 	http.Redirect(w, r, url, http.StatusFound)
// }

// // handleRefreshToken handles refreshing an access token using a stored refresh token
// func (s *Server) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
// 	// Only allow POST requests
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Verify the user is authenticated
// 	jwtToken := r.Header.Get("Authorization")
// 	if !auth.ValidateJWT(jwtToken) {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Get the user ID from the JWT
// 	userID, err := auth.GetIDFromJWT(jwtToken)
// 	if err != nil {
// 		logger.Error(r.Context(), "Failed to get user ID from JWT", "error", err)
// 		http.Error(w, "Invalid token", http.StatusUnauthorized)
// 		return
// 	}

// 	// Get the provider from the JWT
// 	provider, err := auth.GetProviderFromJWT(jwtToken)
// 	if err != nil {
// 		logger.Error(r.Context(), "Failed to get provider from JWT", "error", err)
// 		http.Error(w, "Invalid token", http.StatusUnauthorized)
// 		return
// 	}

// 	// Get the user from the database
// 	user, err := s.db.GetUserByID(r.Context(), userID)
// 	if err != nil {
// 		logger.Error(r.Context(), "Failed to get user from database", "user_id", userID, "error", err)
// 		http.Error(w, "User not found", http.StatusNotFound)
// 		return
// 	}

// 	// Check if the user has a refresh token
// 	if user.RefreshToken == "" {
// 		logger.Error(r.Context(), "User has no refresh token", "user_id", userID)

// 		// Redirect to auth page with user_id to trigger consent screen
// 		redirectURL := fmt.Sprintf("/auth/%s?user_id=%s", provider, userID)

// 		// Return JSON response with redirect info instead of HTTP error
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusUnauthorized)
// 		json.NewEncoder(w).Encode(map[string]interface{}{
// 			"error":        "No refresh token available",
// 			"redirect_url": redirectURL,
// 		})
// 		return
// 	}

// 	// Convert string provider to AuthProvider type
// 	var authProvider auth.AuthProvider
// 	switch provider {
// 	case "google":
// 		authProvider = auth.GoogleProvider
// 	case "microsoft":
// 		authProvider = auth.MicrosoftProvider
// 	default:
// 		logger.Error(r.Context(), "Unknown provider", "provider", provider)
// 		http.Error(w, "Unknown provider", http.StatusBadRequest)
// 		return
// 	}

// 	// Get a new access token using the refresh token
// 	newToken, err := auth.GetNewAccessTokenFromRefreshToken(r.Context(), authProvider, user.RefreshToken)
// 	if err != nil {
// 		logger.Error(r.Context(), "Failed to get new access token", "user_id", userID, "error", err)

// 		// Clear the invalid refresh token in the database
// 		clearErr := s.db.UpdateUserRefreshToken(r.Context(), user.ParselID.String(), "")
// 		if clearErr != nil {
// 			logger.Error(r.Context(), "Failed to clear invalid refresh token", "user_id", userID, "error", clearErr)
// 		}

// 		// Redirect to auth page with user_id to trigger consent screen
// 		redirectURL := fmt.Sprintf("/auth/%s?user_id=%s", provider, userID)

// 		// Return JSON response with redirect info instead of HTTP error
// 		w.Header().Set("Content-Type", "application/json")
// 		w.WriteHeader(http.StatusUnauthorized)
// 		json.NewEncoder(w).Encode(map[string]interface{}{
// 			"error":        "Failed to refresh token, please re-authenticate",
// 			"redirect_url": redirectURL,
// 		})
// 		return
// 	}

// 	// If we get a new refresh token, update it in the database
// 	if newToken.RefreshToken != "" && newToken.RefreshToken != user.RefreshToken {
// 		user.RefreshToken = newToken.RefreshToken
// 		err = s.db.UpdateUserRefreshToken(r.Context(), user.ParselID.String(), newToken.RefreshToken)
// 		if err != nil {
// 			logger.Error(r.Context(), "Failed to update refresh token", "user_id", userID, "error", err)
// 			// This is not critical, so we can continue
// 		}
// 	}

// 	// Generate a new JWT
// 	newJWT, err := auth.GenerateJWT(*user)
// 	if err != nil {
// 		logger.Error(r.Context(), "Failed to generate JWT", "user_id", userID, "error", err)
// 		http.Error(w, "Failed to generate authentication token", http.StatusInternalServerError)
// 		return
// 	}

// 	// Set the new JWT as a cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "jwt",
// 		Value:    newJWT,
// 		Path:     "/",
// 		Secure:   true,
// 		HttpOnly: true,
// 		SameSite: http.SameSiteStrictMode,
// 		MaxAge:   3600, // 1 hour
// 	})

// 	// Return the new token in the response
// 	w.Header().Set("Content-Type", "application/json")
// 	response := map[string]interface{}{
// 		"token":        newJWT,
// 		"access_token": newToken.AccessToken,
// 		"expires_in":   newToken.Expiry.Unix() - time.Now().Unix(),
// 	}

// 	if err := json.NewEncoder(w).Encode(response); err != nil {
// 		logger.Error(r.Context(), "Failed to encode response", "error", err)
// 		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
// 		return
// 	}
// }

// // handleLogout handles user logout by revoking tokens
// func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
// 	// Only allow POST requests
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Get the JWT from the Authorization header
// 	jwtToken := r.Header.Get("Authorization")
// 	if !auth.ValidateJWT(jwtToken) {
// 		// If token is invalid, just clear cookies and return success
// 		http.SetCookie(w, &http.Cookie{
// 			Name:     "jwt",
// 			Value:    "",
// 			Path:     "/",
// 			MaxAge:   -1, // Delete cookie
// 			HttpOnly: true,
// 		})

// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
// 		return
// 	}

// 	// Get the user ID from the JWT
// 	userID, err := auth.GetIDFromJWT(jwtToken)
// 	if err == nil {
// 		// Get the provider from the JWT
// 		provider, err := auth.GetProviderFromJWT(jwtToken)
// 		if err != nil {
// 			logger.Error(r.Context(), "Failed to get provider from JWT", "error", err)
// 			// Continue anyway as we'll clear the token from our database
// 		}

// 		// Get the user from the database
// 		user, err := s.db.GetUserByID(r.Context(), userID)
// 		if err == nil && user.RefreshToken != "" {
// 			// Convert string provider to AuthProvider type
// 			var authProvider auth.AuthProvider
// 			switch provider {
// 			case "google":
// 				authProvider = auth.GoogleProvider
// 			case "microsoft":
// 				authProvider = auth.MicrosoftProvider
// 			default:
// 				logger.Error(r.Context(), "Unknown provider", "provider", provider)
// 				// Continue anyway as we'll clear the token from our database
// 			}

// 			// Revoke the refresh token with the provider
// 			err = auth.RevokeRefreshToken(r.Context(), authProvider, user.RefreshToken)
// 			if err != nil {
// 				logger.Error(r.Context(), "Failed to revoke refresh token", "user_id", userID, "error", err)
// 				// Continue anyway, as we'll clear the token from our database
// 			}

// 			// Clear the refresh token in the database
// 			err = s.db.UpdateUserRefreshToken(r.Context(), user.ParselID.String(), "")
// 			if err != nil {
// 				logger.Error(r.Context(), "Failed to clear refresh token in database", "user_id", userID, "error", err)
// 				// Continue anyway
// 			}
// 		}
// 	}

// 	// Clear the JWT cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "jwt",
// 		Value:    "",
// 		Path:     "/",
// 		MaxAge:   -1, // Delete cookie
// 		HttpOnly: true,
// 		Secure:   true,
// 		SameSite: http.SameSiteStrictMode,
// 	})

// 	// Return success message
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
// }

// // API Key management handlers

// // handleListAPIKeys lists all API keys for a service
// func (s *Server) handleListAPIKeys(w http.ResponseWriter, r *http.Request) {
// 	// Only allow GET requests
// 	if r.Method != http.MethodGet {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Validate the JWT token or API key
// 	if !auth.ValidateJWT(r.Header.Get("Authorization")) {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Get service name from query parameter
// 	serviceName := r.URL.Query().Get("service")
// 	if serviceName == "" {
// 		http.Error(w, "Service name is required", http.StatusBadRequest)
// 		return
// 	}

// 	// Get API keys from the database
// 	apiKeys, err := s.db.GetAPIKeysByService(r.Context(), serviceName)
// 	if err != nil {
// 		logger.Error(r.Context(), "Failed to get API keys", "service", serviceName, "error", err)
// 		http.Error(w, "Failed to retrieve API keys", http.StatusInternalServerError)
// 		return
// 	}

// 	// Don't return sensitive data
// 	type apiKeyResponse struct {
// 		ID          string    `json:"id"`
// 		ServiceName string    `json:"service_name"`
// 		Description string    `json:"description"`
// 		CreatedAt   time.Time `json:"created_at"`
// 		ExpiresAt   time.Time `json:"expires_at"`
// 	}

// 	var result []apiKeyResponse
// 	for _, key := range apiKeys {
// 		description := ""
// 		if key.Description != "" {
// 			description = key.Description
// 		}

// 		result = append(result, apiKeyResponse{
// 			ID:          key.ID.String(),
// 			ServiceName: key.ServiceName,
// 			Description: description,
// 			CreatedAt:   key.CreatedAt,
// 			ExpiresAt:   key.ExpiresAt,
// 		})
// 	}

// 	// Return the response
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"api_keys": result,
// 		"count":    len(result),
// 	})
// }

// // handleCreateAPIKey creates a new API key for a service
// func (s *Server) handleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
// 	// Only allow POST requests
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Validate the JWT token
// 	if !auth.ValidateJWT(r.Header.Get("Authorization")) {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Parse the request body
// 	var req struct {
// 		ServiceName string   `json:"service_name"`
// 		Description string   `json:"description"`
// 		ValidDays   int      `json:"valid_days"`
// 		Roles       []string `json:"roles"`
// 	}

// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil {
// 		logger.Warn(r.Context(), "Invalid request body", "error", err)
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Validate the request
// 	if req.ServiceName == "" {
// 		http.Error(w, "Service name is required", http.StatusBadRequest)
// 		return
// 	}

// 	// Generate an API key
// 	apiKey, err := auth.GenerateAPIKey(req.ServiceName, req.Description, req.ValidDays, req.Roles)
// 	if err != nil {
// 		logger.Error(r.Context(), "Failed to generate API key", "service", req.ServiceName, "error", err)
// 		http.Error(w, "Failed to generate API key", http.StatusInternalServerError)
// 		return
// 	}

// 	// Store the API key
// 	err = auth.StoreAPIKey(r.Context(), apiKey)
// 	if err != nil {
// 		logger.Error(r.Context(), "Failed to store API key", "service", req.ServiceName, "error", err)
// 		http.Error(w, "Failed to store API key", http.StatusInternalServerError)
// 		return
// 	}

// 	// Return the response - include the key, as this is the only time it will be visible
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"id":           apiKey.ID,
// 		"key":          apiKey.Key, // Only return this once
// 		"service_name": apiKey.ServiceName,
// 		"description":  apiKey.Description,
// 		"created_at":   apiKey.CreatedAt.Format(time.RFC3339),
// 		"expires_at":   apiKey.ExpiresAt.Format(time.RFC3339),
// 		"roles":        apiKey.Roles,
// 	})
// }

// // handleRevokeAPIKey revokes an API key
// func (s *Server) handleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
// 	// Only allow POST requests
// 	if r.Method != http.MethodPost {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	// Validate the JWT token
// 	if !auth.ValidateJWT(r.Header.Get("Authorization")) {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	// Parse the request body
// 	var req struct {
// 		ID string `json:"id"`
// 	}

// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil {
// 		logger.Warn(r.Context(), "Invalid request body", "error", err)
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	// Validate the request
// 	if req.ID == "" {
// 		http.Error(w, "API key ID is required", http.StatusBadRequest)
// 		return
// 	}

// 	// Revoke the API key
// 	logger.Info(r.Context(), "Revoking API key", "id", req.ID)

// 	// Return success response
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"message": "API key revoked successfully",
// 		"id":      req.ID,
// 	})
// }

// // handleMicrosoftCallback handles the callback from Microsoft after user authentication
// func (s *Server) handleMicrosoftCallback(w http.ResponseWriter, r *http.Request) {
// 	// Get the callback code from the URL
// 	code := r.URL.Query().Get("code")
// 	if code == "" {
// 		metrics.AuthRequests.WithLabelValues("microsoft", "error").Inc()
// 		metrics.Errors.WithLabelValues("missing_auth_code").Inc()
// 		logger.Warn(r.Context(), "Failed to get auth code from Microsoft callback")
// 		http.Error(w, "Failed to get auth code", http.StatusBadRequest)
// 		return
// 	}

// 	// Convert the auth code to a token
// 	token, err := auth.ConvertAuthCodeToToken(r.Context(), auth.MicrosoftProvider, code)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("microsoft", "error").Inc()
// 		metrics.Errors.WithLabelValues("code_to_token").Inc()
// 		logger.Error(r.Context(), "Failed to convert auth code to token", "error", err)
// 		http.Error(w, "Failed to convert auth code to token", http.StatusInternalServerError)
// 		return
// 	}

// 	client, err := auth.GetClientFromToken(r.Context(), auth.MicrosoftProvider, token)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("microsoft", "error").Inc()
// 		metrics.Errors.WithLabelValues("get_client").Inc()
// 		logger.Error(r.Context(), "Failed to get HTTP client from token", "error", err)
// 		http.Error(w, "Failed to get client from token", http.StatusInternalServerError)
// 		return
// 	}

// 	// Get user details from the Microsoft API using the token
// 	msUser, err := microsoft.GetUserInfo(r.Context(), client)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("microsoft", "error").Inc()
// 		metrics.Errors.WithLabelValues("get_user_info").Inc()
// 		logger.Error(r.Context(), "Failed to get user info from Microsoft", "error", err)
// 		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
// 		return
// 	}

// 	// Check if the token has a refresh token
// 	if token.RefreshToken != "" {
// 		logger.Info(r.Context(), "Setting refresh token", "provider_id", msUser.ProviderID)
// 		msUser.RefreshToken = token.RefreshToken
// 	}

// 	// Track database operations
// 	metrics.DatabaseOperations.WithLabelValues("get_user", "attempt").Inc()
// 	parselUser, err := s.db.GetUserByProviderID(r.Context(), "microsoft", msUser.ProviderID)
// 	if err != nil {
// 		if err != sql.ErrNoRows {
// 			metrics.DatabaseOperations.WithLabelValues("get_user", "error").Inc()
// 			metrics.Errors.WithLabelValues("database_get_user").Inc()
// 			logger.Error(r.Context(), "Failed to get user from database", "provider_id", msUser.ProviderID, "error", err)
// 			http.Error(w, "Failed to get user from the database", http.StatusInternalServerError)
// 			return
// 		} else {
// 			metrics.DatabaseOperations.WithLabelValues("get_user", "not_found").Inc()
// 			logger.Info(r.Context(), "User not found in the database, creating new user", "provider_id", msUser.ProviderID)
// 		}
// 	} else {
// 		metrics.DatabaseOperations.WithLabelValues("get_user", "success").Inc()
// 	}

// 	if parselUser == nil {
// 		logger.Info(r.Context(), "Creating new user in the database", "provider_id", msUser.ProviderID, "email", msUser.Email)
// 		metrics.DatabaseOperations.WithLabelValues("create_user", "attempt").Inc()
// 		createdUser, err := s.db.CreateUser(r.Context(), msUser)
// 		if err != nil {
// 			metrics.DatabaseOperations.WithLabelValues("create_user", "error").Inc()
// 			metrics.Errors.WithLabelValues("database_create_user").Inc()
// 			metrics.AuthRequests.WithLabelValues("microsoft", "error").Inc()
// 			logger.Error(r.Context(), "Failed to save user to the database", "provider_id", msUser.ProviderID, "error", err)
// 			http.Error(w, "Failed to save user to the database", http.StatusInternalServerError)
// 			return
// 		}
// 		metrics.DatabaseOperations.WithLabelValues("create_user", "success").Inc()

// 		parselUser = createdUser
// 		parselUser.RefreshToken = "" // Do not return the refresh token
// 	} else if token.RefreshToken != "" {
// 		logger.Info(r.Context(), "Updating refresh token on user callback", "user_id", parselUser.ParselID)
// 		parselUser.RefreshToken = token.RefreshToken
// 		metrics.DatabaseOperations.WithLabelValues("update_refresh_token", "attempt").Inc()
// 		err = s.db.UpdateUserRefreshToken(r.Context(), parselUser.ParselID.String(), token.RefreshToken)
// 		if err != nil {
// 			metrics.DatabaseOperations.WithLabelValues("update_refresh_token", "error").Inc()
// 			metrics.Errors.WithLabelValues("database_update_token").Inc()
// 			logger.Error(r.Context(), "Failed to update refresh token on user callback", "user_id", parselUser.ParselID, "error", err)
// 			http.Error(w, "Failed to update refresh token", http.StatusInternalServerError)
// 			return
// 		}
// 		metrics.DatabaseOperations.WithLabelValues("update_refresh_token", "success").Inc()
// 	}

// 	// Generate JWT token
// 	jwtToken, err := auth.GenerateJWT(*parselUser)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("microsoft", "error").Inc()
// 		metrics.Errors.WithLabelValues("jwt_generation").Inc()
// 		logger.Error(r.Context(), "Failed to generate JWT", "user_id", parselUser.ParselID, "error", err)
// 		http.Error(w, "Failed to generate JWT", http.StatusInternalServerError)
// 		return
// 	}

// 	// Set JWT as a secure cookie
// 	http.SetCookie(w, &http.Cookie{
// 		Name:     "jwt",
// 		Value:    jwtToken,
// 		Path:     "/",
// 		Secure:   true,
// 		HttpOnly: true,
// 		SameSite: http.SameSiteStrictMode,
// 		MaxAge:   3600, // 1 hour
// 	})

// 	// Redirect with JWT in URL fragment
// 	http.Redirect(w, r, "/home", http.StatusFound)
// }

// // handleMicrosoftAuth initiates the Microsoft authentication flow
// func (s *Server) handleMicrosoftAuth(w http.ResponseWriter, r *http.Request) {
// 	// Check if user has JWT token in the header
// 	if auth.ValidateJWT(r.Header.Get("Authorization")) {
// 		metrics.AuthRequests.WithLabelValues("microsoft", "already_authenticated").Inc()
// 		logger.Info(r.Context(), "User already authenticated, redirecting to home")
// 		http.Redirect(w, r, "/home", http.StatusFound)
// 		return
// 	}

// 	// Get user ID from URL query to check if we should force consent
// 	// This ID would be present if redirected from a refresh token failure
// 	userID := r.URL.Query().Get("user_id")
// 	forceConsent := userID != ""

// 	if forceConsent {
// 		logger.Info(r.Context(), "Forcing consent screen for user", "user_id", userID)
// 	}

// 	url, err := auth.GetAuthCodeURL(auth.MicrosoftProvider, forceConsent)
// 	if err != nil {
// 		metrics.AuthRequests.WithLabelValues("microsoft", "error").Inc()
// 		metrics.Errors.WithLabelValues("auth_code_url").Inc()
// 		logger.Error(r.Context(), "Failed to get Microsoft auth code URL", "error", err)
// 		http.Error(w, "Failed to initiate authentication", http.StatusInternalServerError)
// 		return
// 	}

// 	metrics.AuthRequests.WithLabelValues("microsoft", "initiated").Inc()
// 	http.Redirect(w, r, url, http.StatusFound)
// }
