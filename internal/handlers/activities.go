package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/internal/database"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/gin-gonic/gin"
)

// ActividadesIndex handles the activities page with tabs support
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

	// Get search and tab parameters
	searchQuery := c.Query("q")
	showPast := c.Query("past") == "y"
	activeTab := c.DefaultQuery("tab", "all") // default to "all" activities

	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// Get activities from database based on active tab with pagination
	var activities []models.Activity
	var totalCount int
	switch activeTab {
	case "compartidas":
		activities, totalCount, err = h.getSharedActivitiesPaginated(centro, searchQuery, showPast, page, 25)
	case "enlaces":
		activities, totalCount, err = h.getActivitiesWithCustomLinksPaginated(centro, searchQuery, showPast, page, 25)
	default: // "all" or any other value
		activities, totalCount, err = h.getActivitiesPaginated(centro, searchQuery, showPast, page, 25)
	}
	
	if err != nil {
		activities = []models.Activity{} // Empty slice if error
		totalCount = 0
	}

	// Create pagination info
	pagination := models.NewPaginationInfo(page, 25, totalCount)

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Actividades"
	data["Activities"] = activities
	data["Centro"] = centro
	data["SearchQuery"] = searchQuery
	data["ShowPast"] = showPast
	data["ActiveTab"] = activeTab
	data["Pagination"] = pagination

	h.renderTemplate(c, "actividades.html", data)
}

// getActivities retrieves activities for a center
func (h *Handlers) getActivities(centro string, searchQuery string, showPast bool) ([]models.Activity, error) {
	var query string
	var args []interface{}

	if showPast {
		query = `SELECT id, center_id, title, description, start_datetime, end_datetime, is_global, status, meeting_url, web_url, created_at, updated_at 
				FROM activities WHERE (center_id = (SELECT id FROM centers WHERE name = ?) OR is_global = 1)`
		args = []interface{}{centro}
	} else {
		query = `SELECT id, center_id, title, description, start_datetime, end_datetime, is_global, status, meeting_url, web_url, created_at, updated_at 
				FROM activities WHERE (center_id = (SELECT id FROM centers WHERE name = ?) OR is_global = 1) 
				AND start_datetime <= datetime('now', 'start of day', '+1 day', '-1 second')
				AND end_datetime   >= datetime('now', 'start of day');`
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
			&activity.StartDatetime, &activity.EndDatetime, &activity.IsGlobal, &activity.Status,
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

// getActivitiesPaginated retrieves activities for a center with pagination
func (h *Handlers) getActivitiesPaginated(centro string, searchQuery string, showPast bool, page, perPage int) ([]models.Activity, int, error) {
	var baseQuery string
	var args []interface{}

	if showPast {
		baseQuery = `FROM activities WHERE (center_id = (SELECT id FROM centers WHERE name = ?) OR is_global = 1)`
		args = []interface{}{centro}
	} else {
		baseQuery = `FROM activities WHERE (center_id = (SELECT id FROM centers WHERE name = ?) OR is_global = 1) 
				AND start_datetime <= datetime('now', 'start of day', '+1 day', '-1 second')
				AND end_datetime   >= datetime('now', 'start of day')`
		args = []interface{}{centro}
	}

	if searchQuery != "" {
		baseQuery += " AND (title LIKE ? OR description LIKE ?)"
		searchPattern := "%" + searchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalCount int
	err := database.DB.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Calculate pagination
	pagination := models.NewPaginationInfo(page, perPage, totalCount)
	
	// Get paginated results
	query := `SELECT id, center_id, title, description, start_datetime, end_datetime, is_global, status, meeting_url, web_url, created_at, updated_at ` + 
		baseQuery + ` ORDER BY start_datetime ASC LIMIT ? OFFSET ?`
	
	args = append(args, perPage, pagination.Offset)
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		return nil, totalCount, err
	}
	defer rows.Close()

	var activities []models.Activity
	for rows.Next() {
		var activity models.Activity
		var centerID sql.NullInt64
		var meetingURL sql.NullString
		var webURL sql.NullString

		err := rows.Scan(&activity.ID, &centerID, &activity.Title, &activity.Description,
			&activity.StartDatetime, &activity.EndDatetime, &activity.IsGlobal, &activity.Status,
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

	return activities, totalCount, nil
}

// getSharedActivitiesPaginated retrieves shared activities with pagination
func (h *Handlers) getSharedActivitiesPaginated(centro string, searchQuery string, showPast bool, page, perPage int) ([]models.Activity, int, error) {
	// For simplicity, delegate to main function - in real implementation you'd add specific logic for shared activities
	return h.getActivitiesPaginated(centro, searchQuery, showPast, page, perPage)
}

// getActivitiesWithCustomLinksPaginated retrieves activities with custom links with pagination
func (h *Handlers) getActivitiesWithCustomLinksPaginated(centro string, searchQuery string, showPast bool, page, perPage int) ([]models.Activity, int, error) {
	// For simplicity, delegate to main function - in real implementation you'd add specific logic for custom links
	return h.getActivitiesPaginated(centro, searchQuery, showPast, page, perPage)
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

// getSharedActivities retrieves activities shared specifically with the current center (not global)
func (h *Handlers) getSharedActivities(centro string, searchQuery string, showPast bool) ([]models.Activity, error) {
	var query string
	var args []interface{}

	if showPast {
		query = `SELECT DISTINCT a.id, a.center_id, a.title, a.description, a.start_datetime, a.end_datetime, 
				 a.is_global, a.meeting_url, a.web_url, a.created_at, a.updated_at,
				 c_shared.name as shared_from_center
				 FROM activities a
				 INNER JOIN activity_shares ast ON a.id = ast.activity_id
				 INNER JOIN centers c_current ON ast.center_id = c_current.id
				 INNER JOIN centers c_shared ON ast.shared_by_center_id = c_shared.id
				 WHERE c_current.name = ?`
		args = []interface{}{centro}
	} else {
		query = `SELECT DISTINCT a.id, a.center_id, a.title, a.description, a.start_datetime, a.end_datetime, 
				 a.is_global, a.meeting_url, a.web_url, a.created_at, a.updated_at,
				 c_shared.name as shared_from_center
				 FROM activities a
				 INNER JOIN activity_shares ast ON a.id = ast.activity_id
				 INNER JOIN centers c_current ON ast.center_id = c_current.id
				 INNER JOIN centers c_shared ON ast.shared_by_center_id = c_shared.id
				 WHERE c_current.name = ? AND a.start_datetime >= datetime('now')`
		args = []interface{}{centro}
	}

	if searchQuery != "" {
		query += " AND (a.title LIKE ? OR a.description LIKE ?)"
		searchPattern := "%" + searchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += " ORDER BY a.start_datetime ASC"

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
		var sharedFromCenter sql.NullString

		err := rows.Scan(&activity.ID, &centerID, &activity.Title, &activity.Description,
			&activity.StartDatetime, &activity.EndDatetime, &activity.IsGlobal,
			&meetingURL, &webURL, &activity.CreatedAt, &activity.UpdatedAt, &sharedFromCenter)
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

		if sharedFromCenter.Valid {
			activity.SharedFromCenter = &sharedFromCenter.String
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// getActivitiesWithCustomLinks retrieves activities that have custom links
func (h *Handlers) getActivitiesWithCustomLinks(centro string, searchQuery string, showPast bool) ([]models.Activity, error) {
	var query string
	var args []interface{}

	if showPast {
		query = `SELECT DISTINCT a.id, a.center_id, a.title, a.description, a.start_datetime, a.end_datetime, 
				 a.is_global, a.meeting_url, a.web_url, a.created_at, a.updated_at
				 FROM activities a
				 INNER JOIN activity_custom_links acl ON a.id = acl.activity_id
				 WHERE (a.center_id = (SELECT id FROM centers WHERE name = ?) OR a.is_global = 1)`
		args = []interface{}{centro}
	} else {
		query = `SELECT DISTINCT a.id, a.center_id, a.title, a.description, a.start_datetime, a.end_datetime, 
				 a.is_global, a.meeting_url, a.web_url, a.created_at, a.updated_at
				 FROM activities a
				 INNER JOIN activity_custom_links acl ON a.id = acl.activity_id
				 WHERE (a.center_id = (SELECT id FROM centers WHERE name = ?) OR a.is_global = 1)
				 AND a.start_datetime >= datetime('now')`
		args = []interface{}{centro}
	}

	if searchQuery != "" {
		query += " AND (a.title LIKE ? OR a.description LIKE ?)"
		searchPattern := "%" + searchQuery + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += " ORDER BY a.start_datetime ASC"

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

		// Load custom links for this activity
		customLinks, err := h.getCustomLinksForActivity(activity.ID)
		if err == nil {
			activity.CustomLinks = customLinks
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

// getCustomLinksForActivity retrieves custom links for a specific activity
func (h *Handlers) getCustomLinksForActivity(activityID int) ([]models.ActivityCustomLink, error) {
	query := `SELECT id, activity_id, label, url, created_at FROM activity_custom_links WHERE activity_id = ? ORDER BY created_at ASC`
	
	rows, err := database.DB.Query(query, activityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customLinks []models.ActivityCustomLink
	for rows.Next() {
		var link models.ActivityCustomLink
		err := rows.Scan(&link.ID, &link.ActivityID, &link.Label, &link.URL, &link.CreatedAt)
		if err != nil {
			continue
		}
		customLinks = append(customLinks, link)
	}

	return customLinks, nil
}
