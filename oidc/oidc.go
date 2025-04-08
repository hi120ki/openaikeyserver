package oidc

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

// OIDC handles OpenID Connect authentication and authorization.
type OIDC struct {
	defaultProjectName   string    // Default project name for API key creation
	allowedUsers         *[]string // List of allowed user emails
	allowedDomains       *[]string // List of allowed email domains
	googleTokenIssuerURL string    // Google token issuer URL
	googleTokenJwksURL   string    // Google token JWKS URL
}

// NewOIDC creates a new OIDC client with the specified configuration.
func NewOIDC(defaultProjectName string, allowedUsers *[]string, allowedDomains *[]string, googleTokenIssuerURL string, googleTokenJwksURL string) *OIDC {
	return &OIDC{
		defaultProjectName:   defaultProjectName,
		allowedUsers:         allowedUsers,
		allowedDomains:       allowedDomains,
		googleTokenIssuerURL: googleTokenIssuerURL,
		googleTokenJwksURL:   googleTokenJwksURL,
	}
}

// GetDefaultProjectName returns the configured default project name.
func (o *OIDC) GetDefaultProjectName() string {
	return o.defaultProjectName
}

// GoogleIDTokenClaims represents the claims in a Google ID token.
type GoogleIDTokenClaims struct {
	Aud           string `json:"aud"`
	Azp           string `json:"azp"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Exp           int    `json:"exp"`
	Iat           int    `json:"iat"`
	Iss           string `json:"iss"`
	Sub           string `json:"sub"`
	AtHash        string `json:"at_hash"`
	Hd            string `json:"hd"`
}

// TokenVerifier defines the interface for token verification
type TokenVerifier interface {
	VerifyToken(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error)
}

// DefaultTokenVerifier handles token verification
type DefaultTokenVerifier struct {
	issuerURL string
	jwksURL   string
}

// NewDefaultTokenVerifier creates a new DefaultTokenVerifier
func NewDefaultTokenVerifier(issuerURL, jwksURL string) *DefaultTokenVerifier {
	return &DefaultTokenVerifier{
		issuerURL: issuerURL,
		jwksURL:   jwksURL,
	}
}

// VerifyToken verifies a Google ID token and returns its claims
func (v *DefaultTokenVerifier) VerifyToken(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error) {
	config := &oidc.Config{
		ClientID: aud,
	}

	verifier := oidc.NewVerifier(v.issuerURL, oidc.NewRemoteKeySet(ctx, v.jwksURL), config)

	token, err := verifier.Verify(ctx, idToken)
	if err != nil {
		return nil, err
	}

	var claims GoogleIDTokenClaims
	if err := token.Claims(&claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

// For testing purposes
var createTokenVerifier = func(issuerURL, jwksURL string) TokenVerifier {
	return NewDefaultTokenVerifier(issuerURL, jwksURL)
}

// ExtractGoogleIDToken verifies a Google ID token and extracts the project name and service account email.
// It also checks if the user is allowed to access the service.
func (o *OIDC) ExtractGoogleIDToken(ctx context.Context, aud string, idToken string) (string, string, error) {
	// Create verifier
	verifier := createTokenVerifier(o.googleTokenIssuerURL, o.googleTokenJwksURL)

	// Verify token
	claims, err := verifier.VerifyToken(ctx, aud, idToken)
	if err != nil {
		return "", "", fmt.Errorf("verify id token: %w", err)
	}

	if !claims.EmailVerified {
		return "", "", fmt.Errorf("verify email")
	}

	if !o.isUserAllowed(claims.Email, claims.Hd) {
		return "", "", fmt.Errorf("user not allowed to access the service %s", claims.Email)
	}

	return o.defaultProjectName, claims.Email, nil
}

// isUserAllowed checks if a user is allowed based on email or domain.
func (o *OIDC) isUserAllowed(serviceAccountName, hd string) bool {
	// Check if email is in allowed users list
	if slices.Contains(*o.allowedUsers, serviceAccountName) {
		return true
	}

	// Check if domain is in allowed domains list
	parts := strings.Split(serviceAccountName, "@")
	if len(parts) == 2 {
		domain := parts[1]
		if domain == "" || domain != hd {
			return false
		}
		if slices.Contains(*o.allowedDomains, domain) {
			return true
		}
	}

	return false
}
