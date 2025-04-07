package handler

import (
	"net/http"

	"golang.org/x/oauth2"
)

func (h *Handler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	// Generate the URL for the OAuth2 consent page.
	url := h.oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	// http.StatusFound is general redirect code.
	http.Redirect(w, r, url, http.StatusFound)
}
