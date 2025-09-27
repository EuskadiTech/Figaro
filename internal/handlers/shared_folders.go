package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/EuskadiTech/Figaro/internal/auth"
	"github.com/EuskadiTech/Figaro/internal/database"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/gin-gonic/gin"
)

// CarpetasCompartidasIndex handles the shared folders listing page
func (h *Handlers) CarpetasCompartidasIndex(c *gin.Context) {
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

	// Get shared folders for this center and global ones
	folders, err := h.getSharedFolders(centro)
	if err != nil {
		folders = []models.SharedFolderWithCenter{}
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Carpetas Compartidas"
	data["Centro"] = centro
	data["Folders"] = folders

	h.renderTemplate(c, "carpetas_compartidas.html", data)
}

// CarpetasCompartidasCrear handles shared folder creation
func (h *Handlers) CarpetasCompartidasCrear(c *gin.Context) {
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

	// Get selected center
	centro, err := c.Cookie("centro")
	if err != nil || centro == "" {
		c.Redirect(http.StatusFound, "/elegir_centro")
		return
	}

	if c.Request.Method == http.MethodPost {
		h.handleSharedFolderCreate(c, centro)
		return
	}

	// Show creation form
	centers, err := h.getAllCenters()
	if err != nil {
		centers = []models.Center{}
	}

	data := h.getCommonData(c)
	data["PageTitle"] = "Figaró - Crear Carpeta Compartida"
	data["Centro"] = centro
	data["Centers"] = centers
	data["Action"] = "crear"

	h.renderTemplate(c, "carpeta_compartida_form.html", data)
}

// handleSharedFolderCreate processes shared folder creation
func (h *Handlers) handleSharedFolderCreate(c *gin.Context, centro string) {
	name := c.PostForm("nombre")
	description := c.PostForm("descripcion")
	folderType := c.PostForm("tipo")
	cloudURL := c.PostForm("cloud_url")
	centerIDStr := c.PostForm("center_id")
	isGlobal := c.PostForm("is_global") == "1"

	if name == "" || folderType == "" {
		centers, err := h.getAllCenters()
		if err != nil {
			centers = []models.Center{}
		}

		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Carpeta Compartida"
		data["Centro"] = centro
		data["Centers"] = centers
		data["Action"] = "crear"
		data["ErrorMessage"] = "El nombre y el tipo de carpeta son requeridos"
		data["FormData"] = gin.H{
			"nombre":      name,
			"descripcion": description,
			"tipo":        folderType,
			"cloud_url":   cloudURL,
			"center_id":   centerIDStr,
			"is_global":   isGlobal,
		}
		h.renderTemplate(c, "carpeta_compartida_form.html", data)
		return
	}

	// Validate folder type
	if folderType != "local" && folderType != "cloud" {
		centers, err := h.getAllCenters()
		if err != nil {
			centers = []models.Center{}
		}

		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Carpeta Compartida"
		data["Centro"] = centro
		data["Centers"] = centers
		data["Action"] = "crear"
		data["ErrorMessage"] = "Tipo de carpeta inválido"
		h.renderTemplate(c, "carpeta_compartida_form.html", data)
		return
	}

	// Validate cloud URL if it's a cloud folder
	if folderType == "cloud" && cloudURL == "" {
		centers, err := h.getAllCenters()
		if err != nil {
			centers = []models.Center{}
		}

		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Carpeta Compartida"
		data["Centro"] = centro
		data["Centers"] = centers
		data["Action"] = "crear"
		data["ErrorMessage"] = "La URL de la carpeta en la nube es requerida para carpetas cloud"
		h.renderTemplate(c, "carpeta_compartida_form.html", data)
		return
	}

	var centerID *int
	var localPath *string

	if !isGlobal {
		if centerIDStr == "" {
			centerIDStr = centro
		}
		id, err := strconv.Atoi(centerIDStr)
		if err == nil {
			centerID = &id
		}
	}

	if folderType == "local" {
		// Create local folder path
		folderName := fmt.Sprintf("shared_%s_%s", folderType, name)
		localFolderPath := filepath.Join("shared_folders", folderName)
		fullPath := filepath.Join(h.Config.DataDir, localFolderPath)
		
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			centers, err := h.getAllCenters()
			if err != nil {
				centers = []models.Center{}
			}

			data := h.getCommonData(c)
			data["PageTitle"] = "Figaró - Crear Carpeta Compartida"
			data["Centro"] = centro
			data["Centers"] = centers
			data["Action"] = "crear"
			data["ErrorMessage"] = "Error al crear la carpeta local: " + err.Error()
			h.renderTemplate(c, "carpeta_compartida_form.html", data)
			return
		}
		localPath = &localFolderPath
	}

	// Create the shared folder record
	err := h.createSharedFolder(centerID, name, description, folderType, &cloudURL, localPath)
	if err != nil {
		centers, err := h.getAllCenters()
		if err != nil {
			centers = []models.Center{}
		}

		data := h.getCommonData(c)
		data["PageTitle"] = "Figaró - Crear Carpeta Compartida"
		data["Centro"] = centro
		data["Centers"] = centers
		data["Action"] = "crear"
		data["ErrorMessage"] = "Error al crear la carpeta compartida: " + err.Error()
		h.renderTemplate(c, "carpeta_compartida_form.html", data)
		return
	}

	c.Redirect(http.StatusFound, "/carpetas-compartidas")
}

// getSharedFolders retrieves shared folders for a center and global ones
func (h *Handlers) getSharedFolders(centro string) ([]models.SharedFolderWithCenter, error) {
	// First, get the center ID from the center name
	var centerID int
	centerQuery := `SELECT id FROM centers WHERE name = ?`
	err := database.DB.QueryRow(centerQuery, centro).Scan(&centerID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT sf.id, sf.center_id, COALESCE(c.name, 'Global') as center_name, 
		       sf.name, sf.description, sf.type, sf.cloud_url, sf.local_path, 
		       sf.is_active, sf.created_at, sf.updated_at
		FROM shared_folders sf
		LEFT JOIN centers c ON sf.center_id = c.id
		WHERE (sf.center_id = ? OR sf.center_id IS NULL) AND sf.is_active = 1
		ORDER BY sf.center_id IS NULL DESC, sf.created_at DESC`

	rows, err := database.DB.Query(query, centerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []models.SharedFolderWithCenter
	for rows.Next() {
		var folder models.SharedFolderWithCenter
		err := rows.Scan(
			&folder.ID, &folder.CenterID, &folder.CenterName, &folder.Name,
			&folder.Description, &folder.Type, &folder.CloudURL, &folder.LocalPath,
			&folder.IsActive, &folder.CreatedAt, &folder.UpdatedAt,
		)
		if err != nil {
			continue
		}
		folders = append(folders, folder)
	}

	return folders, nil
}

// createSharedFolder creates a new shared folder
func (h *Handlers) createSharedFolder(centerID *int, name, description, folderType string, cloudURL, localPath *string) error {
	query := `INSERT INTO shared_folders (center_id, name, description, type, cloud_url, local_path, created_at, updated_at) 
	          VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err := database.DB.Exec(query, centerID, name, description, folderType, cloudURL, localPath)
	return err
}

// CarpetasCompartidasEliminar handles shared folder deletion
func (h *Handlers) CarpetasCompartidasEliminar(c *gin.Context) {
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

	folderID := c.Param("id")
	if folderID == "" {
		c.String(http.StatusBadRequest, "ID de carpeta requerido")
		return
	}

	// Get folder info to delete local folder if needed
	folder, err := h.getSharedFolderByID(folderID)
	if err != nil {
		c.String(http.StatusNotFound, "Carpeta no encontrada")
		return
	}

	// Delete from database
	query := `UPDATE shared_folders SET is_active = 0, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err = database.DB.Exec(query, folderID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error al eliminar la carpeta: "+err.Error())
		return
	}

	// If it was a local folder, optionally delete the physical folder
	if folder.Type == "local" && folder.LocalPath != nil {
		fullPath := filepath.Join(h.Config.DataDir, *folder.LocalPath)
		// Note: We don't delete the actual folder to prevent data loss
		// This is a soft delete only
		_ = fullPath // Use the variable to prevent "declared and not used" error
	}

	c.Redirect(http.StatusFound, "/carpetas-compartidas")
}

// getSharedFolderByID gets a shared folder by ID
func (h *Handlers) getSharedFolderByID(folderID string) (models.SharedFolder, error) {
	var folder models.SharedFolder
	query := `SELECT id, center_id, name, description, type, cloud_url, local_path, is_active, created_at, updated_at 
	          FROM shared_folders WHERE id = ?`

	err := database.DB.QueryRow(query, folderID).Scan(
		&folder.ID, &folder.CenterID, &folder.Name, &folder.Description,
		&folder.Type, &folder.CloudURL, &folder.LocalPath, &folder.IsActive,
		&folder.CreatedAt, &folder.UpdatedAt,
	)
	return folder, err
}