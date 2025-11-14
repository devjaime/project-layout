package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-standards/project-layout/internal/app/user-service/handler"
	"github.com/golang-standards/project-layout/internal/app/user-service/repository"
	"github.com/golang-standards/project-layout/internal/app/user-service/service"
	"github.com/golang-standards/project-layout/internal/pkg/config"
	"github.com/golang-standards/project-layout/internal/pkg/database"
	"github.com/golang-standards/project-layout/internal/pkg/logger"
	pb "github.com/golang-standards/project-layout/pkg/api/user/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// @title User Service API
// @version 1.0
// @description Microservice for user management
// @host localhost:50051
// @BasePath /api/v1
func main() {
	// Initialize logger
	log := logger.NewLogger()
	defer log.Sync()

	log.Info("Starting User Service",
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit,
	)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize repository, service, and handler
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, log)
	userHandler := handler.NewUserHandler(userService, log)

	// Create gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logger.UnaryServerInterceptor(log),
			// Add more interceptors here (auth, metrics, etc.)
		),
	)

	// Register services
	pb.RegisterUserServiceServer(grpcServer, userHandler)
	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	// Register reflection service on gRPC server
	reflection.Register(grpcServer)

	// Start gRPC server
	grpcAddr := fmt.Sprintf(":%s", cfg.Server.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatal("Failed to listen", "error", err, "address", grpcAddr)
	}

	// Start HTTP server for health checks and metrics
	httpAddr := fmt.Sprintf(":%s", cfg.Server.HTTPPort)
	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      setupHTTPHandlers(log),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors
	serverErrors := make(chan error, 1)

	// Start gRPC server in a goroutine
	go func() {
		log.Info("gRPC server listening", "address", grpcAddr)
		serverErrors <- grpcServer.Serve(lis)
	}()

	// Start HTTP server in a goroutine
	go func() {
		log.Info("HTTP server listening", "address", httpAddr)
		serverErrors <- httpServer.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		log.Fatal("Server error", "error", err)
	case sig := <-shutdown:
		log.Info("Received shutdown signal", "signal", sig)

		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown HTTP server
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Error("HTTP server shutdown error", "error", err)
			httpServer.Close()
		}

		// Gracefully stop gRPC server
		grpcServer.GracefulStop()

		log.Info("Server stopped gracefully")
	}
}

// setupHTTPHandlers configures HTTP endpoints for health checks and metrics
func setupHTTPHandlers(log logger.Logger) http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Readiness check endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Add your readiness logic here (e.g., check database connection)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	// Metrics endpoint (for Prometheus)
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// Prometheus metrics would be exposed here
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Metrics endpoint\n"))
	})

	// Version info endpoint
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"version":"%s","build_time":"%s","git_commit":"%s"}`,
			Version, BuildTime, GitCommit)
	})

	return mux
}
