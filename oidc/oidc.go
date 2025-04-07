package oidc

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	googleTokenIssuerURL = "https://accounts.google.com"
	googleTokenJwksURL   = "https://www.googleapis.com/oauth2/v3/certs"
)

type OIDC struct {
	defaultProjectName string
	allowedUsers       *[]string
	allowedDomains     *[]string
}

func NewOIDC(defaultProjectName string, allowedUsers *[]string, allowedDomains *[]string) *OIDC {
	return &OIDC{
		defaultProjectName: defaultProjectName,
		allowedUsers:       allowedUsers,
		allowedDomains:     allowedDomains,
	}
}

// GetDefaultProjectName returns the default project name
func (o *OIDC) GetDefaultProjectName() string {
	return o.defaultProjectName
}

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

// isUserAllowed checks if a user is allowed to access the service based on their email
// or domain being in the allowed lists.
func (o *OIDC) isUserAllowed(serviceAccountName, hd string) bool {
	// First check if the user's email is in the allowed users list
	if slices.Contains(*o.allowedUsers, serviceAccountName) {
		return true
	}

	// If not in allowed users, check if the user's email domain is in the allowed domains list
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
