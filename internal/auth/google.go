package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/parsel-email/lib-go/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

// GoogleOAuthProvider implements the OAuthProvider interface for Google
type GoogleOAuthProvider struct {
	config *oauth2.Config
}

// Google OAuth scopes
var googleScopes = []string{
	gmail.GmailReadonlyScope,                           // Read-only access to Gmail
	"https://www.googleapis.com/auth/userinfo.profile", // View your basic profile info
	"https://www.googleapis.com/auth/userinfo.email",   // View your email address
}

// NewGoogleProvider creates a new Google OAuth provider
func NewGoogleProvider() (*GoogleOAuthProvider, error) {
	// Read Google credentials from file
	credentialsFile := os.Getenv("GOOGLE_CREDENTIALS_FILE")
	if credentialsFile == "" {
		logger.Error(context.Background(), "GOOGLE_CREDENTIALS_FILE environment variable not set")
		return nil, fmt.Errorf("GOOGLE_CREDENTIALS_FILE environment variable not set")
	}

	googleCreds, err := os.ReadFile(credentialsFile)
	if err != nil {
		logger.Error(context.Background(), "Unable to read Google client secret file", "error", err)
		return nil, err
	}

	// Configure Google OAuth
	config, err := google.ConfigFromJSON(googleCreds, googleScopes...)
	if err != nil {
		logger.Error(context.Background(), "Unable to parse Google client secret file to config", "error", err)
		return nil, err
	}

	return &GoogleOAuthProvider{
		config: config,
	}, nil
}

// GetAuthCodeURL returns the URL to redirect users to for OAuth flow
func (p *GoogleOAuthProvider) GetAuthCodeURL(forceConsent bool) string {
	options := []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}

	// Only force consent screen when explicitly requested
	if forceConsent {
		options = append(options, oauth2.ApprovalForce)
	}

	return p.config.AuthCodeURL("state-token", options...)
}

// ExchangeCode exchanges an authorization code for an OAuth token
func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// GetClient returns an HTTP client configured with the given token
func (p *GoogleOAuthProvider) GetClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return p.config.Client(ctx, token)
}

// RefreshToken attempts to refresh an OAuth token using the refresh token
func (p *GoogleOAuthProvider) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	tokenSource := p.config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})
	return tokenSource.Token()
}

// RevokeToken revokes a token with Google
func (p *GoogleOAuthProvider) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("empty token")
	}

	// Google's token revocation endpoint
	revokeURL := "https://oauth2.googleapis.com/revoke"

	// Create the form data for the request
	formData := strings.NewReader(fmt.Sprintf("token=%s", token))

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", revokeURL, formData)
	if err != nil {
		return err
	}

	// Set the content type
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to revoke token: %s", resp.Status)
	}

	return nil
}
