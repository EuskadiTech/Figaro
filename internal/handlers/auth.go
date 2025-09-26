package handlers

import (
	"fmt"
	"net/http"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/internal/database"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/gin-gonic/gin"
)

// Index handles the home page
func (h *Handlers) Index(c *gin.Context) {
	data := h.getCommonData(c)
	h.renderTemplate(c, "index.html", data)
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
	h.renderTemplate(c, "login.html", data)
}

// handleLoginPost processes login form submission
func (h *Handlers) handleLoginPost(c *gin.Context) {
	var creds auth.LoginCredentials
	if err := c.ShouldBind(&creds); err != nil {
		h.renderTemplate(c, "login.html", gin.H{
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
			h.renderTemplate(c, "login.html", gin.H{
				"ErrorMessage": "Código QR inválido o caducado",
			})
			return
		}
	} else if creds.Username != "" && creds.Password != "" {
		// Username/password login
		user, err = auth.Login(creds.Username, creds.Password)
		if err != nil {
			h.renderTemplate(c, "login.html", gin.H{
				"ErrorMessage": "Usuario o contraseña incorrectos",
			})
			return
		}
	} else {
		h.renderTemplate(c, "login.html", gin.H{
			"ErrorMessage": "Por favor proporciona credenciales válidas",
		})
		return
	}

	// Set session cookies
	_, err = auth.SetUserSession(c, user, creds.Password, "Web Browser")
	if err != nil {
		h.renderTemplate(c, "login.html", gin.H{
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
	data["Sessions"] = sessions
	data["CurrentSessionID"] = currentSessionID

	h.renderTemplate(c, "profile.html", data)
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
	data["Centers"] = centers
	data["SelectedCentro"] = selectedCentro
	data["Aulas"] = aulas

	h.renderTemplate(c, "elegir_centro.html", data)
}

// getCenters retrieves available centers
func (h *Handlers) getCenters() ([]string, error) {
	query := `SELECT name FROM centers ORDER BY name`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var centers []string
	for rows.Next() {
		var name string
		err := rows.Scan(&name)
		if err != nil {
			continue
		}
		centers = append(centers, name)
	}

	return centers, nil
}

// getAulas retrieves classrooms for a center
func (h *Handlers) getAulas(centro string) ([]string, error) {
	// Get center ID by name first
	var centerID int
	query := `SELECT id FROM centers WHERE name = ?`
	err := database.DB.QueryRow(query, centro).Scan(&centerID)
	if err != nil {
		return []string{}, err
	}

	// Get classrooms from database using the same function as admin
	classrooms, err := h.getClassroomsByCenter(fmt.Sprintf("%d", centerID))
	if err != nil {
		return []string{}, err
	}

	// Convert classroom models to string array for template compatibility
	var aulaNames []string
	for _, classroom := range classrooms {
		aulaNames = append(aulaNames, classroom.Name)
	}

	return aulaNames, nil
}