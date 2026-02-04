package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

// InitLogger initializes the logging system
func InitLogger() {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Cannot create log directory: %v", err)
	}

	dateStr := time.Now().Format("2006-01-02")
	logFile, err := os.OpenFile(filepath.Join(logDir, fmt.Sprintf("app_%s.log", dateStr)), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Cannot open log file: %v", err)
	}

	// MultiWriter to write to both file and console
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	errWriter := io.MultiWriter(os.Stderr, logFile)

	InfoLogger = log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(errWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger.Println("Logger initialized")
}

func LogInfo(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, v...)
	} else {
		fmt.Printf("INFO: "+format+"\n", v...)
	}
}

func LogError(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, v...)
	} else {
		fmt.Printf("ERROR: "+format+"\n", v...)
	}
}
