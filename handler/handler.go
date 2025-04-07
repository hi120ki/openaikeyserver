package handler

import (
	"log/slog"
	"net/http"

	"github.com/hi120ki/monorepo/projects/openaikeyserver/management"
	"github.com/hi120ki/monorepo/projects/openaikeyserver/oidc"
	"golang.org/x/oauth2"
)

// Handler holds configuration for handling OAuth2 flow
type Handler struct {
	AllowedUsers *[]string
	oauth2Config *oauth2.Config
	management   *management.Management
	oidc         *oidc.OIDC
}

// NewHandler creates a new Handler with the given config
func NewHandler(allowedUsers *[]string, clientID, clientSecret, redirectURI string, management *management.Management, oidc *oidc.OIDC) *Handler {
	return &Handler{
		AllowedUsers: allowedUsers,
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

// handleError logs the error and writes an appropriate HTTP response
func (h *Handler) handleError(w http.ResponseWriter, r *http.Request, err error, status int, msg string) {
	slog.Error(msg, "error", err, "path", r.URL.Path, "method", r.Method)
	http.Error(w, msg, status)
}
