package utils

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
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

	// Use utility Now() for filename
	dateStr := Now().Format("2006-01-02")
	logFile, err := os.OpenFile(filepath.Join(logDir, fmt.Sprintf("app_%s.log", dateStr)), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Cannot open log file: %v\n", err)
		os.Exit(1)
	}

	// MultiWriter to write to both file and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Create a new JSON handler with ReplaceAttr to inject correct time
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Override time field
			if a.Key == slog.TimeKey {
				// We need to shift the time. However, slog passes the time of the record.
				// Since we can't easily change the record's time before it hits the handler,
				// we manipulate it here.
				// NOTE: slog's time is from system clock. We apply our offset.
				// t := a.Value.Time()
				// We re-calculate because t is essentially time.Now() when LogInfo was called.
				// But we want it to match our utils.Now()

				// Best way: use utils.Now(). But the log might be slightly delayed.
				// Better way: Apply the offset we know.
				// Since utils.Now() = time.Now() + offset under the hood.

				// Let's just use the time provided by slog and adjust it into VN timezone + offset if needed.
				// Actually, simpler: Just use utils.Now() to overwrite it.
				// But that might lose the micro-difference if logs are queued.
				// For this app, simply overwriting with utils.Now() is acceptable.
				// Use string format to remove milliseconds: 2006-01-02T15:04:05
				return slog.Attr{
					Key:   slog.TimeKey,
					Value: slog.StringValue(Now().Format("2006-01-02T15:04:05")),
				}
			}
			return a
		},
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
