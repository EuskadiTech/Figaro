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
	"github.com/EuskadiTech/Figaro/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize JSON logger
	if err := logger.Initialize(cfg.DataDir); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// Log startup
	logger.Info("Figaro server starting up")
	logger.Info("Data directory: %s", cfg.DataDir)

	// Initialize database
	if err := database.Initialize(cfg.DataDir); err != nil {
		logger.Error("Failed to initialize database: %v", err)
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	logger.Info("Database initialized successfully")

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
	router.GET("/auth/google", h.GoogleOAuthLogin)
	router.GET("/auth/google/callback", h.GoogleOAuthCallback)
	router.GET("/static/*filepath", h.Static)

	// Routes that require authentication
	authGroup := router.Group("/")
	authGroup.Use(auth.RequireAuth())
	{
		authGroup.GET("/", h.Index)
		authGroup.GET("/logout", h.Logout)
		authGroup.GET("/perfil", h.Profile)
		authGroup.POST("/perfil", h.ProfilePost)
		authGroup.GET("/perfil/webdav", h.WebDAVTokens)
		authGroup.POST("/perfil/webdav/crear", h.WebDAVCreateToken)
		authGroup.POST("/perfil/webdav/revocar/:id", h.WebDAVRevokeToken)
		authGroup.GET("/elegir_centro", h.ElegirCentro)
		authGroup.POST("/elegir_centro", h.ElegirCentro)

		// WebDAV server routes (separate from web interface)
		davGroup := router.Group("/dav")
		davGroup.Use(h.WebDAVAuth())
		
		// Personal files WebDAV
		davGroup.Handle("PROPFIND", "/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.Handle("PROPPATCH", "/MisArchivos/*path", h.WebDAVPersonalFiles)  
		davGroup.Handle("MKCOL", "/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.Handle("COPY", "/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.Handle("MOVE", "/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.Handle("LOCK", "/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.Handle("UNLOCK", "/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.GET("/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.PUT("/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.POST("/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.DELETE("/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.HEAD("/MisArchivos/*path", h.WebDAVPersonalFiles)
		davGroup.OPTIONS("/MisArchivos/*path", h.WebDAVPersonalFiles)
		
		// Shared folders WebDAV  
		davGroup.Handle("PROPFIND", "/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.Handle("PROPPATCH", "/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.Handle("MKCOL", "/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.Handle("COPY", "/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.Handle("MOVE", "/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.Handle("LOCK", "/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.Handle("UNLOCK", "/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.GET("/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.PUT("/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.POST("/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.DELETE("/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.HEAD("/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)
		davGroup.OPTIONS("/CarpetasCompartidas/:folder/*path", h.WebDAVSharedFolders)

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

		// Shared folders module
		authGroup.GET("/carpetas-compartidas", h.CarpetasCompartidasIndex)
		authGroup.GET("/carpetas-compartidas/crear", h.CarpetasCompartidasCrear)
		authGroup.POST("/carpetas-compartidas/crear", h.CarpetasCompartidasCrear)
		authGroup.POST("/carpetas-compartidas/eliminar/:id", h.CarpetasCompartidasEliminar)

		// Admin module
		authGroup.GET("/admin", h.AdminIndex)
		authGroup.GET("/admin/usuarios", h.AdminUsuarios)
		authGroup.GET("/admin/usuarios/crear", h.AdminUsuarioCrear)
		authGroup.POST("/admin/usuarios/crear", h.AdminUsuarioCrear)
		authGroup.GET("/admin/usuarios/editar/:id", h.AdminUsuarioEditar)
		authGroup.POST("/admin/usuarios/editar/:id", h.AdminUsuarioEditar)
		authGroup.POST("/admin/usuarios/eliminar/:id", h.AdminUsuarioEliminar)
		authGroup.GET("/admin/centros", h.AdminCentros)
		authGroup.GET("/admin/centros/crear", h.AdminCentroCrear)
		authGroup.POST("/admin/centros/crear", h.AdminCentroCrear)
		authGroup.GET("/admin/centros/editar/:id", h.AdminCentroEditar)
		authGroup.POST("/admin/centros/editar/:id", h.AdminCentroEditar)
		authGroup.GET("/admin/centros/aulas/:center_id", h.AdminCentroAulas)
		authGroup.GET("/admin/centros/aulas/:center_id/crear", h.AdminAulaCrear)
		authGroup.POST("/admin/centros/aulas/:center_id/crear", h.AdminAulaCrear)
		authGroup.GET("/admin/centros/aulas/:center_id/editar/:aula_id", h.AdminAulaEditar)
		authGroup.POST("/admin/centros/aulas/:center_id/editar/:aula_id", h.AdminAulaEditar)
		authGroup.POST("/admin/centros/aulas/:center_id/eliminar/:aula_id", h.AdminAulaEliminar)
		authGroup.GET("/admin/materiales-report", h.AdminMaterialesReport)
		authGroup.GET("/admin/actividades-report", h.AdminActividadesReport)
		authGroup.GET("/admin/files", h.AdminFiles)
		authGroup.GET("/admin/configuracion", h.AdminConfiguracion)
		authGroup.POST("/admin/configuracion/general", h.AdminConfiguracionGeneral)
		authGroup.POST("/admin/configuracion/security", h.AdminConfiguracionSecurity)
		authGroup.POST("/admin/configuracion/oauth", h.AdminConfiguracionOAuth)
		authGroup.POST("/admin/configuracion/email", h.AdminConfiguracionEmail)
		authGroup.POST("/admin/configuracion/backup", h.AdminConfiguracionBackup)
		authGroup.POST("/admin/configuracion/database", h.AdminConfiguracionDatabase)
		authGroup.GET("/admin/configuracion/logs", h.AdminConfiguracionLogs)
	}

	// Start server
	logger.Info("Starting Figaro server on %s:%s", cfg.Host, cfg.Port)
	log.Printf("Starting Figaro server on %s:%s", cfg.Host, cfg.Port)
	log.Printf("Data directory: %s", cfg.DataDir)

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down gracefully...")
		log.Println("Shutting down gracefully...")
		logger.Close()
		database.Close()
		os.Exit(0)
	}()

	if err := router.Run(cfg.Host + ":" + cfg.Port); err != nil {
		logger.Error("Failed to start server: %v", err)
		log.Fatalf("Failed to start server: %v", err)
	}
}
