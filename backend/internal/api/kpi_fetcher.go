package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/chromedp/chromedp"
)

// StationKPI represents station KPI data
type StationKPI struct {
	StationDn                   string  `json:"stationDn"`
	StationName                 string  `json:"stationName"`
	DailyEnergy                 float64 `json:"dailyEnergy"`
	CumulativeEnergy            float64 `json:"cumulativeEnergy"`
	InverterPower               float64 `json:"inverterPower"`
	DailyIncome                 float64 `json:"dailyIncome"`
	TotalChargeEnergy           float64 `json:"totalChargeEnergy"`
	TotalDischargeEnergy        float64 `json:"totalDischargeEnergy"`
	DailyChargeEnergy           float64 `json:"dailyChargeEnergy"`
	DailyOnGridEnergy           float64 `json:"dailyOnGridEnergy"`
	DailyChargeCapacity         float64 `json:"dailyChargeCapacity"`
	DailyDischargeCapacity      float64 `json:"dailyDischargeCapacity"`
	CumulativeChargeCapacity    float64 `json:"cumulativeChargeCapacity"`
	CumulativeDischargeCapacity float64 `json:"cumulativeDischargeCapacity"`
	Currency                    int     `json:"currency"`
	BatteryCapacity             float64 `json:"batteryCapacity"`
	RechargeableEnergy          float64 `json:"rechargeableEnergy"`
	ReDischargeableEnergy       float64 `json:"reDischargeableEnergy"`
	IsPriceConfigured           bool    `json:"isPriceConfigured"`
}

// SocialContribution represents environmental contribution data
type SocialContribution struct {
	StationDn                    string  `json:"stationDn"`
	StationName                  string  `json:"stationName"`
	CO2Reduction                 float64 `json:"co2Reduction"`
	CO2ReductionByYear           float64 `json:"co2ReductionByYear"`
	EquivalentTreePlanting       float64 `json:"equivalentTreePlanting"`
	EquivalentTreePlantingByYear float64 `json:"equivalentTreePlantingByYear"`
	StandardCoalSavings          float64 `json:"standardCoalSavings"`
	StandardCoalSavingsByYear    float64 `json:"standardCoalSavingsByYear"`
	ComponentFlag                int     `json:"componentFlag"`
}

// FetchStationKPI fetches station KPI data (energy, power, income)
func (f *Fetcher) FetchStationKPI(ctx context.Context, stationDn string) (*StationKPI, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	// GET API to /station-kpi-data
	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/station/v1/overview/station-kpi-data';
			url += '?stationDn=%s';
			url += '&_=' + Date.now();
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			return xhr.responseText;
		})()
	`, strings.ReplaceAll(stationDn, "=", "%3D"), token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(result, "{") {
		return nil, fmt.Errorf("invalid response")
	}

	// Parse response
	var response struct {
		KpiData struct {
			DailyEnergy                 string  `json:"dailyEnergy"`
			CumulativeEnergy            string  `json:"cumulativeEnergy"`
			InverterPower               float64 `json:"inverterPower"`
			DailyIncome                 string  `json:"dailyIncome"`
			TotalChargeEnergy           string  `json:"totalChargeEnergy"`
			DailyChargeEnergy           string  `json:"dailyChargeEnergy"`
			DailyOnGridEnergy           string  `json:"dailyOnGridEnergy"`
			ReDischargeableEnergy       string  `json:"reDischargeableEnergy"`
			RechargeableEnergy          string  `json:"rechargeableEnergy"`
			DailyChargeCapacity         string  `json:"dailyChargeCapacity"`
			DailyDischargeCapacity      string  `json:"dailyDischargeCapacity"`
			CumulativeChargeCapacity    string  `json:"cumulativeChargeCapacity"`
			CumulativeDisChargeCapacity string  `json:"cumulativeDisChargeCapacity"`
			Currency                    int     `json:"currency"`
			BatteryCapacity             float64 `json:"batteryCapacity"`
			IsPriceConfigured           bool    `json:"isPriceConfigured"`
		} `json:"kpiData"`
	}

	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert string values to float64
	kpi := &StationKPI{
		StationDn: stationDn,
	}

	// Parse string to float
	fmt.Sscanf(response.KpiData.DailyEnergy, "%f", &kpi.DailyEnergy)
	fmt.Sscanf(response.KpiData.CumulativeEnergy, "%f", &kpi.CumulativeEnergy)
	kpi.InverterPower = response.KpiData.InverterPower
	fmt.Sscanf(response.KpiData.DailyIncome, "%f", &kpi.DailyIncome)
	fmt.Sscanf(response.KpiData.TotalChargeEnergy, "%f", &kpi.TotalChargeEnergy)
	fmt.Sscanf(response.KpiData.DailyChargeEnergy, "%f", &kpi.DailyChargeEnergy)
	fmt.Sscanf(response.KpiData.DailyOnGridEnergy, "%f", &kpi.DailyOnGridEnergy)
	fmt.Sscanf(response.KpiData.DailyChargeCapacity, "%f", &kpi.DailyChargeCapacity)
	fmt.Sscanf(response.KpiData.DailyDischargeCapacity, "%f", &kpi.DailyDischargeCapacity)
	fmt.Sscanf(response.KpiData.CumulativeChargeCapacity, "%f", &kpi.CumulativeChargeCapacity)
	fmt.Sscanf(response.KpiData.CumulativeDisChargeCapacity, "%f", &kpi.CumulativeDischargeCapacity)
	fmt.Sscanf(response.KpiData.ReDischargeableEnergy, "%f", &kpi.ReDischargeableEnergy)
	fmt.Sscanf(response.KpiData.RechargeableEnergy, "%f", &kpi.RechargeableEnergy)
	kpi.Currency = response.KpiData.Currency
	kpi.BatteryCapacity = response.KpiData.BatteryCapacity
	kpi.IsPriceConfigured = response.KpiData.IsPriceConfigured

	return kpi, nil
}

// FetchSocialContribution fetches environmental contribution data
func (f *Fetcher) FetchSocialContribution(ctx context.Context, stationDn string) (*SocialContribution, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	// GET API to /social-contribution
	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var now = Date.now();
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/station/v1/station/social-contribution';
			url += '?dn=%s';
			url += '&clientTime=' + now;
			url += '&timeZone=7';
			url += '&_=' + now;
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			return xhr.responseText;
		})()
	`, strings.ReplaceAll(stationDn, "=", "%3D"), token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(result, "{") {
		return nil, fmt.Errorf("invalid response")
	}

	// Parse response
	var response struct {
		Data struct {
			CO2Reduction                 float64 `json:"co2Reduction"`
			CO2ReductionByYear           float64 `json:"co2ReductionByYear"`
			EquivalentTreePlanting       float64 `json:"equivalentTreePlanting"`
			EquivalentTreePlantingByYear float64 `json:"equivalentTreePlantingByYear"`
			StandardCoalSavings          float64 `json:"standardCoalSavings"`
			StandardCoalSavingsByYear    float64 `json:"standardCoalSavingsByYear"`
			ComponentFlag                int     `json:"componentFlag"`
		} `json:"data"`
		Success bool `json:"success"`
	}

	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API returned success=false")
	}

	sc := &SocialContribution{
		StationDn:                    stationDn,
		CO2Reduction:                 response.Data.CO2Reduction,
		CO2ReductionByYear:           response.Data.CO2ReductionByYear,
		EquivalentTreePlanting:       response.Data.EquivalentTreePlanting,
		EquivalentTreePlantingByYear: response.Data.EquivalentTreePlantingByYear,
		StandardCoalSavings:          response.Data.StandardCoalSavings,
		StandardCoalSavingsByYear:    response.Data.StandardCoalSavingsByYear,
		ComponentFlag:                response.Data.ComponentFlag,
	}

	return sc, nil
}
