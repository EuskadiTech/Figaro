// Package models defines the data structures for Figaro application.
package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// User represents a user in the system
type User struct {
	ID                 int       `json:"id" db:"id"`
	Username           string    `json:"username" db:"username"`
	PasswordHash       string    `json:"-" db:"password_hash"` // Hidden in JSON
	DisplayName        string    `json:"display_name" db:"display_name"`
	Email              string    `json:"email" db:"email"`
	DefaultCenterID    *int      `json:"default_center_id" db:"default_center_id"` // NULL allowed
	ForceDefaultCenter bool      `json:"force_default_center" db:"force_default_center"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
	Permissions        []string  `json:"permissions,omitempty"` // Loaded separately
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
	Category          string    `json:"categoria" db:"category"` // Keep Spanish JSON field for compatibility
	CreatedAt         time.Time `json:"createdAt" db:"created_at"` // Keep existing field name
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// Activity represents an activity or event
type Activity struct {
	ID             int                   `json:"id" db:"id"`
	CenterID       *int                  `json:"center_id" db:"center_id"` // NULL for global activities
	Title          string                `json:"title" db:"title"`
	Description    string                `json:"description" db:"description"`
	StartDatetime  time.Time             `json:"start" db:"start_datetime"`
	EndDatetime    time.Time             `json:"end" db:"end_datetime"`
	IsGlobal       bool                  `json:"_global" db:"is_global"` // Keep underscore prefix for compatibility
	Status         string                `json:"status" db:"status"`
	MeetingURL     *string               `json:"meeting_url" db:"meeting_url"`
	WebURL         *string               `json:"web_url" db:"web_url"`
	CreatedAt      time.Time             `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at" db:"updated_at"`
	// Additional fields for sharing and links (loaded separately)
	Shares         []ActivityShare       `json:"shares,omitempty"`
	CustomLinks    []ActivityCustomLink  `json:"custom_links,omitempty"`
	SharedFromCenter *string             `json:"shared_from_center,omitempty"` // For display purposes
}

// ActivityShare represents an activity shared with a specific center
type ActivityShare struct {
	ID               int    `json:"id" db:"id"`
	ActivityID       int    `json:"activity_id" db:"activity_id"`
	CenterID         int    `json:"center_id" db:"center_id"`
	SharedByCenterID int    `json:"shared_by_center_id" db:"shared_by_center_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// ActivityCustomLink represents a custom link associated with an activity
type ActivityCustomLink struct {
	ID         int    `json:"id" db:"id"`
	ActivityID int    `json:"activity_id" db:"activity_id"`
	Label      string `json:"label" db:"label"`
	URL        string `json:"url" db:"url"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
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

// MaterialWithCenter represents a material with its center name
type MaterialWithCenter struct {
	ID                int       `json:"id"`
	CenterID          int       `json:"center_id"`
	CenterName        string    `json:"center_name"`
	Name              string    `json:"name"`
	PhotoPath         *string   `json:"photo_path"`
	Unit              string    `json:"unit"`
	AvailableQuantity int       `json:"available_quantity"`
	MinimumQuantity   int       `json:"minimum_quantity"`
	Notes             string    `json:"notes"`
	Category          string    `json:"category"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// MaterialStats represents statistics about materials
type MaterialStats struct {
	TotalMaterials     int `json:"total_materials"`
	TotalCenters       int `json:"total_centers"`
	HealthyMaterials   int `json:"healthy_materials"`
	LowStockMaterials  int `json:"low_stock_materials"`
}

// ActivityWithCenter represents an activity with its center name
type ActivityWithCenter struct {
	ID            int       `json:"id"`
	CenterID      *int      `json:"center_id"`
	CenterName    string    `json:"center_name"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	StartDatetime time.Time `json:"start_datetime"`
	EndDatetime   time.Time `json:"end_datetime"`
	IsGlobal      bool      `json:"is_global"`
	Status        string    `json:"status"`
	MeetingURL    *string   `json:"meeting_url"`
	WebURL        *string   `json:"web_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ActivityStats represents statistics about activities
type ActivityStats struct {
	TotalActivities     int `json:"total_activities"`
	PendingActivities   int `json:"pending_activities"`
	InProgressActivities int `json:"in_progress_activities"`
	CompletedActivities int `json:"completed_activities"`
}

// PaginationInfo represents pagination information for lists
type PaginationInfo struct {
	CurrentPage  int `json:"current_page"`
	PerPage      int `json:"per_page"`
	TotalItems   int `json:"total_items"`
	TotalPages   int `json:"total_pages"`
	HasPrevious  bool `json:"has_previous"`
	HasNext      bool `json:"has_next"`
	PreviousPage int `json:"previous_page"`
	NextPage     int `json:"next_page"`
	Offset       int `json:"offset"`
}

// NewPaginationInfo creates a new PaginationInfo with calculated values
func NewPaginationInfo(page, perPage, totalItems int) *PaginationInfo {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25 // Default to 25 items per page
	}

	totalPages := (totalItems + perPage - 1) / perPage // Ceiling division
	if totalPages < 1 {
		totalPages = 1
	}

	// Ensure current page doesn't exceed total pages
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * perPage
	if offset < 0 {
		offset = 0
	}

	return &PaginationInfo{
		CurrentPage:  page,
		PerPage:      perPage,
		TotalItems:   totalItems,
		TotalPages:   totalPages,
		HasPrevious:  page > 1,
		HasNext:      page < totalPages,
		PreviousPage: page - 1,
		NextPage:     page + 1,
		Offset:       offset,
	}
}

// SystemSetting represents a system configuration setting
type SystemSetting struct {
	ID          int       `json:"id" db:"id"`
	Key         string    `json:"key" db:"key"`
	Value       string    `json:"value" db:"value"`
	Category    string    `json:"category" db:"category"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// SharedFolder represents a shared folder with cloud drive links or local files
type SharedFolder struct {
	ID          int       `json:"id" db:"id"`
	CenterID    *int      `json:"center_id" db:"center_id"` // NULL for global folders
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // "local" or "cloud"
	CloudURL    *string   `json:"cloud_url" db:"cloud_url"` // URL for cloud drive folders
	LocalPath   *string   `json:"local_path" db:"local_path"` // Relative path for local folders
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// SharedFolderWithCenter represents a shared folder with its center name
type SharedFolderWithCenter struct {
	ID          int       `json:"id"`
	CenterID    *int      `json:"center_id"`
	CenterName  string    `json:"center_name"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	CloudURL    *string   `json:"cloud_url"`
	LocalPath   *string   `json:"local_path"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}