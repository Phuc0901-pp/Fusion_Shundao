package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

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
	Date string `json:"date"` // Format: "01", "02", etc. (day of month)
	// SHUNDAO 1
	Site1DailyEnergy float64 `json:"site1DailyEnergy"` // kWh
	Site1GridFeedIn  float64 `json:"site1GridFeedIn"`  // kWh
	Site1Irradiation float64 `json:"site1Irradiation"` // MJ/m²
	// SHUNDAO 2
	Site2DailyEnergy float64 `json:"site2DailyEnergy"` // kWh
	Site2GridFeedIn  float64 `json:"site2GridFeedIn"`  // kWh
	Site2Irradiation float64 `json:"site2Irradiation"` // MJ/m²
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

// Cache Mechanism
var (
	apiCache      DashboardData
	cacheMutex    sync.RWMutex
	baseSites     []SiteNode
	baseSitesLock sync.RWMutex
)

// StartServer starts the API server
func StartServer() {
	// Start background data aggregator
	go startBackgroundUpdater()

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
	cacheMutex.RLock()
	data := apiCache
	cacheMutex.RUnlock()

	// Disable Caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	// Create a new slice to avoid modifying the cached data
	// Sort Loggers and Inverters ON THE FLY to ensure it's always sorted
	// (Deep copy or just sort the top level if needed, but here we invoke sorting in the updateCache usually)
	// But to be safe, let's verify sorting here or in updateCache.
	// We already sorted in fetchSitesFromOutput.

	bytes, _ := json.Marshal(data)
	w.Write(bytes)
}

func startBackgroundUpdater() {
	// Initial population
	fmt.Println("⏳ Initializing site structure (Disk Scan)...")
	scanSitesStructure()
	fmt.Println("⏳ Initializing metrics (VM Fetch)...")
	updateMetrics()
	fmt.Println("✅ Background updater started.")

	// Metric Loop (5s)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			updateMetrics()
		}
	}()

	// Structure Loop (60s) for detecting new devices/files
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for range ticker.C {
			scanSitesStructure()
		}
	}()
}

func scanSitesStructure() {
	// Scan output dir manually to build tree STRUCTURE only
	// This is I/O intensive, so run less frequently
	sites := fetchSitesFromOutput("output")
	
	baseSitesLock.Lock()
	baseSites = sites
	baseSitesLock.Unlock()
}

func updateMetrics() {
	// 1. Fetch KPI from VM (Global)
	kpi := fetchKPIFromVM()

	// 2. Clone Sites from Base (Deep Copy to avoid race conditions)
	baseSitesLock.RLock()
	sites := deepCopySites(baseSites)
	baseSitesLock.RUnlock()

	// 3. Enrich Sites with Real-time Data from VM (Power, Strings, Status)
	enrichSitesWithVMData(&sites)
	
	// 4. Enrich Sites with KPI Data (Breakdown)
	enrichSitesWithKPI(&sites)

	// 5. Fetch Chart Data from VM
	chartData := fetchChartDataFromVM()
	siteData := SiteDataMap{
		All:   chartData,
		SiteA: []ChartPoint{},
		SiteB: []ChartPoint{},
	}

	// 6. Fallback: Aggregate KPI from Sites if VM query returned 0
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

	// 7. Fetch Production Data for bar chart
	productionData := fetchProductionDataFromVM()

	newData := DashboardData{
		Alerts: []AlertMessage{
			{ID: "1", Timestamp: time.Now().UnixMilli(), Level: "info", Message: "System Online (VM Backend)", Source: "Backend"},
		},
		Sites:          sites,
		KPI:            kpi,
		Sensors:        []SensorData{},
		Meters:         []MeterData{},
		ChartData:      chartData,
		SiteData:       siteData,
		ProductionData: productionData,
	}

	cacheMutex.Lock()
	apiCache = newData
	cacheMutex.Unlock()
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
		GridSupplyToday:   querySum(endpoint, `last_over_time(shundao_plant{name="daily_ongrid_energy"}[1h])`),
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
				// Normalize Name
				name := strings.ReplaceAll(slName, "_", " ")
				name = strings.ReplaceAll(name, "SmartloggerStation", "Smartlogger Station") // Fix inconsistent naming
				name = strings.ReplaceAll(name, "Sattion", "Station") // Fix typo in folder name
				
				sitesMap[siteName].Loggers = append(sitesMap[siteName].Loggers, LoggerNode{
					ID:        slName,
					Name:      name,
					Inverters: make([]InverterNode, 0),
				})
			}

			// Detect Device (Inverter)
			if len(parts) == 3 && strings.Contains(parts[2], "Inverter") {
				slName := parts[1]
				invName := parts[2]
				
				// Read Status from data.json (if exists)
				status := "Unknown"
				dataBytes, err := os.ReadFile(filepath.Join(path, "data.json"))
				if err == nil {
					// Minimal struct to extract status
					var miniData struct {
						Fields struct {
							DeviceStatus string `json:"device_status"`
						} `json:"fields"`
					}
					if json.Unmarshal(dataBytes, &miniData) == nil {
						if miniData.Fields.DeviceStatus != "" {
							status = strings.TrimSpace(miniData.Fields.DeviceStatus)
						}
					}
				}

				// Find correct logger
				for i := range sitesMap[siteName].Loggers {
					if sitesMap[siteName].Loggers[i].ID == slName {
						// Clean Inverter Name (Remove HFxx)
						cleanName := strings.ReplaceAll(invName, "_", " ")
						// Regex to remove "HF" followed by digits
						re := regexp.MustCompile(`HF\d+\s*`)
						cleanName = re.ReplaceAllString(cleanName, "")
						cleanName = strings.TrimSpace(cleanName)

						sitesMap[siteName].Loggers[i].Inverters = append(sitesMap[siteName].Loggers[i].Inverters, InverterNode{
							ID:           invName,
							Name:         cleanName,
							DeviceStatus: status,
							Strings:      make([]StringData, 0),
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
		// Sort Loggers
		sort.Slice(s.Loggers, func(i, j int) bool {
			return naturalLess(s.Loggers[i].Name, s.Loggers[j].Name)
		})
		// Sort Inverters in each Logger
		for k := range s.Loggers {
			sort.Slice(s.Loggers[k].Inverters, func(i, j int) bool {
				return naturalLess(s.Loggers[k].Inverters[i].Name, s.Loggers[k].Inverters[j].Name)
			})
		}
		sites = append(sites, *s)
	}

	// Sort Sites
	sort.Slice(sites, func(i, j int) bool {
		return naturalLess(sites[i].Name, sites[j].Name)
	})

	return sites
}

// naturalLess compares two strings with embedded numbers naturally
func naturalLess(s1, s2 string) bool {
    // Split into parts (text vs numbers)
    parts1 := splitName(s1)
    parts2 := splitName(s2)
    
    n := len(parts1)
    if len(parts2) < n {
        n = len(parts2)
    }
    
    for i := 0; i < n; i++ {
        // Check if both are numbers
        num1, err1 := strconv.Atoi(parts1[i])
        num2, err2 := strconv.Atoi(parts2[i])
        
        if err1 == nil && err2 == nil {
            if num1 != num2 {
                return num1 < num2
            }
        } else {
			// Case insensitive string comparison
			str1 := strings.ToLower(parts1[i])
			str2 := strings.ToLower(parts2[i])
            if str1 != str2 {
                return str1 < str2
            }
        }
    }
    
    return len(parts1) < len(parts2)
}

func splitName(s string) []string {
    var parts []string
    var current string
    var isDigit bool
    
    for _, r := range s {
        if r >= '0' && r <= '9' {
            if !isDigit && current != "" {
                parts = append(parts, current)
                current = ""
            }
            isDigit = true
            current += string(r)
        } else {
            if isDigit && current != "" {
                parts = append(parts, current)
                current = ""
            }
            isDigit = false
            current += string(r)
        }
    }
    if current != "" {
        parts = append(parts, current)
    }
    return parts
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

func enrichSitesWithKPI(sites *[]SiteNode) {
	endpoint := "http://100.118.142.45:8428"
	
	// Fetch all data points for relevant metrics, grouped by 'device' (SmartLogger name)
	metrics := map[string]string{
		"dailyEnergy":       `last_over_time(shundao_plant{name="daily_energy"}[1h])`,
		"totalEnergy":       `last_over_time(shundao_plant{name="cumulative_energy"}[1h])`,
		"dailyIncome":       `last_over_time(shundao_plant{name="daily_income"}[1h])`,
		"co2Reduction":      `last_over_time(shundao_plant{name="co2_reduction"}[1h])`,
		"treesPlanted":      `last_over_time(shundao_plant{name="equivalent_trees"}[1h])`,
		"standardCoalSaved": `last_over_time(shundao_plant{name="standard_coal_savings"}[1h])`,
		"gridSupplyToday":   `last_over_time(shundao_plant{name="daily_ongrid_energy"}[1h])`,
		"ratedPower":        `last_over_time(shundao_inverter{name="rated_power_kw"}[1h])`,
	}

	// Store results: DeviceID -> Metric -> Value
	dataMap := make(map[string]map[string]float64)

	for header, query := range metrics {
		vals := queryByLabel(endpoint, query) // Returns map[deviceID]value
		for devID, val := range vals {
			if _, ok := dataMap[devID]; !ok {
				dataMap[devID] = make(map[string]float64)
			}
			dataMap[devID][header] += val
		}
	}

	// Distribute to Sites
	for i := range *sites {
		site := &(*sites)[i]
		siteKPI := KPIData{}

		// 1. Assign Site-Level KPI from Plant Metrics (using Site ID/Name)
		var siteMetrics map[string]float64
		var ok bool
		
		if siteMetrics, ok = dataMap[site.ID]; !ok {
			siteMetrics, ok = dataMap[site.Name]
		}
		
		if ok {
			siteKPI.DailyEnergy = siteMetrics["dailyEnergy"]
			siteKPI.TotalEnergy = siteMetrics["totalEnergy"]
			siteKPI.DailyIncome = siteMetrics["dailyIncome"]
			siteKPI.CO2Reduction = siteMetrics["co2Reduction"]
			siteKPI.TreesPlanted = siteMetrics["treesPlanted"]
			siteKPI.StandardCoalSaved = siteMetrics["standardCoalSaved"]
			siteKPI.GridSupplyToday = siteMetrics["gridSupplyToday"]
		}

		// 2. Aggregate Rated Power from Inverters
		var siteRatedPower float64
		
		for j := range site.Loggers {
			logger := &site.Loggers[j]
			loggerKPI := KPIData{}
			
			var loggerRatedPower float64
			for _, inv := range logger.Inverters {
				if m, ok := dataMap[inv.ID]; ok {
					loggerRatedPower += m["ratedPower"]
				}
			}
			loggerKPI.RatedPower = loggerRatedPower
			logger.KPI = loggerKPI
			
			siteRatedPower += loggerRatedPower
		}
		
		siteKPI.RatedPower = siteRatedPower
		site.KPI = siteKPI
	}
}

func queryByLabel(endpoint, query string) map[string]float64 {
	results := make(map[string]float64)
	url := fmt.Sprintf("%s/api/v1/query?query=%s", endpoint, url.QueryEscape(query))
	
	resp, err := http.Get(url)
	if err != nil {
		return results
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return results
	}

	for _, r := range result.Data.Result {
		// Identify ID. Priority: 'device' -> 'station' -> 'site_name' -> 'id'
		id := r.Metric["device"]
		if id == "" {
			id = r.Metric["station"]
		}
		if id == "" {
			id = r.Metric["site_name"]
		}
		if id == "" {
			id = r.Metric["id"]
		}
		
		if id != "" && len(r.Value) > 1 {
			valStr, _ := r.Value[1].(string)
			var val float64
			fmt.Sscanf(valStr, "%f", &val)
			results[id] = val
		}
	}
	return results
}

// fetchProductionDataFromVM fetches daily production data for bar chart (per site)
func fetchProductionDataFromVM() []ProductionDataPoint {
	endpoint := "http://100.118.142.45:8428"
	now := time.Now()
	
	// Start of current month
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	end := now

	pointMap := make(map[string]*ProductionDataPoint)
	
	// Helper to ensure point exists
	getPoint := func(ts time.Time) *ProductionDataPoint {
		dateStr := ts.Format("02")
		if _, ok := pointMap[dateStr]; !ok {
			pointMap[dateStr] = &ProductionDataPoint{Date: dateStr}
		}
		return pointMap[dateStr]
	}

	// 1. Site 1 Energy
	fetchDailyRange(endpoint, `shundao_plant{name="daily_energy", site_name="SHUNDAO_1"}`, start, end, func(ts time.Time, val float64) {
		getPoint(ts).Site1DailyEnergy = val
	})
	// 2. Site 1 Grid
	fetchDailyRange(endpoint, `shundao_plant{name="daily_ongrid_energy", site_name="SHUNDAO_1"}`, start, end, func(ts time.Time, val float64) {
		getPoint(ts).Site1GridFeedIn = val
	})
	// 3. Site 1 Irradiation
	fetchDailyRange(endpoint, `avg(shundao_sensor{name="daily_irradiation1_mjm2", site_name="SHUNDAO_1"})`, start, end, func(ts time.Time, val float64) {
		getPoint(ts).Site1Irradiation = val
	})

	// 4. Site 2 Energy
	fetchDailyRange(endpoint, `shundao_plant{name="daily_energy", site_name="SHUNDAO_2"}`, start, end, func(ts time.Time, val float64) {
		getPoint(ts).Site2DailyEnergy = val
	})
	// 5. Site 2 Grid
	fetchDailyRange(endpoint, `shundao_plant{name="daily_ongrid_energy", site_name="SHUNDAO_2"}`, start, end, func(ts time.Time, val float64) {
		getPoint(ts).Site2GridFeedIn = val
	})
	// 6. Site 2 Irradiation
	fetchDailyRange(endpoint, `avg(shundao_sensor{name="daily_irradiation1_mjm2", site_name="SHUNDAO_2"})`, start, end, func(ts time.Time, val float64) {
		getPoint(ts).Site2Irradiation = val
	})

	// Convert map to slice
	result := make([]ProductionDataPoint, 0, len(pointMap))
	for _, v := range pointMap {
		result = append(result, *v)
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].Date < result[j].Date
	})
	
	return result
}

// deepCopySites creates a deep copy of the sites slice to avoid race conditions
func deepCopySites(src []SiteNode) []SiteNode {
	dst := make([]SiteNode, len(src))
	for i, s := range src {
		dst[i] = s
		// Deep copy Loggers
		dst[i].Loggers = make([]LoggerNode, len(s.Loggers))
		for j, l := range s.Loggers {
			dst[i].Loggers[j] = l
			// Deep copy Inverters
			dst[i].Loggers[j].Inverters = make([]InverterNode, len(s.Loggers[j].Inverters))
			copy(dst[i].Loggers[j].Inverters, l.Inverters)
			// Strings slice is replaced by enrichSitesWithVMData, no need to deep copy
			// KPI is value type, copied automatically
		}
	}
	return dst
}

// fetchDailyRange fetches time-series data and calls callback for each point
func fetchDailyRange(endpoint, query string, start, end time.Time, callback func(ts time.Time, val float64)) {
	// Step = 1 day (86400 seconds)
	u := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=86400",
		endpoint, url.QueryEscape(query), start.Unix(), end.Unix())
	
	resp, err := http.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	var result struct {
		Data struct {
			Result []struct {
				Values [][]interface{} `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	
	if json.Unmarshal(body, &result) != nil {
		return
	}
	
	if len(result.Data.Result) == 0 {
		return
	}
	
	for _, v := range result.Data.Result[0].Values {
		if len(v) >= 2 {
			ts := time.Unix(int64(v[0].(float64)), 0)
			valStr, _ := v[1].(string)
			var val float64
			fmt.Sscanf(valStr, "%f", &val)
			callback(ts, val)
		}
	}
}
