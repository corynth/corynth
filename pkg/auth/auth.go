package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	Authenticate(username, password string) (*User, error)
	ValidateToken(token string) (*User, error)
	GetUser(username string) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(username string) error
	ListUsers() ([]*User, error)
}

// User represents a user in the system
type User struct {
	Username    string            `yaml:"username" json:"username"`
	Email       string            `yaml:"email" json:"email"`
	PasswordHash string           `yaml:"password_hash" json:"-"`
	Roles       []string          `yaml:"roles" json:"roles"`
	Permissions []string          `yaml:"permissions" json:"permissions"`
	Metadata    map[string]string `yaml:"metadata" json:"metadata"`
	CreatedAt   time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `yaml:"updated_at" json:"updated_at"`
	LastLogin   *time.Time        `yaml:"last_login" json:"last_login"`
	Active      bool              `yaml:"active" json:"active"`
}

// Session represents an authenticated session
type Session struct {
	Token     string    `yaml:"token" json:"token"`
	Username  string    `yaml:"username" json:"username"`
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
	ExpiresAt time.Time `yaml:"expires_at" json:"expires_at"`
	Metadata  map[string]string `yaml:"metadata" json:"metadata"`
}

// Permission represents a permission in the system
type Permission struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Resource    string `yaml:"resource" json:"resource"`
	Action      string `yaml:"action" json:"action"`
}

// Role represents a role with a set of permissions
type Role struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Permissions []string `yaml:"permissions" json:"permissions"`
}

// AuthConfig defines authentication configuration
type AuthConfig struct {
	Provider        string            `yaml:"provider" json:"provider"`                 // "local", "ldap", "oauth"
	TokenExpiration time.Duration     `yaml:"token_expiration" json:"token_expiration"` // Default token expiration
	SecretKey       string            `yaml:"secret_key" json:"-"`                      // For signing tokens
	ProviderConfig  map[string]string `yaml:"provider_config" json:"provider_config"`  // Provider-specific config
	DefaultRoles    []string          `yaml:"default_roles" json:"default_roles"`      // Default roles for new users
	RequireAuth     bool              `yaml:"require_auth" json:"require_auth"`         // Whether authentication is required
}

// AuthManager manages authentication and authorization
type AuthManager struct {
	config   *AuthConfig
	Provider AuthProvider
	sessions map[string]*Session
	dataDir  string
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(config *AuthConfig, dataDir string) (*AuthManager, error) {
	if config == nil {
		config = DefaultAuthConfig()
	}

	manager := &AuthManager{
		config:   config,
		sessions: make(map[string]*Session),
		dataDir:  dataDir,
	}

	// Initialize provider based on configuration
	provider, err := createProvider(config, dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth provider: %w", err)
	}
	manager.Provider = provider

	// Load existing sessions
	if err := manager.loadSessions(); err != nil {
		// Log warning but don't fail - sessions will be regenerated
		fmt.Printf("Warning: failed to load sessions: %v\n", err)
	}

	return manager, nil
}

// DefaultAuthConfig returns default authentication configuration
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		Provider:        "local",
		TokenExpiration: 24 * time.Hour,
		SecretKey:       generateSecretKey(),
		DefaultRoles:    []string{"user"},
		RequireAuth:     false, // Start with auth disabled for backward compatibility
	}
}

// Authenticate authenticates a user and returns a session token
func (am *AuthManager) Authenticate(username, password string) (string, error) {
	if !am.config.RequireAuth {
		return "", fmt.Errorf("authentication is disabled")
	}

	user, err := am.Provider.Authenticate(username, password)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	am.Provider.UpdateUser(user)

	// Create session
	session := &Session{
		Token:     generateToken(),
		Username:  user.Username,
		CreatedAt: now,
		ExpiresAt: now.Add(am.config.TokenExpiration),
		Metadata:  make(map[string]string),
	}

	am.sessions[session.Token] = session
	am.saveSessions()

	return session.Token, nil
}

// ValidateToken validates a session token and returns the user
func (am *AuthManager) ValidateToken(token string) (*User, error) {
	if !am.config.RequireAuth {
		// If auth is disabled, create a default admin user
		return &User{
			Username:    "admin",
			Email:       "admin@localhost",
			Roles:       []string{"admin"},
			Permissions: []string{"*"},
			Active:      true,
		}, nil
	}

	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	session, exists := am.sessions[token]
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}

	if time.Now().After(session.ExpiresAt) {
		delete(am.sessions, token)
		am.saveSessions()
		return nil, fmt.Errorf("token expired")
	}

	return am.Provider.GetUser(session.Username)
}

// CreateUser creates a new user
func (am *AuthManager) CreateUser(username, email, password string, roles []string) error {
	if !am.config.RequireAuth {
		return fmt.Errorf("authentication is disabled")
	}

	if roles == nil {
		roles = am.config.DefaultRoles
	}

	user := &User{
		Username:     username,
		Email:        email,
		PasswordHash: hashPassword(password),
		Roles:        roles,
		Permissions:  []string{}, // Will be populated based on roles
		Metadata:     make(map[string]string),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Active:       true,
	}

	return am.Provider.CreateUser(user)
}

// CheckPermission checks if a user has a specific permission
func (am *AuthManager) CheckPermission(user *User, resource, action string) bool {
	if user == nil {
		return false
	}

	// Check for admin role or wildcard permission
	for _, role := range user.Roles {
		if role == "admin" {
			return true
		}
	}

	for _, perm := range user.Permissions {
		if perm == "*" || perm == fmt.Sprintf("%s:%s", resource, action) {
			return true
		}
	}

	return false
}

// EnableAuth enables authentication system
func (am *AuthManager) EnableAuth() error {
	am.config.RequireAuth = true
	return am.saveConfig()
}

// DisableAuth disables authentication system
func (am *AuthManager) DisableAuth() error {
	am.config.RequireAuth = false
	return am.saveConfig()
}

// IsAuthRequired returns whether authentication is required
func (am *AuthManager) IsAuthRequired() bool {
	return am.config.RequireAuth
}

// Helper functions

func createProvider(config *AuthConfig, dataDir string) (AuthProvider, error) {
	switch config.Provider {
	case "local", "":
		return NewLocalProvider(dataDir)
	default:
		return nil, fmt.Errorf("unsupported auth provider: %s", config.Provider)
	}
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateSecretKey() string {
	bytes := make([]byte, 64)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (am *AuthManager) loadSessions() error {
	sessionFile := filepath.Join(am.dataDir, "sessions.yaml")
	
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No sessions file is fine
		}
		return err
	}

	var sessions map[string]*Session
	if err := yaml.Unmarshal(data, &sessions); err != nil {
		return err
	}

	// Filter out expired sessions
	now := time.Now()
	for token, session := range sessions {
		if now.Before(session.ExpiresAt) {
			am.sessions[token] = session
		}
	}

	return nil
}

func (am *AuthManager) saveSessions() error {
	if err := os.MkdirAll(am.dataDir, 0755); err != nil {
		return err
	}

	sessionFile := filepath.Join(am.dataDir, "sessions.yaml")
	
	data, err := yaml.Marshal(am.sessions)
	if err != nil {
		return err
	}

	return os.WriteFile(sessionFile, data, 0600)
}

func (am *AuthManager) saveConfig() error {
	if err := os.MkdirAll(am.dataDir, 0755); err != nil {
		return err
	}

	configFile := filepath.Join(am.dataDir, "auth.yaml")
	
	data, err := yaml.Marshal(am.config)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0600)
}

// Utility functions for working with roles and permissions

// HasRole checks if a user has a specific role
func (u *User) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if a user has a specific permission
func (u *User) HasPermission(permission string) bool {
	for _, p := range u.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// AddRole adds a role to the user
func (u *User) AddRole(role string) {
	if !u.HasRole(role) {
		u.Roles = append(u.Roles, role)
		u.UpdatedAt = time.Now()
	}
}

// RemoveRole removes a role from the user
func (u *User) RemoveRole(role string) {
	for i, r := range u.Roles {
		if r == role {
			u.Roles = append(u.Roles[:i], u.Roles[i+1:]...)
			u.UpdatedAt = time.Now()
			break
		}
	}
}

// AddPermission adds a permission to the user
func (u *User) AddPermission(permission string) {
	if !u.HasPermission(permission) {
		u.Permissions = append(u.Permissions, permission)
		u.UpdatedAt = time.Now()
	}
}

// IsActive returns whether the user account is active
func (u *User) IsActive() bool {
	return u.Active
}

// Activate activates the user account
func (u *User) Activate() {
	u.Active = true
	u.UpdatedAt = time.Now()
}

// Deactivate deactivates the user account
func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now()
}