package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/corynth/corynth/pkg/auth"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// NewAuthCommand creates the authentication command group
func NewAuthCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and user management commands",
		Long: `Authentication commands for managing users, roles, and authentication tokens.

Corynth supports multiple authentication modes:
- Disabled: No authentication required (default)
- Local: File-based user management
- Future: LDAP, OAuth, and other providers`,
		Example: `  # Enable authentication
  corynth auth enable

  # Login and get a token
  corynth auth login

  # Create a new user
  corynth auth create-user --username alice --email alice@example.com

  # List all users
  corynth auth list-users

  # Change password
  corynth auth change-password

  # Logout
  corynth auth logout`,
	}

	cmd.AddCommand(NewAuthStatusCommand())
	cmd.AddCommand(NewAuthEnableCommand())
	cmd.AddCommand(NewAuthDisableCommand())
	cmd.AddCommand(NewAuthLoginCommand())
	cmd.AddCommand(NewAuthLogoutCommand())
	cmd.AddCommand(NewAuthWhoamiCommand())
	cmd.AddCommand(NewAuthCreateUserCommand())
	cmd.AddCommand(NewAuthListUsersCommand())
	cmd.AddCommand(NewAuthDeleteUserCommand())
	cmd.AddCommand(NewAuthChangePasswordCommand())

	return cmd
}

// NewAuthStatusCommand shows authentication status
func NewAuthStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  "Display current authentication configuration and status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthStatus()
		},
	}
}

// NewAuthEnableCommand enables authentication
func NewAuthEnableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable authentication system",
		Long: `Enable the authentication system. This will require users to authenticate
before executing workflows or accessing sensitive operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthEnable()
		},
	}
}

// NewAuthDisableCommand disables authentication
func NewAuthDisableCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable authentication system",
		Long: `Disable the authentication system. This will allow unrestricted access
to all Corynth functionality. Use with caution in production environments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthDisable()
		},
	}
}

// NewAuthLoginCommand creates login command
func NewAuthLoginCommand() *cobra.Command {
	var username string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login and obtain an authentication token",
		Long: `Authenticate with Corynth and store a session token for subsequent commands.
The token will be stored in ~/.corynth/token for automatic use.`,
		Example: `  # Login interactively
  corynth auth login

  # Login with specific username
  corynth auth login --username alice`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogin(username)
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username to login with")

	return cmd
}

// NewAuthLogoutCommand creates logout command
func NewAuthLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout and remove authentication token",
		Long:  "Remove the stored authentication token and logout from Corynth.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogout()
		},
	}
}

// NewAuthWhoamiCommand shows current user
func NewAuthWhoamiCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current authenticated user",
		Long:  "Display information about the currently authenticated user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthWhoami()
		},
	}
}

// NewAuthCreateUserCommand creates user creation command
func NewAuthCreateUserCommand() *cobra.Command {
	var username, email string
	var roles []string
	var admin bool

	cmd := &cobra.Command{
		Use:   "create-user",
		Short: "Create a new user",
		Long:  "Create a new user account with specified username, email, and roles.",
		Example: `  # Create a regular user
  corynth auth create-user --username alice --email alice@example.com

  # Create an admin user
  corynth auth create-user --username bob --email bob@example.com --admin

  # Create user with specific roles
  corynth auth create-user --username charlie --email charlie@example.com --roles user,developer`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if admin {
				roles = append(roles, "admin")
			}
			return runAuthCreateUser(username, email, roles)
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username for the new user (required)")
	cmd.Flags().StringVar(&email, "email", "", "Email address for the new user (required)")
	cmd.Flags().StringSliceVar(&roles, "roles", []string{"user"}, "Roles to assign to the user")
	cmd.Flags().BoolVar(&admin, "admin", false, "Create user with admin privileges")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("email")

	return cmd
}

// NewAuthListUsersCommand lists all users
func NewAuthListUsersCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-users",
		Short: "List all users",
		Long:  "Display a list of all user accounts in the system.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthListUsers()
		},
	}
}

// NewAuthDeleteUserCommand deletes a user
func NewAuthDeleteUserCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete-user <username>",
		Short: "Delete a user account",
		Long:  "Delete the specified user account. This action cannot be undone.",
		Args:  cobra.ExactArgs(1),
		Example: `  # Delete user with confirmation
  corynth auth delete-user alice

  # Delete user without confirmation
  corynth auth delete-user alice --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthDeleteUser(args[0], force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Delete without confirmation")

	return cmd
}

// NewAuthChangePasswordCommand changes user password
func NewAuthChangePasswordCommand() *cobra.Command {
	var username string

	cmd := &cobra.Command{
		Use:   "change-password",
		Short: "Change user password",
		Long:  "Change password for the current user or specified user (admin only).",
		Example: `  # Change own password
  corynth auth change-password

  # Change another user's password (admin only)
  corynth auth change-password --username alice`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthChangePassword(username)
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username to change password for (admin only)")

	return cmd
}

// Implementation functions

func runAuthStatus() error {
	authManager, err := getAuthManager()
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	fmt.Printf("🔐 Authentication Status\n\n")
	
	if authManager.IsAuthRequired() {
		fmt.Printf("Status:           ✅ Enabled\n")
		fmt.Printf("Provider:         local\n")
		
		// Check if user is logged in
		tokenStore := auth.NewTokenStore()
		if tokenStore.HasToken() {
			token, err := tokenStore.LoadToken()
			if err == nil {
				user, err := authManager.ValidateToken(token)
				if err == nil {
					fmt.Printf("Current User:     %s (%s)\n", user.Username, user.Email)
					fmt.Printf("Roles:            %s\n", strings.Join(user.Roles, ", "))
				} else {
					fmt.Printf("Token Status:     ❌ Invalid/Expired\n")
				}
			}
		} else {
			fmt.Printf("Token Status:     ❌ Not logged in\n")
		}
	} else {
		fmt.Printf("Status:           ❌ Disabled\n")
		fmt.Printf("Provider:         none\n")
		fmt.Printf("Access:           unrestricted\n")
	}

	fmt.Printf("\n")
	return nil
}

func runAuthEnable() error {
	authManager, err := getAuthManager()
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	if err := authManager.EnableAuth(); err != nil {
		return fmt.Errorf("failed to enable authentication: %w", err)
	}

	fmt.Printf("✅ Authentication enabled successfully\n")
	fmt.Printf("\n")
	fmt.Printf("Default admin user created:\n")
	fmt.Printf("  Username: admin\n")
	fmt.Printf("  Password: admin123\n")
	fmt.Printf("\n")
	fmt.Printf("⚠️  IMPORTANT: Change the default admin password immediately:\n")
	fmt.Printf("   corynth auth change-password --username admin\n")
	fmt.Printf("\n")
	fmt.Printf("To login:\n")
	fmt.Printf("   corynth auth login\n")

	return nil
}

func runAuthDisable() error {
	authManager, err := getAuthManager()
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	if err := authManager.DisableAuth(); err != nil {
		return fmt.Errorf("failed to disable authentication: %w", err)
	}

	fmt.Printf("❌ Authentication disabled\n")
	fmt.Printf("⚠️  WARNING: All operations now have unrestricted access\n")

	return nil
}

func runAuthLogin(username string) error {
	authManager, err := getAuthManager()
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	if !authManager.IsAuthRequired() {
		return fmt.Errorf("authentication is disabled - enable it with 'corynth auth enable'")
	}

	// Get username if not provided
	if username == "" {
		fmt.Print("Username: ")
		fmt.Scanln(&username)
	}

	// Get password securely
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // New line after password input

	password := string(passwordBytes)

	// Authenticate
	token, err := authManager.Authenticate(username, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save token
	tokenStore := auth.NewTokenStore()
	if err := tokenStore.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Printf("✅ Login successful\n")
	fmt.Printf("Token saved to ~/.corynth/token\n")

	return nil
}

func runAuthLogout() error {
	tokenStore := auth.NewTokenStore()
	
	if !tokenStore.HasToken() {
		fmt.Printf("❌ Not logged in\n")
		return nil
	}

	if err := tokenStore.DeleteToken(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	fmt.Printf("✅ Logged out successfully\n")
	return nil
}

func runAuthWhoami() error {
	authManager, err := getAuthManager()
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	if !authManager.IsAuthRequired() {
		fmt.Printf("Username:     admin (default)\n")
		fmt.Printf("Email:        admin@localhost\n")
		fmt.Printf("Roles:        admin\n")
		fmt.Printf("Permissions:  *\n")
		fmt.Printf("Status:       Authentication disabled\n")
		return nil
	}

	tokenStore := auth.NewTokenStore()
	token, err := tokenStore.LoadToken()
	if err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	user, err := authManager.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	fmt.Printf("Username:     %s\n", user.Username)
	fmt.Printf("Email:        %s\n", user.Email)
	fmt.Printf("Roles:        %s\n", strings.Join(user.Roles, ", "))
	fmt.Printf("Permissions:  %s\n", strings.Join(user.Permissions, ", "))
	fmt.Printf("Active:       %t\n", user.Active)
	fmt.Printf("Created:      %s\n", user.CreatedAt.Format("2006-01-02 15:04:05"))
	if user.LastLogin != nil {
		fmt.Printf("Last Login:   %s\n", user.LastLogin.Format("2006-01-02 15:04:05"))
	}

	return nil
}

func runAuthCreateUser(username, email string, roles []string) error {
	authManager, err := getAuthManager()
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	if !authManager.IsAuthRequired() {
		return fmt.Errorf("authentication is disabled - enable it with 'corynth auth enable'")
	}

	// Get password securely
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // New line

	fmt.Print("Confirm Password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	fmt.Println() // New line

	if string(passwordBytes) != string(confirmBytes) {
		return fmt.Errorf("passwords do not match")
	}

	// Create user
	if err := authManager.CreateUser(username, email, string(passwordBytes), roles); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	fmt.Printf("✅ User '%s' created successfully\n", username)
	fmt.Printf("Email: %s\n", email)
	fmt.Printf("Roles: %s\n", strings.Join(roles, ", "))

	return nil
}

func runAuthListUsers() error {
	authManager, err := getAuthManager()
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	if !authManager.IsAuthRequired() {
		return fmt.Errorf("authentication is disabled - enable it with 'corynth auth enable'")
	}

	users, err := authManager.Provider.ListUsers()
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	if len(users) == 0 {
		fmt.Printf("No users found\n")
		return nil
	}

	fmt.Printf("👥 Users (%d total)\n\n", len(users))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "USERNAME\tEMAIL\tROLES\tSTATUS\tLAST LOGIN")
	fmt.Fprintln(w, "--------\t-----\t-----\t------\t----------")

	for _, user := range users {
		status := "✅ active"
		if !user.Active {
			status = "❌ disabled"
		}

		lastLogin := "never"
		if user.LastLogin != nil {
			lastLogin = user.LastLogin.Format("2006-01-02 15:04")
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			user.Username,
			user.Email,
			strings.Join(user.Roles, ","),
			status,
			lastLogin)
	}

	w.Flush()
	return nil
}

func runAuthDeleteUser(username string, force bool) error {
	authManager, err := getAuthManager()
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	if !authManager.IsAuthRequired() {
		return fmt.Errorf("authentication is disabled - enable it with 'corynth auth enable'")
	}

	if !force {
		fmt.Printf("⚠️  Are you sure you want to delete user '%s'? [y/N]: ", username)
		var response string
		fmt.Scanln(&response)

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Printf("❌ User deletion cancelled\n")
			return nil
		}
	}

	if err := authManager.Provider.DeleteUser(username); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	fmt.Printf("✅ User '%s' deleted successfully\n", username)
	return nil
}

func runAuthChangePassword(username string) error {
	authManager, err := getAuthManager()
	if err != nil {
		return fmt.Errorf("failed to initialize auth manager: %w", err)
	}

	if !authManager.IsAuthRequired() {
		return fmt.Errorf("authentication is disabled - enable it with 'corynth auth enable'")
	}

	localProvider, ok := authManager.Provider.(*auth.LocalProvider)
	if !ok {
		return fmt.Errorf("password change not supported for current auth provider")
	}

	// If no username specified, use current user
	if username == "" {
		tokenStore := auth.NewTokenStore()
		token, err := tokenStore.LoadToken()
		if err != nil {
			return fmt.Errorf("not logged in: %w", err)
		}

		user, err := authManager.ValidateToken(token)
		if err != nil {
			return fmt.Errorf("invalid token: %w", err)
		}

		username = user.Username
	}

	// Get current password
	fmt.Print("Current Password: ")
	oldPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read current password: %w", err)
	}
	fmt.Println()

	// Get new password
	fmt.Print("New Password: ")
	newPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read new password: %w", err)
	}
	fmt.Println()

	// Confirm new password
	fmt.Print("Confirm New Password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password confirmation: %w", err)
	}
	fmt.Println()

	if string(newPasswordBytes) != string(confirmBytes) {
		return fmt.Errorf("new passwords do not match")
	}

	// Change password
	if err := localProvider.ChangePassword(username, string(oldPasswordBytes), string(newPasswordBytes)); err != nil {
		return fmt.Errorf("failed to change password: %w", err)
	}

	fmt.Printf("✅ Password changed successfully for user '%s'\n", username)
	return nil
}

// Helper function to get auth manager
func getAuthManager() (*auth.AuthManager, error) {
	homeDir, _ := os.UserHomeDir()
	authDir := filepath.Join(homeDir, ".corynth", "auth")
	
	return auth.NewAuthManager(auth.DefaultAuthConfig(), authDir)
}