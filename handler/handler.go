package handler

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"

	"github.com/hi120ki/monorepo/projects/openaikeyserver/management"
	"github.com/hi120ki/monorepo/projects/openaikeyserver/oidc"
	"golang.org/x/oauth2"
)

// Handler manages OAuth2 authentication flow and API key operations.
type Handler struct {
	allowedUsers   *[]string          // List of allowed user emails
	allowedDomains *[]string          // List of allowed email domains
	oauth2Config   *oauth2.Config     // OAuth2 configuration
	management     management.Manager // Management interface for API key operations
	oidc           *oidc.OIDC         // OIDC client for authentication
}

// NewHandler initializes a new handler with the provided configuration.
func NewHandler(allowedUsers *[]string, allowedDomains *[]string, clientID, clientSecret, redirectURI string, management management.Manager, oidc *oidc.OIDC) *Handler {
	return &Handler{
		allowedUsers:   allowedUsers,
		allowedDomains: allowedDomains,
		oauth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURI,
			Scopes:       []string{"email", "openid"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
		management: management,
		oidc:       oidc,
	}
}

// handleError logs errors and returns appropriate HTTP responses.
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error, status int, msg string) {
	slog.Error(msg, "error", err, "path", r.URL.Path, "method", r.Method)
	http.Error(w, msg, status)
}

// generateStateOauthCookie creates a secure random state token and stores it in a cookie.
func (h *Handler) generateStateOauthCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	cookie := &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set Secure flag if using HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	return state, nil
}
