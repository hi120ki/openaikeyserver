package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/hi120ki/monorepo/projects/openaikeyserver/config"
	"github.com/hi120ki/monorepo/projects/openaikeyserver/server"
	"github.com/joho/godotenv"
)

func main() {
	// Setup logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		slog.Warn("failed to load .env file", "error", err)
	}

	// Load configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to create configuration: %v", err)
	}

	// Create and start server
	srv, err := server.NewServer(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	// Start server
	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
