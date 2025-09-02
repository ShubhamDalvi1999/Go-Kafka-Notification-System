package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kafka-notify/internal/config"
	"kafka-notify/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Server represents an HTTP server
type Server struct {
	config     *config.ServerConfig
	router     *gin.Engine
	httpServer *http.Server
	stopChan   chan os.Signal
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.ServerConfig) *Server {
	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	server := &Server{
		config:   cfg,
		router:   router,
		stopChan: make(chan os.Signal, 1),
	}

	// Setup health check route
	server.setupHealthCheck()

	return server
}

// GetRouter returns the Gin router for adding routes
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

// setupHealthCheck sets up the health check endpoint
func (s *Server) setupHealthCheck() {
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "notification-service",
		})
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         s.config.Port,
		Handler:      s.router,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting HTTP server on port %s", s.config.Port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	signal.Notify(s.stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-s.stopChan

	log.Println("Shutting down server...")
	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server exited gracefully")
	return nil
}

// AddRoute adds a route to the server
func (s *Server) AddRoute(method, path string, handler gin.HandlerFunc) {
	switch method {
	case "GET":
		s.router.GET(path, handler)
	case "POST":
		s.router.POST(path, handler)
	case "PUT":
		s.router.PUT(path, handler)
	case "DELETE":
		s.router.DELETE(path, handler)
	case "PATCH":
		s.router.PATCH(path, handler)
	default:
		log.Printf("Unsupported HTTP method: %s", method)
	}
}

// AddGroup adds a route group to the server
func (s *Server) AddGroup(prefix string) *gin.RouterGroup {
	return s.router.Group(prefix)
}

// GetStatus returns the server status
func (s *Server) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"port":         s.config.Port,
		"readTimeout":  s.config.ReadTimeout.String(),
		"writeTimeout": s.config.WriteTimeout.String(),
		"idleTimeout":  s.config.IdleTimeout.String(),
		"status":       "running",
	}
}
