package oidc

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	googleTokenIssuerURL = "https://accounts.google.com"
	googleTokenJwksURL   = "https://www.googleapis.com/oauth2/v3/certs"
)

type OIDC struct {
	defaultProjectName string
}

func NewOIDC(defaultProjectName string) *OIDC {
	return &OIDC{
		defaultProjectName: defaultProjectName,
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
