// Package repository: VictoriaMetrics implementation của DashboardRepository.
// File này là "adapter" kết nối usecase layer với VM infrastructure.
package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"fusion/internal/domain"
	"fusion/internal/platform/config"
	"fusion/internal/platform/utils"
)

// VMDashboardRepo implements DashboardRepository bằng cách gọi VictoriaMetrics HTTP API.
type VMDashboardRepo struct {
	endpoint string
}

// NewVMDashboardRepo tạo repo mới, lấy endpoint từ config.
func NewVMDashboardRepo() *VMDashboardRepo {
	return &VMDashboardRepo{endpoint: config.App.System.VMEndpoint}
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func (r *VMDashboardRepo) queryVM(query string, start, end time.Time, step string) ([][]interface{}, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", fmt.Sprintf("%d", start.Unix()))
	params.Set("end", fmt.Sprintf("%d", end.Unix()))
	params.Set("step", step)

	resp, err := http.Get(r.endpoint + "/api/v1/query_range?" + params.Encode())
	if err != nil {
		return nil, err
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
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.Data.Result) == 0 {
		return nil, nil
	}
	return result.Data.Result[0].Values, nil
}

func (r *VMDashboardRepo) queryInstant(query string) (float64, error) {
	params := url.Values{}
	params.Set("query", query)
	resp, err := http.Get(r.endpoint + "/api/v1/query?" + params.Encode())
	if err != nil {
		return 0, err
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
		return 0, err
	}
	if len(result.Data.Result) == 0 || len(result.Data.Result[0].Value) < 2 {
		return 0, nil
	}
	var val float64
	fmt.Sscanf(fmt.Sprintf("%v", result.Data.Result[0].Value[1]), "%f", &val)
	return val, nil
}

// ─── DashboardRepository implementation ───────────────────────────────────────

// FetchKPI truy vấn tổng hợp KPI từ VictoriaMetrics.
func (r *VMDashboardRepo) FetchKPI() domain.KPIData {
	queries := map[string]string{
		"dailyEnergy":       `sum(increase(ac_energy_kwh_total[1d]))`,
		"totalEnergy":       `sum(ac_energy_kwh_total)`,
		"dailyIncome":       `sum(increase(ac_energy_kwh_total[1d])) * 2000`,
		"ratedPower":        `sum(rated_power_kw)`,
		"gridSupplyToday":   `sum(increase(grid_energy_kwh_total[1d]))`,
		"standardCoalSaved": `sum(increase(ac_energy_kwh_total[1d])) * 0.0003215`,
		"co2Reduction":      `sum(increase(ac_energy_kwh_total[1d])) * 0.000997`,
		"treesPlanted":      `sum(increase(ac_energy_kwh_total[1d])) * 0.0002`,
	}

	results := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for key, q := range queries {
		wg.Add(1)
		go func(k, query string) {
			defer wg.Done()
			val, err := r.queryInstant(query)
			if err != nil {
				utils.LogError("[VM-KPI] %s: %v", k, err)
				return
			}
			mu.Lock()
			results[k] = val
			mu.Unlock()
		}(key, q)
	}
	wg.Wait()

	return domain.KPIData{
		DailyEnergy:       results["dailyEnergy"],
		TotalEnergy:       results["totalEnergy"],
		DailyIncome:       results["dailyIncome"],
		RatedPower:        results["ratedPower"],
		GridSupplyToday:   results["gridSupplyToday"],
		StandardCoalSaved: results["standardCoalSaved"],
		CO2Reduction:      results["co2Reduction"],
		TreesPlanted:      results["treesPlanted"],
	}
}

// FetchProductionData truy xuất dữ liệu công suất 5 phút từ VM.
func (r *VMDashboardRepo) FetchProductionData() []domain.ProductionDataPoint {
	loc := time.FixedZone("UTC+7", 7*3600)
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	type rowKey = string
	type rawRow struct {
		Site1Power      float64
		Site1Irradiance float64
		Site2Power      float64
		Site2Irradiance float64
	}
	rawMap := make(map[rowKey]*rawRow)
	var mu sync.Mutex

	getOrCreate := func(ts time.Time) *rawRow {
		key := ts.In(loc).Format("15:04")
		mu.Lock()
		defer mu.Unlock()
		if rawMap[key] == nil {
			rawMap[key] = &rawRow{}
		}
		return rawMap[key]
	}

	vmQueries := []struct {
		q    string
		fill func(*rawRow, float64)
	}{
		{`sum(p_out_kw{site="shundao1"})`, func(row *rawRow, v float64) { row.Site1Power = v }},
		{`avg(irradiance_wm2{site="shundao1"})`, func(row *rawRow, v float64) { row.Site1Irradiance = v }},
		{`sum(p_out_kw{site="shundao2"})`, func(row *rawRow, v float64) { row.Site2Power = v }},
		{`avg(irradiance_wm2{site="shundao2"})`, func(row *rawRow, v float64) { row.Site2Irradiance = v }},
	}

	var wg sync.WaitGroup
	for _, vq := range vmQueries {
		wg.Add(1)
		go func(q string, fill func(*rawRow, float64)) {
			defer wg.Done()
			vals, err := r.queryVM(q, start, now, "300")
			if err != nil || vals == nil {
				return
			}
			for _, pair := range vals {
				if len(pair) < 2 {
					continue
				}
				ts := time.Unix(int64(pair[0].(float64)), 0)
				var v float64
				fmt.Sscanf(fmt.Sprintf("%v", pair[1]), "%f", &v)
				fill(getOrCreate(ts), v)
			}
		}(vq.q, vq.fill)
	}
	wg.Wait()

	// Sort by time key
	keys := make([]string, 0, len(rawMap))
	for k := range rawMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]domain.ProductionDataPoint, 0, len(keys))
	for _, k := range keys {
		row := rawMap[k]
		p1, i1, p2, i2 := row.Site1Power, row.Site1Irradiance, row.Site2Power, row.Site2Irradiance
		result = append(result, domain.ProductionDataPoint{
			Date:            k,
			Site1Power:      &p1,
			Site1Irradiance: &i1,
			Site2Power:      &p2,
			Site2Irradiance: &i2,
		})
	}
	return result
}

// FetchChartData truy xuất dữ liệu biểu đồ thời gian thực.
func (r *VMDashboardRepo) FetchChartData() []domain.ChartPoint {
	loc := time.FixedZone("UTC+7", 7*3600)
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	vals, err := r.queryVM(`sum(p_out_kw)`, start, now, "300")
	if err != nil || vals == nil {
		return []domain.ChartPoint{}
	}

	result := make([]domain.ChartPoint, 0, len(vals))
	for _, pair := range vals {
		if len(pair) < 2 {
			continue
		}
		ts := time.Unix(int64(pair[0].(float64)), 0).In(loc)
		var v float64
		fmt.Sscanf(fmt.Sprintf("%v", pair[1]), "%f", &v)
		result = append(result, domain.ChartPoint{
			Time:    ts.Format("15:04"),
			Power:   v,
			PvPower: v,
		})
	}
	return result
}

// FetchMonthlyProduction truy xuất dữ liệu sản lượng theo tháng.
// Hỗ trợ optional month argument (time.Time).
func (r *VMDashboardRepo) FetchMonthlyProduction(args ...interface{}) []domain.MonthlyDataPoint {
	loc := time.FixedZone("UTC+7", 7*3600)
	now := time.Now().In(loc)
	month := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

	if len(args) > 0 {
		if t, ok := args[0].(time.Time); ok && !t.IsZero() {
			month = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}
	}

	start := month
	end := month.AddDate(0, 1, 0).Add(-time.Second)

	vals, err := r.queryVM(`max_over_time(sum(p_out_kw)[5m:5m])`, start, end, "86400")
	if err != nil || vals == nil {
		return []domain.MonthlyDataPoint{}
	}

	result := make([]domain.MonthlyDataPoint, 0, len(vals))
	for _, pair := range vals {
		if len(pair) < 2 {
			continue
		}
		ts := time.Unix(int64(pair[0].(float64)), 0).In(loc)
		var v float64
		fmt.Sscanf(fmt.Sprintf("%v", pair[1]), "%f", &v)
		result = append(result, domain.MonthlyDataPoint{
			Date:          ts.Format("02/01"),
			Site1MaxPower: &v,
		})
	}
	return result
}

// FetchInverterPowerData truy xuất DC/AC của một inverter theo deviceID.
func (r *VMDashboardRepo) FetchInverterPowerData(deviceID string) []domain.InverterPowerPoint {
	loc := time.FixedZone("UTC+7", 7*3600)
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	type pointMap = map[string]*domain.InverterPowerPoint
	points := make(pointMap)
	var mu sync.Mutex

	getOrCreate := func(ts time.Time) *domain.InverterPowerPoint {
		key := ts.In(loc).Format("15:04")
		mu.Lock()
		defer mu.Unlock()
		if points[key] == nil {
			points[key] = &domain.InverterPowerPoint{Time: key}
		}
		return points[key]
	}

	var wg sync.WaitGroup

	// DC power
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := fmt.Sprintf(`dc_power_kw{device="%s"}`, deviceID)
		vals, err := r.queryVM(q, start, now, "300")
		if err != nil || vals == nil {
			return
		}
		for _, pair := range vals {
			if len(pair) < 2 {
				continue
			}
			ts := time.Unix(int64(pair[0].(float64)), 0)
			var v float64
			fmt.Sscanf(fmt.Sprintf("%v", pair[1]), "%f", &v)
			p := getOrCreate(ts)
			vCopy := v
			p.DcPower = &vCopy
		}
	}()

	// AC power
	wg.Add(1)
	go func() {
		defer wg.Done()
		q := fmt.Sprintf(`p_out_kw{device="%s"}`, deviceID)
		vals, err := r.queryVM(q, start, now, "300")
		if err != nil || vals == nil {
			return
		}
		for _, pair := range vals {
			if len(pair) < 2 {
				continue
			}
			ts := time.Unix(int64(pair[0].(float64)), 0)
			var v float64
			fmt.Sscanf(fmt.Sprintf("%v", pair[1]), "%f", &v)
			p := getOrCreate(ts)
			vCopy := v
			p.AcPower = &vCopy
		}
	}()

	wg.Wait()

	keys := make([]string, 0, len(points))
	for k := range points {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]domain.InverterPowerPoint, 0, len(keys))
	for _, k := range keys {
		result = append(result, *points[k])
	}
	return result
}
