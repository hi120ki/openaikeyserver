package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"golang.org/x/oauth2"
)

func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the OAuth2 authorization code from the request.
	code := r.URL.Query().Get("code")
	if code == "" {
		h.handleError(w, r, errors.New("no authorization code provided"), http.StatusBadRequest, "Authorization code is required")
		return
	}

	// Exchange the authorization code for an access token.
	token, err := h.oauth2Config.Exchange(ctx, code)
	if err != nil {
		var retrieveErr *oauth2.RetrieveError
		if errors.As(err, &retrieveErr) {
			slog.Info("oauth2 retrieve error, redirecting to root", "error", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to exchange authorization code")
		return
	}

	// Extract the ID Token from the access token.
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		h.handleError(w, r, errors.New("id_token not found in token response"), http.StatusInternalServerError, "Invalid token response")
		return
	}

	// Verify the ID Token.
	projectName, serviceAccountName, err := h.oidc.ExtractGoogleIDToken(ctx, h.oauth2Config.ClientID, idToken)
	if err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to verify ID token")
		return
	}

	// Check if the user is allowed to access the service.
	// First check if the user's email is in the allowed users list
	userAllowed := slices.Contains(*h.AllowedUsers, serviceAccountName)

	// If not in allowed users, check if the user's email domain is in the allowed domains list
	if !userAllowed {
		// Extract domain from email
		parts := strings.Split(serviceAccountName, "@")
		if len(parts) == 2 {
			domain := parts[1]
			userAllowed = slices.Contains(*h.AllowedDomains, domain)
		}
	}

	if !userAllowed {
		h.handleError(w, r, errors.New("user not allowed"), http.StatusForbidden, "User is not allowed to access this service")
		return
	}

	// Create a new API key using the management client.
	key, err := h.management.CreateAPIKey(ctx, projectName, serviceAccountName)
	if err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	// Serve HTML page with copyable API key.
	html := fmt.Sprintf(`
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>OpenAI API Key</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body class="bg-light">
  <div class="container py-5">
    <h1 class="mb-4">Your OpenAI API Key</h1>
    <div class="mb-3">
      <textarea id="tokenBox" class="form-control" rows="4" readonly>%s</textarea>
    </div>
    <button class="btn btn-primary" onclick="copyToken()">Copy to Clipboard</button>
  </div>
  <script>
    function copyToken() {
      const box = document.getElementById('tokenBox');
      box.select();
      box.setSelectionRange(0, 99999); // For mobile devices
      document.execCommand('copy');
    }
  </script>
</body>
</html>`, key)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := fmt.Fprint(w, html); err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to write response")
		return
	}
}
