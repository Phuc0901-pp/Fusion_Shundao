// Package domain chứa tất cả các domain types được chia sẻ giữa
// delivery (HTTP), usecase (logic) và repository (data access).
// Không có dependency vào bất kỳ package nào khác trong codebase.
package domain

// ─── Dashboard & Metrics ──────────────────────────────────────────────────────

// DashboardData là payload đầy đủ được trả về cho Frontend.
type DashboardData struct {
	Alerts         []AlertMessage        `json:"alerts"`
	Sites          []SiteNode            `json:"sites"`
	KPI            KPIData               `json:"kpi"`
	Sensors        []SensorData          `json:"sensors"`
	Meters         []MeterData           `json:"meters"`
	ChartData      []ChartPoint          `json:"chartData"`
	SiteData       SiteDataMap           `json:"siteData"`
	ProductionData []ProductionDataPoint `json:"productionData"`
}

// ProductionDataPoint chứa dữ liệu công suất 5-phút theo từng site.
type ProductionDataPoint struct {
	Date            string   `json:"date"`
	Site1Power      *float64 `json:"site1DailyEnergy"`
	Site1Irradiance *float64 `json:"site1Irradiation"`
	Site2Power      *float64 `json:"site2DailyEnergy"`
	Site2Irradiance *float64 `json:"site2Irradiation"`
}

// MonthlyDataPoint chứa giá trị max mỗi ngày trong tháng.
type MonthlyDataPoint struct {
	Date          string   `json:"date"`
	Site1MaxPower *float64 `json:"site1MaxPower"`
	Site1MaxIrrad *float64 `json:"site1MaxIrrad"`
	Site2MaxPower *float64 `json:"site2MaxPower"`
	Site2MaxIrrad *float64 `json:"site2MaxIrrad"`
}

// ChartPoint là một điểm thời gian trên biểu đồ tổng công suất.
type ChartPoint struct {
	Time             string  `json:"time"`
	Power            float64 `json:"power"`
	PvPower          float64 `json:"pvPower"`
	GridPower        float64 `json:"gridPower"`
	ConsumptionPower float64 `json:"consumptionPower"`
}

// SiteDataMap nhóm chart data theo site.
type SiteDataMap struct {
	All   []ChartPoint `json:"all"`
	SiteA []ChartPoint `json:"siteA"`
	SiteB []ChartPoint `json:"siteB"`
}

// InverterPowerPoint là một điểm dữ liệu DC+AC cho inverter.
type InverterPowerPoint struct {
	Time    string   `json:"time"`
	DcPower *float64 `json:"dcPower,omitempty"`
	AcPower *float64 `json:"acPower,omitempty"`
}

// ─── Alert ────────────────────────────────────────────────────────────────────

// AlertMessage là thông báo cảnh báo từ hệ thống.
type AlertMessage struct {
	ID         string `json:"id"`
	Timestamp  int64  `json:"timestamp"`
	Level      string `json:"level"`    // "info" | "warning" | "critical"
	Message    string `json:"message"`
	Source     string `json:"source"`
	DeviceID   string `json:"deviceId,omitempty"`
	DeviceType string `json:"deviceType,omitempty"`
}

// ─── Site Hierarchy ───────────────────────────────────────────────────────────

// SiteNode đại diện cho một nhà máy điện mặt trời.
type SiteNode struct {
	ID          string       `json:"id"`
	DbID        string       `json:"dbId"`
	Name        string       `json:"name"`
	DefaultName string       `json:"defaultName"`
	Loggers     []LoggerNode `json:"loggers"`
	KPI         KPIData      `json:"kpi"`
}

// LoggerNode đại diện cho một bộ thu thập dữ liệu (Smart Logger).
type LoggerNode struct {
	ID          string         `json:"id"`
	DbID        string         `json:"dbId"`
	Name        string         `json:"name"`
	DefaultName string         `json:"defaultName"`
	Inverters   []InverterNode `json:"inverters"`
	KPI         KPIData        `json:"kpi"`
}

// InverterNode đại diện cho một inverter với tất cả thông số kỹ thuật.
type InverterNode struct {
	ID              string       `json:"id"`
	DbID            string       `json:"dbId"`
	Name            string       `json:"name"`
	DefaultName     string       `json:"defaultName"`
	NumberStringSet string       `json:"numberStringSet,omitempty"`
	DeviceStatus    string       `json:"deviceStatus"`
	Strings         []StringData `json:"strings"`
	// Power
	DcPowerKw    float64 `json:"dcPowerKw,omitempty"`
	POutKw       float64 `json:"pOutKw,omitempty"`
	RatedPowerKw float64 `json:"ratedPowerKw,omitempty"`
	PPeakTodayKw float64 `json:"pPeakTodayKw,omitempty"`
	PowerFactor   float64 `json:"powerFactor,omitempty"`
	QOutKvar      float64 `json:"qOutKvar,omitempty"`
	// Energy
	EDailyKwh float64 `json:"eDailyKwh,omitempty"`
	ETotalKwh float64 `json:"eTotalKwh,omitempty"`
	// Grid
	GridFreqHz float64 `json:"gridFreqHz,omitempty"`
	GridVaV    float64 `json:"gridVaV,omitempty"`
	GridVbV    float64 `json:"gridVbV,omitempty"`
	GridVcV    float64 `json:"gridVcV,omitempty"`
	GridIaA    float64 `json:"gridIaA,omitempty"`
	GridIbA    float64 `json:"gridIbA,omitempty"`
	GridIcA    float64 `json:"gridIcA,omitempty"`
	// Other
	InternalTempDegC       float64 `json:"internalTempDegC,omitempty"`
	InsulationResistanceMO float64 `json:"insulationResistanceMO,omitempty"`
	OutputMode             string  `json:"outputMode,omitempty"`
	StartupTime            string  `json:"startupTime,omitempty"`
	ShutdownTime           string  `json:"shutdownTime,omitempty"`
}

// StringData là dữ liệu chuỗi PV.
type StringData struct {
	ID      string  `json:"id"`
	Current float64 `json:"current"`
	Voltage float64 `json:"voltage"`
}

// ─── KPI ──────────────────────────────────────────────────────────────────────

// KPIData chứa các chỉ số hiệu suất chính.
type KPIData struct {
	DailyEnergy       float64 `json:"dailyEnergy"`
	TotalEnergy       float64 `json:"totalEnergy"`
	DailyIncome       float64 `json:"dailyIncome"`
	RatedPower        float64 `json:"ratedPower"`
	GridSupplyToday   float64 `json:"gridSupplyToday"`
	StandardCoalSaved float64 `json:"standardCoalSaved"`
	CO2Reduction      float64 `json:"co2Reduction"`
	TreesPlanted      float64 `json:"treesPlanted"`
}

// ─── Sensors & Meters ─────────────────────────────────────────────────────────

// SensorData chứa dữ liệu cảm biến thời tiết.
type SensorData struct {
	ID            string  `json:"id"`
	SiteID        string  `json:"siteId"`
	Name          string  `json:"name"`
	Irradiance    float64 `json:"irradiance"`
	AmbientString float64 `json:"ambientString"`
	ModuleTemp    float64 `json:"moduleTemp"`
	WindSpeed     float64 `json:"windSpeed"`
}

// MeterData chứa dữ liệu đồng hồ điện.
type MeterData struct {
	ID          string  `json:"id"`
	SiteID      string  `json:"siteId"`
	Name        string  `json:"name"`
	TotalPower  float64 `json:"totalPower"`
	Frequency   float64 `json:"frequency"`
	PowerFactor float64 `json:"powerFactor"`
}
