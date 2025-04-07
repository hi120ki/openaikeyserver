package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hi120ki/monorepo/projects/openaikeyserver/client"
	"github.com/hi120ki/monorepo/projects/openaikeyserver/config"
	"github.com/hi120ki/monorepo/projects/openaikeyserver/handler"
	"github.com/hi120ki/monorepo/projects/openaikeyserver/management"
	"github.com/hi120ki/monorepo/projects/openaikeyserver/oidc"
)

// Server handles HTTP requests and manages the application lifecycle.
type Server struct {
	config     *config.Config
	server     *http.Server
	handler    *handler.Handler
	management *management.Management
	oidc       *oidc.OIDC
	shutdown   chan struct{}
}

// NewServer initializes a new server with the provided configuration.
func NewServer(cfg *config.Config) (*Server, error) {
	openaiClient := client.NewClient(
		cfg.GetOpenAIManagementKey(),
		&http.Client{
			Timeout: cfg.GetTimeout(),
		},
	)
	managementClient := management.NewManagement(
		openaiClient,
		cfg.GetExpiration(),
	)
	oidcClient := oidc.NewOIDC(
		cfg.GetDefaultProjectName(),
		cfg.GetAllowedUsers(),
		cfg.GetAllowedDomains(),
	)

	h := handler.NewHandler(
		cfg.GetAllowedUsers(),
		cfg.GetAllowedDomains(),
		cfg.GetClientID(),
		cfg.GetClientSecret(),
		cfg.GetRedirectURI(),
		managementClient,
		oidcClient,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", h.HandleRoot)
	mux.HandleFunc("/oauth2/callback", h.HandleOAuthCallback)
	mux.HandleFunc("/revoke", h.HandleRevoke)

	server := &http.Server{
		Addr:    ":" + cfg.GetPort(),
		Handler: mux,
	}

	return &Server{
		config:     cfg,
		server:     server,
		handler:    h,
		management: managementClient,
		oidc:       oidcClient,
		shutdown:   make(chan struct{}),
	}, nil
}

// Start launches the HTTP server and sets up graceful shutdown handling.
func (s *Server) Start() error {
	// Graceful shutdown setup
	go s.handleShutdown()

	// Start cleanup routine
	go s.startCleanupRoutine()

	slog.Info("starting server", "port", s.config.GetPort())
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-s.shutdown
	slog.Info("server shutdown gracefully")
	return nil
}

// handleShutdown listens for termination signals and shuts down the server gracefully.
func (s *Server) handleShutdown() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)
	<-sigint

	slog.Info("received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.GetTimeout())
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
	close(s.shutdown)
}

// startCleanupRoutine periodically runs API key cleanup based on the configured interval.
func (s *Server) startCleanupRoutine() {
	ticker := time.NewTicker(s.config.GetCleanupInterval())
	defer ticker.Stop()

	// Run cleanup immediately on startup
	ctx := context.Background()
	if err := s.management.CleanupAPIKey(ctx, s.oidc.GetDefaultProjectName()); err != nil {
		slog.Error("failed to cleanup API keys", "error", err)
	} else {
		slog.Info("API key cleanup completed")
	}

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			if err := s.management.CleanupAPIKey(ctx, s.oidc.GetDefaultProjectName()); err != nil {
				slog.Error("failed to cleanup API keys", "error", err)
			} else {
				slog.Info("API key cleanup completed")
			}
		case <-s.shutdown:
			return
		}
	}
}
