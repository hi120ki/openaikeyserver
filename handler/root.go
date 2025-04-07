package handler

import (
	"net/http"

	"golang.org/x/oauth2"
)

// HandleRoot initiates the OAuth2 authentication flow by redirecting to the consent page.
func (h *Handler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	// Create and store state token in cookie
	state, err := h.generateStateOauthCookie(w, r)
	if err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to generate OAuth state")
		return
	}

	// Build OAuth2 consent page URL
	url := h.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Redirect to OAuth2 consent page
	http.Redirect(w, r, url, http.StatusFound)
}
