package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/internal/database"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/gin-gonic/gin"
)

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
	query := `SELECT id, center_id, name, photo_path, unit, available_quantity, minimum_quantity, notes, category, created_at, updated_at 
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
			&material.Notes, &material.Category, &material.CreatedAt, &material.UpdatedAt)
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
	category := c.PostForm("categoria")
	availableQty := c.PostForm("cantidad_disponible")
	minimumQty := c.PostForm("cantidad_minima")
	notes := c.PostForm("notas")

	if name == "" || unit == "" || category == "" {
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Material"
		data["Centro"] = centro
		data["Action"] = "crear"
		data["ErrorMessage"] = "El nombre, la unidad y la categoría son requeridos"
		data["FormData"] = gin.H{
			"nombre":              name,
			"unidad":              unit,
			"categoria":           category,
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
	query := `INSERT INTO materials (center_id, name, unit, category, available_quantity, minimum_quantity, notes, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))`

	_, err = database.DB.Exec(query, centerID, name, unit, category, availableQtyInt, minimumQtyInt, notes)
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
	category := c.PostForm("categoria")
	availableQty := c.PostForm("cantidad_disponible")
	minimumQty := c.PostForm("cantidad_minima")
	notes := c.PostForm("notas")

	if name == "" || unit == "" || category == "" {
		material, _ := h.getMaterial(materialID, centro)
		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Editar Material"
		data["Centro"] = centro
		data["Action"] = "editar"
		data["Material"] = material
		data["ErrorMessage"] = "El nombre, la unidad y la categoría son requeridos"
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
	query := `UPDATE materials SET name = ?, unit = ?, category = ?, available_quantity = ?, minimum_quantity = ?, notes = ?, updated_at = datetime('now')
			  WHERE id = ? AND center_id = ?`

	result, err := database.DB.Exec(query, name, unit, category, availableQtyInt, minimumQtyInt, notes, materialID, centerID)
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
	query := `SELECT id, center_id, name, photo_path, unit, available_quantity, minimum_quantity, notes, category, created_at, updated_at 
			  FROM materials WHERE id = ? AND center_id = (SELECT id FROM centers WHERE name = ?)`

	var photoPath sql.NullString
	err := database.DB.QueryRow(query, materialID, centro).Scan(
		&material.ID, &material.CenterID, &material.Name, &photoPath,
		&material.Unit, &material.AvailableQuantity, &material.MinimumQuantity,
		&material.Notes, &material.Category, &material.CreatedAt, &material.UpdatedAt)

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