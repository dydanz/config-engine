package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"config-engine/internal/handlers"
	"config-engine/internal/repository"
	"config-engine/internal/service"
	"config-engine/internal/validation"
)

const (
	defaultPort            = "8080"
	shutdownTimeout        = 15 * time.Second
	readTimeout            = 10 * time.Second
	writeTimeout           = 10 * time.Second
	idleTimeout            = 60 * time.Second
	readHeaderTimeout      = 5 * time.Second
)

func main() {
	// Parse command-line flags
	port := flag.String("port", defaultPort, "Server port")
	flag.Parse()

	// Setup logger
	logger := log.New(os.Stdout, "[config-engine] ", log.LstdFlags|log.Lshortfile)

	// Initialize validator
	validator, err := validation.NewValidator()
	if err != nil {
		logger.Fatalf("Failed to initialize validator: %v", err)
	}
	logger.Println("Validator initialized successfully")

	// Initialize repository
	repo := repository.NewInMemoryRepository()
	logger.Println("Repository initialized successfully")

	// Initialize service
	svc := service.NewConfigService(repo, validator)
	logger.Println("Service initialized successfully")

	// Initialize handler
	handler := handlers.NewConfigHandler(svc, logger)

	// Setup router (Gin engine)
	router := handlers.SetupRouter(handler, logger)

	// Configure server
	addr := fmt.Sprintf(":%s", *port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		ErrorLog:          logger,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("Starting server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Printf("Configuration Management Service is running on http://localhost%s", addr)
	logger.Println("Press Ctrl+C to stop the server")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server stopped")
}