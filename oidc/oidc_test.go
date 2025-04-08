package oidc

import (
	"context"
	"errors"
	"testing"
)

func TestNewOIDC(t *testing.T) {
	// Test data
	defaultProjectName := "test-project"
	allowedUsers := &[]string{"user1@example.com", "user2@example.com"}
	allowedDomains := &[]string{"example.com", "test.com"}
	googleTokenIssuerURL := "https://accounts.google.com"
	googleTokenJwksURL := "https://www.googleapis.com/oauth2/v3/certs"

	// Create OIDC instance
	oidcClient := NewOIDC(defaultProjectName, allowedUsers, allowedDomains, googleTokenIssuerURL, googleTokenJwksURL)

	// Verify the instance was created correctly
	if oidcClient == nil {
		t.Fatal("Expected non-nil OIDC instance")
	}

	// Test GetDefaultProjectName
	if name := oidcClient.GetDefaultProjectName(); name != defaultProjectName {
		t.Errorf("GetDefaultProjectName() = %v, want %v", name, defaultProjectName)
	}
}

func TestIsUserAllowed(t *testing.T) {
	// Test data
	allowedUsers := &[]string{"user1@example.com", "user2@example.com"}
	allowedDomains := &[]string{"example.com", "test.com"}
	oidcClient := NewOIDC("test-project", allowedUsers, allowedDomains, "", "")

	tests := []struct {
		name            string
		email           string
		hd              string
		expectedAllowed bool
	}{
		{
			name:            "Allowed user by email",
			email:           "user1@example.com",
			hd:              "example.com",
			expectedAllowed: true,
		},
		{
			name:            "Allowed user by domain",
			email:           "newuser@example.com",
			hd:              "example.com",
			expectedAllowed: true,
		},
		{
			name:            "Not allowed user",
			email:           "user@otherdomain.com",
			hd:              "otherdomain.com",
			expectedAllowed: false,
		},
		{
			name:            "Email with allowed domain but hd mismatch",
			email:           "user@example.com",
			hd:              "otherdomain.com",
			expectedAllowed: false,
		},
		{
			name:            "Email with empty domain part",
			email:           "useronly",
			hd:              "",
			expectedAllowed: false,
		},
		{
			name:            "Empty hd field",
			email:           "user@example.com",
			hd:              "",
			expectedAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := oidcClient.isUserAllowed(tt.email, tt.hd)
			if allowed != tt.expectedAllowed {
				t.Errorf("isUserAllowed(%q, %q) = %v, want %v", tt.email, tt.hd, allowed, tt.expectedAllowed)
			}
		})
	}
}

// MockTokenVerifier is a mock implementation of TokenVerifier for testing
type MockTokenVerifier struct {
	mockVerifyTokenFunc func(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error)
}

// VerifyToken implements the TokenVerifier interface for testing
func (m *MockTokenVerifier) VerifyToken(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error) {
	if m.mockVerifyTokenFunc != nil {
		return m.mockVerifyTokenFunc(ctx, aud, idToken)
	}
	return nil, nil
}

// Save the original function to restore it after tests
var originalCreateTokenVerifier = createTokenVerifier

// Helper function to set up a test with a mock verifier
func setupTokenVerifierTest(mockFunc func(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error)) func() {
	// Create a mock verifier
	mockVerifier := &MockTokenVerifier{
		mockVerifyTokenFunc: mockFunc,
	}

	// Override the createTokenVerifier function
	createTokenVerifier = func(issuerURL, jwksURL string) TokenVerifier {
		return mockVerifier
	}

	// Return a cleanup function
	return func() {
		createTokenVerifier = originalCreateTokenVerifier
	}
}

func TestExtractGoogleIDToken_UserNotAllowed(t *testing.T) {
	// Create OIDC client
	oidcClient := &OIDC{
		defaultProjectName:   "test-project",
		allowedUsers:         &[]string{"user1@example.com", "user2@example.com"},
		allowedDomains:       &[]string{"example.com", "test.com"},
		googleTokenIssuerURL: "https://accounts.google.com",
		googleTokenJwksURL:   "https://www.googleapis.com/oauth2/v3/certs",
	}

	// Setup mock verifier
	cleanup := setupTokenVerifierTest(func(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error) {
		return &GoogleIDTokenClaims{
			Email:         "unauthorized@otherdomain.com",
			EmailVerified: true,
			Hd:            "otherdomain.com",
		}, nil
	})
	defer cleanup()

	// Test ExtractGoogleIDToken with unauthorized user
	_, _, err := oidcClient.ExtractGoogleIDToken(context.Background(), "client-id", "fake-token")
	if err == nil {
		t.Error("Expected error for unauthorized user, got nil")
	}
}

func TestExtractGoogleIDToken_EmailNotVerified(t *testing.T) {
	// Create OIDC client
	oidcClient := &OIDC{
		defaultProjectName:   "test-project",
		allowedUsers:         &[]string{"user1@example.com", "user2@example.com"},
		allowedDomains:       &[]string{"example.com", "test.com"},
		googleTokenIssuerURL: "https://accounts.google.com",
		googleTokenJwksURL:   "https://www.googleapis.com/oauth2/v3/certs",
	}

	// Setup mock verifier
	cleanup := setupTokenVerifierTest(func(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error) {
		return &GoogleIDTokenClaims{
			Email:         "user1@example.com",
			EmailVerified: false,
			Hd:            "example.com",
		}, nil
	})
	defer cleanup()

	// Test ExtractGoogleIDToken with unverified email
	_, _, err := oidcClient.ExtractGoogleIDToken(context.Background(), "client-id", "fake-token")
	if err == nil {
		t.Error("Expected error for unverified email, got nil")
	}
}

func TestExtractGoogleIDToken_VerifierError(t *testing.T) {
	// Create OIDC client
	oidcClient := &OIDC{
		defaultProjectName:   "test-project",
		allowedUsers:         &[]string{"user1@example.com", "user2@example.com"},
		allowedDomains:       &[]string{"example.com", "test.com"},
		googleTokenIssuerURL: "https://accounts.google.com",
		googleTokenJwksURL:   "https://www.googleapis.com/oauth2/v3/certs",
	}

	// Setup mock verifier
	cleanup := setupTokenVerifierTest(func(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error) {
		return nil, errors.New("verification error")
	})
	defer cleanup()

	// Test ExtractGoogleIDToken with verifier error
	_, _, err := oidcClient.ExtractGoogleIDToken(context.Background(), "client-id", "fake-token")
	if err == nil {
		t.Error("Expected error from verifier, got nil")
	}
}

func TestExtractGoogleIDToken_Success(t *testing.T) {
	// Create OIDC client
	oidcClient := &OIDC{
		defaultProjectName:   "test-project",
		allowedUsers:         &[]string{"user1@example.com", "user2@example.com"},
		allowedDomains:       &[]string{"example.com", "test.com"},
		googleTokenIssuerURL: "https://accounts.google.com",
		googleTokenJwksURL:   "https://www.googleapis.com/oauth2/v3/certs",
	}

	// Setup mock verifier
	cleanup := setupTokenVerifierTest(func(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error) {
		return &GoogleIDTokenClaims{
			Email:         "user1@example.com",
			EmailVerified: true,
			Hd:            "example.com",
		}, nil
	})
	defer cleanup()

	// Test ExtractGoogleIDToken with authorized user
	projectName, email, err := oidcClient.ExtractGoogleIDToken(context.Background(), "client-id", "fake-token")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if projectName != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", projectName)
	}
	if email != "user1@example.com" {
		t.Errorf("Expected email 'user1@example.com', got '%s'", email)
	}
}
