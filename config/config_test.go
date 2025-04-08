package config

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	// Save original environment variables
	origAllowedUsers := os.Getenv("ALLOWED_USERS")
	origAllowedDomains := os.Getenv("ALLOWED_DOMAINS")
	origOpenAIManagementKey := os.Getenv("OPENAI_MANAGEMENT_KEY")
	origClientID := os.Getenv("CLIENT_ID")
	origClientSecret := os.Getenv("CLIENT_SECRET")
	origRedirectURI := os.Getenv("REDIRECT_URI")
	origDefaultProjectName := os.Getenv("DEFAULT_PROJECT_NAME")
	origPort := os.Getenv("PORT")
	origExpiration := os.Getenv("EXPIRATION")
	origCleanupInterval := os.Getenv("CLEANUP_INTERVAL")
	origTimeout := os.Getenv("TIMEOUT")

	// Restore environment variables after test
	defer func() {
		os.Setenv("ALLOWED_USERS", origAllowedUsers)
		os.Setenv("ALLOWED_DOMAINS", origAllowedDomains)
		os.Setenv("OPENAI_MANAGEMENT_KEY", origOpenAIManagementKey)
		os.Setenv("CLIENT_ID", origClientID)
		os.Setenv("CLIENT_SECRET", origClientSecret)
		os.Setenv("REDIRECT_URI", origRedirectURI)
		os.Setenv("DEFAULT_PROJECT_NAME", origDefaultProjectName)
		os.Setenv("PORT", origPort)
		os.Setenv("EXPIRATION", origExpiration)
		os.Setenv("CLEANUP_INTERVAL", origCleanupInterval)
		os.Setenv("TIMEOUT", origTimeout)
	}()

	tests := []struct {
		name          string
		envSetup      func()
		expectedError bool
	}{
		{
			name: "Valid configuration with allowed users",
			envSetup: func() {
				os.Setenv("ALLOWED_USERS", "user@example.com")
				os.Setenv("ALLOWED_DOMAINS", "")
				os.Setenv("OPENAI_MANAGEMENT_KEY", "test-key")
				os.Setenv("CLIENT_ID", "test-client-id")
				os.Setenv("CLIENT_SECRET", "test-client-secret")
				os.Setenv("REDIRECT_URI", "http://localhost:8080/callback")
			},
			expectedError: false,
		},
		{
			name: "Valid configuration with allowed domains",
			envSetup: func() {
				os.Setenv("ALLOWED_USERS", "")
				os.Setenv("ALLOWED_DOMAINS", "example.com")
				os.Setenv("OPENAI_MANAGEMENT_KEY", "test-key")
				os.Setenv("CLIENT_ID", "test-client-id")
				os.Setenv("CLIENT_SECRET", "test-client-secret")
				os.Setenv("REDIRECT_URI", "http://localhost:8080/callback")
			},
			expectedError: false,
		},
		{
			name: "Valid configuration with both allowed users and domains",
			envSetup: func() {
				os.Setenv("ALLOWED_USERS", "user@example.com")
				os.Setenv("ALLOWED_DOMAINS", "example.com")
				os.Setenv("OPENAI_MANAGEMENT_KEY", "test-key")
				os.Setenv("CLIENT_ID", "test-client-id")
				os.Setenv("CLIENT_SECRET", "test-client-secret")
				os.Setenv("REDIRECT_URI", "http://localhost:8080/callback")
			},
			expectedError: false,
		},
		{
			name: "Missing allowed users and domains",
			envSetup: func() {
				os.Setenv("ALLOWED_USERS", "")
				os.Setenv("ALLOWED_DOMAINS", "")
				os.Setenv("OPENAI_MANAGEMENT_KEY", "test-key")
				os.Setenv("CLIENT_ID", "test-client-id")
				os.Setenv("CLIENT_SECRET", "test-client-secret")
				os.Setenv("REDIRECT_URI", "http://localhost:8080/callback")
			},
			expectedError: true,
		},
		{
			name: "Missing OpenAI management key",
			envSetup: func() {
				os.Setenv("ALLOWED_USERS", "user@example.com")
				os.Setenv("ALLOWED_DOMAINS", "")
				os.Setenv("OPENAI_MANAGEMENT_KEY", "")
				os.Setenv("CLIENT_ID", "test-client-id")
				os.Setenv("CLIENT_SECRET", "test-client-secret")
				os.Setenv("REDIRECT_URI", "http://localhost:8080/callback")
			},
			expectedError: true,
		},
		{
			name: "Missing client ID",
			envSetup: func() {
				os.Setenv("ALLOWED_USERS", "user@example.com")
				os.Setenv("ALLOWED_DOMAINS", "")
				os.Setenv("OPENAI_MANAGEMENT_KEY", "test-key")
				os.Setenv("CLIENT_ID", "")
				os.Setenv("CLIENT_SECRET", "test-client-secret")
				os.Setenv("REDIRECT_URI", "http://localhost:8080/callback")
			},
			expectedError: true,
		},
		{
			name: "Missing client secret",
			envSetup: func() {
				os.Setenv("ALLOWED_USERS", "user@example.com")
				os.Setenv("ALLOWED_DOMAINS", "")
				os.Setenv("OPENAI_MANAGEMENT_KEY", "test-key")
				os.Setenv("CLIENT_ID", "test-client-id")
				os.Setenv("CLIENT_SECRET", "")
				os.Setenv("REDIRECT_URI", "http://localhost:8080/callback")
			},
			expectedError: true,
		},
		{
			name: "Missing redirect URI",
			envSetup: func() {
				os.Setenv("ALLOWED_USERS", "user@example.com")
				os.Setenv("ALLOWED_DOMAINS", "")
				os.Setenv("OPENAI_MANAGEMENT_KEY", "test-key")
				os.Setenv("CLIENT_ID", "test-client-id")
				os.Setenv("CLIENT_SECRET", "test-client-secret")
				os.Setenv("REDIRECT_URI", "")
			},
			expectedError: true,
		},
		{
			name: "With custom values for optional parameters",
			envSetup: func() {
				os.Setenv("ALLOWED_USERS", "user@example.com")
				os.Setenv("ALLOWED_DOMAINS", "")
				os.Setenv("OPENAI_MANAGEMENT_KEY", "test-key")
				os.Setenv("CLIENT_ID", "test-client-id")
				os.Setenv("CLIENT_SECRET", "test-client-secret")
				os.Setenv("REDIRECT_URI", "http://localhost:8080/callback")
				os.Setenv("DEFAULT_PROJECT_NAME", "custom-project")
				os.Setenv("PORT", "9000")
				os.Setenv("EXPIRATION", "43200")
				os.Setenv("CLEANUP_INTERVAL", "1800")
				os.Setenv("TIMEOUT", "30")
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Unsetenv("ALLOWED_USERS")
			os.Unsetenv("ALLOWED_DOMAINS")
			os.Unsetenv("OPENAI_MANAGEMENT_KEY")
			os.Unsetenv("CLIENT_ID")
			os.Unsetenv("CLIENT_SECRET")
			os.Unsetenv("REDIRECT_URI")
			os.Unsetenv("DEFAULT_PROJECT_NAME")
			os.Unsetenv("PORT")
			os.Unsetenv("EXPIRATION")
			os.Unsetenv("CLEANUP_INTERVAL")
			os.Unsetenv("TIMEOUT")

			// Set up test environment
			tt.envSetup()

			// Test NewConfig
			cfg, err := NewConfig()
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Verify config values
			if cfg == nil {
				t.Errorf("Expected non-nil config but got nil")
				return
			}

			// Check that Get returns the same config
			if cfg.Get() != cfg {
				t.Errorf("Expected Get() to return the same config instance")
			}
		})
	}
}

func TestConfigGetters(t *testing.T) {
	// Create a test config
	cfg := &Config{
		AllowedUsers:         "user1@example.com,user2@example.com",
		AllowedDomains:       "example.com,test.com",
		OpenAIManagementKey:  "test-key",
		ClientID:             "test-client-id",
		ClientSecret:         "test-client-secret",
		RedirectURI:          "http://localhost:8080/callback",
		DefaultProjectName:   "test-project",
		Port:                 "9000",
		Expiration:           43200,
		CleanupInterval:      1800,
		Timeout:              30,
		GoogleTokenIssuerURL: "https://accounts.google.com",
		GoogleTokenJwksURL:   "https://www.googleapis.com/oauth2/v3/certs",
	}

	// Test GetAllowedUsers
	allowedUsers := cfg.GetAllowedUsers()
	if len(*allowedUsers) != 2 || (*allowedUsers)[0] != "user1@example.com" || (*allowedUsers)[1] != "user2@example.com" {
		t.Errorf("GetAllowedUsers() = %v, want [user1@example.com user2@example.com]", *allowedUsers)
	}

	// Test GetAllowedDomains
	allowedDomains := cfg.GetAllowedDomains()
	if len(*allowedDomains) != 2 || (*allowedDomains)[0] != "example.com" || (*allowedDomains)[1] != "test.com" {
		t.Errorf("GetAllowedDomains() = %v, want [example.com test.com]", *allowedDomains)
	}

	// Test GetOpenAIManagementKey
	if key := cfg.GetOpenAIManagementKey(); key != "test-key" {
		t.Errorf("GetOpenAIManagementKey() = %v, want test-key", key)
	}

	// Test GetClientID
	if id := cfg.GetClientID(); id != "test-client-id" {
		t.Errorf("GetClientID() = %v, want test-client-id", id)
	}

	// Test GetClientSecret
	if secret := cfg.GetClientSecret(); secret != "test-client-secret" {
		t.Errorf("GetClientSecret() = %v, want test-client-secret", secret)
	}

	// Test GetRedirectURI
	if uri := cfg.GetRedirectURI(); uri != "http://localhost:8080/callback" {
		t.Errorf("GetRedirectURI() = %v, want http://localhost:8080/callback", uri)
	}

	// Test GetDefaultProjectName
	if name := cfg.GetDefaultProjectName(); name != "test-project" {
		t.Errorf("GetDefaultProjectName() = %v, want test-project", name)
	}

	// Test GetPort
	if port := cfg.GetPort(); port != "9000" {
		t.Errorf("GetPort() = %v, want 9000", port)
	}

	// Test GetExpiration
	if exp := cfg.GetExpiration(); exp != 43200*time.Second {
		t.Errorf("GetExpiration() = %v, want %v", exp, 43200*time.Second)
	}

	// Test GetCleanupInterval
	if interval := cfg.GetCleanupInterval(); interval != 1800*time.Second {
		t.Errorf("GetCleanupInterval() = %v, want %v", interval, 1800*time.Second)
	}

	// Test GetTimeout
	if timeout := cfg.GetTimeout(); timeout != 30*time.Second {
		t.Errorf("GetTimeout() = %v, want %v", timeout, 30*time.Second)
	}

	// Test GetGoogleTokenIssuerURL
	if url := cfg.GetGoogleTokenIssuerURL(); url != "https://accounts.google.com" {
		t.Errorf("GetGoogleTokenIssuerURL() = %v, want https://accounts.google.com", url)
	}

	// Test GetGoogleTokenJwksURL
	if url := cfg.GetGoogleTokenJwksURL(); url != "https://www.googleapis.com/oauth2/v3/certs" {
		t.Errorf("GetGoogleTokenJwksURL() = %v, want https://www.googleapis.com/oauth2/v3/certs", url)
	}

	// Test empty allowed users and domains
	emptyCfg := &Config{
		AllowedUsers:   "",
		AllowedDomains: "",
	}

	emptyUsers := emptyCfg.GetAllowedUsers()
	if len(*emptyUsers) != 0 {
		t.Errorf("GetAllowedUsers() with empty string = %v, want []", *emptyUsers)
	}

	emptyDomains := emptyCfg.GetAllowedDomains()
	if len(*emptyDomains) != 0 {
		t.Errorf("GetAllowedDomains() with empty string = %v, want []", *emptyDomains)
	}
}
