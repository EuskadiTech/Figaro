package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/EuskadiTech/Figaro/internal/database"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"
)

// WebDAV file system that filters out dot files
type filteredFileSystem struct {
	fs webdav.FileSystem
}

func (f *filteredFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	return f.fs.Mkdir(ctx, name, perm)
}

func (f *filteredFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	return f.fs.OpenFile(ctx, name, flag, perm)
}

func (f *filteredFileSystem) RemoveAll(ctx context.Context, name string) error {
	return f.fs.RemoveAll(ctx, name)
}

func (f *filteredFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	return f.fs.Rename(ctx, oldName, newName)
}

func (f *filteredFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	// Block access to hidden files
	if isDotFile(name) {
		return nil, os.ErrNotExist
	}
	return f.fs.Stat(ctx, name)
}

// Filtered directory wrapper that hides dot files
type filteredDir struct {
	webdav.File
}

func (f *filteredDir) Readdir(count int) ([]os.FileInfo, error) {
	entries, err := f.File.Readdir(count)
	if err != nil {
		return nil, err
	}
	
	var filtered []os.FileInfo
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

func isDotFile(name string) bool {
	base := path.Base(name)
	return strings.HasPrefix(base, ".")
}

// WebDAV authentication middleware
func (h *Handlers) WebDAVAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for Basic Auth token
		_, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="Figaro WebDAV"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Validate WebDAV token (token is in password field)
		user, err := h.validateWebDAVToken(password)
		if err != nil {
			c.Header("WWW-Authenticate", `Basic realm="Figaro WebDAV"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Update last used timestamp
		h.updateWebDAVTokenUsage(password)

		// Store user in context
		c.Set("webdav_user", user)
		c.Next()
	}
}

// WebDAV handler for shared folders
func (h *Handlers) WebDAVSharedFolders(c *gin.Context) {
	user := c.MustGet("webdav_user").(*models.User)
	folderName := c.Param("folder")

	// Get shared folder details
	folder, err := h.getSharedFolderByName(folderName, user.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var rootDir string
	if folder.Type == "local" && folder.LocalPath != nil {
		rootDir = filepath.Join(h.Config.DataDir, *folder.LocalPath)
	} else {
		// Cloud folders are not accessible via WebDAV
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Ensure directory exists
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Create WebDAV handler
	handler := &webdav.Handler{
		FileSystem: &filteredFileSystem{fs: webdav.Dir(rootDir)},
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				fmt.Printf("WebDAV error for user %d folder %s: %v\n", user.ID, folderName, err)
			}
		},
	}

	// Remove the path prefix to get relative path for WebDAV
	originalURL := c.Request.URL.Path
	c.Request.URL.Path = strings.TrimPrefix(originalURL, "/dav/CarpetasCompartidas/"+folderName)
	if c.Request.URL.Path == originalURL {
		c.Request.URL.Path = "/"
	}

	handler.ServeHTTP(c.Writer, c.Request)
}

// WebDAV handler for personal files
func (h *Handlers) WebDAVPersonalFiles(c *gin.Context) {
	user := c.MustGet("webdav_user").(*models.User)
	
	// Personal files directory
	userDir := filepath.Join(h.Config.DataDir, "ArchivosPersonales", strconv.Itoa(user.ID))
	
	// Ensure directory exists
	if err := os.MkdirAll(userDir, 0755); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Create WebDAV handler
	handler := &webdav.Handler{
		FileSystem: &filteredFileSystem{fs: webdav.Dir(userDir)},
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				fmt.Printf("WebDAV error for user %d personal files: %v\n", user.ID, err)
			}
		},
	}

	// Remove the path prefix to get relative path for WebDAV
	originalURL := c.Request.URL.Path
	c.Request.URL.Path = strings.TrimPrefix(originalURL, "/dav/MisArchivos")
	if c.Request.URL.Path == originalURL {
		c.Request.URL.Path = "/"
	}

	handler.ServeHTTP(c.Writer, c.Request)
}

// Generate WebDAV token
func (h *Handlers) generateWebDAVToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Create WebDAV token for device
func (h *Handlers) createWebDAVToken(userID int, deviceName string) (*models.WebDAVToken, error) {
	token, err := h.generateWebDAVToken()
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO webdav_tokens (user_id, device_name, token, created_at, last_used, expires_at) 
	          VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, datetime('now', '+30 days'))`
	
	result, err := database.DB.Exec(query, userID, deviceName, token)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	
	return &models.WebDAVToken{
		ID:         int(id),
		UserID:     userID,
		DeviceName: deviceName,
		Token:      token,
		CreatedAt:  time.Now(),
		LastUsed:   time.Now(),
		ExpiresAt:  time.Now().AddDate(0, 0, 30),
		IsActive:   true,
	}, nil
}

// Validate WebDAV token and return user
func (h *Handlers) validateWebDAVToken(token string) (*models.User, error) {
	// First check if token exists and is valid
	var userID int
	query := `SELECT user_id FROM webdav_tokens 
	          WHERE token = ? AND is_active = 1 AND expires_at > datetime('now')`
	
	err := database.DB.QueryRow(query, token).Scan(&userID)
	if err != nil {
		return nil, err
	}

	// Get user details
	userQuery := `SELECT id, username, display_name, email, created_at, updated_at 
	              FROM users WHERE id = ?`
	
	var user models.User
	err = database.DB.QueryRow(userQuery, userID).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.Email, 
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Update token last used timestamp
func (h *Handlers) updateWebDAVTokenUsage(token string) error {
	query := `UPDATE webdav_tokens SET last_used = CURRENT_TIMESTAMP WHERE token = ?`
	_, err := database.DB.Exec(query, token)
	return err
}

// Get shared folder by name that user has access to
func (h *Handlers) getSharedFolderByName(name string, userID int) (*models.SharedFolder, error) {
	// Get user's current center
	var centerName string
	userQuery := `SELECT c.name FROM users u 
	              LEFT JOIN centers c ON u.default_center_id = c.id 
	              WHERE u.id = ?`
	database.DB.QueryRow(userQuery, userID).Scan(&centerName)

	// Convert center name to ID if available
	var centerID *int
	if centerName != "" {
		var id int
		centerQuery := `SELECT id FROM centers WHERE name = ?`
		if err := database.DB.QueryRow(centerQuery, centerName).Scan(&id); err == nil {
			centerID = &id
		}
	}

	// Get shared folder accessible to user
	var folder models.SharedFolder
	query := `SELECT id, center_id, name, description, type, cloud_url, local_path, is_active, created_at, updated_at
	          FROM shared_folders 
	          WHERE name = ? AND is_active = 1 AND (center_id IS NULL OR center_id = ?)`
	
	err := database.DB.QueryRow(query, name, centerID).Scan(
		&folder.ID, &folder.CenterID, &folder.Name, &folder.Description,
		&folder.Type, &folder.CloudURL, &folder.LocalPath, &folder.IsActive,
		&folder.CreatedAt, &folder.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &folder, nil
}

// Get user's WebDAV tokens
func (h *Handlers) getUserWebDAVTokens(userID int) ([]models.WebDAVToken, error) {
	query := `SELECT id, user_id, device_name, token, created_at, last_used, expires_at, is_active
	          FROM webdav_tokens 
	          WHERE user_id = ? AND is_active = 1 AND expires_at > datetime('now')
	          ORDER BY last_used DESC`
	
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.WebDAVToken
	for rows.Next() {
		var token models.WebDAVToken
		err := rows.Scan(
			&token.ID, &token.UserID, &token.DeviceName, &token.Token,
			&token.CreatedAt, &token.LastUsed, &token.ExpiresAt, &token.IsActive,
		)
		if err != nil {
			continue
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// Revoke WebDAV token
func (h *Handlers) revokeWebDAVToken(tokenID int, userID int) error {
	query := `UPDATE webdav_tokens SET is_active = 0 WHERE id = ? AND user_id = ?`
	_, err := database.DB.Exec(query, tokenID, userID)
	return err
}