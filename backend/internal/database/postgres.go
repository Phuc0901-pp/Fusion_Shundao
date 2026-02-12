package database

import (
	"fmt"
	"log"
	"time"

	"fusion/internal/platform/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the database connection
func InitDB() error {
	cfg := config.App.Database

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)

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
	// log.Println("⏳ Running AutoMigrate...")
	// err = DB.AutoMigrate(&Site{}, &SmartLogger{}, &Device{})
	// if err != nil {
	// 	return fmt.Errorf("failed to migrate database: %w", err)
	// }
	// log.Println("✅ Database migration completed!")

	return nil
}
