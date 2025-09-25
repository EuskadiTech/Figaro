// Package utils provides utility functions for Figaro application.
package utils

import (
	"fmt"
	"strings"
	"time"
)

// IsoToSpanish converts ISO date format to Spanish format (DD/MM/YYYY)
func IsoToSpanish(isoDate string) string {
	// Extract date part before T
	datePart := strings.Split(isoDate, "T")[0]
	parts := strings.Split(datePart, "-")
	
	if len(parts) != 3 {
		return isoDate // Return original if format is unexpected
	}
	
	// Reverse array (YYYY-MM-DD to DD/MM/YYYY)
	return fmt.Sprintf("%s/%s/%s", parts[2], parts[1], parts[0])
}

// FormatTimeForDisplay formats time for display in Spanish format
func FormatTimeForDisplay(t time.Time) string {
	return t.Format("02/01/2006 15:04")
}

// FormatDateForDisplay formats date for display in Spanish format
func FormatDateForDisplay(t time.Time) string {
	return t.Format("02/01/2006")
}

// FormatTimeOnly formats time only (HH:MM)
func FormatTimeOnly(t time.Time) string {
	return t.Format("15:04")
}

// ParseSpanishDate parses Spanish date format (DD/MM/YYYY) to time.Time
func ParseSpanishDate(dateStr string) (time.Time, error) {
	return time.Parse("02/01/2006", dateStr)
}

// ParseSpanishDateTime parses Spanish datetime format (DD/MM/YYYY HH:MM) to time.Time
func ParseSpanishDateTime(dateTimeStr string) (time.Time, error) {
	return time.Parse("02/01/2006 15:04", dateTimeStr)
}

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// StringPtr returns a pointer to the string value
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StringValue returns the string value from a pointer, or empty string if nil
func StringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// IntPtr returns a pointer to the int value
func IntPtr(i int) *int {
	return &i
}

// IntValue returns the int value from a pointer, or 0 if nil
func IntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}