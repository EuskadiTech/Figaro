// Figaro - Educational center management system
// Main application entry point
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/internal/database"
	"github.com/EuskadiTech/Figaro/internal/handlers"
	"github.com/EuskadiTech/Figaro/pkg/config"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Initialize database
	if err := database.Initialize(cfg.DataDir); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create handlers
	h := handlers.New(cfg)

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode) // Production mode by default
	}
	
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Routes that don't require authentication
	router.GET("/login", h.Login)
	router.POST("/login", h.Login)
	router.GET("/static/*filepath", h.Static)

	// Routes that require authentication
	authGroup := router.Group("/")
	authGroup.Use(auth.RequireAuth())
	{
		authGroup.GET("/", h.Index)
		authGroup.GET("/logout", h.Logout)
		
		// TODO: Add other authenticated routes here
		// authGroup.GET("/materiales", h.MaterialesIndex)
		// authGroup.GET("/actividades", h.ActividadesIndex)
		// authGroup.GET("/admin", h.AdminIndex)
		// etc.
	}

	// Start server
	log.Printf("Starting Figaro server on %s:%s", cfg.Host, cfg.Port)
	log.Printf("Data directory: %s", cfg.DataDir)
	
	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		database.Close()
		os.Exit(0)
	}()

	if err := router.Run(cfg.Host + ":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}