package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/mitchellh/go-homedir"
)

// Config represents the complete Corynth configuration
type Config struct {
	Version       string                `hcl:"version,optional"`
	Notifications *NotificationConfig   `hcl:"notifications,block"`
	Plugins       *PluginConfig        `hcl:"plugins,block"`
	State         *StateConfig         `hcl:"state,block"`
	Variables     map[string]string    `hcl:"variables,optional"`
	Environments  []EnvironmentConfig  `hcl:"environment,block"`
}

// NotificationConfig defines notification settings
type NotificationConfig struct {
	Enabled      bool             `hcl:"enabled,optional"`
	DefaultLevel string           `hcl:"default_level,optional"`
	Slack        []SlackConfig    `hcl:"slack,block"`
	Email        []EmailConfig    `hcl:"email,block"`
	Webhook      []WebhookConfig  `hcl:"webhook,block"`
	Discord      []DiscordConfig  `hcl:"discord,block"`
	Teams        []TeamsConfig    `hcl:"teams,block"`
}

// SlackConfig defines Slack notification settings
type SlackConfig struct {
	Name             string            `hcl:"name,label"`
	WebhookURL       string            `hcl:"webhook_url"`
	Channel          string            `hcl:"channel,optional"`
	Username         string            `hcl:"username,optional"`
	IconEmoji        string            `hcl:"icon_emoji,optional"`
	NotifyOnStart    bool              `hcl:"notify_on_start,optional"`
	NotifyOnSuccess  bool              `hcl:"notify_on_success,optional"`
	NotifyOnFailure  bool              `hcl:"notify_on_failure,optional"`
	NotifyOnStep     bool              `hcl:"notify_on_step_failure,optional"`
	Templates        *TemplateConfig   `hcl:"templates,block"`
	Formatting       *FormattingConfig `hcl:"formatting,block"`
}

// EmailConfig defines email notification settings
type EmailConfig struct {
	Name         string   `hcl:"name,label"`
	SMTPServer   string   `hcl:"smtp_server"`
	SMTPPort     int      `hcl:"smtp_port"`
	Username     string   `hcl:"username"`
	Password     string   `hcl:"password"`
	FromAddress  string   `hcl:"from_address"`
	ToAddresses  []string `hcl:"to_addresses"`
	UseSSL       bool     `hcl:"use_ssl,optional"`
	UseTLS       bool     `hcl:"use_tls,optional"`
}

// WebhookConfig defines webhook notification settings
type WebhookConfig struct {
	Name        string            `hcl:"name,label"`
	URL         string            `hcl:"url"`
	Method      string            `hcl:"method,optional"`
	Headers     map[string]string `hcl:"headers,optional"`
	ContentType string            `hcl:"content_type,optional"`
	Template    string            `hcl:"template,optional"`
}

// DiscordConfig defines Discord notification settings
type DiscordConfig struct {
	Name            string `hcl:"name,label"`
	WebhookURL      string `hcl:"webhook_url"`
	Username        string `hcl:"username,optional"`
	AvatarURL       string `hcl:"avatar_url,optional"`
	NotifyOnStart   bool   `hcl:"notify_on_start,optional"`
	NotifyOnSuccess bool   `hcl:"notify_on_success,optional"`
	NotifyOnFailure bool   `hcl:"notify_on_failure,optional"`
}

// TeamsConfig defines Microsoft Teams notification settings
type TeamsConfig struct {
	Name            string `hcl:"name,label"`
	WebhookURL      string `hcl:"webhook_url"`
	NotifyOnStart   bool   `hcl:"notify_on_start,optional"`
	NotifyOnSuccess bool   `hcl:"notify_on_success,optional"`
	NotifyOnFailure bool   `hcl:"notify_on_failure,optional"`
}

// TemplateConfig defines notification templates
type TemplateConfig struct {
	FlowStart       string `hcl:"flow_start,optional"`
	FlowSuccess     string `hcl:"flow_success,optional"`
	FlowFailure     string `hcl:"flow_failure,optional"`
	StepStart       string `hcl:"step_start,optional"`
	StepSuccess     string `hcl:"step_success,optional"`
	StepFailure     string `hcl:"step_failure,optional"`
}

// FormattingConfig defines formatting options
type FormattingConfig struct {
	UseBlocks          bool   `hcl:"use_blocks,optional"`
	AddContext         bool   `hcl:"add_context,optional"`
	IncludeRunMetadata bool   `hcl:"include_run_metadata,optional"`
	IncludeLinkToLogs  bool   `hcl:"include_link_to_logs,optional"`
	ColorSuccess       string `hcl:"color_success,optional"`
	ColorFailure       string `hcl:"color_failure,optional"`
	ColorWarning       string `hcl:"color_warning,optional"`
}

// PluginConfig defines plugin settings
type PluginConfig struct {
	LocalPath    string              `hcl:"local_path,optional"`
	Repositories []PluginRepository  `hcl:"repository,block"`
	Cache        *PluginCacheConfig  `hcl:"cache,block"`
	AutoInstall  bool                `hcl:"auto_install,optional"`
}

// PluginRepository defines a plugin repository
type PluginRepository struct {
	Name     string `hcl:"name,label"`
	URL      string `hcl:"url"`
	Branch   string `hcl:"branch,optional"`
	Token    string `hcl:"token,optional"`
	Priority int    `hcl:"priority,optional"`
}

// PluginCacheConfig defines plugin cache settings
type PluginCacheConfig struct {
	Enabled    bool   `hcl:"enabled,optional"`
	Path       string `hcl:"path,optional"`
	MaxSize    string `hcl:"max_size,optional"`
	TTL        string `hcl:"ttl,optional"`
}

// StateConfig defines state management settings
type StateConfig struct {
	Backend      string            `hcl:"backend"`
	BackendConfig map[string]string `hcl:"backend_config,optional"`
	Locking      bool              `hcl:"locking,optional"`
	Encryption   bool              `hcl:"encryption,optional"`
}

// EnvironmentConfig defines environment-specific settings
type EnvironmentConfig struct {
	Name      string            `hcl:"name,label"`
	Variables map[string]string `hcl:"variables,optional"`
	Secrets   map[string]string `hcl:"secrets,optional"`
	Tags      []string          `hcl:"tags,optional"`
}

// Load loads configuration from HCL files
func Load(path string) (*Config, error) {
	if path == "" {
		path = findConfigFile()
	}

	config := &Config{}
	
	// Expand home directory if present
	expandedPath, err := homedir.Expand(path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}

	// Read and parse the HCL file
	err = hclsimple.DecodeFile(expandedPath, nil, config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Apply environment variable overrides
	config.applyEnvironmentOverrides()

	// Set defaults
	config.setDefaults()

	return config, nil
}

// findConfigFile searches for configuration file in standard locations
func findConfigFile() string {
	// Search order:
	// 1. corynth.hcl in current directory
	// 2. .corynth/config.hcl in current directory
	// 3. ~/.corynth/config.hcl
	// 4. /etc/corynth/config.hcl

	searchPaths := []string{
		"corynth.hcl",
		".corynth/config.hcl",
		"~/.corynth/config.hcl",
		"/etc/corynth/config.hcl",
	}

	for _, path := range searchPaths {
		expandedPath, _ := homedir.Expand(path)
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath
		}
	}

	return "corynth.hcl" // Default to current directory
}

// applyEnvironmentOverrides applies environment variable overrides
func (c *Config) applyEnvironmentOverrides() {
	// Override configuration with environment variables
	// Format: CORYNTH_<SECTION>_<KEY>
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "CORYNTH_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimPrefix(parts[0], "CORYNTH_")
				value := parts[1]
				c.applyOverride(key, value)
			}
		}
	}
}

// applyOverride applies a single environment variable override
func (c *Config) applyOverride(key, value string) {
	// Parse the key and apply the override
	// Example: SLACK_WEBHOOK_URL -> notifications.slack[0].webhook_url
	parts := strings.Split(strings.ToLower(key), "_")
	
	if len(parts) >= 2 {
		switch parts[0] {
		case "slack":
			if c.Notifications != nil && len(c.Notifications.Slack) > 0 {
				if parts[1] == "webhook" && len(parts) > 2 && parts[2] == "url" {
					c.Notifications.Slack[0].WebhookURL = value
				}
			}
		case "email":
			if c.Notifications != nil && len(c.Notifications.Email) > 0 {
				switch parts[1] {
				case "smtp":
					if len(parts) > 2 && parts[2] == "server" {
						c.Notifications.Email[0].SMTPServer = value
					}
				case "username":
					c.Notifications.Email[0].Username = value
				case "password":
					c.Notifications.Email[0].Password = value
				}
			}
		}
	}
}

// setDefaults sets default values for configuration
func (c *Config) setDefaults() {
	if c.Version == "" {
		c.Version = "1.0"
	}

	if c.Notifications != nil {
		if c.Notifications.DefaultLevel == "" {
			c.Notifications.DefaultLevel = "info"
		}
	}

	if c.Plugins != nil {
		if c.Plugins.LocalPath == "" {
			c.Plugins.LocalPath = "bin/plugins"
		}
		
		// Add default repository if none configured
		if len(c.Plugins.Repositories) == 0 {
			c.Plugins.Repositories = []PluginRepository{
				{
					Name:     "official",
					URL:      "https://github.com/corynth/plugins",
					Branch:   "main",
					Priority: 1,
				},
			}
		}
		
		if c.Plugins.Cache != nil {
			if c.Plugins.Cache.Path == "" {
				home, _ := homedir.Dir()
				c.Plugins.Cache.Path = filepath.Join(home, ".corynth", "cache")
			}
			if c.Plugins.Cache.MaxSize == "" {
				c.Plugins.Cache.MaxSize = "1GB"
			}
			if c.Plugins.Cache.TTL == "" {
				c.Plugins.Cache.TTL = "24h"
			}
		} else {
			// Initialize cache config with defaults
			home, _ := homedir.Dir()
			c.Plugins.Cache = &PluginCacheConfig{
				Enabled: true,
				Path:    filepath.Join(home, ".corynth", "cache"),
				TTL:     "24h",
				MaxSize: "1GB",
			}
		}
	} else {
		// Initialize plugins config with defaults
		home, _ := homedir.Dir()
		c.Plugins = &PluginConfig{
			LocalPath:   "bin/plugins",
			AutoInstall: true,
			Repositories: []PluginRepository{
				{
					Name:     "official",
					URL:      "https://github.com/corynth/plugins",
					Branch:   "main",
					Priority: 1,
				},
			},
			Cache: &PluginCacheConfig{
				Enabled: true,
				Path:    filepath.Join(home, ".corynth", "cache"),
				TTL:     "24h",
				MaxSize: "1GB",
			},
		}
	}

	if c.State != nil {
		if c.State.Backend == "" {
			c.State.Backend = "local"
		}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate notification configuration
	if c.Notifications != nil {
		for _, slack := range c.Notifications.Slack {
			if slack.WebhookURL == "" {
				return fmt.Errorf("slack notification '%s' requires webhook_url", slack.Name)
			}
		}
		
		for _, email := range c.Notifications.Email {
			if email.SMTPServer == "" || email.FromAddress == "" {
				return fmt.Errorf("email notification '%s' requires smtp_server and from_address", email.Name)
			}
		}
	}

	// Validate plugin configuration
	if c.Plugins != nil {
		for _, repo := range c.Plugins.Repositories {
			if repo.URL == "" {
				return fmt.Errorf("plugin repository '%s' requires url", repo.Name)
			}
		}
	}

	return nil
}

// GetEnvironment returns environment-specific configuration
func (c *Config) GetEnvironment(name string) *EnvironmentConfig {
	for _, env := range c.Environments {
		if env.Name == name {
			return &env
		}
	}
	return nil
}