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
		authGroup.GET("/perfil", h.Profile)
		authGroup.POST("/perfil", h.ProfilePost)
		authGroup.GET("/elegir_centro", h.ElegirCentro)
		authGroup.POST("/elegir_centro", h.ElegirCentro)

		// Materials module
		authGroup.GET("/materiales", h.MaterialesIndex)
		authGroup.GET("/materiales/crear", h.MaterialesCrear)
		authGroup.POST("/materiales/crear", h.MaterialesCrear)
		authGroup.GET("/materiales/editar/:id", h.MaterialesEditar)
		authGroup.POST("/materiales/editar/:id", h.MaterialesEditar)
		authGroup.POST("/materiales/eliminar/:id", h.MaterialesEliminar)

		// Activities module
		authGroup.GET("/actividades", h.ActividadesIndex)
		authGroup.GET("/actividades/crear", h.ActividadesCrear)
		authGroup.POST("/actividades/crear", h.ActividadesCrear)
		authGroup.GET("/actividades/editar/:id", h.ActividadesEditar)
		authGroup.POST("/actividades/editar/:id", h.ActividadesEditar)
		authGroup.POST("/actividades/eliminar/:id", h.ActividadesEliminar)

		// Admin module
		authGroup.GET("/admin", h.AdminIndex)
		authGroup.GET("/admin/usuarios", h.AdminUsuarios)
		authGroup.GET("/admin/usuarios/crear", h.AdminUsuarioCrear)
		authGroup.POST("/admin/usuarios/crear", h.AdminUsuarioCrear)
		authGroup.GET("/admin/usuarios/editar/:id", h.AdminUsuarioEditar)
		authGroup.POST("/admin/usuarios/editar/:id", h.AdminUsuarioEditar)
		authGroup.POST("/admin/usuarios/eliminar/:id", h.AdminUsuarioEliminar)
		authGroup.GET("/admin/centros", h.AdminCentros)
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
