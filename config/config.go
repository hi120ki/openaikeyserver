package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	AllowedUsers         string `envconfig:"ALLOWED_USERS"`
	AllowedDomains       string `envconfig:"ALLOWED_DOMAINS"`
	OpenAIManagementKey  string `envconfig:"OPENAI_MANAGEMENT_KEY"`
	ClientID             string `envconfig:"CLIENT_ID"`
	ClientSecret         string `envconfig:"CLIENT_SECRET"`
	RedirectURI          string `envconfig:"REDIRECT_URI"`
	DefaultProjectName   string `envconfig:"DEFAULT_PROJECT_NAME" default:"personal"`
	Port                 string `envconfig:"PORT" default:"8080"`
	Expiration           int    `envconfig:"EXPIRATION" default:"86400"`      // 24 hours
	CleanupInterval      int    `envconfig:"CLEANUP_INTERVAL" default:"3600"` // 1 hour
	Timeout              int    `envconfig:"TIMEOUT" default:"10"`            // 10 seconds
	GoogleTokenIssuerURL string `envconfig:"GOOGLE_TOKEN_ISSUER_URL" default:"https://accounts.google.com"`
	GoogleTokenJwksURL   string `envconfig:"GOOGLE_TOKEN_AUDIENCE" default:"https://www.googleapis.com/oauth2/v3/certs"`
}

// NewConfig creates and validates a new configuration from environment variables.
func NewConfig() (*Config, error) {
	config := &Config{}
	if err := envconfig.Process("", config); err != nil {
		return nil, fmt.Errorf("failed to process env: %w", err)
	}
	if config.AllowedUsers == "" && config.AllowedDomains == "" {
		return nil, fmt.Errorf("either ALLOWED_USERS or ALLOWED_DOMAINS (or both) is required")
	}
	if config.OpenAIManagementKey == "" {
		return nil, fmt.Errorf("OPENAI_MANAGEMENT_KEY is required")
	}
	if config.ClientID == "" {
		return nil, fmt.Errorf("CLIENT_ID is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("CLIENT_SECRET is required")
	}
	if config.RedirectURI == "" {
		return nil, fmt.Errorf("REDIRECT_URI is required")
	}
	return config, nil
}

// Get returns the config instance.
func (c *Config) Get() *Config {
	return c
}

// GetAllowedUsers returns the list of allowed user emails.
func (c *Config) GetAllowedUsers() *[]string {
	if c.AllowedUsers == "" {
		empty := []string{}
		return &empty
	}
	result := strings.Split(c.AllowedUsers, ",")
	return &result
}

// GetAllowedDomains returns the list of allowed email domains.
func (c *Config) GetAllowedDomains() *[]string {
	if c.AllowedDomains == "" {
		empty := []string{}
		return &empty
	}
	result := strings.Split(c.AllowedDomains, ",")
	return &result
}

// GetOpenAIManagementKey returns the OpenAI management API key.
func (c *Config) GetOpenAIManagementKey() string {
	return c.OpenAIManagementKey
}

// GetClientID returns the OAuth client ID.
func (c *Config) GetClientID() string {
	return c.ClientID
}

// GetClientSecret returns the OAuth client secret.
func (c *Config) GetClientSecret() string {
	return c.ClientSecret
}

// GetRedirectURI returns the OAuth redirect URI.
func (c *Config) GetRedirectURI() string {
	return c.RedirectURI
}

// GetDefaultProjectName returns the default OpenAI project name.
func (c *Config) GetDefaultProjectName() string {
	return c.DefaultProjectName
}

// GetPort returns the HTTP server port.
func (c *Config) GetPort() string {
	return c.Port
}

// GetExpiration returns the API key expiration duration.
func (c *Config) GetExpiration() time.Duration {
	return time.Duration(c.Expiration) * time.Second
}

// GetCleanupInterval returns the interval for API key cleanup operations.
func (c *Config) GetCleanupInterval() time.Duration {
	return time.Duration(c.CleanupInterval) * time.Second
}

// GetTimeout returns the HTTP client timeout duration.
func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// GetGoogleTokenIssuerURL returns the Google token issuer URL.
func (c *Config) GetGoogleTokenIssuerURL() string {
	return c.GoogleTokenIssuerURL
}

// GetGoogleTokenJwksURL returns the Google token JWKS URL.
func (c *Config) GetGoogleTokenJwksURL() string {
	return c.GoogleTokenJwksURL
}
