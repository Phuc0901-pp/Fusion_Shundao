package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// UpsertSite creates or updates a site. Returns true if created, false if skipped.
func UpsertSite(id, name string) (bool, error) {
	var count int64
	if err := DB.Model(&Site{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check site existence: %w", err)
	}
	if count > 0 {
		return false, nil // Exists, skip
	}

	site := Site{
		ID:   id,
		Name: name,
	}

	// Create new
	if err := DB.Create(&site).Error; err != nil {
		return false, fmt.Errorf("failed to create site %s: %w", name, err)
	}
	return true, nil
}

// UpsertSmartLogger creates or updates a smart logger. Returns true if created, false if skipped.
func UpsertSmartLogger(id, siteID, name string) (bool, error) {
	var count int64
	if err := DB.Model(&SmartLogger{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check smart logger existence: %w", err)
	}
	if count > 0 {
		return false, nil // Exists, skip
	}

	sl := SmartLogger{
		ID:     id,
		SiteID: siteID,
		Name:   name,
	}

	if err := DB.Create(&sl).Error; err != nil {
		return false, fmt.Errorf("failed to create smart logger %s: %w", name, err)
	}
	return true, nil
}

// UpsertDevice creates or updates a device. Returns true if created, false if skipped.
func UpsertDevice(id, smartLoggerID, name, deviceType, model, sn string) (bool, error) {
	// Simple validation to prevent empty IDs from crashing
	if id == "" {
		log.Printf("[WARNING] Skipping UpsertDevice with empty ID for %s", name)
		return false, nil
	}

	var count int64
	if err := DB.Model(&Device{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("[ERROR] Failed to check device existence: %w", err)
	}
	if count > 0 {
		return false, nil // Exists, skip
	}

	dev := Device{
		ID:            id,
		SmartLoggerID: smartLoggerID,
		Name:          name,
		DeviceType:    deviceType,
		Model:         model,
		SN:            sn,
	}

	if err := DB.Create(&dev).Error; err != nil {
		return false, fmt.Errorf("[ERROR] Failed to create device %s (ID: %s): %w", name, id, err)
	}
	return true, nil
}

// EntityConfig holds custom configuration for an entity
type EntityConfig struct {
	Name      string
	StringSet string
}

// GetAllEntityConfigs returns a map of ID -> EntityConfig for all entities
func GetAllEntityConfigs() (map[string]EntityConfig, error) {
	configs := make(map[string]EntityConfig)

	// 1. Sites
	var sites []Site
	if err := DB.Select("id, name_change").Where("name_change IS NOT NULL AND name_change != ''").Find(&sites).Error; err != nil {
		return nil, err
	}
	for _, s := range sites {
		configs[s.ID] = EntityConfig{Name: *s.NameChange}
	}

	// 2. Smart Loggers
	var loggers []SmartLogger
	if err := DB.Select("id, name_change").Where("name_change IS NOT NULL AND name_change != ''").Find(&loggers).Error; err != nil {
		return nil, err
	}
	for _, l := range loggers {
		configs[l.ID] = EntityConfig{Name: *l.NameChange}
	}

	// 3. Devices
	var devices []Device
	if err := DB.Select("id, name_change, number_set_up_string").Where("(name_change IS NOT NULL AND name_change != '') OR (number_set_up_string IS NOT NULL AND number_set_up_string != '')").Find(&devices).Error; err != nil {
		return nil, err
	}
	for _, d := range devices {
		cfg := configs[d.ID] // Get existing or empty
		if d.NameChange != nil {
			cfg.Name = *d.NameChange
		}
		if d.NumberStringSet != nil {
			cfg.StringSet = *d.NumberStringSet
		}
		configs[d.ID] = cfg
	}

	return configs, nil
}

// UpdateNameChange updates the name_change column for a given entity
func UpdateNameChange(entityType, id, newName string) error {
	var result *gorm.DB
	switch entityType {
	case "site":
		result = DB.Model(&Site{}).Where("id = ?", id).Update("name_change", newName)
	case "logger":
		result = DB.Model(&SmartLogger{}).Where("id = ?", id).Update("name_change", newName)
	case "device":
		result = DB.Model(&Device{}).Where("id = ?", id).Update("name_change", newName)
	default:
		return fmt.Errorf("unknown entity type: %s", entityType)
	}

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no record found or no update needed for %s ID %s", entityType, id)
	}
	return nil
}

// UpdateDeviceStringSet updates the number_string_set_up column for a device
// It correctly handles empty string input by setting the column to NULL or empty string
func UpdateDeviceStringSet(id, stringSet string) error {
	result := DB.Model(&Device{}).Where("id = ?", id).Update("number_set_up_string", stringSet)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no device found with ID %s", id)
	}
	return nil
}
