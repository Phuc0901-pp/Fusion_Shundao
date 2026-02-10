package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"fusion/internal/platform/config"
)

// fetchProductionDataFromVM fetches 5-minute interval production data for today (per site)
func fetchProductionDataFromVM() []ProductionDataPoint {
	endpoint := config.App.System.VMEndpoint
	now := time.Now()
	
	// Start of TODAY (00:00)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	end := now

	// 1. Fetch RAW Instantaneous Data (step=5m = 300s)
	// We use a map to store populated raw points
	type RawPoint struct {
		Date            string
		Site1Power      float64
		Site1Irradiance float64
		Site2Power      float64
		Site2Irradiance float64
	}
	
	rawMap := make(map[string]*RawPoint)
	getRaw := func(ts time.Time) *RawPoint {
		key := ts.Format("15:04") // "HH:mm"
		if _, ok := rawMap[key]; !ok {
			rawMap[key] = &RawPoint{Date: key}
		}
		return rawMap[key]
	}

	// 1. Site 1 Power (kW)
	fetchDailyRange(endpoint, `sum(shundao_inverter{name="p_out_kw", site_name="SHUNDAO_1"})`, start, end, 300, func(ts time.Time, val float64) {
		getRaw(ts).Site1Power = val
	})
	// 2. Site 1 Irradiance (W/m²) - using total_irradiance_wm2
	fetchDailyRange(endpoint, `avg(shundao_sensor{name="total_irradiance_wm2", site_name="SHUNDAO_1"})`, start, end, 300, func(ts time.Time, val float64) {
		getRaw(ts).Site1Irradiance = val
	})

	// 3. Site 2 Power (kW)
	fetchDailyRange(endpoint, `sum(shundao_inverter{name="p_out_kw", site_name="SHUNDAO_2"})`, start, end, 300, func(ts time.Time, val float64) {
		getRaw(ts).Site2Power = val
	})
	// 4. Site 2 Irradiance (W/m²) - using total_irradiance_wm2
	fetchDailyRange(endpoint, `avg(shundao_sensor{name="total_irradiance_wm2", site_name="SHUNDAO_2"})`, start, end, 300, func(ts time.Time, val float64) {
		getRaw(ts).Site2Irradiance = val
	})

	// Process & Create Full 24h Result (5-minute intervals)
	// 24 hours * 12 points/hour = 288 points
	totalPoints := 24 * 12
	result := make([]ProductionDataPoint, totalPoints)
	
	// Initialize time slots
	slotTime := start
	for i := 0; i < totalPoints; i++ {
		h := slotTime.Format("15:04")
		result[i] = ProductionDataPoint{Date: h} 
		slotTime = slotTime.Add(5 * time.Minute)
	}

	// Helper to safely get raw values (return nil if missing)
	getRawVal := func(key string) *RawPoint {
		if p, ok := rawMap[key]; ok {
			return p
		}
		return nil
	}
	
	// Loop to fill result
	// We only show data up to current time
	
	minutesSinceMidnight := now.Hour()*60 + now.Minute()
	currentSlotIndex := minutesSinceMidnight / 5

	for i := 0; i < totalPoints; i++ {
		// If slot is in future (relative to now), skip (leave as nil)
		if i > currentSlotIndex {
			continue
		}
		
		key := result[i].Date // "HH:mm"
		curr := getRawVal(key)
		
		// If data exists, assign it. If not, it remains nil (gap in chart, which is correct for missing data)
		if curr != nil {
			// Helper to create float pointer
			toPtr := func(v float64) *float64 { return &v }
			
			// For instantaneous values, we just take the value directly.
			// No delta calculation needed.
			
			if curr.Site1Power > 0 { result[i].Site1Power = toPtr(curr.Site1Power) }
			if curr.Site1Irradiance > 0 { result[i].Site1Irradiance = toPtr(curr.Site1Irradiance) }
			
			if curr.Site2Power > 0 { result[i].Site2Power = toPtr(curr.Site2Power) }
			if curr.Site2Irradiance > 0 { result[i].Site2Irradiance = toPtr(curr.Site2Irradiance) }
		}
	}
	return result
}

// fetchMonthlyProductionData fetches daily max power and irradiance from Feb 9 to today
func fetchMonthlyProductionData() []MonthlyDataPoint {
	endpoint := config.App.System.VMEndpoint
	now := time.Now()
	
	// Start from Feb 9, 2026 (or adjust as needed)
	startDate := time.Date(2026, 2, 9, 0, 0, 0, 0, time.Local)
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)
	
	// Calculate number of days
	numDays := int(endDate.Sub(startDate).Hours()/24) + 1
	if numDays < 1 {
		numDays = 1
	}
	if numDays > 31 {
		numDays = 31 // Limit to 31 days
	}
	
	result := make([]MonthlyDataPoint, numDays)
	
	// Initialize dates
	for i := 0; i < numDays; i++ {
		day := startDate.AddDate(0, 0, i)
		result[i] = MonthlyDataPoint{
			Date: day.Format("02/01"), // DD/MM
		}
	}
	
	// Helper to convert to pointer
	toPtr := func(v float64) *float64 { return &v }
	
	// Fetch max power per day for Site 1
	// Using max_over_time with 1d step
	fetchMaxPerDay := func(query string, start, end time.Time, setter func(dayIndex int, val float64)) {
		// Query: max_over_time(metric[1d])
		// Step: 86400 (1 day)
		u := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=86400",
			endpoint, url.QueryEscape(query), start.Unix(), end.Unix())
		
		resp, err := http.Get(u)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		
		body, _ := io.ReadAll(resp.Body)
		
		var res struct {
			Data struct {
				Result []struct {
					Values [][]interface{} `json:"values"`
				} `json:"result"`
			} `json:"data"`
		}
		
		if json.Unmarshal(body, &res) != nil {
			return
		}
		
		if len(res.Data.Result) == 0 {
			return
		}
		
		for _, v := range res.Data.Result[0].Values {
			if len(v) >= 2 {
				ts := time.Unix(int64(v[0].(float64)), 0)
				dayIndex := int(ts.Sub(startDate).Hours() / 24)
				if dayIndex < 0 || dayIndex >= numDays {
					continue
				}
				valStr, _ := v[1].(string)
				var val float64
				fmt.Sscanf(valStr, "%f", &val)
				setter(dayIndex, val)
			}
		}
	}
	
	// Fetch Site 1 Max Power - use subquery syntax for max of sum
	fetchMaxPerDay(
		`max_over_time(sum(shundao_inverter{name="p_out_kw", site_name="SHUNDAO_1"})[1d:5m])`,
		startDate, endDate,
		func(i int, v float64) { if v > 0 { result[i].Site1MaxPower = toPtr(v) } },
	)
	
	// Fetch Site 1 Max Irradiance
	fetchMaxPerDay(
		`max_over_time(avg(shundao_sensor{name="total_irradiance_wm2", site_name="SHUNDAO_1"})[1d:5m])`,
		startDate, endDate,
		func(i int, v float64) { if v > 0 { result[i].Site1MaxIrrad = toPtr(v) } },
	)
	
	// Fetch Site 2 Max Power
	fetchMaxPerDay(
		`max_over_time(sum(shundao_inverter{name="p_out_kw", site_name="SHUNDAO_2"})[1d:5m])`,
		startDate, endDate,
		func(i int, v float64) { if v > 0 { result[i].Site2MaxPower = toPtr(v) } },
	)
	
	// Fetch Site 2 Max Irradiance
	fetchMaxPerDay(
		`max_over_time(avg(shundao_sensor{name="total_irradiance_wm2", site_name="SHUNDAO_2"})[1d:5m])`,
		startDate, endDate,
		func(i int, v float64) { if v > 0 { result[i].Site2MaxIrrad = toPtr(v) } },
	)
	
	return result
}

func fetchKPIFromVM() KPIData {
	endpoint := config.App.System.VMEndpoint

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
	endpoint := config.App.System.VMEndpoint
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

// fetchDailyRange fetches time-series data and calls callback for each point
func fetchDailyRange(endpoint, query string, start, end time.Time, step int, callback func(ts time.Time, val float64)) {
	// Step defined by caller (e.g., 3600 for 1h, 86400 for 1d)
	u := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=%d",
		endpoint, url.QueryEscape(query), start.Unix(), end.Unix(), step)
	
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
