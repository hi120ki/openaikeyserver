package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AllowedUsers        string `envconfig:"ALLOWED_USERS"`
	AllowedDomains      string `envconfig:"ALLOWED_DOMAINS"`
	OpenAIManagementKey string `envconfig:"OPENAI_MANAGEMENT_KEY"`
	ClientID            string `envconfig:"CLIENT_ID"`
	ClientSecret        string `envconfig:"CLIENT_SECRET"`
	RedirectURI         string `envconfig:"REDIRECT_URI"`
	DefaultProjectName  string `envconfig:"DEFAULT_PROJECT_NAME" default:"personal"`
	Port                string `envconfig:"PORT" default:"8080"`
	Expiration          int    `envconfig:"EXPIRATION" default:"86400"`      // 24 hours
	CleanupInterval     int    `envconfig:"CLEANUP_INTERVAL" default:"3600"` // 1 hour
	Timeout             int    `envconfig:"TIMEOUT" default:"10"`            // 10 seconds
}

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

func (c *Config) Get() *Config {
	return c
}

func (c *Config) GetAllowedUsers() *[]string {
	if c.AllowedUsers == "" {
		empty := []string{}
		return &empty
	}
	result := strings.Split(c.AllowedUsers, ",")
	return &result
}

func (c *Config) GetAllowedDomains() *[]string {
	if c.AllowedDomains == "" {
		empty := []string{}
		return &empty
	}
	result := strings.Split(c.AllowedDomains, ",")
	return &result
}

func (c *Config) GetOpenAIManagementKey() string {
	return c.OpenAIManagementKey
}

func (c *Config) GetClientID() string {
	return c.ClientID
}

func (c *Config) GetClientSecret() string {
	return c.ClientSecret
}

func (c *Config) GetRedirectURI() string {
	return c.RedirectURI
}

func (c *Config) GetDefaultProjectName() string {
	return c.DefaultProjectName
}

func (c *Config) GetPort() string {
	return c.Port
}

func (c *Config) GetExpiration() time.Duration {
	return time.Duration(c.Expiration) * time.Second
}

func (c *Config) GetCleanupInterval() time.Duration {
	return time.Duration(c.CleanupInterval) * time.Second
}

func (c *Config) GetTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}
