package handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// HandleOAuthCallback processes OAuth2 callback requests, verifies tokens, and issues API keys.
func (h *Handler) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		h.handleError(w, r, errors.New("no authorization code provided"), http.StatusBadRequest, "Authorization code is required")
		return
	}

	// Verify state parameter
	receivedState := r.URL.Query().Get("state")
	if receivedState == "" {
		h.handleError(w, r, errors.New("no state provided"), http.StatusBadRequest, "State parameter is required")
		return
	}

	// Retrieve state cookie
	stateCookie, err := r.Cookie("oauthstate")
	if err != nil {
		h.handleError(w, r, err, http.StatusBadRequest, "State cookie not found")
		return
	}

	// Validate state parameter against cookie
	if stateCookie.Value != receivedState {
		http.Redirect(w, r, "/", http.StatusFound)
		slog.Debug("state mismatch, redirecting to root", "receivedState", receivedState, "cookieState", stateCookie.Value)
		return
	}

	// Remove state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// Exchange code for token
	token, err := h.oauth2Config.Exchange(ctx, code)
	if err != nil {
		var retrieveErr *oauth2.RetrieveError
		if errors.As(err, &retrieveErr) {
			slog.Debug("oauth2 retrieve error, redirecting to root", "error", err)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to exchange authorization code")
		return
	}

	// Extract ID token
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		h.handleError(w, r, errors.New("id_token not found in token response"), http.StatusInternalServerError, "Invalid token response")
		return
	}

	// Verify ID token and extract user info
	projectName, serviceAccountName, err := h.oidc.ExtractGoogleIDToken(ctx, h.oauth2Config.ClientID, idToken)
	if err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to verify ID token")
		return
	}

	// Generate API key
	key, expiration, err := h.management.CreateAPIKey(ctx, projectName, serviceAccountName)
	if err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to create API key")
		return
	}

	// Calculate and format expiration time in JST
	jst, _ := time.LoadLocation("Asia/Tokyo")
	expirationDateStr := expiration.In(jst).Format("2006/01/02 15:04:05")

	// Render response page with API key
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
    <p class="text-muted mb-3">Expires on: %s (JST)</p>
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
</html>`, expirationDateStr, key)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := fmt.Fprint(w, html); err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to write response")
		return
	}
}
