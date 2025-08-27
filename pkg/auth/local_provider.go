package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LocalProvider implements file-based authentication
type LocalProvider struct {
	usersFile string
	users     map[string]*User
}

// UserStore represents the structure of the users file
type UserStore struct {
	Users map[string]*User `yaml:"users"`
}

// NewLocalProvider creates a new local file-based auth provider
func NewLocalProvider(dataDir string) (*LocalProvider, error) {
	usersFile := filepath.Join(dataDir, "users.yaml")
	
	provider := &LocalProvider{
		usersFile: usersFile,
		users:     make(map[string]*User),
	}

	// Load existing users
	if err := provider.loadUsers(); err != nil {
		return nil, fmt.Errorf("failed to load users: %w", err)
	}

	// Create default admin user if no users exist
	if len(provider.users) == 0 {
		if err := provider.createDefaultAdmin(); err != nil {
			return nil, fmt.Errorf("failed to create default admin: %w", err)
		}
	}

	return provider, nil
}

// Authenticate validates username and password
func (lp *LocalProvider) Authenticate(username, password string) (*User, error) {
	user, exists := lp.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	if !user.IsActive() {
		return nil, fmt.Errorf("user account is disabled: %s", username)
	}

	expectedHash := hashPassword(password)
	if user.PasswordHash != expectedHash {
		return nil, fmt.Errorf("invalid password for user: %s", username)
	}

	return user, nil
}

// ValidateToken validates a token (not used in local provider, handled by AuthManager)
func (lp *LocalProvider) ValidateToken(token string) (*User, error) {
	return nil, fmt.Errorf("token validation should be handled by AuthManager")
}

// GetUser retrieves a user by username
func (lp *LocalProvider) GetUser(username string) (*User, error) {
	user, exists := lp.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	// Return a copy to prevent external modification
	userCopy := *user
	return &userCopy, nil
}

// CreateUser creates a new user
func (lp *LocalProvider) CreateUser(user *User) error {
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}

	if _, exists := lp.users[user.Username]; exists {
		return fmt.Errorf("user already exists: %s", user.Username)
	}

	// Validate required fields
	if user.Email == "" {
		return fmt.Errorf("email is required")
	}

	if user.PasswordHash == "" {
		return fmt.Errorf("password hash is required")
	}

	// Set default roles if none provided
	if len(user.Roles) == 0 {
		user.Roles = []string{"user"}
	}

	// Initialize metadata if nil
	if user.Metadata == nil {
		user.Metadata = make(map[string]string)
	}

	// Store user
	lp.users[user.Username] = user

	return lp.saveUsers()
}

// UpdateUser updates an existing user
func (lp *LocalProvider) UpdateUser(user *User) error {
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}

	if _, exists := lp.users[user.Username]; !exists {
		return fmt.Errorf("user not found: %s", user.Username)
	}

	// Store updated user
	lp.users[user.Username] = user

	return lp.saveUsers()
}

// DeleteUser deletes a user
func (lp *LocalProvider) DeleteUser(username string) error {
	if username == "" {
		return fmt.Errorf("username is required")
	}

	if _, exists := lp.users[username]; !exists {
		return fmt.Errorf("user not found: %s", username)
	}

	// Prevent deletion of admin user if it's the only admin
	if lp.users[username].HasRole("admin") {
		adminCount := 0
		for _, user := range lp.users {
			if user.HasRole("admin") {
				adminCount++
			}
		}
		if adminCount <= 1 {
			return fmt.Errorf("cannot delete the last admin user")
		}
	}

	delete(lp.users, username)

	return lp.saveUsers()
}

// ListUsers returns all users
func (lp *LocalProvider) ListUsers() ([]*User, error) {
	users := make([]*User, 0, len(lp.users))
	
	for _, user := range lp.users {
		// Return copies to prevent external modification
		userCopy := *user
		users = append(users, &userCopy)
	}

	return users, nil
}

// ChangePassword changes a user's password
func (lp *LocalProvider) ChangePassword(username, oldPassword, newPassword string) error {
	user, exists := lp.users[username]
	if !exists {
		return fmt.Errorf("user not found: %s", username)
	}

	// Verify old password
	expectedHash := hashPassword(oldPassword)
	if user.PasswordHash != expectedHash {
		return fmt.Errorf("invalid current password")
	}

	// Update password
	user.PasswordHash = hashPassword(newPassword)

	return lp.saveUsers()
}

// Helper methods

func (lp *LocalProvider) loadUsers() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(lp.usersFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// If users file doesn't exist, that's fine - we'll create it
	if _, err := os.Stat(lp.usersFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(lp.usersFile)
	if err != nil {
		return err
	}

	var store UserStore
	if err := yaml.Unmarshal(data, &store); err != nil {
		return err
	}

	if store.Users != nil {
		lp.users = store.Users
	}

	return nil
}

func (lp *LocalProvider) saveUsers() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(lp.usersFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	store := UserStore{
		Users: lp.users,
	}

	data, err := yaml.Marshal(store)
	if err != nil {
		return err
	}

	// Write with restricted permissions
	return os.WriteFile(lp.usersFile, data, 0600)
}

func (lp *LocalProvider) createDefaultAdmin() error {
	// Create default admin user with password "admin123"
	// In production, this should be changed immediately
	admin := &User{
		Username:     "admin",
		Email:        "admin@localhost",
		PasswordHash: hashPassword("admin123"),
		Roles:        []string{"admin"},
		Permissions:  []string{"*"},
		Metadata:     make(map[string]string),
		Active:       true,
	}
	admin.Metadata["created_by"] = "system"
	admin.Metadata["note"] = "Default admin user - change password immediately"

	return lp.CreateUser(admin)
}