package handler

import (
	"net/http"

	"golang.org/x/oauth2"
)

func (h *Handler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	// Generate a secure random state and store it in a cookie
	state, err := h.generateStateOauthCookie(w, r)
	if err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to generate OAuth state")
		return
	}

	// Generate the URL for the OAuth2 consent page with the secure state
	url := h.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// http.StatusFound is general redirect code.
	http.Redirect(w, r, url, http.StatusFound)
}
