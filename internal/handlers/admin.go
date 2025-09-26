package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/internal/database"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/gin-gonic/gin"
)

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

	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// Get all users with pagination
	users, totalCount, err := h.getAllUsersPaginated(page, 25)
	if err != nil {
		users = []models.User{} // Empty slice if error
		totalCount = 0
	}

	// Create pagination info
	pagination := models.NewPaginationInfo(page, 25, totalCount)

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Gestión de Usuarios"
	data["Users"] = users
	data["Pagination"] = pagination

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
	centers, err := h.getAllCenters()
	if err != nil {
		centers = []models.Center{}
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Crear Usuario"
	data["Action"] = "crear"
	data["Centers"] = centers
	data["DefaultCenterID"] = 0 // Default to no center selected

	h.renderTemplate(c, "admin_usuario_form.html", data)
}

// handleUserCreate processes user creation
func (h *Handlers) handleUserCreate(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	displayName := c.PostForm("display_name")
	email := c.PostForm("email")
	permissions := c.PostFormArray("permissions")
	defaultCenterID := c.PostForm("default_center_id")
	forceDefaultCenter := c.PostForm("force_default_center") == "on"

	if username == "" || password == "" || displayName == "" || email == "" {
		centers, _ := h.getAllCenters()
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Usuario"
		data["Action"] = "crear"
		data["Centers"] = centers
		data["ErrorMessage"] = "Todos los campos son requeridos"
		data["FormData"] = gin.H{
			"username":              username,
			"display_name":          displayName,
			"email":                 email,
			"permissions":           permissions,
			"default_center_id":     defaultCenterID,
			"force_default_center":  forceDefaultCenter,
		}
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		centers, _ := h.getAllCenters()
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Usuario"
		data["Action"] = "crear"
		data["Centers"] = centers
		data["ErrorMessage"] = "Error al procesar la contraseña"
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		centers, _ := h.getAllCenters()
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Usuario"
		data["Action"] = "crear"
		data["Centers"] = centers
		data["ErrorMessage"] = "Error en la base de datos"
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Prepare default_center_id for database (NULL if empty)
	var defaultCenterIDPtr *int
	if defaultCenterID != "" && defaultCenterID != "0" {
		if id, err := strconv.Atoi(defaultCenterID); err == nil {
			defaultCenterIDPtr = &id
		}
	}

	// Insert user
	userQuery := `INSERT INTO users (username, password_hash, display_name, email, default_center_id, force_default_center, updated_at) 
				  VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`

	result, err := tx.Exec(userQuery, username, hashedPassword, displayName, email, defaultCenterIDPtr, forceDefaultCenter)
	if err != nil {
		tx.Rollback()
		centers, _ := h.getAllCenters()
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Usuario"
		data["Action"] = "crear"
		data["Centers"] = centers
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

	// Get all centers for the select dropdown
	centers, err := h.getAllCenters()
	if err != nil {
		centers = []models.Center{}
	}

	// Convert default center ID from pointer to simple int for template
	defaultCenterID := 0
	if editUser.DefaultCenterID != nil {
		defaultCenterID = *editUser.DefaultCenterID
	}

	// Show edit form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Editar Usuario"
	data["Action"] = "editar"
	data["EditUser"] = editUser
	data["UserPermissions"] = permissions
	data["Centers"] = centers
	data["DefaultCenterID"] = defaultCenterID

	h.renderTemplate(c, "admin_usuario_form.html", data)
}

// handleUserUpdate processes user updates
func (h *Handlers) handleUserUpdate(c *gin.Context, userID string) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	displayName := c.PostForm("display_name")
	email := c.PostForm("email")
	permissions := c.PostFormArray("permissions")
	defaultCenterID := c.PostForm("default_center_id")
	forceDefaultCenter := c.PostForm("force_default_center") == "on"

	if username == "" || displayName == "" || email == "" {
		editUser, _ := h.getUserByID(userID)
		userPermissions, _ := h.getUserPermissions(userID)
		centers, _ := h.getAllCenters()
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Usuario"
		data["Action"] = "editar"
		data["EditUser"] = editUser
		data["UserPermissions"] = userPermissions
		data["Centers"] = centers
		data["ErrorMessage"] = "Username, nombre y email son requeridos"
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Start transaction
	tx, err := database.DB.Begin()
	if err != nil {
		editUser, _ := h.getUserByID(userID)
		userPermissions, _ := h.getUserPermissions(userID)
		centers, _ := h.getAllCenters()
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Usuario"
		data["Action"] = "editar"
		data["EditUser"] = editUser
		data["UserPermissions"] = userPermissions
		data["Centers"] = centers
		data["ErrorMessage"] = "Error en la base de datos"
		h.renderTemplate(c, "admin_usuario_form.html", data)
		return
	}

	// Prepare default_center_id for database (NULL if empty)
	var defaultCenterIDPtr *int
	if defaultCenterID != "" && defaultCenterID != "0" {
		if id, err := strconv.Atoi(defaultCenterID); err == nil {
			defaultCenterIDPtr = &id
		}
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
			centers, _ := h.getAllCenters()
			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Editar Usuario"
			data["Action"] = "editar"
			data["EditUser"] = editUser
			data["UserPermissions"] = userPermissions
			data["Centers"] = centers
			data["ErrorMessage"] = "Error al procesar la contraseña"
			h.renderTemplate(c, "admin_usuario_form.html", data)
			return
		}
		userQuery = `UPDATE users SET username = ?, password_hash = ?, display_name = ?, email = ?, default_center_id = ?, force_default_center = ?, updated_at = datetime('now') WHERE id = ?`
		args = []interface{}{username, hashedPassword, displayName, email, defaultCenterIDPtr, forceDefaultCenter, userID}
	} else {
		userQuery = `UPDATE users SET username = ?, display_name = ?, email = ?, default_center_id = ?, force_default_center = ?, updated_at = datetime('now') WHERE id = ?`
		args = []interface{}{username, displayName, email, defaultCenterIDPtr, forceDefaultCenter, userID}
	}

	_, err = tx.Exec(userQuery, args...)
	if err != nil {
		tx.Rollback()
		editUser, _ := h.getUserByID(userID)
		userPermissions, _ := h.getUserPermissions(userID)
		centers, _ := h.getAllCenters()
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Usuario"
		data["Action"] = "editar"
		data["EditUser"] = editUser
		data["UserPermissions"] = userPermissions
		data["Centers"] = centers
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
		centers, _ := h.getAllCenters()
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Usuario"
		data["Action"] = "editar"
		data["EditUser"] = editUser
		data["UserPermissions"] = userPermissions
		data["Centers"] = centers
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
			centers, _ := h.getAllCenters()
			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Editar Usuario"
			data["Action"] = "editar"
			data["EditUser"] = editUser
			data["UserPermissions"] = userPermissions
			data["Centers"] = centers
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

// AdminCentroCrear handles center creation
func (h *Handlers) AdminCentroCrear(c *gin.Context) {
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
		h.handleCenterCreate(c)
		return
	}

	// Show creation form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Crear Centro"
	data["Action"] = "crear"

	h.renderTemplate(c, "admin_centro_form.html", data)
}

// handleCenterCreate processes center creation
func (h *Handlers) handleCenterCreate(c *gin.Context) {
	name := c.PostForm("nombre")
	timezone := c.PostForm("timezone")

	if name == "" {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Centro"
		data["Action"] = "crear"
		data["ErrorMessage"] = "El nombre del centro es obligatorio"
		data["FormData"] = map[string]string{"nombre": name, "timezone": timezone}
		h.renderTemplate(c, "admin_centro_form.html", data)
		return
	}

	// Set default timezone if none provided
	if timezone == "" {
		timezone = "Europe/Madrid"
	}

	// Create center
	err := h.createCenter(name, timezone)
	if err != nil {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Centro"
		data["Action"] = "crear"
		data["ErrorMessage"] = "Error al crear el centro: " + err.Error()
		data["FormData"] = map[string]string{"nombre": name, "timezone": timezone}
		h.renderTemplate(c, "admin_centro_form.html", data)
		return
	}

	c.Redirect(http.StatusFound, "/admin/centros?success=Centro creado correctamente")
}

// AdminCentroEditar handles center editing
func (h *Handlers) AdminCentroEditar(c *gin.Context) {
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

	centerID := c.Param("id")
	if centerID == "" {
		c.Redirect(http.StatusFound, "/admin/centros?error=ID de centro requerido")
		return
	}

	if c.Request.Method == http.MethodPost {
		h.handleCenterUpdate(c, centerID)
		return
	}

	// Get center data
	center, err := h.getCenterByID(centerID)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/centros?error=Centro no encontrado")
		return
	}

	// Show edit form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Editar Centro"
	data["Action"] = "editar"
	data["Center"] = center

	h.renderTemplate(c, "admin_centro_form.html", data)
}

// handleCenterUpdate processes center updates
func (h *Handlers) handleCenterUpdate(c *gin.Context, centerID string) {
	name := c.PostForm("nombre")
	timezone := c.PostForm("timezone")

	if name == "" {
		center, _ := h.getCenterByID(centerID)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Centro"
		data["Action"] = "editar"
		data["Center"] = center
		data["ErrorMessage"] = "El nombre del centro es obligatorio"
		data["FormData"] = map[string]string{"nombre": name, "timezone": timezone}
		h.renderTemplate(c, "admin_centro_form.html", data)
		return
	}

	// Set default timezone if none provided
	if timezone == "" {
		timezone = "Europe/Madrid"
	}

	// Update center
	err := h.updateCenter(centerID, name, timezone)
	if err != nil {
		center, _ := h.getCenterByID(centerID)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Centro"
		data["Action"] = "editar"
		data["Center"] = center
		data["ErrorMessage"] = "Error al actualizar el centro: " + err.Error()
		data["FormData"] = map[string]string{"nombre": name, "timezone": timezone}
		h.renderTemplate(c, "admin_centro_form.html", data)
		return
	}

	c.Redirect(http.StatusFound, "/admin/centros?success=Centro actualizado correctamente")
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

// AdminFiles handles file management page
func (h *Handlers) AdminFiles(c *gin.Context) {
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
	data["PageTitle"] = "Figaró - Gestión de Archivos"

	h.renderTemplate(c, "admin_files.html", data)
}

// AdminCentroAulas handles classroom management for a specific center
func (h *Handlers) AdminCentroAulas(c *gin.Context) {
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

	centerID := c.Param("center_id")

	// Get center information
	center, err := h.getCenterByID(centerID)
	if err != nil {
		c.String(http.StatusNotFound, "Centro no encontrado")
		return
	}

	// Get classrooms for this center
	classrooms, err := h.getClassroomsByCenter(centerID)
	if err != nil {
		classrooms = []models.Classroom{} // Empty slice if error
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Aulas de " + center.Name
	data["Center"] = center
	data["Classrooms"] = classrooms

	h.renderTemplate(c, "admin_aulas.html", data)
}

// AdminAulaCrear handles classroom creation
func (h *Handlers) AdminAulaCrear(c *gin.Context) {
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

	centerID := c.Param("center_id")

	// Get center information
	center, err := h.getCenterByID(centerID)
	if err != nil {
		c.String(http.StatusNotFound, "Centro no encontrado")
		return
	}

	if c.Request.Method == "POST" {
		// Handle form submission
		name := c.PostForm("nombre")
		if name == "" {
			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Crear Aula"
			data["Center"] = center
			data["ErrorMessage"] = "El nombre del aula es obligatorio"
			data["FormData"] = map[string]string{"nombre": name}
			h.renderTemplate(c, "admin_aula_form.html", data)
			return
		}

		// Create classroom
		err := h.createClassroom(centerID, name)
		if err != nil {
			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Crear Aula"
			data["Center"] = center
			data["ErrorMessage"] = "Error al crear el aula: " + err.Error()
			data["FormData"] = map[string]string{"nombre": name}
			h.renderTemplate(c, "admin_aula_form.html", data)
			return
		}

		c.Redirect(http.StatusFound, "/admin/centros/aulas/"+centerID)
		return
	}

	// Show form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Crear Aula"
	data["Center"] = center
	data["Action"] = "crear"

	h.renderTemplate(c, "admin_aula_form.html", data)
}

// AdminAulaEditar handles classroom editing
func (h *Handlers) AdminAulaEditar(c *gin.Context) {
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

	centerID := c.Param("center_id")
	aulaID := c.Param("aula_id")

	// Get center information
	center, err := h.getCenterByID(centerID)
	if err != nil {
		c.String(http.StatusNotFound, "Centro no encontrado")
		return
	}

	// Get classroom information
	classroom, err := h.getClassroomByID(aulaID)
	if err != nil {
		c.String(http.StatusNotFound, "Aula no encontrada")
		return
	}

	if c.Request.Method == "POST" {
		// Handle form submission
		name := c.PostForm("nombre")
		if name == "" {
			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Editar Aula"
			data["Center"] = center
			data["Classroom"] = classroom
			data["ErrorMessage"] = "El nombre del aula es obligatorio"
			data["FormData"] = map[string]string{"nombre": name}
			data["Action"] = "editar"
			h.renderTemplate(c, "admin_aula_form.html", data)
			return
		}

		// Update classroom
		err := h.updateClassroom(aulaID, name)
		if err != nil {
			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Editar Aula"
			data["Center"] = center
			data["Classroom"] = classroom
			data["ErrorMessage"] = "Error al actualizar el aula: " + err.Error()
			data["FormData"] = map[string]string{"nombre": name}
			data["Action"] = "editar"
			h.renderTemplate(c, "admin_aula_form.html", data)
			return
		}

		c.Redirect(http.StatusFound, "/admin/centros/aulas/"+centerID)
		return
	}

	// Show form
	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Editar Aula"
	data["Center"] = center
	data["Classroom"] = classroom
	data["Action"] = "editar"

	h.renderTemplate(c, "admin_aula_form.html", data)
}

// AdminAulaEliminar handles classroom deletion
func (h *Handlers) AdminAulaEliminar(c *gin.Context) {
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

	centerID := c.Param("center_id")
	aulaID := c.Param("aula_id")

	// Delete classroom
	err := h.deleteClassroom(aulaID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error al eliminar el aula: "+err.Error())
		return
	}

	c.Redirect(http.StatusFound, "/admin/centros/aulas/"+centerID)
}

// Helper functions for admin module
func (h *Handlers) getAllUsers() ([]models.User, error) {
	users, _, err := h.getAllUsersPaginated(1, 1000) // Large limit for backward compatibility
	return users, err
}

// getAllUsersPaginated retrieves users with pagination
func (h *Handlers) getAllUsersPaginated(page, perPage int) ([]models.User, int, error) {
	// First get total count
	countQuery := `SELECT COUNT(*) FROM users`
	var totalCount int
	err := database.DB.QueryRow(countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	pagination := models.NewPaginationInfo(page, perPage, totalCount)

	// Get paginated results
	query := `SELECT id, username, display_name, email, default_center_id, force_default_center, created_at, updated_at FROM users ORDER BY username LIMIT ? OFFSET ?`

	rows, err := database.DB.Query(query, perPage, pagination.Offset)
	if err != nil {
		return nil, totalCount, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.DefaultCenterID, &user.ForceDefaultCenter, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue
		}

		// Get user permissions
		permissions, _ := h.getUserPermissions(strconv.Itoa(user.ID))
		user.Permissions = permissions

		users = append(users, user)
	}

	return users, totalCount, nil
}

func (h *Handlers) getUserByID(userID string) (models.User, error) {
	var user models.User
	query := `SELECT id, username, display_name, email, default_center_id, force_default_center, created_at, updated_at FROM users WHERE id = ?`

	err := database.DB.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.DefaultCenterID, &user.ForceDefaultCenter, &user.CreatedAt, &user.UpdatedAt)
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
			   m.available_quantity, m.minimum_quantity, m.notes, m.category, m.created_at, m.updated_at
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
			&material.Notes, &material.Category, &material.CreatedAt, &material.UpdatedAt)
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
			   a.start_datetime, a.end_datetime, a.is_global, a.status, a.meeting_url, a.web_url,
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
			&activity.Status, &activity.MeetingURL, &activity.WebURL, &activity.CreatedAt, &activity.UpdatedAt)
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

// getCenterByID gets a center by ID
func (h *Handlers) getCenterByID(centerID string) (models.Center, error) {
	var center models.Center
	query := `SELECT id, name, timezone, created_at, updated_at FROM centers WHERE id = ?`

	err := database.DB.QueryRow(query, centerID).Scan(
		&center.ID, &center.Name, &center.Timezone, &center.CreatedAt, &center.UpdatedAt)
	return center, err
}

// getClassroomsByCenter gets all classrooms for a center
func (h *Handlers) getClassroomsByCenter(centerID string) ([]models.Classroom, error) {
	query := `SELECT id, center_id, name, created_at, updated_at FROM classrooms WHERE center_id = ? ORDER BY name`

	rows, err := database.DB.Query(query, centerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classrooms []models.Classroom
	for rows.Next() {
		var classroom models.Classroom
		err := rows.Scan(&classroom.ID, &classroom.CenterID, &classroom.Name, &classroom.CreatedAt, &classroom.UpdatedAt)
		if err != nil {
			continue
		}
		classrooms = append(classrooms, classroom)
	}

	return classrooms, nil
}

// getClassroomByID gets a classroom by ID
func (h *Handlers) getClassroomByID(classroomID string) (models.Classroom, error) {
	var classroom models.Classroom
	query := `SELECT id, center_id, name, created_at, updated_at FROM classrooms WHERE id = ?`

	err := database.DB.QueryRow(query, classroomID).Scan(
		&classroom.ID, &classroom.CenterID, &classroom.Name, &classroom.CreatedAt, &classroom.UpdatedAt)
	return classroom, err
}

// createClassroom creates a new classroom
func (h *Handlers) createClassroom(centerID, name string) error {
	query := `INSERT INTO classrooms (center_id, name, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := database.DB.Exec(query, centerID, name)
	return err
}

// updateClassroom updates a classroom
func (h *Handlers) updateClassroom(classroomID, name string) error {
	query := `UPDATE classrooms SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := database.DB.Exec(query, name, classroomID)
	return err
}

// deleteClassroom deletes a classroom
func (h *Handlers) deleteClassroom(classroomID string) error {
	query := `DELETE FROM classrooms WHERE id = ?`
	_, err := database.DB.Exec(query, classroomID)
	return err
}

// createCenter creates a new center
func (h *Handlers) createCenter(name, timezone string) error {
	query := `INSERT INTO centers (name, timezone, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := database.DB.Exec(query, name, timezone)
	return err
}

// updateCenter updates a center
func (h *Handlers) updateCenter(centerID, name, timezone string) error {
	query := `UPDATE centers SET name = ?, timezone = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := database.DB.Exec(query, name, timezone, centerID)
	return err
}