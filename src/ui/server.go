package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DashboardData structure matching frontend needs
type DashboardData struct {
	Alerts    []AlertMessage `json:"alerts"`
	Sites     []SiteNode     `json:"sites"`
	KPI       KPIData        `json:"kpi"`
	Sensors   []SensorData   `json:"sensors"`
	Meters    []MeterData    `json:"meters"`
	ChartData []ChartPoint   `json:"chartData"`
	SiteData  SiteDataMap    `json:"siteData"`
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
}

type LoggerNode struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Inverters []InverterNode `json:"inverters"`
}

type InverterNode struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Strings []StringData `json:"strings"`
}

type StringData struct {
	ID      string  `json:"id"`
	Current float64 `json:"current"`
	Voltage float64 `json:"voltage"`
}

type KPIData struct {
	DailyEnergy       float64 `json:"dailyEnergy"`
	DailyIncome       float64 `json:"dailyIncome"`
	TotalEnergy       float64 `json:"totalEnergy"`
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

// StartServer starts the API server
func StartServer() {
	http.HandleFunc("/api/dashboard", handleDashboard)

	// Enable CORS for development
	// corsHandler := corsMiddleware(http.DefaultServeMux)

	port := ":5039"
	fmt.Printf("🚀 Starting UI Backend Server on %s...\n", port)
	if err := http.ListenAndServe(port, corsMiddleware(http.DefaultServeMux)); err != nil {
		fmt.Printf("❌ Server failed: %v\n", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	// 1. Fetch KPI from VM (using correct field names)
	kpi := fetchKPIFromVM()

	// 2. Fetch Sites Tree from Output Directory (Structure Only)
	sites := fetchSitesFromOutput("output")

	// 3. Enrich Sites with Real-time Data from VM (Power, Strings, Status)
	enrichSitesWithVMData(&sites)

	// 4. Fetch Chart Data from VM
	chartData := fetchChartDataFromVM()
	siteData := SiteDataMap{
		All:   chartData,
		SiteA: []ChartPoint{}, // TODO: breakdown by site
		SiteB: []ChartPoint{},
	}

	data := DashboardData{
		Alerts: []AlertMessage{
			{ID: "1", Timestamp: time.Now().UnixMilli(), Level: "info", Message: "System Online (VM Backend)", Source: "Backend"},
		},
		Sites:     sites,
		KPI:       kpi,
		Sensors:   []SensorData{},
		Meters:    []MeterData{},
		ChartData: chartData,
		SiteData:  siteData,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func fetchKPIFromVM() KPIData {
	endpoint := "http://100.118.142.45:8428"

	// Metric names match JSON fields: daily_energy, daily_income, etc.
	// Metric names match JSON fields: daily_energy, daily_income, etc.
	// Use last_over_time[1h] to handle stale data (up to 1 hour old)
	return KPIData{
		DailyEnergy:       querySum(endpoint, `last_over_time(shundao_plant{name="daily_energy"}[1h])`),
		TotalEnergy:       querySum(endpoint, `last_over_time(shundao_plant{name="cumulative_energy"}[1h])`),
		DailyIncome:       querySum(endpoint, `last_over_time(shundao_plant{name="daily_income"}[1h])`),
		RatedPower:        querySum(endpoint, `last_over_time(shundao_inverter{name="rated_power_kw"}[1h])`),
		CO2Reduction:      querySum(endpoint, `last_over_time(shundao_plant{name="co2_reduction"}[1h])`),
		TreesPlanted:      querySum(endpoint, `last_over_time(shundao_plant{name="equivalent_trees"}[1h])`),
		StandardCoalSaved: querySum(endpoint, `last_over_time(shundao_plant{name="standard_coal_savings"}[1h])`),
	}
}

func fetchChartDataFromVM() []ChartPoint {
	endpoint := "http://100.118.142.45:8428"
	// Query AC Power (p_out_kw) over last 24h
	end := time.Now()
	start := end.Add(-24 * time.Hour)

	// step = 300s (5 min)
	query := `sum(shundao_inverter{name="p_out_kw"})`
	u := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=300",
		endpoint, url.QueryEscape(query), start.Unix(), end.Unix())

	resp, err := http.Get(u)
	if err != nil {
		return []ChartPoint{}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Struct for query_range response
	var result struct {
		Data struct {
			Result []struct {
				Values [][]interface{} `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return []ChartPoint{}
	}

	points := []ChartPoint{}
	if len(result.Data.Result) > 0 {
		for _, v := range result.Data.Result[0].Values {
			if len(v) >= 2 {
				// v[0] is timestamp (float), v[1] is value (string)
				tsFloat, _ := v[0].(float64)
				valStr, _ := v[1].(string)

				var val float64
				fmt.Sscanf(valStr, "%f", &val)

				// Frontend expects time format "HH:MM"
				// Convert to Vietnam Time (UTC+7)
				loc := time.FixedZone("UTC+7", 7*60*60)
				ts := time.Unix(int64(tsFloat), 0).In(loc)

				points = append(points, ChartPoint{
					Time:    ts.Format("15:04"),
					Power:   val,
					PvPower: val, // Assuming all is PV for now
				})
			}
		}
	}
	return points
}

func fetchSitesFromOutput(rootDir string) []SiteNode {
	// Scan output dir manually to build tree STRUCTURE only
	sitesMap := make(map[string]*SiteNode)

	if err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}

		rel, _ := filepath.Rel(rootDir, path)
		parts := strings.Split(filepath.ToSlash(rel), "/")

		if len(parts) >= 1 && parts[0] != "." {
			siteName := parts[0]
			if _, exists := sitesMap[siteName]; !exists {
				sitesMap[siteName] = &SiteNode{
					ID:      siteName,
					Name:    strings.ReplaceAll(siteName, "_", " "),
					Loggers: make([]LoggerNode, 0),
				}
			}

			// Detect SmartLogger (Station)
			if len(parts) == 2 && (strings.Contains(parts[1], "Smartlogger") || strings.Contains(parts[1], "Station")) {
				slName := parts[1]
				sitesMap[siteName].Loggers = append(sitesMap[siteName].Loggers, LoggerNode{
					ID:        slName,
					Name:      strings.ReplaceAll(slName, "_", " "),
					Inverters: make([]InverterNode, 0),
				})
			}

			// Detect Device (Inverter)
			if len(parts) == 3 && strings.Contains(parts[2], "Inverter") {
				slName := parts[1]
				invName := parts[2]

				// Find correct logger
				for i := range sitesMap[siteName].Loggers {
					if sitesMap[siteName].Loggers[i].ID == slName {
						sitesMap[siteName].Loggers[i].Inverters = append(sitesMap[siteName].Loggers[i].Inverters, InverterNode{
							ID:      invName,
							Name:    strings.ReplaceAll(invName, "_", " "),
							Strings: make([]StringData, 0),
						})
						break
					}
				}
			}
		}
		return nil
	}); err != nil {
		fmt.Printf("Warning scanning sites: %v\n", err)
	}

	// Convert map to slice
	sites := make([]SiteNode, 0)
	for _, s := range sitesMap {
		sites = append(sites, *s)
	}
	return sites
}

func enrichSitesWithVMData(sites *[]SiteNode) {
	endpoint := "http://100.118.142.45:8428"
	// Query all PV strings data: shundao_inverter{name=~"pv.*"}
	// Use last_over_time[1h] to get latest value in last hour if recent scrape failed
	query := `last_over_time(shundao_inverter{name=~"pv.*"}[1h])`
	url := fmt.Sprintf("%s/api/v1/query?query=%s", endpoint, url.QueryEscape(query))

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error querying VM for strings:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"` // [timestamp, "value"]
			} `json:"result"`
		} `json:"data"`
	}
	json.Unmarshal(body, &result)

	// Map data to structure
	// Key: DeviceName -> Map[PV_ID] -> {v, a}
	type PVVal struct {
		V float64
		A float64
	}
	deviceMap := make(map[string]map[int]*PVVal)

	for _, r := range result.Data.Result {
		deviceName := r.Metric["device"] // e.g., "HF1_Inverter_1"
		fieldName := r.Metric["name"]    // e.g., "pv01_volt_v"

		if deviceName == "" || fieldName == "" {
			continue
		}

		// Parse pv index
		var pvIdx int
		var unit string
		// Format: pv01_volt_v or pv01_amp_a
		if n, err := fmt.Sscanf(fieldName, "pv%d_%s", &pvIdx, &unit); err == nil && n == 2 {
			if _, ok := deviceMap[deviceName]; !ok {
				deviceMap[deviceName] = make(map[int]*PVVal)
			}
			if _, ok := deviceMap[deviceName][pvIdx]; !ok {
				deviceMap[deviceName][pvIdx] = &PVVal{}
			}

			// Get Value
			if len(r.Value) >= 2 {
				valStr, _ := r.Value[1].(string)
				val, _ := strconv.ParseFloat(valStr, 64)

				if strings.Contains(unit, "volt") {
					deviceMap[deviceName][pvIdx].V = val
				} else if strings.Contains(unit, "amp") {
					deviceMap[deviceName][pvIdx].A = val
				}
			}
		}
	}

	// Assign to sites
	for i := range *sites {
		for j := range (*sites)[i].Loggers {
			for k := range (*sites)[i].Loggers[j].Inverters {
				inv := &(*sites)[i].Loggers[j].Inverters[k]
				// Match device name (ID matches folder name which matches label)
				if pvMap, ok := deviceMap[inv.ID]; ok {
					stringsData := make([]StringData, 0)
					for idx := 1; idx <= 24; idx++ {
						if val, exists := pvMap[idx]; exists {
							// Always show string if data exists in VM, even if 0 (night time)
							stringsData = append(stringsData, StringData{
								ID:      fmt.Sprintf("PV%02d", idx),
								Voltage: val.V,
								Current: val.A,
							})
						}
					}
					inv.Strings = stringsData
				} else {
					// Ensure not nil even if no data found in VM
					if inv.Strings == nil {
						inv.Strings = make([]StringData, 0)
					}
				}
			}
		}
	}
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func querySum(endpoint, query string) float64 {
	url := fmt.Sprintf("%s/api/v1/query?query=sum(%s)", endpoint, url.QueryEscape(query))
	resp, err := http.Get(url)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data struct {
			Result []struct {
				Value []interface{} `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0
	}

	if len(result.Data.Result) > 0 && len(result.Data.Result[0].Value) > 1 {
		valStr, ok := result.Data.Result[0].Value[1].(string)
		if ok {
			var val float64
			fmt.Sscanf(valStr, "%f", &val)
			return val
		}
	}
	return 0
}
