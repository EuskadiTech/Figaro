// Package handlers provides HTTP request handlers for Figaro application.
package handlers

import (
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/EuskadiTech/Figaro/pkg/config"
	"github.com/gin-gonic/gin"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// Handlers holds the application handlers and dependencies
type Handlers struct {
	Config    *config.Config
	templates *template.Template
}

// SessionData holds session information
type SessionData struct {
	Centro string
	Aula   string
}

// TemplateData holds data passed to templates
type TemplateData struct {
	User       interface{}
	Session    *SessionData
	Flash      string
	Data       interface{}
	HasAccess  func(string) bool
	ErrorMessage string
}

// New creates a new Handlers instance with loaded templates
func New(cfg *config.Config) *Handlers {
	// Create custom template functions
	funcMap := template.FuncMap{
		"hasAccess": func() string { return "" }, // Placeholder, will be replaced per request
	}

	// Load templates from embedded filesystem
	templates := template.Must(template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html"))

	return &Handlers{
		Config:    cfg,
		templates: templates,
	}
}

// getTemplateData creates template data with common fields
func (h *Handlers) getTemplateData(c *gin.Context) *TemplateData {
	user := auth.GetCurrentUser(c)
	
	// Get session data
	session := &SessionData{}
	if centro, err := c.Cookie("centro"); err == nil {
		session.Centro = centro
	}
	if aula, err := c.Cookie("aula"); err == nil {
		session.Aula = aula
	}

	// Create hasAccess function for this specific request
	hasAccessFunc := func(permission string) bool {
		return auth.UserHasAccess(c, permission)
	}

	return &TemplateData{
		User:      user,
		Session:   session,
		Flash:     c.Query("flash"),
		HasAccess: hasAccessFunc,
	}
}

// renderTemplate renders a template with the given data
func (h *Handlers) renderTemplate(c *gin.Context, templateName string, data *TemplateData) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	
	// Create a new template instance with the current hasAccess function
	tmpl := template.Must(h.templates.Clone())
	tmpl = tmpl.Funcs(template.FuncMap{
		"call": func(fn func(string) bool, arg string) bool {
			return fn(arg)
		},
	})
	
	if err := tmpl.ExecuteTemplate(c.Writer, templateName, data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Template rendering failed"})
		return
	}
}

// Index handles the home page
func (h *Handlers) Index(c *gin.Context) {
	data := h.getTemplateData(c)
	h.renderTemplate(c, "base.html", data)
}

// Login handles the login page
func (h *Handlers) Login(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		h.handleLoginPost(c)
		return
	}

	// Show login form
	data := &TemplateData{}
	h.renderTemplate(c, "base.html", data)
}

// handleLoginPost processes login form submission
func (h *Handlers) handleLoginPost(c *gin.Context) {
	var creds auth.LoginCredentials
	if err := c.ShouldBind(&creds); err != nil {
		data := &TemplateData{ErrorMessage: "Datos de formulario inválidos"}
		h.renderTemplate(c, "base.html", data)
		return
	}

	var user *models.User
	var err error

	if creds.QRData != "" {
		// QR login
		user, err = auth.LoginWithQR(creds.QRData)
		if err != nil {
			data := &TemplateData{ErrorMessage: "Código QR inválido o caducado"}
			h.renderTemplate(c, "base.html", data)
			return
		}
	} else if creds.Username != "" && creds.Password != "" {
		// Username/password login
		user, err = auth.Login(creds.Username, creds.Password)
		if err != nil {
			data := &TemplateData{ErrorMessage: "Usuario o contraseña incorrectos"}
			h.renderTemplate(c, "base.html", data)
			return
		}
	} else {
		data := &TemplateData{ErrorMessage: "Por favor proporciona credenciales válidas"}
		h.renderTemplate(c, "base.html", data)
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