package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"fusion/internal/platform/config"
	"fusion/internal/platform/utils"
)

// fetchProductionDataFromVM fetches 5-minute interval production data for today (per site)
func fetchProductionDataFromVM() []ProductionDataPoint {
	endpoint := config.App.System.VMEndpoint
	now := time.Now()
	loc := time.FixedZone("UTC+7", 7*60*60)

	// Start of TODAY (00:00) in UTC+7
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
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
	var rawMu sync.Mutex
	getRaw := func(ts time.Time) *RawPoint {
		key := ts.In(loc).Format("15:04")
		rawMu.Lock()
		defer rawMu.Unlock()
		if _, ok := rawMap[key]; !ok {
			rawMap[key] = &RawPoint{Date: key}
		}
		return rawMap[key]
	}

	// === Concurrent fetching of all 4 time-series (WaitGroup + Mutex) ===
	var wg sync.WaitGroup
	type rangeTask struct {
		query  string
		setter func(ts time.Time, val float64)
	}
	rangeTasks := []rangeTask{
		{`sum(shundao_inverter{name="p_out_kw", site_name="SHUNDAO_1"})`, func(ts time.Time, val float64) { getRaw(ts).Site1Power = val }},
		{`avg(shundao_sensor{name="total_irradiance_wm2", site_name="SHUNDAO_1"})`, func(ts time.Time, val float64) { getRaw(ts).Site1Irradiance = val }},
		{`sum(shundao_inverter{name="p_out_kw", site_name="SHUNDAO_2"})`, func(ts time.Time, val float64) { getRaw(ts).Site2Power = val }},
		{`avg(shundao_sensor{name="total_irradiance_wm2", site_name="SHUNDAO_2"})`, func(ts time.Time, val float64) { getRaw(ts).Site2Irradiance = val }},
	}
	for _, t := range rangeTasks {
		wg.Add(1)
		go func(task rangeTask) {
			defer wg.Done()
			fetchDailyRange(endpoint, task.query, start, end, 300, task.setter)
		}(t)
	}
	wg.Wait()


	// Helper to safely get raw values (return nil if missing)
	getRawVal := func(key string) *RawPoint {
		if p, ok := rawMap[key]; ok {
			return p
		}
		return nil
	}

	// Process & Create Result (5-minute intervals)
	// Rules:
	// 1. Start at 06:00
	// 2. End at MIN(Now, 18:00) => Dynamic effect

	// Create a new start time at 06:00
	startAt6 := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, loc)

	// Determine end time limit (18:00)
	endAt18 := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, loc)

	// Determine actual limit: Min(Now, 18:00)
	limitTime := now
	if limitTime.After(endAt18) {
		limitTime = endAt18
	}

	// If now is before 6am, return empty or just one point?
	// Let's return empty if before 6am to avoid negative slice size
	if limitTime.Before(startAt6) {
		return []ProductionDataPoint{}
	}

	// Calculate number of 5-min slots
	duration := limitTime.Sub(startAt6)
	totalSlots := int(duration.Minutes() / 5)

	// Create result slice
	result := make([]ProductionDataPoint, totalSlots+1)

	slotTime := startAt6
	for i := 0; i <= totalSlots; i++ {
		h := slotTime.Format("15:04")
		result[i] = ProductionDataPoint{Date: h}

		// Fill data if available
		curr := getRawVal(h)
		if curr != nil {
			toPtr := func(v float64) *float64 { return &v }
			if curr.Site1Power > 0 {
				result[i].Site1Power = toPtr(curr.Site1Power)
			}
			if curr.Site1Irradiance > 0 {
				result[i].Site1Irradiance = toPtr(curr.Site1Irradiance)
			}
			if curr.Site2Power > 0 {
				result[i].Site2Power = toPtr(curr.Site2Power)
			}
			if curr.Site2Irradiance > 0 {
				result[i].Site2Irradiance = toPtr(curr.Site2Irradiance)
			}
		}

		slotTime = slotTime.Add(5 * time.Minute)
	}

	return result
}

// fetchMonthlyProductionData fetches daily total energy and total irradiance for each day of the given month.
// selectedMonth: any time.Time in the desired month (uses its Year+Month). Zero value = current month.
func fetchMonthlyProductionData(selectedMonth time.Time) []MonthlyDataPoint {
	endpoint := config.App.System.VMEndpoint
	now := time.Now()

	// Default to current month if zero
	if selectedMonth.IsZero() {
		selectedMonth = now
	}

	// First day of selected month → last day (or today if current month)
	startDate := time.Date(selectedMonth.Year(), selectedMonth.Month(), 1, 0, 0, 0, 0, time.Local)
	lastDay := startDate.AddDate(0, 1, -1) // last day of the month
	endDate := time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), 23, 59, 59, 0, time.Local)
	// If the selected month is the current month, cap end at today
	today := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.Local)
	if endDate.After(today) {
		endDate = today
	}

	// Number of days to display
	numDays := int(endDate.Sub(startDate).Hours()/24) + 1
	if numDays < 1 {
		numDays = 1
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

	// fetchPerDay: uses query_range with step=86400 (1 day) to get one value per day
	fetchPerDay := func(query string, start, end time.Time, setter func(dayIndex int, val float64)) {
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

	// --- Site 1 ---
	// Tổng sản lượng ngày (kWh): sum của max edaily_kwh trong ngày (= giá trị cuối ngày của từng inverter)
	fetchPerDay(
		`sum(max_over_time(shundao_inverter{name="edaily_kwh", site_name="SHUNDAO_1"}[1d:5m]))`,
		startDate, endDate,
		func(i int, v float64) {
			if v > 0 {
				result[i].Site1MaxPower = toPtr(v)
			}
		},
	)
	
	fetchPerDay(
        // Lấy giá trị lớn nhất trong ngày của cảm biến (giá trị chốt cuối ngày)
        `avg(max_over_time(shundao_sensor{name="daily_irradiation1_mjm2", site_name="SHUNDAO_1"}[1d:5m]))`,
        startDate, endDate,
        func(i int, v float64) {
            if v >= 0 {
                result[i].Site1MaxIrrad = toPtr(v)
            }
        },
    )

	// --- Site 2 ---
	fetchPerDay(
		`sum(max_over_time(shundao_inverter{name="edaily_kwh", site_name="SHUNDAO_2"}[1d:5m]))`,
		startDate, endDate,
		func(i int, v float64) {
			if v > 0 {
				result[i].Site2MaxPower = toPtr(v)
			}
		},
	)

	fetchPerDay(
        // Lấy giá trị lớn nhất trong ngày của cảm biến (giá trị chốt cuối ngày)
        `avg(max_over_time(shundao_sensor{name="daily_irradiation1_mjm2", site_name="SHUNDAO_2"}[1d:5m]))`,
        startDate, endDate,
        func(i int, v float64) {
            if v >= 0 {
                result[i].Site2MaxIrrad = toPtr(v)
            }
        },
    )

	return result
}

func fetchKPIFromVM() KPIData {
	endpoint := config.App.System.VMEndpoint

	// === Concurrent KPI Fetching (Goroutines + WaitGroup) ===
	// All 8 HTTP requests fire simultaneously. Total latency = slowest single request.
	var mu sync.Mutex
	var wg sync.WaitGroup
	kpi := KPIData{}

	type kpiTask struct {
		query  string
		setter func(v float64)
	}

	tasks := []kpiTask{
		{`last_over_time(shundao_plant{name="daily_energy"}[1h])`, func(v float64) { mu.Lock(); kpi.DailyEnergy = v; mu.Unlock() }},
		{`last_over_time(shundao_plant{name="cumulative_energy"}[1h])`, func(v float64) { mu.Lock(); kpi.TotalEnergy = v; mu.Unlock() }},
		{`last_over_time(shundao_plant{name="daily_income"}[1h])`, func(v float64) { mu.Lock(); kpi.DailyIncome = v; mu.Unlock() }},
		{`last_over_time(shundao_inverter{name="rated_power_kw"}[1h])`, func(v float64) { mu.Lock(); kpi.RatedPower = v; mu.Unlock() }},
		{`last_over_time(shundao_plant{name="daily_ongrid_energy"}[1h])`, func(v float64) { mu.Lock(); kpi.GridSupplyToday = v; mu.Unlock() }},
		{`last_over_time(shundao_plant{name="co2_reduction"}[1h])`, func(v float64) { mu.Lock(); kpi.CO2Reduction = v; mu.Unlock() }},
		{`last_over_time(shundao_plant{name="equivalent_trees"}[1h])`, func(v float64) { mu.Lock(); kpi.TreesPlanted = v; mu.Unlock() }},
		{`last_over_time(shundao_plant{name="standard_coal_savings"}[1h])`, func(v float64) { mu.Lock(); kpi.StandardCoalSaved = v; mu.Unlock() }},
	}

	for _, t := range tasks {
		wg.Add(1)
		go func(task kpiTask) {
			defer wg.Done()
			v := querySum(endpoint, task.query)
			task.setter(v)
		}(t)
	}
	wg.Wait()

	return kpi
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
	urlQuery := fmt.Sprintf("%s/api/v1/query?query=%s", endpoint, url.QueryEscape(query))

	resp, err := http.Get(urlQuery)
	if err != nil {
		utils.LogError("[ERROR] Failed to fetch data for queryByLabel: %v", err)
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
		utils.LogError("[ERROR] Failed to unmarshal queryByLabel response: %v", err)
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

// fetchInverterPowerData fetches today's DC + AC power data for a specific inverter
func fetchInverterPowerData(deviceID string) []InverterPowerPoint {
	endpoint := config.App.System.VMEndpoint
	now := time.Now()
	loc := time.FixedZone("UTC+7", 7*60*60)
	// Force Timezone to UTC+7
	now = now.In(loc)

	// Start at 06:00, end at 18:00 (solar hours only)
	start := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, loc)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, loc)
	end := now
	if end.After(endOfDay) {
		end = endOfDay
	}

	// Map: time string -> InverterPowerPoint
	pointsMap := make(map[string]*InverterPowerPoint)
	var timeKeys []string // preserve order

	// Helper to fetch a single metric and populate the map
	fetchMetric := func(metricName string, setter func(p *InverterPowerPoint, val float64)) {
		query := fmt.Sprintf(`shundao_inverter{name="%s", id="%s"}`, metricName, deviceID)
		u := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%d&end=%d&step=300",
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

		if len(result.Data.Result) > 0 {
			for _, v := range result.Data.Result[0].Values {
				if len(v) >= 2 {
					tsFloat, _ := v[0].(float64)
					valStr, _ := v[1].(string)

					var val float64
					fmt.Sscanf(valStr, "%f", &val)

					ts := time.Unix(int64(tsFloat), 0).In(loc)
					key := ts.Format("15:04")

					if _, exists := pointsMap[key]; !exists {
						pointsMap[key] = &InverterPowerPoint{Time: key}
						timeKeys = append(timeKeys, key)
					}
					setter(pointsMap[key], val)
				}
			}
		}
	}

	// Công suất thuần (DC) → đường cam
	fetchMetric("dc_power_kw", func(p *InverterPowerPoint, val float64) {
		p.DcPower = &val
	})
	// Tổng công suất đầu vào → đường xanh
	fetchMetric("p_out_kw", func(p *InverterPowerPoint, val float64) {
		p.AcPower = &val
	})

	// Build sorted result
	sort.Strings(timeKeys)
	result := make([]InverterPowerPoint, 0, len(timeKeys))
	for _, key := range timeKeys {
		result = append(result, *pointsMap[key])
	}
	return result
}
