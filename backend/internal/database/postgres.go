package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"fusion/internal/platform/config"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the database connection
func InitDB() error {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		cfg := config.App.Database
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
			cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("[ERROR] Failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	// Connection Pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("[SUCCESS] Connected to PostgreSQL successfully!")

	// Auto Migrate
	log.Println("[WAITING] Running AutoMigrate...")
	err = DB.AutoMigrate(&Site{}, &SmartLogger{}, &Device{}, &Account{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("[SUCCESS] Database migration completed!")

	// Seed default admin account if none exists
	SeedAdminAccount()

	return nil
}

// SeedAdminAccount creates a default admin account if the account table is empty.
func SeedAdminAccount() {
	var count int64
	DB.Model(&Account{}).Count(&count)
	if count > 0 {
		log.Println("[AUTH] Admin account already exists, skipping seed.")
		return
	}

	defaultPassword := "shundao2026"
	hashed, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Failed to hash admin password: %v", err)
		return
	}

	admin := Account{
		Username:     "admin",
		PasswordHash: string(hashed),
		Role:         "admin",
		FullName:     "Quản trị viên hệ thống",
	}
	if result := DB.Create(&admin); result.Error != nil {
		log.Printf("[ERROR] Failed to seed admin account: %v", result.Error)
		return
	}
	log.Println("[AUTH] [SUCCESS] Default admin account created: username=admin, password=shundao2026")
}
