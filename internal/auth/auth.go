// Package auth provides authentication and authorization functionality for Figaro.
package auth

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/EuskadiTech/Figaro/internal/database"
	"github.com/EuskadiTech/Figaro/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidQRData      = errors.New("invalid QR code data")
)

// LoginCredentials represents login form data
type LoginCredentials struct {
	Username string `form:"username" json:"username"`
	Password string `form:"password" json:"password"`
	QRData   string `form:"qr_data" json:"qr_data"`
}

// GetUser retrieves a user by username from the database
func GetUser(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, password_hash, display_name, email, created_at, updated_at 
			  FROM users WHERE username = ?`
	
	err := database.DB.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, 
		&user.DisplayName, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Load user permissions
	permissions, err := GetUserPermissions(user.ID)
	if err != nil {
		log.Printf("Warning: failed to load permissions for user %s: %v", username, err)
	} else {
		user.Permissions = permissions
	}

	return user, nil
}

// GetUserPermissions retrieves all permissions for a user
func GetUserPermissions(userID int) ([]string, error) {
	query := `SELECT permission FROM user_permissions WHERE user_id = ?`
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// VerifyPassword checks if the provided password matches the user's password hash
func VerifyPassword(user *models.User, password string) error {
	// First try bcrypt (standard format works with $2y$ from PHP)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err == nil {
		return nil
	}

	// Fallback to PHP password_verify format (if needed for migration)
	// This is for compatibility with existing PHP hashes
	if strings.HasPrefix(user.PasswordHash, "$2y$") {
		// Convert $2y$ to $2a$ for Go's bcrypt compatibility
		modifiedHash := strings.Replace(user.PasswordHash, "$2y$", "$2a$", 1)
		if err := bcrypt.CompareHashAndPassword([]byte(modifiedHash), []byte(password)); err == nil {
			return nil
		}
	}

	return ErrInvalidCredentials
}

// Login authenticates a user with username and password
func Login(username, password string) (*models.User, error) {
	user, err := GetUser(username)
	if err != nil {
		return nil, err
	}

	if err := VerifyPassword(user, password); err != nil {
		return nil, err
	}

	return user, nil
}

// LoginWithQR authenticates a user using QR code data
func LoginWithQR(qrData string) (*models.User, error) {
	parts := strings.Split(qrData, ":")
	if len(parts) != 3 {
		return nil, ErrInvalidQRData
	}

	username := parts[0]
	password, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidQRData
	}
	hash := parts[2]

	// Verify the hash
	expectedHash := fmt.Sprintf("%x", sha256.Sum256([]byte(username+":"+string(password))))
	if expectedHash != hash {
		return nil, ErrInvalidQRData
	}

	return Login(username, string(password))
}

// SetUserSession sets user session cookies
func SetUserSession(c *gin.Context, user *models.User, password string) {
	c.SetCookie("username", user.Username, 3600, "/", "", false, false)
	c.SetCookie("password", base64.StdEncoding.EncodeToString([]byte(password)), 3600, "/", "", false, false)
	c.SetCookie("loggedin", "yes", 3600, "/", "", false, false)
}

// ClearUserSession clears user session cookies
func ClearUserSession(c *gin.Context) {
	c.SetCookie("username", "", -1, "/", "", false, false)
	c.SetCookie("password", "", -1, "/", "", false, false)
	c.SetCookie("loggedin", "", -1, "/", "", false, false)
}

// IsLoggedIn checks if a user is logged in by verifying cookies
func IsLoggedIn(c *gin.Context) bool {
	username, err := c.Cookie("username")
	if err != nil {
		return false
	}

	passwordB64, err := c.Cookie("password")
	if err != nil {
		return false
	}

	loggedin, err := c.Cookie("loggedin")
	if err != nil || loggedin != "yes" {
		return false
	}

	password, err := base64.StdEncoding.DecodeString(passwordB64)
	if err != nil {
		return false
	}

	// Verify credentials
	user, err := Login(username, string(password))
	if err != nil {
		return false
	}

	// Store user in context for later use
	c.Set("user", user)
	return true
}

// GetCurrentUser returns the current logged-in user from context
func GetCurrentUser(c *gin.Context) *models.User {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			return u
		}
	}
	return nil
}

// UserHasAccess checks if the current user has access to a specific module
func UserHasAccess(c *gin.Context, module string) bool {
	user := GetCurrentUser(c)
	if user == nil {
		return false
	}

	// Admin users have access to everything
	for _, perm := range user.Permissions {
		if perm == "ADMIN" {
			return true
		}
		if perm == module {
			return true
		}
	}

	return false
}

// RequireAuth is middleware that requires authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsLoggedIn(c) {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePermission is middleware that requires a specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !UserHasAccess(c, permission) {
			c.Redirect(http.StatusFound, fmt.Sprintf("/?flash=No+tienes+permiso+para+acceder+a+esta+página"))
			c.Abort()
			return
		}
		c.Next()
	}
}