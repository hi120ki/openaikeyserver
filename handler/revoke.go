package handler

import (
	"log/slog"
	"net/http"
)

func (h *Handler) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Clean up old API keys
	if err := h.management.CleanupAPIKey(ctx, h.oidc.GetDefaultProjectName()); err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to cleanup API keys")
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("API key cleanup completed successfully.")); err != nil {
		h.handleError(w, r, err, http.StatusInternalServerError, "Failed to write response")
		return
	}
	slog.Info("api key cleanup completed successfully")
}
