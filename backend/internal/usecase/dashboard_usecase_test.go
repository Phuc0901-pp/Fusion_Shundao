package usecase_test

import (
	"testing"
	"time"

	"fusion/internal/domain"
	"fusion/internal/usecase"
)

// ─── Mock Repository ───────────────────────────────────────────────────────────

// mockDashboardRepo là mock implementation của DashboardRepository cho unit test.
// Không cần VictoriaMetrics thật – đây là điểm mạnh của Clean Architecture.
type mockDashboardRepo struct{}

func (m *mockDashboardRepo) FetchKPI() domain.KPIData {
	return domain.KPIData{DailyEnergy: 100.0, RatedPower: 200.0}
}

func (m *mockDashboardRepo) FetchProductionData() []domain.ProductionDataPoint {
	p1, i1 := 50.0, 600.0
	return []domain.ProductionDataPoint{
		{Date: "08:00", Site1Power: &p1, Site1Irradiance: &i1},
	}
}

func (m *mockDashboardRepo) FetchChartData() []domain.ChartPoint {
	return []domain.ChartPoint{{Time: "08:00", Power: 50.0}}
}

func (m *mockDashboardRepo) FetchMonthlyProduction(args ...interface{}) []domain.MonthlyDataPoint {
	v := 100.0
	return []domain.MonthlyDataPoint{{Date: "01/03", Site1MaxPower: &v}}
}

func (m *mockDashboardRepo) FetchInverterPowerData(deviceID string) []domain.InverterPowerPoint {
	dc, ac := 48.0, 45.0
	return []domain.InverterPowerPoint{{Time: "08:00", DcPower: &dc, AcPower: &ac}}
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestBuildDashboard_WithValidSites_ReturnsData(t *testing.T) {
	repo := &mockDashboardRepo{}
	uc := usecase.NewDashboardUsecase(repo)

	sites := []domain.SiteNode{
		{ID: "site1", Name: "Shundao 1"},
	}
	kpi := repo.FetchKPI()

	result := uc.BuildDashboard(sites, kpi)

	if len(result.Sites) != 1 {
		t.Errorf("Expected 1 site, got %d", len(result.Sites))
	}
	if result.KPI.DailyEnergy != 100.0 {
		t.Errorf("Expected DailyEnergy 100.0, got %f", result.KPI.DailyEnergy)
	}
	if len(result.ProductionData) == 0 {
		t.Error("Expected production data, got empty")
	}
}

func TestBuildDashboard_KPIFallback_AggregatesFromSites(t *testing.T) {
	// Khi VM trả về 0, KPI nên được tổng hợp từ Sites
	repo := &mockDashboardRepo{}
	uc := usecase.NewDashboardUsecase(repo)

	sites := []domain.SiteNode{
		{
			ID:   "site1",
			Name: "Shundao 1",
			KPI:  domain.KPIData{DailyEnergy: 50.0, RatedPower: 100.0},
		},
		{
			ID:   "site2",
			Name: "Shundao 2",
			KPI:  domain.KPIData{DailyEnergy: 75.0, RatedPower: 150.0},
		},
	}

	// KPI rỗng → trigger fallback
	zeroKPI := domain.KPIData{}
	result := uc.BuildDashboard(sites, zeroKPI)

	if result.KPI.DailyEnergy != 125.0 {
		t.Errorf("Expected fallback DailyEnergy 125.0, got %f", result.KPI.DailyEnergy)
	}
	if result.KPI.RatedPower != 250.0 {
		t.Errorf("Expected fallback RatedPower 250.0, got %f", result.KPI.RatedPower)
	}
}

func TestGetMonthlyProduction_ReturnsCorrectData(t *testing.T) {
	repo := &mockDashboardRepo{}
	uc := usecase.NewDashboardUsecase(repo)

	result := uc.GetMonthlyProduction(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))

	if len(result) == 0 {
		t.Error("Expected monthly production data, got empty")
	}
	if result[0].Date != "01/03" {
		t.Errorf("Expected date '01/03', got '%s'", result[0].Date)
	}
}

func TestGetInverterPower_ReturnsDCAndAC(t *testing.T) {
	repo := &mockDashboardRepo{}
	uc := usecase.NewDashboardUsecase(repo)

	result := uc.GetInverterPower("NE=123456")

	if len(result) == 0 {
		t.Error("Expected inverter power data, got empty")
	}
	if result[0].DcPower == nil || *result[0].DcPower != 48.0 {
		t.Error("Expected DC power 48.0")
	}
	if result[0].AcPower == nil || *result[0].AcPower != 45.0 {
		t.Error("Expected AC power 45.0")
	}
}
