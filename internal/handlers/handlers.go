// Package handlers provides HTTP request handlers for Figaro application.
package handlers

import (
	"embed"
	"log"
	"net/http"
	"strings"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/EuskadiTech/Figaro/pkg/config"
	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFS embed.FS

// Handlers holds the application handlers and dependencies
type Handlers struct {
	Config *config.Config
}

// SessionData holds session information
type SessionData struct {
	Centro string
	Aula   string
}

// New creates a new Handlers instance
func New(cfg *config.Config) *Handlers {
	return &Handlers{
		Config: cfg,
	}
}

// getCommonData creates common template data
func (h *Handlers) getCommonData(c *gin.Context) gin.H {
	data := gin.H{
		"Flash": c.Query("flash"),
	}

	// Add user if logged in
	if user := auth.GetCurrentUser(c); user != nil {
		data["User"] = user
		
		// Add session info
		session := gin.H{}
		if centro, err := c.Cookie("centro"); err == nil {
			session["Centro"] = centro
		}
		if aula, err := c.Cookie("aula"); err == nil {
			session["Aula"] = aula
		}
		data["Session"] = session

		// Add permissions check function
		data["HasAccess"] = func(permission string) bool {
			return auth.UserHasAccess(c, permission)
		}
	}

	return data
}

// Index handles the home page
func (h *Handlers) Index(c *gin.Context) {
	data := h.getCommonData(c)
	c.HTML(http.StatusOK, "index.html", data)
}

// Login handles the login page
func (h *Handlers) Login(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		h.handleLoginPost(c)
		return
	}

	// Show login form
	c.HTML(http.StatusOK, "login.html", gin.H{})
}

// handleLoginPost processes login form submission
func (h *Handlers) handleLoginPost(c *gin.Context) {
	var creds auth.LoginCredentials
	if err := c.ShouldBind(&creds); err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorMessage": "Datos de formulario inválidos",
		})
		return
	}

	var user *models.User
	var err error

	if creds.QRData != "" {
		// QR login
		user, err = auth.LoginWithQR(creds.QRData)
		if err != nil {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"ErrorMessage": "Código QR inválido o caducado",
			})
			return
		}
	} else if creds.Username != "" && creds.Password != "" {
		// Username/password login
		user, err = auth.Login(creds.Username, creds.Password)
		if err != nil {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{
				"ErrorMessage": "Usuario o contraseña incorrectos",
			})
			return
		}
	} else {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorMessage": "Por favor proporciona credenciales válidas",
		})
		return
	}

	// Set session cookies
	auth.SetUserSession(c, user, creds.Password)
	
	c.Redirect(http.StatusFound, "/")
}

// Logout handles user logout
func (h *Handlers) Logout(c *gin.Context) {
	auth.ClearUserSession(c)
	c.Redirect(http.StatusFound, "/login")
}

// Static serves static files from embedded filesystem
func (h *Handlers) Static(c *gin.Context) {
	// Get the file path from URL
	path := strings.TrimPrefix(c.Request.URL.Path, "/static/")
	
	// Read file from embedded filesystem
	data, err := staticFS.ReadFile("static/" + path)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// Set appropriate content type based on file extension
	if strings.HasSuffix(path, ".css") {
		c.Header("Content-Type", "text/css")
	} else if strings.HasSuffix(path, ".js") {
		c.Header("Content-Type", "application/javascript")
	} else if strings.HasSuffix(path, ".png") {
		c.Header("Content-Type", "image/png")
	} else if strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") {
		c.Header("Content-Type", "image/jpeg")
	} else if strings.HasSuffix(path, ".gif") {
		c.Header("Content-Type", "image/gif")
	}

	c.Data(http.StatusOK, c.GetHeader("Content-Type"), data)
}