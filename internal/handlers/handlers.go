// Package handlers provides HTTP request handlers for Figaro application.
package handlers

import (
	"database/sql"
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// loadTemplate loads a specific template by name from the embedded filesystem
func (h *Handlers) loadTemplate(templateName string) (*template.Template, error) {
	// Create a new template with base as the root
	tmpl := template.New("base.html")

	// Add template functions
	tmpl = tmpl.Funcs(template.FuncMap{
		"call": func(fn interface{}, args ...interface{}) interface{} {
			if f, ok := fn.(func(string) bool); ok && len(args) > 0 {
				if arg, ok := args[0].(string); ok {
					return f(arg)
				}
			}
			return false
		},
		"contains": func(slice interface{}, item string) bool {
			switch s := slice.(type) {
			case []string:
				for _, v := range s {
					if v == item {
						return true
					}
				}
			case string:
				return strings.Contains(s, item)
			}
			return false
		},
		"now": func() time.Time {
			return time.Now()
		},
	})

	// First, always load base.html
	baseContent, err := templateFS.ReadFile("templates/base.html")
	if err != nil {
		return nil, err
	}

	// Parse base template
	tmpl, err = tmpl.Parse(string(baseContent))
	if err != nil {
		return nil, err
	}

	// Then load the specific template if it's not base.html
	if templateName != "base.html" {
		content, err := templateFS.ReadFile("templates/" + templateName)
		if err != nil {
			return nil, err
		}

		// Parse the specific template
		tmpl, err = tmpl.Parse(string(content))
		if err != nil {
			return nil, err
		}
	}

	return tmpl, nil
} // renderTemplate renders a template with the given data
func (h *Handlers) renderTemplate(c *gin.Context, templateName string, data interface{}) {
	tmpl, err := h.loadTemplate(templateName)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: %v", err)
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	// Always execute base.html as the root template
	err = tmpl.ExecuteTemplate(c.Writer, "base.html", data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template execution error: %v", err)
		return
	}
}

// getCommonData creates common template data
func (h *Handlers) getCommonData(c *gin.Context) gin.H {
	data := gin.H{
		"Flash":     c.Query("flash"),
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
} // ProfilePost handles profile form submissions
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
	data["Materials"] = materials
	data["Centro"] = centro

	h.renderTemplate(c, "materiales.html", data)
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

// MaterialesCrear handles material creation (GET shows form, POST processes it)
func (h *Handlers) MaterialesCrear(c *gin.Context) {
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

	if c.Request.Method == http.MethodPost {
		h.handleMaterialCreate(c, centro)
		return
	}

	// Show creation form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Crear Material"
	data["Centro"] = centro
	data["Action"] = "crear"

	h.renderTemplate(c, "material_form.html", data)
}

// handleMaterialCreate processes material creation
func (h *Handlers) handleMaterialCreate(c *gin.Context, centro string) {
	name := c.PostForm("nombre")
	unit := c.PostForm("unidad")
	availableQty := c.PostForm("cantidad_disponible")
	minimumQty := c.PostForm("cantidad_minima")
	notes := c.PostForm("notas")

	if name == "" || unit == "" {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Material"
		data["Centro"] = centro
		data["Action"] = "crear"
		data["ErrorMessage"] = "El nombre y la unidad son requeridos"
		data["FormData"] = gin.H{
			"nombre":              name,
			"unidad":              unit,
			"cantidad_disponible": availableQty,
			"cantidad_minima":     minimumQty,
			"notas":               notes,
		}
		h.renderTemplate(c, "material_form.html", data)
		return
	}

	// Get center ID
	centerID, err := h.getCenterID(centro)
	if err != nil {
		c.Redirect(http.StatusFound, "/materiales?error=Centro no encontrado")
		return
	}

	// Convert quantities to int
	var availableQtyInt, minimumQtyInt int
	if availableQty != "" {
		if val, err := parseIntSafe(availableQty); err == nil {
			availableQtyInt = val
		}
	}
	if minimumQty != "" {
		if val, err := parseIntSafe(minimumQty); err == nil {
			minimumQtyInt = val
		}
	}

	// Insert into database
	query := `INSERT INTO materials (center_id, name, unit, available_quantity, minimum_quantity, notes, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`

	_, err = database.DB.Exec(query, centerID, name, unit, availableQtyInt, minimumQtyInt, notes)
	if err != nil {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Material"
		data["Centro"] = centro
		data["Action"] = "crear"
		data["ErrorMessage"] = "Error al crear el material: " + err.Error()
		data["FormData"] = gin.H{
			"nombre":              name,
			"unidad":              unit,
			"cantidad_disponible": availableQty,
			"cantidad_minima":     minimumQty,
			"notas":               notes,
		}
		h.renderTemplate(c, "material_form.html", data)
		return
	}

	c.Redirect(http.StatusFound, "/materiales?success=Material creado correctamente")
}

// MaterialesEditar handles material editing
func (h *Handlers) MaterialesEditar(c *gin.Context) {
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

	materialID := c.Param("id")
	if materialID == "" {
		c.Redirect(http.StatusFound, "/materiales?error=ID de material requerido")
		return
	}

	if c.Request.Method == http.MethodPost {
		h.handleMaterialUpdate(c, centro, materialID)
		return
	}

	// Get material data
	material, err := h.getMaterial(materialID, centro)
	if err != nil {
		c.Redirect(http.StatusFound, "/materiales?error=Material no encontrado")
		return
	}

	// Show edit form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Editar Material"
	data["Centro"] = centro
	data["Action"] = "editar"
	data["Material"] = material

	h.renderTemplate(c, "material_form.html", data)
}

// handleMaterialUpdate processes material updates
func (h *Handlers) handleMaterialUpdate(c *gin.Context, centro, materialID string) {
	name := c.PostForm("nombre")
	unit := c.PostForm("unidad")
	availableQty := c.PostForm("cantidad_disponible")
	minimumQty := c.PostForm("cantidad_minima")
	notes := c.PostForm("notas")

	if name == "" || unit == "" {
		material, _ := h.getMaterial(materialID, centro)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Material"
		data["Centro"] = centro
		data["Action"] = "editar"
		data["Material"] = material
		data["ErrorMessage"] = "El nombre y la unidad son requeridos"
		h.renderTemplate(c, "material_form.html", data)
		return
	}

	// Get center ID
	centerID, err := h.getCenterID(centro)
	if err != nil {
		c.Redirect(http.StatusFound, "/materiales?error=Centro no encontrado")
		return
	}

	// Convert quantities to int
	var availableQtyInt, minimumQtyInt int
	if availableQty != "" {
		if val, err := parseIntSafe(availableQty); err == nil {
			availableQtyInt = val
		}
	}
	if minimumQty != "" {
		if val, err := parseIntSafe(minimumQty); err == nil {
			minimumQtyInt = val
		}
	}

	// Update in database
	query := `UPDATE materials SET name = ?, unit = ?, available_quantity = ?, minimum_quantity = ?, notes = ?, updated_at = datetime('now')
			  WHERE id = ? AND center_id = ?`

	result, err := database.DB.Exec(query, name, unit, availableQtyInt, minimumQtyInt, notes, materialID, centerID)
	if err != nil {
		material, _ := h.getMaterial(materialID, centro)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Material"
		data["Centro"] = centro
		data["Action"] = "editar"
		data["Material"] = material
		data["ErrorMessage"] = "Error al actualizar el material: " + err.Error()
		h.renderTemplate(c, "material_form.html", data)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.Redirect(http.StatusFound, "/materiales?error=Material no encontrado o sin permisos")
		return
	}

	c.Redirect(http.StatusFound, "/materiales?success=Material actualizado correctamente")
}

// MaterialesEliminar handles material deletion
func (h *Handlers) MaterialesEliminar(c *gin.Context) {
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

	materialID := c.Param("id")
	if materialID == "" {
		c.Redirect(http.StatusFound, "/materiales?error=ID de material requerido")
		return
	}

	// Get center ID for security
	centerID, err := h.getCenterID(centro)
	if err != nil {
		c.Redirect(http.StatusFound, "/materiales?error=Centro no encontrado")
		return
	}

	// Delete material
	query := `DELETE FROM materials WHERE id = ? AND center_id = ?`
	result, err := database.DB.Exec(query, materialID, centerID)
	if err != nil {
		c.Redirect(http.StatusFound, "/materiales?error=Error al eliminar el material")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.Redirect(http.StatusFound, "/materiales?error=Material no encontrado o sin permisos")
		return
	}

	c.Redirect(http.StatusFound, "/materiales?success=Material eliminado correctamente")
}

// getMaterial retrieves a single material by ID and center
func (h *Handlers) getMaterial(materialID, centro string) (models.Material, error) {
	var material models.Material
	query := `SELECT id, center_id, name, photo_path, unit, available_quantity, minimum_quantity, notes, created_at, updated_at 
			  FROM materials WHERE id = ? AND center_id = (SELECT id FROM centers WHERE name = ?)`

	var photoPath sql.NullString
	err := database.DB.QueryRow(query, materialID, centro).Scan(
		&material.ID, &material.CenterID, &material.Name, &photoPath,
		&material.Unit, &material.AvailableQuantity, &material.MinimumQuantity,
		&material.Notes, &material.CreatedAt, &material.UpdatedAt)

	if photoPath.Valid {
		material.PhotoPath = &photoPath.String
	}

	return material, err
}

// getCenterID gets the center ID by name
func (h *Handlers) getCenterID(centerName string) (int, error) {
	var centerID int
	query := `SELECT id FROM centers WHERE name = ?`
	err := database.DB.QueryRow(query, centerName).Scan(&centerID)
	return centerID, err
}

// parseIntSafe safely parses a string to int
func parseIntSafe(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
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
	data["Activities"] = activities
	data["Centro"] = centro
	data["SearchQuery"] = searchQuery
	data["ShowPast"] = showPast

	h.renderTemplate(c, "actividades.html", data)
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

// ActividadesCrear handles activity creation
func (h *Handlers) ActividadesCrear(c *gin.Context) {
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

	if c.Request.Method == http.MethodPost {
		h.handleActivityCreate(c, centro)
		return
	}

	// Show creation form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Crear Actividad"
	data["Centro"] = centro
	data["Action"] = "crear"

	h.renderTemplate(c, "actividad_form.html", data)
}

// handleActivityCreate processes activity creation
func (h *Handlers) handleActivityCreate(c *gin.Context, centro string) {
	title := c.PostForm("titulo")
	description := c.PostForm("descripcion")
	startDate := c.PostForm("fecha_inicio")
	startTime := c.PostForm("hora_inicio")
	endDate := c.PostForm("fecha_fin")
	endTime := c.PostForm("hora_fin")
	isGlobal := c.PostForm("global") == "1"
	meetingURL := c.PostForm("meeting_url")
	webURL := c.PostForm("web_url")

	if title == "" || startDate == "" || startTime == "" || endDate == "" || endTime == "" {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Actividad"
		data["Centro"] = centro
		data["Action"] = "crear"
		data["ErrorMessage"] = "Título, fecha y hora de inicio y fin son requeridos"
		data["FormData"] = gin.H{
			"titulo":       title,
			"descripcion":  description,
			"fecha_inicio": startDate,
			"hora_inicio":  startTime,
			"fecha_fin":    endDate,
			"hora_fin":     endTime,
			"global":       isGlobal,
			"meeting_url":  meetingURL,
			"web_url":      webURL,
		}
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	// Parse dates
	startDatetime, err := parseDateTime(startDate, startTime)
	if err != nil {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Actividad"
		data["Centro"] = centro
		data["Action"] = "crear"
		data["ErrorMessage"] = "Fecha/hora de inicio inválida"
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	endDatetime, err := parseDateTime(endDate, endTime)
	if err != nil {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Actividad"
		data["Centro"] = centro
		data["Action"] = "crear"
		data["ErrorMessage"] = "Fecha/hora de fin inválida"
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	if endDatetime.Before(startDatetime) {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Actividad"
		data["Centro"] = centro
		data["Action"] = "crear"
		data["ErrorMessage"] = "La fecha de fin debe ser posterior a la fecha de inicio"
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	var centerID *int
	if !isGlobal {
		id, err := h.getCenterID(centro)
		if err != nil {
			c.Redirect(http.StatusFound, "/actividades?error=Centro no encontrado")
			return
		}
		centerID = &id
	}

	// Insert into database
	query := `INSERT INTO activities (center_id, title, description, start_datetime, end_datetime, is_global, meeting_url, web_url, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`

	var meetingURLPtr, webURLPtr *string
	if meetingURL != "" {
		meetingURLPtr = &meetingURL
	}
	if webURL != "" {
		webURLPtr = &webURL
	}

	_, err = database.DB.Exec(query, centerID, title, description, startDatetime, endDatetime, isGlobal, meetingURLPtr, webURLPtr)
	if err != nil {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Actividad"
		data["Centro"] = centro
		data["Action"] = "crear"
		data["ErrorMessage"] = "Error al crear la actividad: " + err.Error()
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	c.Redirect(http.StatusFound, "/actividades?success=Actividad creada correctamente")
}

// ActividadesEditar handles activity editing
func (h *Handlers) ActividadesEditar(c *gin.Context) {
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

	activityID := c.Param("id")
	if activityID == "" {
		c.Redirect(http.StatusFound, "/actividades?error=ID de actividad requerido")
		return
	}

	if c.Request.Method == http.MethodPost {
		h.handleActivityUpdate(c, centro, activityID)
		return
	}

	// Get activity data
	activity, err := h.getActivity(activityID, centro)
	if err != nil {
		c.Redirect(http.StatusFound, "/actividades?error=Actividad no encontrada")
		return
	}

	// Show edit form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Editar Actividad"
	data["Centro"] = centro
	data["Action"] = "editar"
	data["Activity"] = activity

	h.renderTemplate(c, "actividad_form.html", data)
}

// handleActivityUpdate processes activity updates
func (h *Handlers) handleActivityUpdate(c *gin.Context, centro, activityID string) {
	title := c.PostForm("titulo")
	description := c.PostForm("descripcion")
	startDate := c.PostForm("fecha_inicio")
	startTime := c.PostForm("hora_inicio")
	endDate := c.PostForm("fecha_fin")
	endTime := c.PostForm("hora_fin")
	isGlobal := c.PostForm("global") == "1"
	meetingURL := c.PostForm("meeting_url")
	webURL := c.PostForm("web_url")

	if title == "" || startDate == "" || startTime == "" || endDate == "" || endTime == "" {
		activity, _ := h.getActivity(activityID, centro)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Actividad"
		data["Centro"] = centro
		data["Action"] = "editar"
		data["Activity"] = activity
		data["ErrorMessage"] = "Título, fecha y hora de inicio y fin son requeridos"
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	// Parse dates
	startDatetime, err := parseDateTime(startDate, startTime)
	if err != nil {
		activity, _ := h.getActivity(activityID, centro)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Actividad"
		data["Centro"] = centro
		data["Action"] = "editar"
		data["Activity"] = activity
		data["ErrorMessage"] = "Fecha/hora de inicio inválida"
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	endDatetime, err := parseDateTime(endDate, endTime)
	if err != nil {
		activity, _ := h.getActivity(activityID, centro)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Actividad"
		data["Centro"] = centro
		data["Action"] = "editar"
		data["Activity"] = activity
		data["ErrorMessage"] = "Fecha/hora de fin inválida"
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	if endDatetime.Before(startDatetime) {
		activity, _ := h.getActivity(activityID, centro)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Actividad"
		data["Centro"] = centro
		data["Action"] = "editar"
		data["Activity"] = activity
		data["ErrorMessage"] = "La fecha de fin debe ser posterior a la fecha de inicio"
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	var centerID *int
	if !isGlobal {
		id, err := h.getCenterID(centro)
		if err != nil {
			c.Redirect(http.StatusFound, "/actividades?error=Centro no encontrado")
			return
		}
		centerID = &id
	}

	var meetingURLPtr, webURLPtr *string
	if meetingURL != "" {
		meetingURLPtr = &meetingURL
	}
	if webURL != "" {
		webURLPtr = &webURL
	}

	// Update in database - allow editing global activities or activities from the current center
	query := `UPDATE activities SET center_id = ?, title = ?, description = ?, start_datetime = ?, end_datetime = ?, 
			  is_global = ?, meeting_url = ?, web_url = ?, updated_at = datetime('now')
			  WHERE id = ? AND (is_global = 1 OR center_id = (SELECT id FROM centers WHERE name = ?))`

	result, err := database.DB.Exec(query, centerID, title, description, startDatetime, endDatetime,
		isGlobal, meetingURLPtr, webURLPtr, activityID, centro)
	if err != nil {
		activity, _ := h.getActivity(activityID, centro)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Actividad"
		data["Centro"] = centro
		data["Action"] = "editar"
		data["Activity"] = activity
		data["ErrorMessage"] = "Error al actualizar la actividad: " + err.Error()
		h.renderTemplate(c, "actividad_form.html", data)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.Redirect(http.StatusFound, "/actividades?error=Actividad no encontrada o sin permisos")
		return
	}

	c.Redirect(http.StatusFound, "/actividades?success=Actividad actualizada correctamente")
}

// ActividadesEliminar handles activity deletion
func (h *Handlers) ActividadesEliminar(c *gin.Context) {
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

	activityID := c.Param("id")
	if activityID == "" {
		c.Redirect(http.StatusFound, "/actividades?error=ID de actividad requerido")
		return
	}

	// Delete activity - allow deleting global activities or activities from the current center
	query := `DELETE FROM activities WHERE id = ? AND (is_global = 1 OR center_id = (SELECT id FROM centers WHERE name = ?))`
	result, err := database.DB.Exec(query, activityID, centro)
	if err != nil {
		c.Redirect(http.StatusFound, "/actividades?error=Error al eliminar la actividad")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.Redirect(http.StatusFound, "/actividades?error=Actividad no encontrada o sin permisos")
		return
	}

	c.Redirect(http.StatusFound, "/actividades?success=Actividad eliminada correctamente")
}

// getActivity retrieves a single activity by ID
func (h *Handlers) getActivity(activityID, centro string) (models.Activity, error) {
	var activity models.Activity
	query := `SELECT id, center_id, title, description, start_datetime, end_datetime, is_global, meeting_url, web_url, created_at, updated_at 
			  FROM activities WHERE id = ? AND (is_global = 1 OR center_id = (SELECT id FROM centers WHERE name = ?))`

	var centerID sql.NullInt64
	var meetingURL sql.NullString
	var webURL sql.NullString

	err := database.DB.QueryRow(query, activityID, centro).Scan(
		&activity.ID, &centerID, &activity.Title, &activity.Description,
		&activity.StartDatetime, &activity.EndDatetime, &activity.IsGlobal,
		&meetingURL, &webURL, &activity.CreatedAt, &activity.UpdatedAt)

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

	return activity, err
}

// parseDateTime parses date and time strings into a time.Time
func parseDateTime(date, timeStr string) (time.Time, error) {
	dateTimeStr := date + " " + timeStr
	return time.Parse("2006-01-02 15:04", dateTimeStr)
}

// AdminIndex handles the admin panel page
func (h *Handlers) AdminIndex(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check admin permissions
	if !auth.UserHasAccess(c, "ADMIN") {
		c.String(http.StatusForbidden, "Acceso denegado")
		return
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Panel de Administración"

	h.renderTemplate(c, "admin.html", data)
}

// AdminUsuarios handles user management
func (h *Handlers) AdminUsuarios(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check admin permissions
	if !auth.UserHasAccess(c, "ADMIN") {
		c.String(http.StatusForbidden, "Acceso denegado")
		return
	}

	// Get all users
	users, err := h.getAllUsers()
	if err != nil {
		users = []models.User{} // Empty slice if error
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Gestión de Usuarios"
	data["Users"] = users

	h.renderTemplate(c, "admin_usuarios.html", data)
}

// AdminUsuarioCrear handles user creation
func (h *Handlers) AdminUsuarioCrear(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check admin permissions
	if !auth.UserHasAccess(c, "ADMIN") {
		c.String(http.StatusForbidden, "Acceso denegado")
		return
	}

	if c.Request.Method == http.MethodPost {
		h.handleUserCreate(c)
		return
	}

	// Show creation form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Crear Usuario"
	data["Action"] = "crear"

	h.renderTemplate(c, "admin_usuario_form.html", data)
}

// handleUserCreate processes user creation
func (h *Handlers) handleUserCreate(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	displayName := c.PostForm("display_name")
	email := c.PostForm("email")
	permissions := c.PostFormArray("permissions")

	if username == "" || password == "" || displayName == "" || email == "" {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Usuario"
		data["Action"] = "crear"
		data["ErrorMessage"] = "Todos los campos son requeridos"
		data["FormData"] = gin.H{
			"username":     username,
			"display_name": displayName,
			"email":        email,
			"permissions":  permissions,
		}
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Usuario"
		data["Action"] = "crear"
		data["ErrorMessage"] = "Error al procesar la contraseña"
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Usuario"
		data["Action"] = "crear"
		data["ErrorMessage"] = "Error en la base de datos"
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Insert user
	userQuery := `INSERT INTO users (username, password_hash, display_name, email, updated_at) 
				  VALUES (?, ?, ?, ?, datetime('now'))`

	result, err := tx.Exec(userQuery, username, hashedPassword, displayName, email)
	if err != nil {
		tx.Rollback()
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Usuario"
		data["Action"] = "crear"
		data["ErrorMessage"] = "Error al crear el usuario: " + err.Error()
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	userID, _ := result.LastInsertId()

	// Insert permissions
	for _, permission := range permissions {
		permQuery := `INSERT INTO user_permissions (user_id, permission) VALUES (?, ?)`
		_, err := tx.Exec(permQuery, userID, permission)
		if err != nil {
			tx.Rollback()
			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Crear Usuario"
			data["Action"] = "crear"
			data["ErrorMessage"] = "Error al asignar permisos: " + err.Error()
			h.renderTemplate(c, "admin_usuario_form.html", data)
			return
		}
	}

	tx.Commit()
	c.Redirect(http.StatusFound, "/admin/usuarios?success=Usuario creado correctamente")
}

// AdminUsuarioEditar handles user editing
func (h *Handlers) AdminUsuarioEditar(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check admin permissions
	if !auth.UserHasAccess(c, "ADMIN") {
		c.String(http.StatusForbidden, "Acceso denegado")
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.Redirect(http.StatusFound, "/admin/usuarios?error=ID de usuario requerido")
		return
	}

	if c.Request.Method == http.MethodPost {
		h.handleUserUpdate(c, userID)
		return
	}

	// Get user data
	editUser, err := h.getUserByID(userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/usuarios?error=Usuario no encontrado")
		return
	}

	// Get user permissions
	permissions, _ := h.getUserPermissions(userID)

	// Show edit form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Editar Usuario"
	data["Action"] = "editar"
	data["EditUser"] = editUser
	data["UserPermissions"] = permissions

	h.renderTemplate(c, "admin_usuario_form.html", data)
}

// handleUserUpdate processes user updates
func (h *Handlers) handleUserUpdate(c *gin.Context, userID string) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	displayName := c.PostForm("display_name")
	email := c.PostForm("email")
	permissions := c.PostFormArray("permissions")

	if username == "" || displayName == "" || email == "" {
		editUser, _ := h.getUserByID(userID)
		userPermissions, _ := h.getUserPermissions(userID)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Usuario"
		data["Action"] = "editar"
		data["EditUser"] = editUser
		data["UserPermissions"] = userPermissions
		data["ErrorMessage"] = "Username, nombre y email son requeridos"
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		editUser, _ := h.getUserByID(userID)
		userPermissions, _ := h.getUserPermissions(userID)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Usuario"
		data["Action"] = "editar"
		data["EditUser"] = editUser
		data["UserPermissions"] = userPermissions
		data["ErrorMessage"] = "Error en la base de datos"
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Update user
	var userQuery string
	var args []interface{}

	if password != "" {
		// Hash new password if provided
		hashedPassword, err := auth.HashPassword(password)
		if err != nil {
			tx.Rollback()
			editUser, _ := h.getUserByID(userID)
			userPermissions, _ := h.getUserPermissions(userID)
			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Editar Usuario"
			data["Action"] = "editar"
			data["EditUser"] = editUser
			data["UserPermissions"] = userPermissions
			data["ErrorMessage"] = "Error al procesar la contraseña"
			h.renderTemplate(c, "admin_usuario_form.html", data)
			return
		}
		userQuery = `UPDATE users SET username = ?, password_hash = ?, display_name = ?, email = ?, updated_at = datetime('now') WHERE id = ?`
		args = []interface{}{username, hashedPassword, displayName, email, userID}
	} else {
		userQuery = `UPDATE users SET username = ?, display_name = ?, email = ?, updated_at = datetime('now') WHERE id = ?`
		args = []interface{}{username, displayName, email, userID}
	}

	_, err = tx.Exec(userQuery, args...)
	if err != nil {
		tx.Rollback()
		editUser, _ := h.getUserByID(userID)
		userPermissions, _ := h.getUserPermissions(userID)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Usuario"
		data["Action"] = "editar"
		data["EditUser"] = editUser
		data["UserPermissions"] = userPermissions
		data["ErrorMessage"] = "Error al actualizar el usuario: " + err.Error()
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Delete existing permissions
	_, err = tx.Exec(`DELETE FROM user_permissions WHERE user_id = ?`, userID)
	if err != nil {
		tx.Rollback()
		editUser, _ := h.getUserByID(userID)
		userPermissions, _ := h.getUserPermissions(userID)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Usuario"
		data["Action"] = "editar"
		data["EditUser"] = editUser
		data["UserPermissions"] = userPermissions
		data["ErrorMessage"] = "Error al actualizar permisos"
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Insert new permissions
	for _, permission := range permissions {
		permQuery := `INSERT INTO user_permissions (user_id, permission) VALUES (?, ?)`
		_, err := tx.Exec(permQuery, userID, permission)
		if err != nil {
			tx.Rollback()
			editUser, _ := h.getUserByID(userID)
			userPermissions, _ := h.getUserPermissions(userID)
			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Editar Usuario"
			data["Action"] = "editar"
			data["EditUser"] = editUser
			data["UserPermissions"] = userPermissions
			data["ErrorMessage"] = "Error al asignar permisos: " + err.Error()
			h.renderTemplate(c, "admin_usuario_form.html", data)
			return
		}
	}

	tx.Commit()
	c.Redirect(http.StatusFound, "/admin/usuarios?success=Usuario actualizado correctamente")
}

// AdminUsuarioEliminar handles user deletion
func (h *Handlers) AdminUsuarioEliminar(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check admin permissions
	if !auth.UserHasAccess(c, "ADMIN") {
		c.String(http.StatusForbidden, "Acceso denegado")
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.Redirect(http.StatusFound, "/admin/usuarios?error=ID de usuario requerido")
		return
	}

	// Prevent deleting yourself
	if strconv.Itoa(user.ID) == userID {
		c.Redirect(http.StatusFound, "/admin/usuarios?error=No puedes eliminar tu propio usuario")
		return
	}

	// Delete user (CASCADE will delete permissions and sessions)
	query := `DELETE FROM users WHERE id = ?`
	result, err := database.DB.Exec(query, userID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/usuarios?error=Error al eliminar el usuario")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.Redirect(http.StatusFound, "/admin/usuarios?error=Usuario no encontrado")
		return
	}

	c.Redirect(http.StatusFound, "/admin/usuarios?success=Usuario eliminado correctamente")
}

// AdminCentros handles center management
func (h *Handlers) AdminCentros(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check admin permissions
	if !auth.UserHasAccess(c, "ADMIN") {
		c.String(http.StatusForbidden, "Acceso denegado")
		return
	}

	// Get all centers
	centers, err := h.getAllCenters()
	if err != nil {
		centers = []models.Center{} // Empty slice if error
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Gestión de Centros"
	data["Centers"] = centers

	h.renderTemplate(c, "admin_centros.html", data)
}

// AdminMaterialesReport handles materials report page
func (h *Handlers) AdminMaterialesReport(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check admin permissions
	if !auth.UserHasAccess(c, "ADMIN") {
		c.String(http.StatusForbidden, "Acceso denegado")
		return
	}

	// Get all centers for filter dropdown
	centers, err := h.getAllCenters()
	if err != nil {
		centers = []models.Center{}
	}

	// Get all materials with center information
	materials, err := h.getAllMaterialsWithCenter()
	if err != nil {
		materials = []models.MaterialWithCenter{}
	}

	// Calculate statistics
	stats := h.calculateMaterialStats(materials)

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Informe de Materiales por Centro"
	data["Centers"] = centers
	data["Materials"] = materials
	data["Stats"] = stats

	h.renderTemplate(c, "admin_materiales_report.html", data)
}

// AdminActividadesReport handles activities report page
func (h *Handlers) AdminActividadesReport(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check admin permissions
	if !auth.UserHasAccess(c, "ADMIN") {
		c.String(http.StatusForbidden, "Acceso denegado")
		return
	}

	// Get all centers for filter dropdown
	centers, err := h.getAllCenters()
	if err != nil {
		centers = []models.Center{}
	}

	// Get recent activities with center information
	activities, err := h.getAllActivitiesWithCenter()
	if err != nil {
		activities = []models.ActivityWithCenter{}
	}

	// Calculate statistics
	stats := h.calculateActivityStats()

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Informe de Actividades"
	data["Centers"] = centers
	data["Activities"] = activities
	data["Stats"] = stats

	h.renderTemplate(c, "admin_actividades_report.html", data)
}

// AdminConfiguracion handles system configuration page
func (h *Handlers) AdminConfiguracion(c *gin.Context) {
	user := auth.GetCurrentUser(c)
	if user == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Check admin permissions
	if !auth.UserHasAccess(c, "ADMIN") {
		c.String(http.StatusForbidden, "Acceso denegado")
		return
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Configuración del Sistema"

	h.renderTemplate(c, "admin_configuracion.html", data)
}

// Helper functions for admin module
func (h *Handlers) getAllUsers() ([]models.User, error) {
	query := `SELECT id, username, display_name, email, created_at, updated_at FROM users ORDER BY username`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}

		// Get user permissions
		permissions, _ := h.getUserPermissions(strconv.Itoa(user.ID))
		user.Permissions = permissions

		users = append(users, user)
	}

	return users, nil
}

func (h *Handlers) getUserByID(userID string) (models.User, error) {
	var user models.User
	query := `SELECT id, username, display_name, email, created_at, updated_at FROM users WHERE id = ?`

	err := database.DB.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (h *Handlers) getUserPermissions(userID string) ([]string, error) {
	query := `SELECT permission FROM user_permissions WHERE user_id = ?`

	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			continue
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

func (h *Handlers) getAllCenters() ([]models.Center, error) {
	query := `SELECT id, name, timezone, created_at, updated_at FROM centers ORDER BY name`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var centers []models.Center
	for rows.Next() {
		var center models.Center
		err := rows.Scan(&center.ID, &center.Name, &center.Timezone, &center.CreatedAt, &center.UpdatedAt)
		if err != nil {
			continue
		}
		centers = append(centers, center)
	}

	return centers, nil
}

// getAllMaterialsWithCenter gets all materials with their center names
func (h *Handlers) getAllMaterialsWithCenter() ([]models.MaterialWithCenter, error) {
	query := `
		SELECT m.id, m.center_id, c.name as center_name, m.name, m.photo_path, m.unit, 
			   m.available_quantity, m.minimum_quantity, m.notes, m.created_at, m.updated_at
		FROM materials m
		JOIN centers c ON m.center_id = c.id
		ORDER BY c.name, m.name
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []models.MaterialWithCenter
	for rows.Next() {
		var material models.MaterialWithCenter
		err := rows.Scan(&material.ID, &material.CenterID, &material.CenterName, &material.Name,
			&material.PhotoPath, &material.Unit, &material.AvailableQuantity, &material.MinimumQuantity,
			&material.Notes, &material.CreatedAt, &material.UpdatedAt)
		if err != nil {
			continue
		}
		materials = append(materials, material)
	}

	return materials, nil
}

// calculateMaterialStats calculates statistics from materials data
func (h *Handlers) calculateMaterialStats(materials []models.MaterialWithCenter) models.MaterialStats {
	stats := models.MaterialStats{}
	stats.TotalMaterials = len(materials)
	
	centersMap := make(map[int]bool)
	for _, material := range materials {
		centersMap[material.CenterID] = true
		
		if material.AvailableQuantity >= material.MinimumQuantity {
			stats.HealthyMaterials++
		} else {
			stats.LowStockMaterials++
		}
	}
	stats.TotalCenters = len(centersMap)
	
	return stats
}

// getAllActivitiesWithCenter gets all activities with their center names
func (h *Handlers) getAllActivitiesWithCenter() ([]models.ActivityWithCenter, error) {
	query := `
		SELECT a.id, a.center_id, COALESCE(c.name, 'Global') as center_name, a.title, a.description,
			   a.start_datetime, a.end_datetime, a.is_global, a.meeting_url, a.web_url,
			   a.created_at, a.updated_at
		FROM activities a
		LEFT JOIN centers c ON a.center_id = c.id
		ORDER BY a.start_datetime DESC
		LIMIT 10
	`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.ActivityWithCenter
	for rows.Next() {
		var activity models.ActivityWithCenter
		err := rows.Scan(&activity.ID, &activity.CenterID, &activity.CenterName, &activity.Title,
			&activity.Description, &activity.StartDatetime, &activity.EndDatetime, &activity.IsGlobal,
			&activity.MeetingURL, &activity.WebURL, &activity.CreatedAt, &activity.UpdatedAt)
		if err != nil {
			continue
		}
		activities = append(activities, activity)
	}

	return activities, nil
}

// calculateActivityStats calculates statistics from activities data
func (h *Handlers) calculateActivityStats() models.ActivityStats {
	stats := models.ActivityStats{}
	
	// Count total activities
	database.DB.QueryRow("SELECT COUNT(*) FROM activities").Scan(&stats.TotalActivities)
	
	// Count activities by status based on current time
	now := time.Now()
	
	// Pending (future start date)
	database.DB.QueryRow("SELECT COUNT(*) FROM activities WHERE start_datetime > ?", now).Scan(&stats.PendingActivities)
	
	// In Progress (current time between start and end)
	database.DB.QueryRow("SELECT COUNT(*) FROM activities WHERE start_datetime <= ? AND end_datetime >= ?", now, now).Scan(&stats.InProgressActivities)
	
	// Completed (past end date)
	database.DB.QueryRow("SELECT COUNT(*) FROM activities WHERE end_datetime < ?", now).Scan(&stats.CompletedActivities)
	
	return stats
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
