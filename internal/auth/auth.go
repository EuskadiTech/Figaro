// Package auth provides authentication and authorization functionality for Figaro.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

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

// GetUser retrieves a user by ID from the database
func GetUserByID(userID int) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, password_hash, display_name, email, created_at, updated_at 
			  FROM users WHERE id = ?`
	
	err := database.DB.QueryRow(query, userID).Scan(
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
		log.Printf("Warning: failed to load permissions for user %d: %v", userID, err)
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

// generateSessionToken generates a secure random session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateUserSession creates a new user session with device information
func CreateUserSession(userID int, deviceName, ipAddress, userAgent string) (*models.UserSession, error) {
	token, err := generateSessionToken()
	if err != nil {
		return nil, err
	}

	sessionID, err := generateSessionToken()
	if err != nil {
		return nil, err
	}

	session := &models.UserSession{
		ID:         sessionID[:16], // Use first 16 chars as ID
		UserID:     userID,
		Token:      token,
		DeviceName: deviceName,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(24 * time.Hour * 30), // 30 days
		IsActive:   true,
	}

	query := `INSERT INTO user_sessions (id, user_id, token, device_name, ip_address, user_agent, created_at, updated_at, expires_at, is_active)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err = database.DB.Exec(query, session.ID, session.UserID, session.Token, session.DeviceName,
		session.IPAddress, session.UserAgent, session.CreatedAt, session.UpdatedAt, session.ExpiresAt, session.IsActive)
	
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSessionByToken retrieves a session by token
func GetSessionByToken(token string) (*models.UserSession, error) {
	session := &models.UserSession{}
	query := `SELECT id, user_id, token, device_name, ip_address, user_agent, created_at, updated_at, expires_at, is_active 
			  FROM user_sessions WHERE token = ? AND is_active = 1 AND expires_at > datetime('now')`
	
	err := database.DB.QueryRow(query, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.DeviceName,
		&session.IPAddress, &session.UserAgent, &session.CreatedAt, &session.UpdatedAt,
		&session.ExpiresAt, &session.IsActive)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("session not found or expired")
		}
		return nil, err
	}

	return session, nil
}

// GetUserSessions retrieves all active sessions for a user
func GetUserSessions(userID int) ([]models.UserSession, error) {
	query := `SELECT id, user_id, token, device_name, ip_address, user_agent, created_at, updated_at, expires_at, is_active 
			  FROM user_sessions WHERE user_id = ? AND is_active = 1 AND expires_at > datetime('now') ORDER BY updated_at DESC`
	
	rows, err := database.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.UserSession
	for rows.Next() {
		var session models.UserSession
		err := rows.Scan(&session.ID, &session.UserID, &session.Token, &session.DeviceName,
			&session.IPAddress, &session.UserAgent, &session.CreatedAt, &session.UpdatedAt,
			&session.ExpiresAt, &session.IsActive)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeactivateSession deactivates a session by ID
func DeactivateSession(sessionID string) error {
	query := `UPDATE user_sessions SET is_active = 0, updated_at = datetime('now') WHERE id = ?`
	_, err := database.DB.Exec(query, sessionID)
	return err
}

// DeactivateAllUserSessions deactivates all sessions for a user except the current one
func DeactivateAllUserSessions(userID int, exceptSessionID string) error {
	query := `UPDATE user_sessions SET is_active = 0, updated_at = datetime('now') WHERE user_id = ? AND id != ?`
	_, err := database.DB.Exec(query, userID, exceptSessionID)
	return err
}

// UpdateSessionActivity updates the last activity time for a session
func UpdateSessionActivity(token string) error {
	query := `UPDATE user_sessions SET updated_at = datetime('now') WHERE token = ? AND is_active = 1`
	_, err := database.DB.Exec(query, token)
	return err
}

// SetUserSession sets user session cookies with token-based authentication
func SetUserSession(c *gin.Context, user *models.User, password, deviceName string) (*models.UserSession, error) {
	// Get client information
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	if deviceName == "" {
		deviceName = "Web Browser"
	}

	// Create session
	session, err := CreateUserSession(user.ID, deviceName, ipAddress, userAgent)
	if err != nil {
		return nil, err
	}

	// Set cookies
	c.SetCookie("session_token", session.Token, int(24*time.Hour*30/time.Second), "/", "", false, true) // HttpOnly for security
	c.SetCookie("username", user.Username, int(24*time.Hour*30/time.Second), "/", "", false, false)
	c.SetCookie("loggedin", "yes", int(24*time.Hour*30/time.Second), "/", "", false, false)
	
	return session, nil
}

// ClearUserSession clears user session cookies and deactivates session
func ClearUserSession(c *gin.Context) {
	// Get current session token and deactivate it
	if sessionToken, err := c.Cookie("session_token"); err == nil {
		if session, err := GetSessionByToken(sessionToken); err == nil {
			DeactivateSession(session.ID)
		}
	}

	// Clear cookies
	c.SetCookie("session_token", "", -1, "/", "", false, true)
	c.SetCookie("username", "", -1, "/", "", false, false)
	c.SetCookie("loggedin", "", -1, "/", "", false, false)
}

// IsLoggedIn checks if a user is logged in by verifying session token
func IsLoggedIn(c *gin.Context) bool {
	sessionToken, err := c.Cookie("session_token")
	if err != nil {
		return false
	}

	loggedin, err := c.Cookie("loggedin")
	if err != nil || loggedin != "yes" {
		return false
	}

	// Verify session token
	session, err := GetSessionByToken(sessionToken)
	if err != nil {
		return false
	}

	// Get user
	user, err := GetUserByID(session.UserID)
	if err != nil {
		return false
	}

	// Update session activity
	UpdateSessionActivity(sessionToken)

	// Store user and session in context
	c.Set("user", user)
	c.Set("session", session)
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