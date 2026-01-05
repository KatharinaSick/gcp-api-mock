package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/KatharinaSick/gcp-api-mock/internal/api"
	"github.com/KatharinaSick/gcp-api-mock/internal/service"
	"github.com/KatharinaSick/gcp-api-mock/internal/store"
	"github.com/KatharinaSick/gcp-api-mock/internal/store/memory"
)

// version is set at build time via ldflags
var version = "dev"

func main() {
	// Configuration from environment variables
	port := getEnv("PORT", "8080")
	projectID := getEnv("PROJECT_ID", "playground")
	logLevel := getEnv("LOG_LEVEL", "info")
	shutdownTimeout := getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second)

	// Configure logging based on log level
	configureLogging(logLevel)

	// Create store factory with in-memory stores
	storeFactory := store.NewStoreFactory(func() store.Store {
		return memory.New()
	})

	// Create services
	bucketService := service.NewBucketService(storeFactory)
	objectService := service.NewObjectService(storeFactory)

	// Link services for bucket-object relationship
	bucketService.SetObjectService(objectService)
	objectService.SetBucketService(bucketService)

	// Create request logger for dashboard
	requestLogger := api.NewRequestLogger(50)

	// Create router
	router := api.NewRouter()

	// Create and register handlers
	bucketHandler := api.NewBucketHandler(bucketService)
	bucketHandler.RegisterRoutes(router)

	objectHandler := api.NewObjectHandler(objectService)
	objectHandler.RegisterRoutes(router)

	// Register health check endpoint (before dashboard to avoid conflicts)
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Create and register dashboard handler
	dashboardHandler := api.NewDashboardHandler(bucketService, objectService, requestLogger, projectID)
	dashboardHandler.RegisterRoutes(router)

	// Root path redirects to dashboard
	router.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})

	// Apply middleware chain: CORS -> Logging with RequestLogger -> Router
	handler := api.CORSMiddleware(api.LoggingMiddlewareWithLogger(router, requestLogger))

	// Create server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel for shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		log.Printf("Starting GCP API Mock server v%s on port %s", version, port)
		log.Printf("Log level: %s", logLevel)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-stop
	log.Println("Shutdown signal received, initiating graceful shutdown...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Graceful shutdown failed: %v", err)
		log.Println("Forcing server close...")
		if err := server.Close(); err != nil {
			log.Printf("Server close error: %v", err)
		}
	} else {
		log.Println("Server gracefully stopped")
	}
}

// getEnv returns the value of an environment variable or a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvDuration returns a duration from an environment variable or a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Invalid duration for %s: %s, using default", key, value)
	}
	return defaultValue
}

// configureLogging sets up logging based on the log level
func configureLogging(level string) {
	level = strings.ToLower(level)
	switch level {
	case "debug":
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	case "info":
		log.SetFlags(log.Ldate | log.Ltime)
	case "warn", "error":
		log.SetFlags(log.Ldate | log.Ltime)
	default:
		log.SetFlags(log.Ldate | log.Ltime)
	}
}
