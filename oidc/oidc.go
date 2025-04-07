package oidc

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	// googleTokenIssuerURL is the issuer URL for Google OIDC tokens.
	googleTokenIssuerURL = "https://accounts.google.com"
	// googleTokenJwksURL is the JWKS URL for Google OIDC tokens.
	googleTokenJwksURL = "https://www.googleapis.com/oauth2/v3/certs"
)

// OIDC handles OpenID Connect authentication and authorization.
type OIDC struct {
	defaultProjectName string    // Default project name for API key creation
	allowedUsers       *[]string // List of allowed user emails
	allowedDomains     *[]string // List of allowed email domains
}

// NewOIDC creates a new OIDC client with the specified configuration.
func NewOIDC(defaultProjectName string, allowedUsers *[]string, allowedDomains *[]string) *OIDC {
	return &OIDC{
		defaultProjectName: defaultProjectName,
		allowedUsers:       allowedUsers,
		allowedDomains:     allowedDomains,
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

// ExtractGoogleIDToken verifies a Google ID token and extracts the project name and service account email.
// It also checks if the user is allowed to access the service.
func (o *OIDC) ExtractGoogleIDToken(ctx context.Context, aud string, idToken string) (string, string, error) {
	claims, err := o.verifyGoogleOIDCToken(ctx, aud, idToken)
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

// verifyGoogleOIDCToken verifies a Google ID token and returns its claims.
func (o *OIDC) verifyGoogleOIDCToken(ctx context.Context, aud string, idToken string) (*GoogleIDTokenClaims, error) {
	config := &oidc.Config{
		ClientID: aud,
	}

	verifier := oidc.NewVerifier(googleTokenIssuerURL, oidc.NewRemoteKeySet(ctx, googleTokenJwksURL), config)

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
