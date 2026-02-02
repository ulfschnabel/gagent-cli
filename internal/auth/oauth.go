package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthConfig holds the OAuth2 configuration.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
	RedirectURL  string
}

// NewOAuthConfig creates a new OAuth2 config from client credentials.
func NewOAuthConfig(clientID, clientSecret string, scopeType ScopeType) *oauth2.Config {
	scopes := GetScopes(scopeType)
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
}

// generateState creates a random state string for OAuth security.
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// findAvailablePort finds a random available port.
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port, nil
}

// AuthResult holds the result of an OAuth authorization flow.
type AuthResult struct {
	Token *oauth2.Token
	Error error
}

// PerformOAuthFlow runs the OAuth2 authorization flow with a local callback server.
// It opens the browser for user authorization and waits for the callback.
func PerformOAuthFlow(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Find an available port for the callback server
	port, err := findAvailablePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find available port: %w", err)
	}

	// Set the redirect URL with the found port
	config.RedirectURL = fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	// Generate state for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Channel to receive the auth result
	resultCh := make(chan AuthResult, 1)

	// Create the callback server
	server := &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", port),
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// Verify state
		if r.URL.Query().Get("state") != state {
			resultCh <- AuthResult{Error: fmt.Errorf("invalid state parameter")}
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			return
		}

		// Check for errors
		if errParam := r.URL.Query().Get("error"); errParam != "" {
			errDesc := r.URL.Query().Get("error_description")
			resultCh <- AuthResult{Error: fmt.Errorf("authorization error: %s - %s", errParam, errDesc)}
			http.Error(w, fmt.Sprintf("Authorization failed: %s", errDesc), http.StatusBadRequest)
			return
		}

		// Get the authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			resultCh <- AuthResult{Error: fmt.Errorf("no authorization code received")}
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			return
		}

		// Exchange code for token
		token, err := config.Exchange(ctx, code)
		if err != nil {
			resultCh <- AuthResult{Error: fmt.Errorf("failed to exchange code for token: %w", err)}
			http.Error(w, "Token exchange failed", http.StatusInternalServerError)
			return
		}

		resultCh <- AuthResult{Token: token}

		// Send success response to browser
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>gagent-cli Authorization</title></head>
<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
<h1>✓ Authorization Successful</h1>
<p>You can close this window and return to the terminal.</p>
</body>
</html>`)
	})

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			resultCh <- AuthResult{Error: fmt.Errorf("callback server error: %w", err)}
		}
	}()

	// Generate the authorization URL
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	// Open the browser
	fmt.Printf("\nOpening browser for authorization...\n")
	fmt.Printf("If the browser doesn't open, visit this URL:\n%s\n\n", authURL)

	if err := browser.OpenURL(authURL); err != nil {
		fmt.Printf("Warning: could not open browser automatically: %v\n", err)
	}

	// Wait for the result with a timeout
	select {
	case result := <-resultCh:
		// Shutdown the server
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)

		if result.Error != nil {
			return nil, result.Error
		}
		return result.Token, nil

	case <-time.After(5 * time.Minute):
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
		return nil, fmt.Errorf("authorization timed out after 5 minutes")

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
		return nil, ctx.Err()
	}
}

// GetClient returns an HTTP client with the appropriate token for the scope type.
// If the token needs refreshing, it will be refreshed and saved.
func GetClient(ctx context.Context, configDir, clientID, clientSecret string, scopeType ScopeType) (*http.Client, error) {
	tokenPath := TokenPath(configDir, scopeType)
	token, err := LoadToken(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("scope '%s' not authorized: %w", scopeType, err)
	}

	config := NewOAuthConfig(clientID, clientSecret, scopeType)
	tokenSource := config.TokenSource(ctx, token)

	// Get a potentially refreshed token
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// If the token was refreshed, save it
	if newToken.AccessToken != token.AccessToken {
		if err := SaveToken(tokenPath, newToken); err != nil {
			// Log warning but don't fail - we can still use the token
			fmt.Printf("Warning: failed to save refreshed token: %v\n", err)
		}
	}

	return oauth2.NewClient(ctx, tokenSource), nil
}

// RequireScope checks if the required scope is authorized and returns an error if not.
func RequireScope(configDir string, scopeType ScopeType) error {
	if !TokenExists(configDir, scopeType) {
		return fmt.Errorf("scope '%s' not authorized. Run: gagent-cli auth login --scope %s", scopeType, scopeType)
	}
	return nil
}

// PerformOAuthFlowManual runs the OAuth2 flow for remote/headless environments.
// The user manually visits the auth URL and pastes back the callback URL.
func PerformOAuthFlowManual(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Use a fixed redirect URL - Google will redirect there but it won't load
	config.RedirectURL = "http://127.0.0.1:8085/callback"

	// Generate state for CSRF protection
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Generate the authorization URL
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)

	fmt.Println()
	fmt.Println("Open this URL in your browser:")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("After authorizing, your browser will redirect to a localhost URL that won't load.")
	fmt.Println("Copy the ENTIRE URL from your browser's address bar and paste it below.")
	fmt.Println()
	fmt.Print("Paste the callback URL: ")

	reader := bufio.NewReader(os.Stdin)
	callbackURL, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %w", err)
	}
	callbackURL = strings.TrimSpace(callbackURL)

	// Parse the callback URL
	parsed, err := url.Parse(callbackURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Verify state
	if parsed.Query().Get("state") != state {
		return nil, fmt.Errorf("invalid state parameter - possible CSRF attack or URL mismatch")
	}

	// Check for errors
	if errParam := parsed.Query().Get("error"); errParam != "" {
		errDesc := parsed.Query().Get("error_description")
		return nil, fmt.Errorf("authorization error: %s - %s", errParam, errDesc)
	}

	// Get the authorization code
	code := parsed.Query().Get("code")
	if code == "" {
		return nil, fmt.Errorf("no authorization code in URL")
	}

	// Exchange code for token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token, nil
}
