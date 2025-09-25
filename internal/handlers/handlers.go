// Package handlers provides HTTP request handlers for Figaro application.
package handlers

import (
	"database/sql"
	"embed"
	"html/template"
	"net/http"
	"strings"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/internal/database"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/EuskadiTech/Figaro/pkg/config"
	"github.com/gin-gonic/gin"
)

//go:embed static/*
var staticFS embed.FS

//go:embed templates/*
var templateFS embed.FS

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

// CreateHTMLTemplate creates HTML template from embedded templates
func CreateHTMLTemplate() *template.Template {
	tmpl := template.New("base")
	
	// Parse all template files from embedded filesystem
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		panic("Failed to read templates directory: " + err.Error())
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
			content, err := templateFS.ReadFile("templates/" + entry.Name())
			if err != nil {
				panic("Failed to read template " + entry.Name() + ": " + err.Error())
			}
			
			// Parse the template content
			_, err = tmpl.New(entry.Name()).Parse(string(content))
			if err != nil {
				panic("Failed to parse template " + entry.Name() + ": " + err.Error())
			}
		}
	}
	
	return tmpl
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

// MaterialesIndex handles the materials inventory page
func (h *Handlers) MaterialesIndex(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get selected center
	centro, err := c.Cookie("centro")
	if err != nil || centro == "" {
		c.Redirect(http.StatusFound, "/elegir_centro")
		return
	}

	// Get materials from database
	materials, err := h.getMaterials(centro)
	if err != nil {
		materials = []models.Material{} // Empty slice if error
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Inventario de Materiales"
	data["ContentTemplate"] = "materiales"
	data["Materials"] = materials
	data["Centro"] = centro
	
	c.HTML(http.StatusOK, "base.html", data)
}

// getMaterials retrieves materials for a center
func (h *Handlers) getMaterials(centro string) ([]models.Material, error) {
	query := `SELECT id, center_id, name, photo_path, unit, available_quantity, minimum_quantity, notes, created_at, updated_at 
			  FROM materials WHERE center_id = (SELECT id FROM centers WHERE name = ?) ORDER BY name`
	
	rows, err := database.DB.Query(query, centro)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []models.Material
	for rows.Next() {
		var material models.Material
		var photoPath sql.NullString
		
		err := rows.Scan(&material.ID, &material.CenterID, &material.Name, &photoPath, 
			&material.Unit, &material.AvailableQuantity, &material.MinimumQuantity, 
			&material.Notes, &material.CreatedAt, &material.UpdatedAt)
		if err != nil {
			continue
		}
		
		if photoPath.Valid {
			material.PhotoPath = &photoPath.String
		}
		
		materials = append(materials, material)
	}
	
	return materials, nil
}

// ActividadesIndex handles the activities page
func (h *Handlers) ActividadesIndex(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Get selected center
	centro, err := c.Cookie("centro")
	if err != nil || centro == "" {
		c.Redirect(http.StatusFound, "/elegir_centro")
		return
	}

	// Get search parameters
	searchQuery := c.Query("q")
	showPast := c.Query("past") == "y"

	// Get activities from database
	activities, err := h.getActivities(centro, searchQuery, showPast)
	if err != nil {
		activities = []models.Activity{} // Empty slice if error
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Actividades"
	data["ContentTemplate"] = "actividades"
	data["Activities"] = activities
	data["Centro"] = centro
	data["SearchQuery"] = searchQuery
	data["ShowPast"] = showPast
	
	c.HTML(http.StatusOK, "base.html", data)
}

// getActivities retrieves activities for a center
func (h *Handlers) getActivities(centro string, searchQuery string, showPast bool) ([]models.Activity, error) {
	var query string
	var args []interface{}

	if showPast {
		query = `SELECT id, center_id, title, description, start_datetime, end_datetime, is_global, meeting_url, web_url, created_at, updated_at 
				FROM activities WHERE (center_id = (SELECT id FROM centers WHERE name = ?) OR is_global = 1)`
		args = []interface{}{centro}
	} else {
		query = `SELECT id, center_id, title, description, start_datetime, end_datetime, is_global, meeting_url, web_url, created_at, updated_at 
				FROM activities WHERE (center_id = (SELECT id FROM centers WHERE name = ?) OR is_global = 1) 
				AND start_datetime >= datetime('now')`
		args = []interface{}{centro}
	}

	if searchQuery != "" {
		query += " AND (title LIKE ? OR description LIKE ?)"
		searchPattern := "%" + searchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += " ORDER BY start_datetime ASC"

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.Activity
	for rows.Next() {
		var activity models.Activity
		var centerID sql.NullInt64
		var meetingURL sql.NullString
		var webURL sql.NullString
		
		err := rows.Scan(&activity.ID, &centerID, &activity.Title, &activity.Description, 
			&activity.StartDatetime, &activity.EndDatetime, &activity.IsGlobal,
			&meetingURL, &webURL, &activity.CreatedAt, &activity.UpdatedAt)
		if err != nil {
			continue
		}
		
		if centerID.Valid {
			centerIDInt := int(centerID.Int64)
			activity.CenterID = &centerIDInt
		}
		
		if meetingURL.Valid {
			activity.MeetingURL = &meetingURL.String
		}
		
		if webURL.Valid {
			activity.WebURL = &webURL.String
		}
		
		activities = append(activities, activity)
	}
	
	return activities, nil
}

// AdminIndex handles the admin panel page
func (h *Handlers) AdminIndex(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Panel de Administración"
	data["ContentTemplate"] = "admin"
	
	c.HTML(http.StatusOK, "base.html", data)
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