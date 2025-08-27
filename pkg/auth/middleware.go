package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// ContextKey type for context keys
type ContextKey string

const (
	// UserContextKey is the key for storing user in context
	UserContextKey ContextKey = "user"
	// TokenContextKey is the key for storing token in context
	TokenContextKey ContextKey = "token"
)

// AuthMiddleware provides authentication middleware functionality
type AuthMiddleware struct {
	authManager *AuthManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authManager *AuthManager) *AuthMiddleware {
	return &AuthMiddleware{
		authManager: authManager,
	}
}

// RequireAuth middleware that requires authentication
func (am *AuthMiddleware) RequireAuth(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		user, err := am.authenticateRequest(ctx)
		if err != nil {
			return fmt.Errorf("authentication required: %w", err)
		}

		// Add user to context
		ctx = context.WithValue(ctx, UserContextKey, user)
		
		return next(ctx)
	}
}

// RequirePermission middleware that requires a specific permission
func (am *AuthMiddleware) RequirePermission(resource, action string) func(func(ctx context.Context) error) func(ctx context.Context) error {
	return func(next func(ctx context.Context) error) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			user, err := am.authenticateRequest(ctx)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			if !am.authManager.CheckPermission(user, resource, action) {
				return fmt.Errorf("permission denied: %s:%s", resource, action)
			}

			// Add user to context
			ctx = context.WithValue(ctx, UserContextKey, user)
			
			return next(ctx)
		}
	}
}

// RequireRole middleware that requires a specific role
func (am *AuthMiddleware) RequireRole(role string) func(func(ctx context.Context) error) func(ctx context.Context) error {
	return func(next func(ctx context.Context) error) func(ctx context.Context) error {
		return func(ctx context.Context) error {
			user, err := am.authenticateRequest(ctx)
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			if !user.HasRole(role) {
				return fmt.Errorf("role required: %s", role)
			}

			// Add user to context
			ctx = context.WithValue(ctx, UserContextKey, user)
			
			return next(ctx)
		}
	}
}

// OptionalAuth middleware that adds user to context if authenticated
func (am *AuthMiddleware) OptionalAuth(next func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		user, _ := am.authenticateRequest(ctx)
		if user != nil {
			ctx = context.WithValue(ctx, UserContextKey, user)
		}
		
		return next(ctx)
	}
}

// authenticateRequest extracts and validates authentication from the request context
func (am *AuthMiddleware) authenticateRequest(ctx context.Context) (*User, error) {
	// If auth is not required, return a default admin user
	if !am.authManager.IsAuthRequired() {
		return &User{
			Username:    "admin",
			Email:       "admin@localhost",
			Roles:       []string{"admin"},
			Permissions: []string{"*"},
			Active:      true,
		}, nil
	}

	// Try to get token from various sources
	token := am.getTokenFromContext(ctx)
	if token == "" {
		return nil, fmt.Errorf("no authentication token provided")
	}

	return am.authManager.ValidateToken(token)
}

// getTokenFromContext extracts authentication token from context
func (am *AuthMiddleware) getTokenFromContext(ctx context.Context) string {
	// First check if token is directly in context
	if token, ok := ctx.Value(TokenContextKey).(string); ok && token != "" {
		return token
	}

	// Check environment variable
	if token := os.Getenv("CORYNTH_TOKEN"); token != "" {
		return token
	}

	// Check for token file
	if tokenFile := os.Getenv("CORYNTH_TOKEN_FILE"); tokenFile != "" {
		if data, err := os.ReadFile(tokenFile); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	// Check default token file location
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" {
		tokenFile := fmt.Sprintf("%s/.corynth/token", homeDir)
		if data, err := os.ReadFile(tokenFile); err == nil {
			return strings.TrimSpace(string(data))
		}
	}

	return ""
}

// GetUserFromContext retrieves the authenticated user from context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(UserContextKey).(*User)
	return user, ok
}

// GetTokenFromContext retrieves the authentication token from context
func GetTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(TokenContextKey).(string)
	return token, ok
}

// SetTokenInContext adds a token to the context
func SetTokenInContext(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, TokenContextKey, token)
}

// TokenStore provides functionality for storing and retrieving tokens
type TokenStore struct {
	tokenFile string
}

// NewTokenStore creates a new token store
func NewTokenStore() *TokenStore {
	homeDir, _ := os.UserHomeDir()
	tokenFile := fmt.Sprintf("%s/.corynth/token", homeDir)
	
	return &TokenStore{
		tokenFile: tokenFile,
	}
}

// SaveToken saves a token to the token file
func (ts *TokenStore) SaveToken(token string) error {
	// Create directory if it doesn't exist
	dir := fmt.Sprintf("%s/.corynth", os.Getenv("HOME"))
	if homeDir, _ := os.UserHomeDir(); homeDir != "" {
		dir = fmt.Sprintf("%s/.corynth", homeDir)
	}
	
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Write token with restricted permissions
	if err := os.WriteFile(ts.tokenFile, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// LoadToken loads a token from the token file
func (ts *TokenStore) LoadToken() (string, error) {
	data, err := os.ReadFile(ts.tokenFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no token found - please login with 'corynth auth login'")
		}
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// DeleteToken removes the token file
func (ts *TokenStore) DeleteToken() error {
	if err := os.Remove(ts.tokenFile); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

// HasToken checks if a token file exists
func (ts *TokenStore) HasToken() bool {
	_, err := os.Stat(ts.tokenFile)
	return err == nil
}