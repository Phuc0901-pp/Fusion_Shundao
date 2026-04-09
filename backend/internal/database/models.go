package database

import (
	"time"

	"gorm.io/gorm"
)

// Site represents a solar power plant site
// Table: site_table
type Site struct {
	ID         string         `gorm:"primaryKey;type:varchar(255);column:id"` // Changed from uuid to varchar
	Name       string         `gorm:"column:name;not null"`
	NameChange *string        `gorm:"column:name_change"` // Nullable
	CreatedAt  time.Time      `gorm:"column:created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index;column:deleted_at"`

	// Relationships
	SmartLoggers []SmartLogger `gorm:"foreignKey:SiteID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// TableName overrides the table name
func (Site) TableName() string {
	return "site_table"
}

// SmartLogger represents a data logger device
// Table: smart_logger_table
type SmartLogger struct {
	ID        string         `gorm:"primaryKey;type:varchar(255);column:id"`      // Unique ID (e.g., DN or UUID)
	SiteID    string         `gorm:"type:varchar(255);column:site_id"` // Foreign Key to Site (varchar)
	Name      string         `gorm:"column:name;not null"`
	NameChange *string       `gorm:"column:name_change"` // Nullable
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at"`

	// Relationships
	Devices []Device `gorm:"foreignKey:SmartLoggerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// TableName overrides the table name
func (SmartLogger) TableName() string {
	return "smart_logger_table"
}

// Device represents an inverter, meter, or sensor
// Table: device_table
type Device struct {
	ID               string         `gorm:"primaryKey;type:varchar(255);column:id"`    // Unique ID (e.g., DN or UUID)
	SmartLoggerID    string         `gorm:"type:varchar(255);column:smart_logger_id"` // Foreign Key to SmartLogger (varchar)
	Name             string         `gorm:"column:name;not null"`
	NameChange       *string        `gorm:"column:name_change"`                        // Nullable
	NumberStringSet  *string        `gorm:"column:number_set_up_string"`               // Nullable - total string count
	ExcludedStrings  *string        `gorm:"column:excluded_strings"`                   // Nullable - comma-separated indices e.g. "4,8"
	DeviceType       string         `gorm:"column:device_type"`                        // e.g., Inverter, Meter
	Model            string         `gorm:"column:model"`
	SN               string         `gorm:"column:sn"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

// TableName overrides the table name
func (Device) TableName() string {
	return "device_table"
}

// Account represents a user account for system authentication
// Table: account_table
type Account struct {
	ID           uint           `gorm:"primaryKey;autoIncrement;column:id"`
	Username     string         `gorm:"column:username;uniqueIndex;not null;type:varchar(100)"`
	PasswordHash string         `gorm:"column:password_hash;not null"`
	Role         string         `gorm:"column:role;not null;default:'viewer'"` // e.g., admin, viewer
	FullName     string         `gorm:"column:full_name;type:varchar(255)"`
	LastLoginAt  *time.Time     `gorm:"column:last_login_at"` // Nullable
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

// TableName overrides the table name
func (Account) TableName() string {
	return "account_table"
}
