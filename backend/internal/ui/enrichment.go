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

	"fusion/internal/platform/config"
	"fusion/internal/platform/utils"
)

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
				
				// Read Status and other fields from data.json
				status := "Unknown"
				var startupTime, shutdownTime, outputMode string
				var dcPower, pOut, ratedPower, pPeak, powerFactor, qOut float64
				var eDaily, eTotal float64
				var gridFreq, gridVa, gridVb, gridVc, gridIa, gridIb, gridIc float64
				var temp, resistance float64
				
				dataBytes, err := os.ReadFile(filepath.Join(path, "data.json"))
				if err == nil {
					// Struct to extract all fields
					var miniData struct {
						Fields struct {
							DeviceStatus string `json:"device_status"`
							StartupTime  string `json:"startup_time"`
							ShutdownTime string `json:"shutdown_time"`
							OutputMode   string `json:"output_mode"`
							
							DcPowerKw    float64 `json:"dc_power_kw"`
							POutKw       float64 `json:"p_out_kw"`
							RatedPowerKw float64 `json:"rated_power_kw"`
							PPeakTodayKw float64 `json:"p_peak_today_kw"`
							PowerFactor  float64 `json:"power_factor"`
							QOutKvar     float64 `json:"q_out_kvar"`
							
							EDailyKwh float64 `json:"edaily_kwh"`
							ETotalKwh float64 `json:"etotal_kwh"`
							
							GridFreqHz float64 `json:"grid_freq_hz"`
							GridVaV    float64 `json:"grid_va_v"`
							GridVbV    float64 `json:"grid_vb_v"`
							GridVcV    float64 `json:"grid_vc_v"`
							GridIaA    float64 `json:"grid_ia_a"`
							GridIbA    float64 `json:"grid_ib_a"`
							GridIcA    float64 `json:"grid_ic_a"`
							
							InternalTempDegC       float64 `json:"internal_temp_degC"`
							InsulationResistanceMO float64 `json:"insulation_resistance_MΩ"`
						} `json:"fields"`
					}
					
					if json.Unmarshal(dataBytes, &miniData) == nil {
						if miniData.Fields.DeviceStatus != "" {
							status = strings.TrimSpace(miniData.Fields.DeviceStatus)
						}
						startupTime = miniData.Fields.StartupTime
						shutdownTime = miniData.Fields.ShutdownTime
						outputMode = miniData.Fields.OutputMode
						
						dcPower = miniData.Fields.DcPowerKw
						pOut = miniData.Fields.POutKw
						ratedPower = miniData.Fields.RatedPowerKw
						pPeak = miniData.Fields.PPeakTodayKw
						powerFactor = miniData.Fields.PowerFactor
						qOut = miniData.Fields.QOutKvar
						
						eDaily = miniData.Fields.EDailyKwh
						eTotal = miniData.Fields.ETotalKwh
						
						gridFreq = miniData.Fields.GridFreqHz
						gridVa = miniData.Fields.GridVaV
						gridVb = miniData.Fields.GridVbV
						gridVc = miniData.Fields.GridVcV
						gridIa = miniData.Fields.GridIaA
						gridIb = miniData.Fields.GridIbA
						gridIc = miniData.Fields.GridIcA
						
						temp = miniData.Fields.InternalTempDegC
						resistance = miniData.Fields.InsulationResistanceMO
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
							StartupTime:  startupTime,
							ShutdownTime: shutdownTime,
							OutputMode:   outputMode,
							
							DcPowerKw:      dcPower,
							POutKw:         pOut,
							RatedPowerKw:   ratedPower,
							PPeakTodayKw:   pPeak,
							PowerFactor:    powerFactor,
							QOutKvar:       qOut,
							EDailyKwh:      eDaily,
							ETotalKwh:      eTotal,
							GridFreqHz:     gridFreq,
							GridVaV:        gridVa,
							GridVbV:        gridVb,
							GridVcV:        gridVc,
							GridIaA:        gridIa,
							GridIbA:        gridIb,
							GridIcA:        gridIc,
							InternalTempDegC:       temp,
							InsulationResistanceMO: resistance,
							
							Strings:      make([]StringData, 0),
						})
						break
					}
				}
			}
		}
		return nil
	}); err != nil {
		utils.LogWarn("Warning scanning sites: %v", err)
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
	endpoint := config.App.System.VMEndpoint
	// Query ALL inverter data: shundao_inverter (without name filter to get all fields)
	// Use last_over_time[1h] to get latest value
	query := `last_over_time(shundao_inverter[1h])`
	url := fmt.Sprintf("%s/api/v1/query?query=%s", endpoint, url.QueryEscape(query))

	resp, err := http.Get(url)
	if err != nil {
		utils.LogError("Error querying VM for inverter data: %v", err)
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
	// Key: DeviceName -> FieldName -> Value
	deviceDataMap := make(map[string]map[string]float64)
	
	for _, r := range result.Data.Result {
		deviceName := r.Metric["device"] // e.g., "HF1_Inverter_1"
		fieldName := r.Metric["name"]    // e.g., "pv01_volt_v", "p_out_kw", etc.

		if deviceName == "" || fieldName == "" {
			continue
		}

		if _, ok := deviceDataMap[deviceName]; !ok {
			deviceDataMap[deviceName] = make(map[string]float64)
		}

		// Get Value
		if len(r.Value) >= 2 {
			valStr, _ := r.Value[1].(string)
			val, _ := strconv.ParseFloat(valStr, 64)
			deviceDataMap[deviceName][fieldName] = val
		}
	}

	// Assign to sites
	for i := range *sites {
		for j := range (*sites)[i].Loggers {
			for k := range (*sites)[i].Loggers[j].Inverters {
				inv := &(*sites)[i].Loggers[j].Inverters[k]
				
				// Get data map for this inverter
				if dataMap, ok := deviceDataMap[inv.ID]; ok {
					// 1. Populate Strings
					stringsData := make([]StringData, 0)
					for idx := 1; idx <= 24; idx++ {
						vKey := fmt.Sprintf("pv%02d_volt_v", idx)
						aKey := fmt.Sprintf("pv%02d_amp_a", idx)
						
						v, hasV := dataMap[vKey]
						a, hasA := dataMap[aKey]
						
						if hasV || hasA {
							stringsData = append(stringsData, StringData{
								ID:      fmt.Sprintf("PV%02d", idx),
								Voltage: v,
								Current: a,
							})
						}
					}
					inv.Strings = stringsData

					// 2. Populate Inverter Metrics
					// 2. Populate Inverter Metrics (Only overwrite if key exists in VM response)
					if v, ok := dataMap["dc_power_kw"]; ok { inv.DcPowerKw = v }
					if v, ok := dataMap["p_out_kw"]; ok { inv.POutKw = v }
					if v, ok := dataMap["rated_power_kw"]; ok { inv.RatedPowerKw = v }
					if v, ok := dataMap["p_peak_today_kw"]; ok { inv.PPeakTodayKw = v }
					if v, ok := dataMap["power_factor"]; ok { inv.PowerFactor = v }
					if v, ok := dataMap["q_out_kvar"]; ok { inv.QOutKvar = v }
					
					if v, ok := dataMap["edaily_kwh"]; ok { inv.EDailyKwh = v }
					if v, ok := dataMap["etotal_kwh"]; ok { inv.ETotalKwh = v }
					
					if v, ok := dataMap["grid_freq_hz"]; ok { inv.GridFreqHz = v }
					if v, ok := dataMap["grid_va_v"]; ok { inv.GridVaV = v }
					if v, ok := dataMap["grid_vb_v"]; ok { inv.GridVbV = v }
					if v, ok := dataMap["grid_vc_v"]; ok { inv.GridVcV = v }
					if v, ok := dataMap["grid_ia_a"]; ok { inv.GridIaA = v }
					if v, ok := dataMap["grid_ib_a"]; ok { inv.GridIbA = v }
					if v, ok := dataMap["grid_ic_a"]; ok { inv.GridIcA = v }
					
					if v, ok := dataMap["internal_temp_degC"]; ok { inv.InternalTempDegC = v }
					
					// Try both likely keys for resistance due to special char handling
					if v, ok := dataMap["insulation_resistance_MΩ"]; ok { 
						inv.InsulationResistanceMO = v 
					} else if v, ok := dataMap["insulation_resistance_MO"]; ok {
						inv.InsulationResistanceMO = v
					}
					
				} else {
					if inv.Strings == nil {
						inv.Strings = make([]StringData, 0)
					}
				}
			}
		}
	}
}

func enrichSitesWithKPI(sites *[]SiteNode) {
	endpoint := config.App.System.VMEndpoint
	
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
