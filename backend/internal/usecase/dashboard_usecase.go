// Package usecase chứa business logic của Dashboard.
// Layer này không biết tới HTTP, GORM hay VictoriaMetrics;
// chỉ làm việc thông qua repository interfaces.
package usecase

import (
	"time"

	"fusion/internal/domain"
	"fusion/internal/repository"
)

// DashboardUsecase thực thi business logic cho Dashboard.
type DashboardUsecase struct {
	repo repository.DashboardRepository
}

// NewDashboardUsecase khởi tạo usecase với repository được inject vào (Dependency Injection).
func NewDashboardUsecase(repo repository.DashboardRepository) *DashboardUsecase {
	return &DashboardUsecase{repo: repo}
}

// BuildDashboard tập hợp toàn bộ dữ liệu cần thiết cho Dashboard.
// Đây là hàm business logic chính – không biết gì về HTTP hay DB implementation.
func (u *DashboardUsecase) BuildDashboard(sites []domain.SiteNode, kpi domain.KPIData) domain.DashboardData {
	chartData := u.repo.FetchChartData()
	productionData := u.repo.FetchProductionData()

	siteData := domain.SiteDataMap{
		All:   chartData,
		SiteA: []domain.ChartPoint{},
		SiteB: []domain.ChartPoint{},
	}

	// Fallback: tổng hợp KPI từ Sites nếu VM trả về 0
	if kpi.DailyEnergy == 0 || kpi.RatedPower == 0 {
		for _, site := range sites {
			kpi.DailyEnergy += site.KPI.DailyEnergy
			kpi.DailyIncome += site.KPI.DailyIncome
			kpi.TotalEnergy += site.KPI.TotalEnergy
			kpi.RatedPower += site.KPI.RatedPower
			kpi.GridSupplyToday += site.KPI.GridSupplyToday
			kpi.StandardCoalSaved += site.KPI.StandardCoalSaved
			kpi.CO2Reduction += site.KPI.CO2Reduction
			kpi.TreesPlanted += site.KPI.TreesPlanted
		}
	}

	baseAlerts := []domain.AlertMessage{
		{
			ID:        "1",
			Timestamp: time.Now().UnixMilli(),
			Level:     "info",
			Message:   "System Online (VM Backend)",
			Source:    "Backend",
		},
	}

	return domain.DashboardData{
		Alerts:         baseAlerts,
		Sites:          sites,
		KPI:            kpi,
		Sensors:        []domain.SensorData{},
		Meters:         []domain.MeterData{},
		ChartData:      chartData,
		SiteData:       siteData,
		ProductionData: productionData,
	}
}

// GetMonthlyProduction lấy dữ liệu sản lượng theo tháng.
func (u *DashboardUsecase) GetMonthlyProduction(month ...interface{}) []domain.MonthlyDataPoint {
	return u.repo.FetchMonthlyProduction(month...)
}

// GetInverterPower lấy dữ liệu DC/AC của một inverter.
func (u *DashboardUsecase) GetInverterPower(deviceID string) []domain.InverterPowerPoint {
	return u.repo.FetchInverterPowerData(deviceID)
}
