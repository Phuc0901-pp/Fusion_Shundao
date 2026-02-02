package formatter

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"fusion/src/api"
)

// FormatInverterData converts raw inverter response to FusionFormattedData
func FormatInverterData(raw map[string]interface{}, deviceName, deviceID string) *FusionFormattedData {
	output := &FusionFormattedData{
		Timestamp:  time.Now().Unix(),
		DeviceName: html.UnescapeString(deviceName),
		DeviceID:   deviceID,
		Data:       make(OrderedDataMap),
	}

	// Try to extract signals map
	signals := extractSignals(raw)
	if signals == nil {
		return output
	}

	// Map known signals
	for id, fieldName := range InverterSignalMap {
		if val, ok := getSignalValue(signals, id); ok {
			output.Data[fieldName] = val
		}
	}

	return output
}

// FormatStringData converts raw string/MPPT response to FusionFormattedData
func FormatStringData(raw map[string]interface{}, deviceName, deviceID string) *FusionFormattedData {
	output := &FusionFormattedData{
		Timestamp:  time.Now().Unix(),
		DeviceName: html.UnescapeString(deviceName),
		DeviceID:   deviceID,
		Data:       make(OrderedDataMap),
	}

	signals := extractSignals(raw)
	if signals == nil {
		return output
	}

	// Helper to check if value is non-zero
	isNonZero := func(val interface{}) bool {
		if v, ok := val.(float64); ok {
			return v != 0
		}
		if v, ok := val.(string); ok {
			f, err := strconv.ParseFloat(v, 64)
			return err == nil && f != 0
		}
		return false
	}

	// Logic from main.go: Process strings 1-24
	// Group 1: 11001-11067 (step 3)
	for i := 0; i < 24; i++ {
		strIndex := i + 1
		volID := fmt.Sprintf("%d", 11001+i*3)
		curID := fmt.Sprintf("%d", 11002+i*3)
		statusID := fmt.Sprintf("%d", 14000+strIndex)

		volVal, volOk := getSignalValue(signals, volID)
		curVal, curOk := getSignalValue(signals, curID)
		statusVal, statusOk := getSignalValue(signals, statusID)

		// Include if status exists OR if there are non-zero values
		if statusOk || (volOk && isNonZero(volVal)) || (curOk && isNonZero(curVal)) {
			if volOk {
				output.Data[GetStringPVField(strIndex, "voltage")] = volVal
			}
			if curOk {
				output.Data[GetStringPVField(strIndex, "current")] = curVal
			}
			if statusOk {
				output.Data[GetStringPVField(strIndex, "status")] = statusVal
			}
		}
	}

	// Logic from main.go: Process strings 25-48
	// Group 2: 11070-11118 (step 2)
	for i := 0; i < 24; i++ {
		strIndex := i + 25
		volID := fmt.Sprintf("%d", 11070+i*2)
		curID := fmt.Sprintf("%d", 11071+i*2)
		statusID := fmt.Sprintf("%d", 14000+strIndex)

		volVal, volOk := getSignalValue(signals, volID)
		curVal, curOk := getSignalValue(signals, curID)
		statusVal, statusOk := getSignalValue(signals, statusID)

		if statusOk || (volOk && isNonZero(volVal)) || (curOk && isNonZero(curVal)) {
			if volOk {
				output.Data[GetStringPVField(strIndex, "voltage")] = volVal
			}
			if curOk {
				output.Data[GetStringPVField(strIndex, "current")] = curVal
			}
			if statusOk {
				output.Data[GetStringPVField(strIndex, "status")] = statusVal
			}
		}
	}

	return output
}

// FormatPowerMeterData converts raw meter response
func FormatPowerMeterData(raw map[string]interface{}, deviceName, deviceID string) *FusionFormattedData {
	output := &FusionFormattedData{
		Timestamp:  time.Now().Unix(),
		DeviceName: html.UnescapeString(deviceName),
		DeviceID:   deviceID,
		Data:       make(OrderedDataMap),
	}

	signals := extractSignals(raw)
	if signals == nil {
		return output
	}

	for id, fieldName := range PowerMeterSignalMap {
		if val, ok := getSignalValue(signals, id); ok {
			output.Data[fieldName] = val
		}
	}

	return output
}

// FormatStationOverview combines KPI and Social data
// FormatStationOverview combines KPI and Social data
func FormatStationOverview(kpi *api.StationKPI, social *api.SocialContribution) *StationFormattedData {
	output := &StationFormattedData{
		Timestamp: time.Now().Unix(),
		SiteName:  kpi.StationName,
		SiteID:    kpi.StationDn,
		Data:      make(OrderedDataMap),
	}

	// KPI Data
	if kpi != nil {
		output.Data["daily_energy"] = kpi.DailyEnergy
		output.Data["cumulative_energy"] = kpi.CumulativeEnergy
		output.Data["daily_income"] = kpi.DailyIncome
		output.Data["inverter_power"] = kpi.InverterPower
	}

	// Social Data
	if social != nil {
		output.Data["co2_reduction"] = social.CO2Reduction
		output.Data["equivalent_trees"] = social.EquivalentTreePlanting
		output.Data["standard_coal_savings"] = social.StandardCoalSavings
	}

	return output
}

// Helpers

// extractSignals attempts to find the signals map/list from various response structures
func extractSignals(raw map[string]interface{}) map[string]interface{} {
	// Root "data" can be a Map or List
	rawDat := raw["data"]
	if rawDat == nil {
		return nil
	}

	allSignals := make(map[string]interface{})

	// Case 1: data is a List (FetchInverterRealtimeData structure)
	// { "data": [ { "signals": [...] }, { "signals": [...] } ] }
	// Case 1: data is a List
	if dataList, ok := rawDat.([]interface{}); ok {
		// Approach: Scan list to see if ANY item contains "signals" key.
		// If yes -> It's Nested Structure (Inverter/Meter).
		// If no -> Assume it's Flat Structure (SmartLogger) if items have "id".

		isNested := false
		for _, item := range dataList {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if _, hasSignals := itemMap["signals"]; hasSignals {
					isNested = true
					break
				}
			}
		}

		if isNested {
			fmt.Println("DEBUG: extractSignals detected NESTED structure")
			// Structure A: Nested (Inverter)
			for _, item := range dataList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if sigList, ok := itemMap["signals"].([]interface{}); ok {
						mergeSignals(allSignals, sigList)
					}
				}
			}
		} else {
			fmt.Println("DEBUG: extractSignals detected FLAT structure")
			// Structure B: Flat (SmartLogger)
			// Treat the whole list as signals
			mergeSignals(allSignals, dataList)
		}

		return allSignals
	}

	// Case 2: data is a Map (FetchInverterStringData structure)
	// { "data": { "signals": { "11001": {...}, "11002": {...} } } }
	if dataMap, ok := rawDat.(map[string]interface{}); ok {
		// Sub-case 2a: data.signals is Map
		if sigMap, ok := dataMap["signals"].(map[string]interface{}); ok {
			return sigMap
		}

		// Sub-case 2b: data.signals is List (rare but possible)
		if sigList, ok := dataMap["signals"].([]interface{}); ok {
			mergeSignals(allSignals, sigList)
			return allSignals
		}

		// Sub-case 2c: Nested data.data logic (Legacy support if needed)
		if subData, ok := dataMap["data"].([]interface{}); ok {
			for _, item := range subData {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if sigList, ok := itemMap["signals"].([]interface{}); ok {
						mergeSignals(allSignals, sigList)
					}
				}
			}
			return allSignals
		}
	}

	return nil
}

// mergeSignals helper to add list-based signals into the map
func mergeSignals(target map[string]interface{}, sourceList []interface{}) {
	fmt.Printf("DEBUG: mergeSignals processing %d items\n", len(sourceList))
	for i, s := range sourceList {
		if sMap, ok := s.(map[string]interface{}); ok {
			// Extract ID
			var idStr string
			if idNum, ok := sMap["id"].(float64); ok {
				idStr = fmt.Sprintf("%.0f", idNum) // 10025.0 -> "10025"
			} else if idStrVal, ok := sMap["id"].(string); ok {
				idStr = idStrVal
			} else {
				// Debug missing ID
				if i < 3 { // Log first few failures only
					fmt.Printf("DEBUG: Item %d has no valid ID. Keys: %v\n", i, getKeys(sMap))
				}
			}

			// Add to map if valid ID
			if idStr != "" {
				target[idStr] = sMap
			}
		} else {
			if i < 3 {
				fmt.Printf("DEBUG: Item %d is not a map. Type: %T\n", i, s)
			}
		}
	}
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// getSignalValue extracts the value from a signal item
func getSignalValue(signals map[string]interface{}, id string) (interface{}, bool) {
	item, ok := signals[id]
	if !ok {
		return nil, false
	}

	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return nil, false
	}

	val, ok := itemMap["value"]
	if !ok {
		return nil, false
	}

	// Convert numerical strings to float if possible, else return string
	if strVal, ok := val.(string); ok {
		// Ignore "Unidentified" or empty
		if strVal == "Unidentified" || strVal == "" || strVal == "--" {
			return nil, false
		}
		if fVal, err := strconv.ParseFloat(strVal, 64); err == nil {
			return fVal, true
		}
		return strVal, true
	}

	return val, true
}

// FormatSmartLoggerData converts raw data (list of parameters) to formatted map
func FormatSmartLoggerData(raw map[string]interface{}, deviceName, deviceID string, children []api.ChildDevice) *FusionFormattedData {
	output := &FusionFormattedData{
		Timestamp:  time.Now().Unix(),
		DeviceName: html.UnescapeString(deviceName),
		DeviceID:   deviceID,
		Data:       make(OrderedDataMap),
	}

	// Debug
	fmt.Printf("DEBUG: FormatSmartLoggerData Raw keys: %v\n", len(raw))

	// SmartLogger data like { "data": [ {"name": "IP", "value": "..."}, ... ] }
	rawKV := GetKeyValues(raw)
	fmt.Printf("DEBUG: GetKeyValues returned %d items\n", len(rawKV))

	for key, val := range rawKV {
		if stdKey, ok := SmartLoggerFieldMap[key]; ok {
			output.Data[stdKey] = val
		} else {
			// Fallback: lowercase and check again or just lowercase
			if stdKey, ok := SensorFieldMap[key]; ok {
				output.Data[stdKey] = val
			} else {
				// Automatic snake_case conversion for unknown keys
				lowerKey := strings.ToLower(key)
				// Replace spaces and special chars
				lowerKey = strings.ReplaceAll(lowerKey, " ", "_")
				lowerKey = strings.ReplaceAll(lowerKey, "(°)", "")
				lowerKey = strings.ReplaceAll(lowerKey, "(%)", "")
				lowerKey = strings.TrimSpace(lowerKey)
				output.Data[lowerKey] = val
			}
		}
	}

	// Process child devices
	if len(children) > 0 {
		var childList []map[string]interface{}
		for _, child := range children {
			// Extract details based on user provided keys
			sn := getStringVal(child.ParamValues, "50012")
			version := getStringVal(child.ParamValues, "50010")
			model := getStringVal(child.ParamValues, "50009") // Inverter logic often uses 2000X but user said 50009

			// If 50009 is empty, try others just in case or keep as is?
			// User explicitly said "5009: model", JSON shows "50009": "SUN2000..." or "EMI"

			childData := map[string]interface{}{
				"name":          html.UnescapeString(child.Name),
				"status":        child.Status,
				"type":          child.MocTypeName,
				"model":         model,
				"version":       version,
				"serial_number": sn,
			}
			childList = append(childList, childData)
		}
		output.Data["child_devices"] = childList
	}

	return output
}

// FormatSensorData converts raw sensor/EMI data to formatted map with standardized keys
func FormatSensorData(raw map[string]interface{}, deviceName, deviceID string) *FusionFormattedData {
	output := &FusionFormattedData{
		Timestamp:  time.Now().Unix(),
		DeviceName: html.UnescapeString(deviceName),
		DeviceID:   deviceID,
		Data:       make(OrderedDataMap),
	}

	rawKV := GetKeyValues(raw)

	for key, val := range rawKV {
		if stdKey, ok := SensorFieldMap[key]; ok {
			output.Data[stdKey] = val
		} else {
			// Automatic snake_case conversion
			lowerKey := strings.ToLower(key)
			lowerKey = strings.ReplaceAll(lowerKey, " ", "_")
			lowerKey = strings.ReplaceAll(lowerKey, "(°)", "")
			lowerKey = strings.ReplaceAll(lowerKey, "(%)", "")
			lowerKey = strings.TrimSpace(lowerKey)
			output.Data[lowerKey] = val
		}
	}

	return output
}

func getStringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
