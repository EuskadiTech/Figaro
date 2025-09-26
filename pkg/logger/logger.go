// Package logger provides JSON logging functionality for Figaro application.
package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity of a log entry
type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelDebug LogLevel = "DEBUG"
)

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Module    string    `json:"module,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	IP        string    `json:"ip,omitempty"`
	Extra     any       `json:"extra,omitempty"`
}

// Logger handles JSON logging to file
type Logger struct {
	file   *os.File
	mutex  sync.Mutex
	logDir string
}

// Global logger instance
var defaultLogger *Logger

// Initialize creates and initializes the global logger
func Initialize(dataDir string) error {
	logDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := filepath.Join(logDir, "figaro.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	defaultLogger = &Logger{
		file:   file,
		logDir: logDir,
	}

	return nil
}

// Close closes the log file
func Close() error {
	if defaultLogger != nil && defaultLogger.file != nil {
		return defaultLogger.file.Close()
	}
	return nil
}

// writeLog writes a log entry to the file
func (l *Logger) writeLog(entry LogEntry) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = l.file.WriteString(string(data) + "\n")
	if err != nil {
		return err
	}

	return l.file.Sync()
}

// Info logs an info level message
func Info(message string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelInfo,
		Message:   fmt.Sprintf(message, args...),
	}
	defaultLogger.writeLog(entry)
}

// InfoWithContext logs an info level message with context
func InfoWithContext(module, userID, ip, message string, extra any) {
	if defaultLogger == nil {
		return
	}
	
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelInfo,
		Message:   message,
		Module:    module,
		UserID:    userID,
		IP:        ip,
		Extra:     extra,
	}
	defaultLogger.writeLog(entry)
}

// Warn logs a warning level message
func Warn(message string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelWarn,
		Message:   fmt.Sprintf(message, args...),
	}
	defaultLogger.writeLog(entry)
}

// WarnWithContext logs a warning level message with context
func WarnWithContext(module, userID, ip, message string, extra any) {
	if defaultLogger == nil {
		return
	}
	
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelWarn,
		Message:   message,
		Module:    module,
		UserID:    userID,
		IP:        ip,
		Extra:     extra,
	}
	defaultLogger.writeLog(entry)
}

// Error logs an error level message
func Error(message string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelError,
		Message:   fmt.Sprintf(message, args...),
	}
	defaultLogger.writeLog(entry)
}

// ErrorWithContext logs an error level message with context
func ErrorWithContext(module, userID, ip, message string, extra any) {
	if defaultLogger == nil {
		return
	}
	
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelError,
		Message:   message,
		Module:    module,
		UserID:    userID,
		IP:        ip,
		Extra:     extra,
	}
	defaultLogger.writeLog(entry)
}

// Debug logs a debug level message
func Debug(message string, args ...interface{}) {
	if defaultLogger == nil {
		return
	}
	
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LogLevelDebug,
		Message:   fmt.Sprintf(message, args...),
	}
	defaultLogger.writeLog(entry)
}

// ReadLogs reads log entries from the log file
func ReadLogs(limit int, level LogLevel, date time.Time) ([]LogEntry, error) {
	if defaultLogger == nil {
		return nil, fmt.Errorf("logger not initialized")
	}

	logFile := filepath.Join(defaultLogger.logDir, "figaro.log")
	file, err := os.Open(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var entries []LogEntry
	decoder := json.NewDecoder(file)
	
	for decoder.More() {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			continue // Skip malformed entries
		}

		// Filter by level if specified
		if level != "" && entry.Level != level {
			continue
		}

		// Filter by date if specified (same day)
		if !date.IsZero() {
			entryDate := entry.Timestamp.Format("2006-01-02")
			filterDate := date.Format("2006-01-02")
			if entryDate != filterDate {
				continue
			}
		}

		entries = append(entries, entry)
	}

	// Return most recent entries first, limited by the limit parameter
	if len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	// Reverse to show newest first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	return entries, nil
}