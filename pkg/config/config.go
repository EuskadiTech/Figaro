// Package config provides configuration management for Figaro application.
package config

import (
	"os"
	"path/filepath"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	Port     string
	Host     string
	DataDir  string
	
	// Database configuration
	DatabasePath string
	
	// File upload configuration
	MaxUploadSize int64
	UploadDir     string
}

// Load loads the configuration from environment variables and defaults
func Load() *Config {
	config := &Config{
		Port:          getEnv("PORT", "8080"),
		Host:          getEnv("HOST", "0.0.0.0"),
		DataDir:       getEnv("DATA_DIR", "./data"),
		MaxUploadSize: getEnvInt64("MAX_UPLOAD_SIZE", 10*1024*1024), // 10MB default
	}
	
	// Ensure data directory exists
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		panic("Failed to create data directory: " + err.Error())
	}
	
	// Set derived paths
	config.DatabasePath = filepath.Join(config.DataDir, "figaro.db")
	config.UploadDir = filepath.Join(config.DataDir, "uploads")
	
	// Ensure upload directory exists
	if err := os.MkdirAll(config.UploadDir, 0755); err != nil {
		panic("Failed to create upload directory: " + err.Error())
	}
	
	return config
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt64 gets an environment variable as int64 or returns a default value
func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}