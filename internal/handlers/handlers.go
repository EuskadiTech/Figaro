// Package handlers provides HTTP request handlers for Figaro application.
package handlers

import (
	"embed"
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
		"PageTitle": "Figaró",
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
	data["ContentTemplate"] = "index"
	c.HTML(http.StatusOK, "base.html", data)
}

// Login handles the login page
func (h *Handlers) Login(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		h.handleLoginPost(c)
		return
	}

	// Show login form
	data := gin.H{
		"PageTitle": "Figaró - Iniciar Sesión",
	}
	c.HTML(http.StatusOK, "login.html", data)
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
	_, err = auth.SetUserSession(c, user, creds.Password, "Web Browser")
	if err != nil {
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"ErrorMessage": "Error al crear la sesión",
		})
		return
	}
	
	c.Redirect(http.StatusFound, "/")
}

// Logout handles user logout
func (h *Handlers) Logout(c *gin.Context) {
	auth.ClearUserSession(c)
	c.Redirect(http.StatusFound, "/login")
}

// Profile handles the user profile page
func (h *Handlers) Profile(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get user sessions
	sessions, err := auth.GetUserSessions(user.ID)
	if err != nil {
		sessions = []models.UserSession{} // Empty slice if error
	}

	// Get current session
	var currentSessionID string
	if sessionVal, exists := c.Get("session"); exists {
		if session, ok := sessionVal.(*models.UserSession); ok {
			currentSessionID = session.ID
		}
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Perfil de Usuario"
	data["ContentTemplate"] = "profile"
	data["Sessions"] = sessions
	data["CurrentSessionID"] = currentSessionID
	
	c.HTML(http.StatusOK, "base.html", data)
}

// ProfilePost handles profile form submissions
func (h *Handlers) ProfilePost(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	action := c.PostForm("action")
	
	switch action {
	case "logout_session":
		sessionID := c.PostForm("session_id")
		if sessionID != "" {
			auth.DeactivateSession(sessionID)
		}
		c.Redirect(http.StatusFound, "/perfil?success=Sesión cerrada correctamente")
		
	case "logout_all_sessions":
		// Get current session ID to preserve it
		var currentSessionID string
		if sessionVal, exists := c.Get("session"); exists {
			if session, ok := sessionVal.(*models.UserSession); ok {
				currentSessionID = session.ID
			}
		}
		
		auth.DeactivateAllUserSessions(user.ID, currentSessionID)
		c.Redirect(http.StatusFound, "/perfil?success=Todas las demás sesiones han sido cerradas")
		
	default:
		c.Redirect(http.StatusFound, "/perfil")
	}
}

// ElegirCentro handles center selection
func (h *Handlers) ElegirCentro(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		centro := c.PostForm("centro")
		aula := c.PostForm("aula")
		
		if centro != "" && aula != "" {
			c.SetCookie("centro", centro, 86400*30, "/", "", false, false) // 30 days
			c.SetCookie("aula", aula, 86400*30, "/", "", false, false)     // 30 days
			c.Redirect(http.StatusFound, "/")
			return
		}
	}

	// Get available centers
	centers, err := h.getCenters()
	if err != nil {
		centers = []string{} // Empty slice if error
	}

	selectedCentro := c.Query("centro")
	var aulas []string
	if selectedCentro != "" {
		aulas, _ = h.getAulas(selectedCentro)
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Elegir Centro"
	data["ContentTemplate"] = "elegir_centro"
	data["Centers"] = centers
	data["SelectedCentro"] = selectedCentro
	data["Aulas"] = aulas
	
	c.HTML(http.StatusOK, "base.html", data)
}

// getCenters retrieves available centers
func (h *Handlers) getCenters() ([]string, error) {
	// This would typically come from database, but for now simulate with existing data
	return []string{"Centro Demo", "Centro Demo 2"}, nil
}

// getAulas retrieves classrooms for a center
func (h *Handlers) getAulas(centro string) ([]string, error) {
	// This would typically come from database, but for now simulate
	switch centro {
	case "Centro Demo":
		return []string{"Aula 1", "Aula 2", "Aula 3"}, nil
	case "Centro Demo 2":
		return []string{"Laboratorio", "Sala de Informática"}, nil
	default:
		return []string{}, nil
	}
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