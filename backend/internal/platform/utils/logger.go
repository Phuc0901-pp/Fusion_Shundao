package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *slog.Logger
)

// InitLogger initializes the production-grade logging system.
//   - Console: Human-readable Text format (easy to read during development)
//   - File:    Structured JSON (machine-parseable for post-mortems)
//   - Rotation: 50MB max / 14 days / 7 backups / gzip compressed
func InitLogger() {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Cannot create log directory: %v\n", err)
		os.Exit(1)
	}

	// ─── File Handler (JSON, rotation via Lumberjack) ────────────────────────
	logFile := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "app.log"),
		MaxSize:    50,   // megabytes – production volume
		MaxBackups: 7,    // keep last 7 rotated files
		MaxAge:     14,   // days – retain 2 weeks of history
		Compress:   true, // gzip old segments to save disk space
	}

	fileHandler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Time(slog.TimeKey, GetNow())
			}
			return a
		},
	})

	// ─── Console Handler (Text, colour-friendly) ─────────────────────────────
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, GetNow().Format("15:04:05"))
			}
			return a
		},
	})

	// ─── Fan-out: write to BOTH console and file ──────────────────────────────
	_ = io.MultiWriter(os.Stdout, logFile) // kept for reference; we use typed handlers below
	Logger = slog.New(&multiHandler{file: fileHandler, console: consoleHandler})
	slog.SetDefault(Logger)

	Logger.Info("Logger initialized", "rotation_mb", 50, "retention_days", 14, "compress", true)
}

// multiHandler fans a log record out to two independent slog.Handler instances.
type multiHandler struct {
	file    slog.Handler
	console slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return m.file.Enabled(ctx, level) || m.console.Enabled(ctx, level)
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	_ = m.file.Handle(ctx, r.Clone())
	return m.console.Handle(ctx, r)
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &multiHandler{file: m.file.WithAttrs(attrs), console: m.console.WithAttrs(attrs)}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	return &multiHandler{file: m.file.WithGroup(name), console: m.console.WithGroup(name)}
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
