package utils

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

var (
	Logger *slog.Logger
)

// InitLogger initializes the logging system using slog
func InitLogger() {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Cannot create log directory: %v\n", err)
		os.Exit(1)
	}

	dateStr := time.Now().Format("2006-01-02")
	logFile, err := os.OpenFile(filepath.Join(logDir, fmt.Sprintf("app_%s.log", dateStr)), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Cannot open log file: %v\n", err)
		os.Exit(1)
	}

	// MultiWriter to write to both file and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Create a new JSON handler
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	Logger = slog.New(handler)
	slog.SetDefault(Logger)

	Logger.Info("Logger initialized", "env", "production")
}

// LogInfo logs an info message
func LogInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if Logger != nil {
		Logger.Info(msg)
	} else {
		// Fallback if logger not initialized
		fmt.Println("INFO: " + msg)
	}
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if Logger != nil {
		Logger.Error(msg)
	} else {
		fmt.Println("ERROR: " + msg)
	}
}

// LogDebug logs a debug message
func LogDebug(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if Logger != nil {
		Logger.Debug(msg)
	} else {
		fmt.Println("DEBUG: " + msg)
	}
}

// LogWarn logs a warning message
func LogWarn(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	if Logger != nil {
		Logger.Warn(msg)
	} else {
		fmt.Println("WARN: " + msg)
	}
}
