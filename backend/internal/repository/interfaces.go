// Package repository định nghĩa interfaces cho data access layer.
// Các implementation cụ thể nằm ở internal/database và internal/victoriametrics.
package repository

import "fusion/internal/domain"

// DashboardRepository là interface truy vấn dữ liệu Dashboard từ DB và VictoriaMetrics.
// Tuân thủ Dependency Inversion: usecase phụ thuộc vào interface, không phụ thuộc implementation.
type DashboardRepository interface {
	// FetchKPI truy vấn các chỉ số KPI tổng hợp từ VictoriaMetrics.
	FetchKPI() domain.KPIData

	// FetchProductionData truy xuất dữ liệu công suất 5 phút từ VictoriaMetrics.
	FetchProductionData() []domain.ProductionDataPoint

	// FetchChartData truy xuất dữ liệu biểu đồ thời gian thực.
	FetchChartData() []domain.ChartPoint

	// FetchMonthlyProduction truy xuất dữ liệu sản lượng hàng tháng.
	FetchMonthlyProduction(month ...interface{}) []domain.MonthlyDataPoint

	// FetchInverterPowerData truy xuất dữ liệu DC/AC theo thời gian cho một inverter.
	FetchInverterPowerData(deviceID string) []domain.InverterPowerPoint
}

// EntityRepository là interface quản lý cấu hình entity từ PostgreSQL.
type EntityRepository interface {
	// GetAllEntityConfigs trả về map ID → EntityConfig cho tất cả entities.
	GetAllEntityConfigs() (map[string]EntityConfig, error)

	// UpdateNameChange cập nhật tên hiển thị của một entity.
	UpdateNameChange(entityType, id, newName string) error

	// UpdateDeviceStringSet cập nhật cấu hình chuỗi PV của một device.
	UpdateDeviceStringSet(id, stringSet string) error

	// UpdateDeviceExcludedStrings cập nhật danh sách chuỗi PV không sử dụng.
	UpdateDeviceExcludedStrings(id, excludedStrings string) error
}

// EntityConfig chứa cấu hình hiển thị của một entity.
type EntityConfig struct {
	Name            string
	StringSet       string
	ExcludedStrings string // comma-separated string indices e.g. "4,8"
}
