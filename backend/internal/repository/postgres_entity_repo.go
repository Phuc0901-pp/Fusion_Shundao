// Package repository: PostgreSQL implementation của EntityRepository.
package repository

import (
	"fmt"

	"fusion/internal/database"
)

// PostgresEntityRepo implements EntityRepository bằng GORM + PostgreSQL.
type PostgresEntityRepo struct{}

// NewPostgresEntityRepo khởi tạo repo.
func NewPostgresEntityRepo() *PostgresEntityRepo {
	return &PostgresEntityRepo{}
}

// GetAllEntityConfigs đọc custom names và string sets từ DB.
func (r *PostgresEntityRepo) GetAllEntityConfigs() (map[string]EntityConfig, error) {
	dbConfigs, err := database.GetAllEntityConfigs()
	if err != nil {
		return nil, err
	}

	// Convert database.EntityConfig → repository.EntityConfig
	result := make(map[string]EntityConfig, len(dbConfigs))
	for id, cfg := range dbConfigs {
		result[id] = EntityConfig{
			Name:            cfg.Name,
			StringSet:       cfg.StringSet,
			ExcludedStrings: cfg.ExcludedStrings,
		}
	}
	return result, nil
}

// UpdateNameChange cập nhật tên hiển thị của một entity qua GORM.
func (r *PostgresEntityRepo) UpdateNameChange(entityType, id, newName string) error {
	return database.UpdateNameChange(entityType, id, newName)
}

// UpdateDeviceStringSet cập nhật số lượng chuỗi PV của một device.
func (r *PostgresEntityRepo) UpdateDeviceStringSet(id, stringSet string) error {
	return database.UpdateDeviceStringSet(id, stringSet)
}

// UpdateDeviceExcludedStrings cập nhật danh sách chuỗi PV không sử dụng.
func (r *PostgresEntityRepo) UpdateDeviceExcludedStrings(id, excludedStrings string) error {
	return database.UpdateDeviceExcludedStrings(id, excludedStrings)
}

// Verify at compile-time that PostgresEntityRepo implements EntityRepository.
var _ EntityRepository = (*PostgresEntityRepo)(nil)

// ─── Compile-time check for VM repo ───────────────────────────────────────────
var _ DashboardRepository = (*VMDashboardRepo)(nil)

// ─── NewEntityConfigFromDB helper (for server.go compatibility) ───────────────

// GetEntityConfigRaw returns the raw database EntityConfig map.
// Used by server.go cache that hasn't been migrated yet.
func GetEntityConfigRaw() (map[string]database.EntityConfig, error) {
	return database.GetAllEntityConfigs()
}

// RenameEntity updates an entity name and string set in the DB.
func RenameEntity(entityType, id, newName, stringSet string) error {
	if err := database.UpdateNameChange(entityType, id, newName); err != nil {
		return fmt.Errorf("rename failed: %w", err)
	}
	if stringSet != "" {
		if err := database.UpdateDeviceStringSet(id, stringSet); err != nil {
			return fmt.Errorf("stringset update failed: %w", err)
		}
	}
	return nil
}
