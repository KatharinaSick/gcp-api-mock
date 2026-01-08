// Package main is the entry point for the GCP API Mock server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/katharinasick/gcp-api-mock/internal/config"
	"github.com/katharinasick/gcp-api-mock/internal/server"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create and configure server
	srv := server.New(cfg)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting GCP API Mock server on %s", cfg.Address())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
