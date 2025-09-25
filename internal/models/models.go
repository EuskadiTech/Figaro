// Package models defines the data structures for Figaro application.
package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	ID          int       `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	PasswordHash string   `json:"-" db:"password_hash"` // Hidden in JSON
	DisplayName string    `json:"display_name" db:"display_name"`
	Email       string    `json:"email" db:"email"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Permissions []string  `json:"permissions,omitempty"` // Loaded separately
}

// Center represents an educational center
type Center struct {
	ID           int                 `json:"id" db:"id"`
	Name         string              `json:"name" db:"name"`
	Timezone     string              `json:"timezone" db:"timezone"`
	CreatedAt    time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at" db:"updated_at"`
	WorkingHours []CenterWorkingHour `json:"working_hours,omitempty"` // Loaded separately
}

// CenterWorkingHour represents working hours for a center on a specific day
type CenterWorkingHour struct {
	ID        int     `json:"id" db:"id"`
	CenterID  int     `json:"center_id" db:"center_id"`
	DayOfWeek int     `json:"day_of_week" db:"day_of_week"` // 0=Sunday, 1=Monday, etc.
	StartTime *string `json:"start_time" db:"start_time"`   // NULL for closed days
	EndTime   *string `json:"end_time" db:"end_time"`       // NULL for closed days
}

// Classroom represents a classroom within a center
type Classroom struct {
	ID        int       `json:"id" db:"id"`
	CenterID  int       `json:"center_id" db:"center_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Material represents a material in the inventory
type Material struct {
	ID                int       `json:"id" db:"id"`
	CenterID          int       `json:"center_id" db:"center_id"`
	Name              string    `json:"nombre" db:"name"` // Keep Spanish JSON field for compatibility
	PhotoPath         *string   `json:"foto" db:"photo_path"` // Keep Spanish JSON field for compatibility
	Unit              string    `json:"unidad" db:"unit"` // Keep Spanish JSON field for compatibility
	AvailableQuantity int       `json:"cantidad_disponible" db:"available_quantity"` // Keep Spanish JSON field for compatibility
	MinimumQuantity   int       `json:"cantidad_minima" db:"minimum_quantity"` // Keep Spanish JSON field for compatibility
	Notes             string    `json:"notas" db:"notes"` // Keep Spanish JSON field for compatibility
	CreatedAt         time.Time `json:"createdAt" db:"created_at"` // Keep existing field name
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Activity represents an activity or event
type Activity struct {
	ID             int       `json:"id" db:"id"`
	CenterID       *int      `json:"center_id" db:"center_id"` // NULL for global activities
	Title          string    `json:"title" db:"title"`
	Description    string    `json:"description" db:"description"`
	StartDatetime  time.Time `json:"start" db:"start_datetime"`
	EndDatetime    time.Time `json:"end" db:"end_datetime"`
	IsGlobal       bool      `json:"_global" db:"is_global"` // Keep underscore prefix for compatibility
	MeetingURL     *string   `json:"meeting_url" db:"meeting_url"`
	WebURL         *string   `json:"web_url" db:"web_url"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// UserPermission represents a user's permission
type UserPermission struct {
	ID         int    `json:"id" db:"id"`
	UserID     int    `json:"user_id" db:"user_id"`
	Permission string `json:"permission" db:"permission"`
}

// NullTime handles NULL time values in database
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Scan implements the sql.Scanner interface
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Valid = false
		return nil
	}
	nt.Valid = true
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
		return nil
	case string:
		t, err := time.Parse("2006-01-02 15:04:05", v)
		if err != nil {
			return err
		}
		nt.Time = t
		return nil
	}
	return fmt.Errorf("cannot scan %T into NullTime", value)
}

// UserSession represents a user session with device information
type UserSession struct {
	ID          string    `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	Token       string    `json:"token" db:"token"`
	DeviceName  string    `json:"device_name" db:"device_name"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

// Value implements the driver.Valuer interface
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}