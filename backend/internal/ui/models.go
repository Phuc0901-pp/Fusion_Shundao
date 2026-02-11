package ui

// DashboardData structure matching frontend needs
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

// ProductionDataPoint for daily production bar chart (per site)
type ProductionDataPoint struct {
	Date string `json:"date"` // Format: "HH:mm"
	// SHUNDAO 1
	Site1Power      *float64 `json:"site1Power"`      // kW
	Site1Irradiance *float64 `json:"site1Irradiance"` // W/m²
	// SHUNDAO 2
	Site2Power      *float64 `json:"site2Power"`      // kW
	Site2Irradiance *float64 `json:"site2Irradiance"` // W/m²
}

// MonthlyDataPoint for monthly production chart (max per day)
type MonthlyDataPoint struct {
	Date            string   `json:"date"` // Format: "09/02" (DD/MM)
	Site1MaxPower   *float64 `json:"site1MaxPower"`
	Site1MaxIrrad   *float64 `json:"site1MaxIrrad"`
	Site2MaxPower   *float64 `json:"site2MaxPower"`
	Site2MaxIrrad   *float64 `json:"site2MaxIrrad"`
}

type ChartPoint struct {
	Time             string  `json:"time"`
	Power            float64 `json:"power"`
	PvPower          float64 `json:"pvPower"`
	GridPower        float64 `json:"gridPower"`
	ConsumptionPower float64 `json:"consumptionPower"`
}

type SiteDataMap struct {
	All   []ChartPoint `json:"all"`
	SiteA []ChartPoint `json:"siteA"` // Legacy
	SiteB []ChartPoint `json:"siteB"` // Legacy
}

type AlertMessage struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Source    string `json:"source"`
}

type SiteNode struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Loggers []LoggerNode `json:"loggers"`
	KPI     KPIData      `json:"kpi"`
}

type LoggerNode struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Inverters []InverterNode `json:"inverters"`
	KPI       KPIData        `json:"kpi"`
}

type InverterNode struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	DeviceStatus string       `json:"deviceStatus"`
	Strings      []StringData `json:"strings"`
	// Power metrics
	DcPowerKw      float64 `json:"dcPowerKw,omitempty"`
	POutKw         float64 `json:"pOutKw,omitempty"`
	RatedPowerKw   float64 `json:"ratedPowerKw,omitempty"`
	PPeakTodayKw   float64 `json:"pPeakTodayKw,omitempty"`
	PowerFactor    float64 `json:"powerFactor,omitempty"`
	QOutKvar       float64 `json:"qOutKvar,omitempty"`
	// Energy metrics
	EDailyKwh float64 `json:"eDailyKwh,omitempty"`
	ETotalKwh float64 `json:"eTotalKwh,omitempty"`
	// Grid metrics
	GridFreqHz float64 `json:"gridFreqHz,omitempty"`
	GridVaV    float64 `json:"gridVaV,omitempty"`
	GridVbV    float64 `json:"gridVbV,omitempty"`
	GridVcV    float64 `json:"gridVcV,omitempty"`
	GridIaA    float64 `json:"gridIaA,omitempty"`
	GridIbA    float64 `json:"gridIbA,omitempty"`
	GridIcA    float64 `json:"gridIcA,omitempty"`
	// Temperature & other
	InternalTempDegC       float64 `json:"internalTempDegC,omitempty"`
	InsulationResistanceMO float64 `json:"insulationResistanceMO,omitempty"`
	OutputMode             string  `json:"outputMode,omitempty"`
	StartupTime            string  `json:"startupTime,omitempty"`
	ShutdownTime           string  `json:"shutdownTime,omitempty"`
}

type StringData struct {
	ID      string  `json:"id"`
	Current float64 `json:"current"`
	Voltage float64 `json:"voltage"`
}

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

type SensorData struct {
	ID            string  `json:"id"`
	SiteID        string  `json:"siteId"`
	Name          string  `json:"name"`
	Irradiance    float64 `json:"irradiance"`
	AmbientString float64 `json:"ambientString"` // ambient temp
	ModuleTemp    float64 `json:"moduleTemp"`
	WindSpeed     float64 `json:"windSpeed"`
}

type MeterData struct {
	ID          string  `json:"id"`
	SiteID      string  `json:"siteId"`
	Name        string  `json:"name"`
	TotalPower  float64 `json:"totalPower"`
	Frequency   float64 `json:"frequency"`
	PowerFactor float64 `json:"powerFactor"`
}
