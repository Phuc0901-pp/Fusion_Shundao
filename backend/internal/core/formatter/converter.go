package formatter

import (
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"fusion/internal/api"
	"fusion/internal/platform/utils"
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
// FormatStationOverview converts station KPI and Social data to formatted struct
func FormatStationOverview(kpi *api.StationKPI, social *api.SocialContribution) *StationFormattedData {
	output := &StationFormattedData{
		Timestamp:   time.Now().Unix(),
		SiteName:    kpi.StationName,
		SiteID:      utils.GenerateUUID(kpi.StationDn),
		Measurement: "plant",
		Fields:      make(OrderedDataMap),
	}

	// KPI Data
	if kpi != nil {
		output.Fields["daily_energy"] = kpi.DailyEnergy
		output.Fields["cumulative_energy"] = kpi.CumulativeEnergy
		output.Fields["daily_income"] = kpi.DailyIncome
		output.Fields["daily_charge_capacity"] = kpi.DailyChargeCapacity
		output.Fields["daily_discharge_capacity"] = kpi.DailyDischargeCapacity
		output.Fields["total_charge_energy"] = kpi.TotalChargeEnergy
		output.Fields["total_discharge_energy"] = kpi.TotalDischargeEnergy
		output.Fields["cumulative_charge_capacity"] = kpi.CumulativeChargeCapacity
		output.Fields["cumulative_discharge_capacity"] = kpi.CumulativeDischargeCapacity
		output.Fields["inverter_power"] = kpi.InverterPower
		output.Fields["battery_capacity"] = kpi.BatteryCapacity
		output.Fields["currency"] = kpi.Currency
		output.Fields["is_price_configured"] = kpi.IsPriceConfigured
		output.Fields["daily_charge_energy"] = kpi.DailyChargeEnergy
		output.Fields["daily_ongrid_energy"] = kpi.DailyOnGridEnergy
		output.Fields["rechargeable_energy"] = kpi.RechargeableEnergy
		output.Fields["redischargeable_energy"] = kpi.ReDischargeableEnergy
	}

	// Social Data
	if social != nil {
		output.Fields["co2_reduction"] = social.CO2Reduction
		output.Fields["co2_reduction_by_year"] = social.CO2ReductionByYear
		output.Fields["equivalent_trees"] = social.EquivalentTreePlanting
		output.Fields["equivalent_trees_by_year"] = social.EquivalentTreePlantingByYear
		output.Fields["standard_coal_savings"] = social.StandardCoalSavings
		output.Fields["standard_coal_savings_by_year"] = social.StandardCoalSavingsByYear
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
			// Structure A: Nested (Inverter)
			for _, item := range dataList {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if sigList, ok := itemMap["signals"].([]interface{}); ok {
						mergeSignals(allSignals, sigList)
					}
				}
			}
		} else {
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
					utils.LogDebug("[DEBUG] Item %d has no valid ID. Keys: %v", i, getKeys(sMap))
				}
			}

			// Add to map if valid ID
			if idStr != "" {
				target[idStr] = sMap
			}
		} else {
			if i < 3 {
				utils.LogDebug("[DEBUG] Item %d is not a map. Type: %T", i, s)
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

	// SmartLogger data like { "data": [ {"name": "IP", "value": "..."}, ... ] }
	rawKV := GetKeyValues(raw)

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

// FormatUnifiedInverterData merges realtime, string, and static data into unified format
func FormatUnifiedInverterData(
	rtRaw map[string]interface{},
	strRaw map[string]interface{},
	staticInfo map[string]string,
	siteInfo map[string]string,
	deviceName, deviceID string,
) *UnifiedInverterData {
	output := &UnifiedInverterData{
		Timestamp:   time.Now().UnixNano() / 1e6, // Milliseconds as per example
		SiteName:    siteInfo["name"],
		SiteID:      utils.GenerateUUID(siteInfo["id"]),
		Name:        deviceName,
		ID:          utils.GenerateUUID(deviceID), // Use deviceID (NE=...) or mapping if available
		Model:       staticInfo["model"],
		SN:          staticInfo["sn"],
		Measurement: "inverter",
		Fields:      make(OrderedDataMap),
	}

	// Helper to set field if exists
	setField := func(field string, val interface{}) {
		output.Fields[field] = val
	}

	// 1. Process String Data (PVs)
	// Calculate DC Power sum
	var dcPower float64

	strSignals := extractSignals(strRaw)
	if strSignals != nil {
		for i := 1; i <= 24; i++ { // Support up to 24 strings
			// Check if Status exists (ID 14000 + i)
			statusID := fmt.Sprintf("%d", 14000+i)
			_, statusOk := getSignalValue(strSignals, statusID)

			// Filter: Only process if Status exists
			if statusOk {
				var volID, curID string
				// Calculate IDs: 11001/11002 for string 1, etc.
				// Pattern: 11001 + (i-1)*3
				volID = fmt.Sprintf("%d", 11001+(i-1)*3)
				curID = fmt.Sprintf("%d", 11002+(i-1)*3)

				volVal, volOk := getSignalValue(strSignals, volID)
				curVal, curOk := getSignalValue(strSignals, curID)

				// Status exists, so we keep this string.
				// If value missing, default to 0.
				if volOk {
					setField(GetUnifiedPVField(i, "volt_v"), volVal)
				} else {
					setField(GetUnifiedPVField(i, "volt_v"), 0)
				}

				if curOk {
					setField(GetUnifiedPVField(i, "amp_a"), curVal)
				} else {
					setField(GetUnifiedPVField(i, "amp_a"), 0)
				}

				// Add to DC power calculation (V * A / 1000 for kW)
				if v, ok := toFloat(volVal); ok {
					if a, ok := toFloat(curVal); ok {
						dcPower += (v * a) / 1000.0
					}
				}
			}
		}
	}

	setField("dc_power_kw", dcPower)

	// 2. Process Realtime Data (AC + Status)
	rtSignals := extractSignals(rtRaw)
	if rtSignals != nil {
		// Use the loaded map from signals.json (via mapper)
		for id, key := range UnifiedInverterSignalMap {
			if val, ok := getSignalValue(rtSignals, id); ok {
				setField(key, val)
			} else {
				// Default 0 if missing (safer for metrics)
				setField(key, 0)
			}
		}
	}

	// 3. Other fields
	// Fill defaults if missing
	defaults := []string{"p_peak_today_kw"}
	for _, k := range defaults {
		if _, ok := output.Fields[k]; !ok {
			setField(k, 0)
		}
	}

	return output
}

// FormatUnifiedSensorData converts raw sensor data to unified format
func FormatUnifiedSensorData(
	raw map[string]interface{},
	staticInfo map[string]string,
	siteInfo map[string]string,
	deviceName, deviceID string,
) *UnifiedSensorData {
	output := &UnifiedSensorData{
		Timestamp:   time.Now().UnixNano() / 1e6,
		SiteName:    siteInfo["name"],
		SiteID:      utils.GenerateUUID(siteInfo["id"]),
		Name:        staticInfo["name"], // Favor static name or fallback
		ID:          utils.GenerateUUID(deviceID),
		Model:       staticInfo["model"],
		SN:          staticInfo["sn"],
		Measurement: "sensor",
		Fields:      make(OrderedDataMap),
	}
	if output.Name == "" {
		output.Name = deviceName
	}

	rawKV := GetKeyValues(raw)

	// Pre-fill all expected fields with 0 or default values
	expectedFields := []string{
		"wind_speed_ms",
		"wind_direction_deg",
		"pv_module_temperature_c",
		"ambient_temperature_c",
		"total_irradiance_wm2",
		"daily_irradiation1_mjm2",
		"total_irradiance2_wm2",
		"daily_irradiation2_mjm2",
		"custom1",
		"custom2",
		"daily_irradiation1_kwhm2",
		"daily_irradiation2_kwhm2",
	}

	// Initialize map with mapped values
	foundFields := make(map[string]bool)

	for key, val := range rawKV {
		if stdKey, ok := SensorFieldMap[key]; ok {
			output.Fields[stdKey] = val
			foundFields[stdKey] = true
		} else {
			// Handle custom mapping or pass through?
			// User example shows generic keys being ignored or mapped specific ways
			// For now, only map known ones.
		}
	}

	// Fill missing fields with defaults matching the user example
	for _, field := range expectedFields {
		if !foundFields[field] {
			// Special handling for some distinct default values in example?
			// User example:
			// "daily_irradiation2_mjm2": -0.001
			// "daily_irradiation2_kwhm2": -0.001
			// Others likely 0 or 3276.7 (invalid value marker for many Modbus devices)

			// We will just use 0 for now as safe default, user can refine if they need specific "invalid" markers
			if strings.Contains(field, "daily_irradiation2") {
				output.Fields[field] = 0.0
			} else {
				// Check if user wants 3276.7 as 'invalid' or defaults?
				// Example has 3276.7 for wind_speed_ms (likely sensor error/disconnected)
				// We won't inject 3276.7 unless we know it's "invalid", we just put 0.
				// Wait, if source data is missing, it's safer to put 0.
				output.Fields[field] = 0.0
			}
		}
	}

	// Add custom fields hardcoded for now as per example if not present
	if _, ok := output.Fields["custom1"]; !ok {
		output.Fields["custom1"] = 3276.7
	}
	if _, ok := output.Fields["custom2"]; !ok {
		output.Fields["custom2"] = 3276.7
	}

	return output
}

// FormatUnifiedPowerMeterData converts raw meter data to unified format
func FormatUnifiedPowerMeterData(
	raw map[string]interface{},
	staticInfo map[string]string,
	siteInfo map[string]string,
	deviceName, deviceID string,
) *UnifiedPowerMeterData {
	output := &UnifiedPowerMeterData{
		Timestamp:   time.Now().UnixNano() / 1e6,
		SiteName:    siteInfo["name"],
		SiteID:      utils.GenerateUUID(siteInfo["id"]),
		Name:        staticInfo["name"],
		ID:          utils.GenerateUUID(deviceID),
		Model:       staticInfo["model"],
		SN:          staticInfo["sn"],
		Measurement: "zonemeter", // As requested
		Fields:      make(OrderedDataMap),
	}
	if output.Name == "" {
		output.Name = deviceName
	}

	// Pre-fill expected fields with 0
	expectedFields := []string{
		"phase_a_voltage_v", "phase_b_voltage_v", "phase_c_voltage_v",
		"line_ab_voltage_v", "line_bc_voltage_v", "line_ca_voltage_v",
		"phase_a_current_a", "phase_b_current_a", "phase_c_current_a",
		"phase_a_active_power_kw", "phase_b_active_power_kw", "phase_c_active_power_kw",
		"active_power_kw", "reactive_power_kvar", "power_factor",
		"total_active_energy_kwh", "total_reactive_energy_kvarh",
		"total_positive_active_energy_kwh", "total_positive_reactive_energy_kvarh",
		"total_negative_active_energy_kwh", "total_negative_reactive_energy_kvarh",
	}

	foundFields := make(map[string]bool)

	// Use extractSignals to handle nested signal structure (same as Inverter/Sensor)
	signals := extractSignals(raw)
	if signals != nil {
		// keys in signals are "IDs" (e.g. "10004")
		for id, stdKey := range PowerMeterSignalMap {
			// Get value from signals map using helper
			if val, ok := getSignalValue(signals, id); ok {
				// Check if we need unit conversion
				finalVal := val

				// Simple heuristic: if key ends in _kw or _kvar but value is W/var, divide by 1000
				if strings.HasSuffix(stdKey, "_kw") || strings.HasSuffix(stdKey, "_kvar") {
					if fVal, ok := toFloat(val); ok {
						finalVal = fVal / 1000.0
					}
				}

				output.Fields[stdKey] = finalVal
				foundFields[stdKey] = true
			}
		}
	}

	// Fill missing fields with 0
	for _, field := range expectedFields {
		if !foundFields[field] {
			output.Fields[field] = 0.0
		}
	}

	return output
}

// Helper to reliably get float
func toFloat(v interface{}) (float64, bool) {
	if f, ok := v.(float64); ok {
		return f, true
	}
	if s, ok := v.(string); ok {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}
