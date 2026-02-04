package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"fusion/src/config"
)

// SiteNode represents a node in the site tree (full data)
type SiteNode struct {
	ChildList    []SiteNode `json:"childList"`
	ElementDn    string     `json:"elementDn"`
	ElementId    string     `json:"elementId"`
	HasMoreChild bool       `json:"hasMoreChild"`
	IsParent     bool       `json:"isParent"`
	MocId        int        `json:"mocId"`
	NodeIcon     string     `json:"nodeIcon"`
	NodeName     string     `json:"nodeName"`
	ParentDn     string     `json:"parentDn"`
	Status       string     `json:"status"`
	TypeId       int        `json:"typeId"`
}

// SiteTree represents the root structure (full data)
type SiteTree struct {
	ChildList    []SiteNode `json:"childList"`
	HasMoreChild bool       `json:"hasMoreChild"`
}

// SimpleSite represents simplified site data for output
type SimpleSite struct {
	ElementDn string `json:"elementDn"`
	NodeName  string `json:"nodeName"`
	ParentDn  string `json:"parentDn"`
	Status    string `json:"status"`
}

// SimpleOutput represents the simplified output structure
type SimpleOutput struct {
	ChildList []SimpleSite `json:"childList"`
}

// FilterSites filters the site data to only include target sites with simplified fields
func FilterSites(jsonData string) (string, error) {
	var tree SiteTree
	if err := json.Unmarshal([]byte(jsonData), &tree); err != nil {
		return "", err
	}

	// Get target site IDs from config
	targetIDs := make(map[string]bool)
	for _, s := range config.App.Sites {
		targetIDs[s.ID] = true
	}

	// Filter to only target sites
	var filtered []SiteNode
	filterNodes(tree.ChildList, targetIDs, &filtered)

	// Convert to simplified output
	var simplified []SimpleSite
	for _, node := range filtered {
		simplified = append(simplified, SimpleSite{
			ElementDn: node.ElementDn,
			NodeName:  node.NodeName,
			ParentDn:  node.ParentDn,
			Status:    node.Status,
		})
	}

	result := SimpleOutput{
		ChildList: simplified,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// filterNodes recursively searches for target sites
func filterNodes(nodes []SiteNode, targetIDs map[string]bool, result *[]SiteNode) {
	for _, node := range nodes {
		if targetIDs[node.ElementDn] {
			*result = append(*result, node)
		}
		// Also search in children
		if len(node.ChildList) > 0 {
			filterNodes(node.ChildList, targetIDs, result)
		}
	}
}

// Fetcher handles API data fetching
type Fetcher struct {
	roarand    string
	requestIDs []network.RequestID
	mu         sync.Mutex
}

// NewFetcher creates a new Fetcher instance
func NewFetcher() *Fetcher {
	return &Fetcher{
		requestIDs: make([]network.RequestID, 0),
	}
}

// ClearToken resets the Roarand token
func (f *Fetcher) ClearToken() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.roarand = ""
	f.requestIDs = make([]network.RequestID, 0)
}

// HasValidToken checks if the fetcher has a token and verifies it
// It returns true if token exists and simple API call succeeds
// It returns false if token is missing or expired
func (f *Fetcher) HasValidToken(ctx context.Context) bool {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return false
	}

	// Try a lightweight API validation (e.g., list sites simplified)
	// Just check if we can get a 200 OK from a known endpoint
	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/station/v1/overview/station-kpi-data';
			// Use a dummy station DN or just check if we get 401
			// Actually, let's use the first target site if available
			url += '?stationDn=NE=00000000'; 
			url += '&_=' + Date.now();
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			return xhr.status;
		})()
	`, token)

	var status int
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &status))
	if err != nil {
		return false
	}

	// 200 OK means valid.
	// 401/403 means invalid.
	// Even if station DN is wrong, we should get 200 with empty data or specific error, NOT 401.
	if status == 401 || status == 403 {
		return false
	}

	// Also if we get 302 redirect to login, it wraps to something else usually?
	// For now assume non-401 is "session alive"
	return true
}

// SetupNetworkListener listens for network events to capture Roarand token
func (f *Fetcher) SetupNetworkListener(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			// Try to capture Roarand from any request headers
			for key, val := range e.Request.Headers {
				if strings.ToLower(key) == "roarand" {
					f.mu.Lock()
					if f.roarand == "" {
						f.roarand = val.(string)
						fmt.Printf(">>> Bắt được Roarand từ request: %s...\n", f.roarand[:min(30, len(f.roarand))])
					}
					f.mu.Unlock()
					break
				}
			}

		case *network.EventResponseReceived:
			if strings.Contains(e.Response.URL, "locate-tree") && e.Response.Status == 200 {
				f.mu.Lock()
				f.requestIDs = append(f.requestIDs, e.RequestID)
				f.mu.Unlock()
			}

			// Also try to capture Roarand from response headers
			for key, val := range e.Response.Headers {
				if strings.ToLower(key) == "roarand" {
					f.mu.Lock()
					if f.roarand == "" {
						f.roarand = val.(string)
						fmt.Printf(">>> Bắt được Roarand từ response: %s...\n", f.roarand[:min(30, len(f.roarand))])
					}
					f.mu.Unlock()
					break
				}
			}
		}
	})
}

// EnableNetwork enables network monitoring
func (f *Fetcher) EnableNetwork(ctx context.Context) error {
	return chromedp.Run(ctx, network.Enable())
}

// WaitAndFetchSiteData waits for page load and fetches site data
func (f *Fetcher) WaitAndFetchSiteData(ctx context.Context) (string, error) {
	var siteData string

	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		fmt.Println("[Chờ] Đang load trang (15 giây)...")
		time.Sleep(15 * time.Second)

		// Try to extract Roarand from browser if not captured
		f.mu.Lock()
		hasToken := f.roarand != ""
		f.mu.Unlock()

		if !hasToken {
			// Try to get Roarand from cookies or window object
			var token string
			chromedp.Evaluate(`
				(function() {
					// Try window.roarand
					if (window.roarand) return window.roarand;
					
					// Try to find in cookies
					var cookies = document.cookie.split(';');
					for (var i = 0; i < cookies.length; i++) {
						var cookie = cookies[i].trim();
						if (cookie.toLowerCase().startsWith('roarand=')) {
							return cookie.substring(8);
						}
					}
					
					// Try localStorage
					var ls = localStorage.getItem('roarand');
					if (ls) return ls;
					
					// Try sessionStorage
					var ss = sessionStorage.getItem('roarand');
					if (ss) return ss;
					
					return '';
				})()
			`, &token).Do(ctx)

			if token != "" {
				f.mu.Lock()
				f.roarand = token
				f.mu.Unlock()
				fmt.Printf(">>> Lấy được Roarand từ browser: %s...\n", token[:min(30, len(token))])
			} else {
				fmt.Println(">>> Không tìm thấy Roarand token trong browser")
			}
		}

		// Try to get response from captured requests
		f.mu.Lock()
		ids := make([]network.RequestID, len(f.requestIDs))
		copy(ids, f.requestIDs)
		f.mu.Unlock()

		fmt.Printf("      Số API locate-tree đã phát hiện: %d\n", len(ids))

		for i, reqID := range ids {
			fmt.Printf("      Thử lấy body từ request %d...\n", i+1)

			body, err := network.GetResponseBody(reqID).Do(ctx)
			if err != nil {
				continue
			}

			if len(body) > 0 && body[0] == '{' {
				siteData = string(body)
				fmt.Printf("      ✓ Lấy được response! Length: %d bytes\n", len(body))
				break
			}
		}

		// If no data, try calling API directly with Roarand
		if siteData == "" {
			f.mu.Lock()
			token := f.roarand
			f.mu.Unlock()

			if token != "" {
				fmt.Println("      Thử gọi API trực tiếp...")
				siteData = f.callAPIDirectly(ctx, token, "NE=50987774")
			}
		}

		return nil
	}))

	return siteData, err
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// FetchSiteData fetches data for a specific site
func (f *Fetcher) FetchSiteData(ctx context.Context, siteID string) (string, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return "", fmt.Errorf("no Roarand token available")
	}

	return f.callAPIDirectly(ctx, token, siteID), nil
}

func (f *Fetcher) callAPIDirectly(ctx context.Context, token, targetDn string) string {
	endpoint := config.App.API.Endpoints["locate_tree"]
	subNodeTypes := config.App.API.Params.SubNodeTypeIDs
	typeIDInclude := config.App.API.Params.TypeIDInclude

	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var url = '%s';
			url += '?targetDn=%s';
			url += '&subNodeTypeIds=%s';
			url += '&typeIdInclude=%s';
			url += '&_=' + Date.now();
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			return xhr.responseText;
		})()
	`, endpoint, strings.ReplaceAll(targetDn, "=", "%3D"), subNodeTypes, typeIDInclude, token)

	var result string
	chromedp.Evaluate(js, &result).Do(ctx)

	if strings.HasPrefix(result, "{") && len(result) > 100 {
		return result
	}

	return ""
}

// SmartLoggerDevice represents a SmartLogger device
type SmartLoggerDevice struct {
	ElementDn string `json:"elementDn"`
	NodeName  string `json:"nodeName"`
	Status    string `json:"status"`
	TypeId    int    `json:"typeId"`
}

// SmartLoggerOutput represents the output for SmartLogger list
type SmartLoggerOutput struct {
	Site    string              `json:"site"`
	Devices []SmartLoggerDevice `json:"devices"`
}

// FetchSmartLoggers fetches SmartLogger devices for a site using POST API
func (f *Fetcher) FetchSmartLoggers(ctx context.Context, parentDn string) ([]SmartLoggerDevice, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	// POST API to /organization/v1/tree
	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var url = 'https://intl.fusionsolar.huawei.com/rest/dp/pvms/organization/v1/tree';
			
			var payload = {
				"parentDn": "%s",
				"treeDepth": "device",
				"pageParam": {"pageId": 1, "pageSize": 100, "needPage": true},
				"displayCond": {"self": true, "status": true},
				"filterCond": {
					"nameType": "device",
					"mocIdInclude": [],
					"typeIdInclude": [23089, 23091]
				}
			};
			
			xhr.open('POST', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('Content-Type', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send(JSON.stringify(payload));
			return xhr.responseText;
		})()
	`, parentDn, token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return nil, err
	}


	if !strings.HasPrefix(result, "{") {
		// Save for debugging
		if len(result) > 0 {
			os.WriteFile("output/debug_response.txt", []byte(result), 0644)
		}
		return nil, fmt.Errorf("invalid response: starts with '%s'", result[:min(20, len(result))])
	}

	// Parse nested response structure
	var response struct {
		ChildList json.RawMessage `json:"childList"`
	}

	if err := json.Unmarshal([]byte(result), &response); err != nil {
		SaveJSON(result, "", "smartlogger_raw.json")
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Extract devices recursively

	// Extract devices recursively
	var devices []SmartLoggerDevice
	extractDevicesFromJSON(response.ChildList, &devices)


	return devices, nil
}

// DeviceNode represents a node in the device tree
type DeviceNode struct {
	ChildList []DeviceNode `json:"childList"`
	ElementDn string       `json:"elementDn"`
	NodeName  string       `json:"nodeName"`
	Status    string       `json:"status"`
	TypeId    int          `json:"typeId"`
	IsParent  bool         `json:"isParent"`
}

// extractDevicesFromJSON recursively extracts devices from nested JSON
func extractDevicesFromJSON(data json.RawMessage, devices *[]SmartLoggerDevice) {
	var nodes []DeviceNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return
	}

	for _, node := range nodes {
		// Check if it's a SmartLogger (nodeName contains "Smartlogger" or "Logger")
		if strings.Contains(strings.ToLower(node.NodeName), "smartlogger") ||
			strings.Contains(strings.ToLower(node.NodeName), "logger") {
			*devices = append(*devices, SmartLoggerDevice{
				ElementDn: node.ElementDn,
				NodeName:  node.NodeName,
				Status:    node.Status,
				TypeId:    node.TypeId,
			})
		}

		// Recursively check children
		if len(node.ChildList) > 0 {
			childData, _ := json.Marshal(node.ChildList)
			extractDevicesFromJSON(childData, devices)
		}
	}
}

// Device represents a device connected to SmartLogger
type Device struct {
	ElementDn string `json:"elementDn"`
	NodeName  string `json:"nodeName"`
	Status    string `json:"status"`
	TypeId    int    `json:"typeId"`
	MocId     int    `json:"mocId"`
}

// DeviceOutput represents output for device list
type DeviceOutput struct {
	SmartLogger string   `json:"smartLogger"`
	Site        string   `json:"site"`
	Devices     []Device `json:"devices"`
}

// FetchDevicesForSmartLogger fetches devices for a SmartLogger using POST API
func (f *Fetcher) FetchDevicesForSmartLogger(ctx context.Context, parentDn string) ([]Device, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	// POST API to /organization/v1/tree with SmartLogger's elementDn as parentDn
	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var url = 'https://intl.fusionsolar.huawei.com/rest/dp/pvms/organization/v1/tree';
			
			var payload = {
				"parentDn": "%s",
				"treeDepth": "device",
				"pageParam": {"pageId": 1, "pageSize": 100, "needPage": true},
				"displayCond": {"self": true, "status": true},
				"filterCond": {
					"nameType": "device",
					"mocIdInclude": [],
					"typeIdInclude": [23089, 23091]
				}
			};
			
			xhr.open('POST', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('Content-Type', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send(JSON.stringify(payload));
			return xhr.responseText;
		})()
	`, parentDn, token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(result, "{") {
		return nil, fmt.Errorf("invalid response")
	}

	// Parse nested response
	var response struct {
		ChildList json.RawMessage `json:"childList"`
	}

	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Extract devices recursively
	var devices []Device
	extractAllDevices(response.ChildList, &devices)

	return devices, nil
}

// extractAllDevices recursively extracts all devices from nested JSON
func extractAllDevices(data json.RawMessage, devices *[]Device) {
	var nodes []DeviceNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return
	}

	for _, node := range nodes {
		// Add all non-parent nodes as devices
		if !node.IsParent && node.ElementDn != "" {
			*devices = append(*devices, Device{
				ElementDn: node.ElementDn,
				NodeName:  node.NodeName,
				Status:    node.Status,
				TypeId:    node.TypeId,
			})
		}

		// Recursively check children
		if len(node.ChildList) > 0 {
			childData, _ := json.Marshal(node.ChildList)
			extractAllDevices(childData, devices)
		}
	}
}

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
	DailyChargeCapacity         float64 `json:"dailyChargeCapacity"`
	DailyDischargeCapacity      float64 `json:"dailyDischargeCapacity"`
	CumulativeChargeCapacity    float64 `json:"cumulativeChargeCapacity"`
	CumulativeDischargeCapacity float64 `json:"cumulativeDischargeCapacity"`
	Currency                    int     `json:"currency"`
	BatteryCapacity             float64 `json:"batteryCapacity"`
	IsPriceConfigured           bool    `json:"isPriceConfigured"`
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
			ReDischargeableEnergy       string  `json:"reDischargeableEnergy"`
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
	fmt.Sscanf(response.KpiData.DailyChargeCapacity, "%f", &kpi.DailyChargeCapacity)
	fmt.Sscanf(response.KpiData.DailyDischargeCapacity, "%f", &kpi.DailyDischargeCapacity)
	fmt.Sscanf(response.KpiData.CumulativeChargeCapacity, "%f", &kpi.CumulativeChargeCapacity)
	fmt.Sscanf(response.KpiData.CumulativeDisChargeCapacity, "%f", &kpi.CumulativeDischargeCapacity)
	kpi.Currency = response.KpiData.Currency
	kpi.BatteryCapacity = response.KpiData.BatteryCapacity
	kpi.IsPriceConfigured = response.KpiData.IsPriceConfigured

	return kpi, nil
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

// EMICData represents EMIC device data
type EMICData struct {
	DeviceDn   string                 `json:"deviceDn"`
	DeviceName string                 `json:"deviceName"`
	SiteName   string                 `json:"siteName"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  int64                  `json:"timestamp"`
}

// FetchEMICData fetches realtime data for an EMIC device
func (f *Fetcher) FetchEMICData(ctx context.Context, deviceDn string) (*EMICData, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	// GET API to /device-realtime-data
	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var now = Date.now();
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/device/v1/device-realtime-data';
			url += '?deviceDn=%s';
			url += '&_=' + now;
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			return xhr.responseText;
		})()
	`, strings.ReplaceAll(deviceDn, "=", "%3D"), token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(result, "{") {
		return nil, fmt.Errorf("invalid response")
	}

	// Parse response - the structure may vary, so we'll use a flexible approach
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	emic := &EMICData{
		DeviceDn:  deviceDn,
		Data:      response,
		Timestamp: time.Now().Unix(),
	}

	return emic, nil
}

// StringDataSignalIds are the signal IDs for string data (MPPT data)
var StringDataSignalIds = []int{
	// Device info
	10032, 10025, 10029, 10019, 10022, 10006, 10020, 10021, 10027, 10028, 21029, 10018,
	10008, 10009, 10010, 10012, 10013, 10011, 10014, 10015, 10016, 10113, 10114, 10115,
	10023, 10024, 10047, 10051,
	// String data (MPPT) - voltage/current pairs
	11001, 11002, 11004, 11005, 11007, 11008, 11010, 11011, 11013, 11014, 11016, 11017,
	11019, 11020, 11022, 11023, 11025, 11026, 11028, 11029, 11031, 11032, 11034, 11035,
	11037, 11038, 11040, 11041, 11043, 11044, 11046, 11047, 11049, 11050, 11052, 11053,
	11055, 11056, 11058, 11059, 11061, 11062, 11064, 11065, 11067, 11068, 11070, 11071,
	11072, 11073, 11074, 11075, 11076, 11077, 11078, 11079, 11080, 11081, 11082, 11083,
	11084, 11085, 11086, 11087, 11088, 11089, 11090, 11091, 11092, 11093, 11094, 11095,
	11096, 11097, 11098, 11099, 11100, 11101, 11102, 11103, 11104, 11105, 11106, 11107,
	11108, 11109, 11110, 11111, 11112, 11113, 11114, 11115, 11116, 11117, 11118, 11119,
	// String status IDs
	14001, 14002, 14003, 14004, 14005, 14006, 14007, 14008, 14009, 14010, 14011, 14012,
	14013, 14014, 14015, 14016, 14017, 14018, 14019, 14020, 14021, 14022, 14023, 14024,
	14025, 14026, 14027, 14028, 14029, 14030, 14031, 14032, 14033, 14034, 14035, 14036,
	14037, 14038, 14039, 14040, 14041, 14042, 14043, 14044, 14045, 14046, 14047, 14048,
}

// FetchInverterStringData fetches string/MPPT data for an inverter
func (f *Fetcher) FetchInverterStringData(ctx context.Context, deviceDn string) (map[string]interface{}, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	// Build signalIds query string
	signalIdsStr := ""
	for i, id := range StringDataSignalIds {
		if i > 0 {
			signalIdsStr += "&signalIds=" + fmt.Sprint(id)
		} else {
			signalIdsStr += fmt.Sprint(id)
		}
	}

	// GET API to /device-real-kpi
	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var now = Date.now();
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/device/v1/device-real-kpi';
			url += '?signalIds=%s';
			url += '&deviceDn=%s';
			url += '&_=' + now;
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			return xhr.responseText;
		})()
	`, signalIdsStr, strings.ReplaceAll(deviceDn, "=", "%3D"), token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(result, "{") {
		return nil, fmt.Errorf("invalid response")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return response, nil
}

// FetchInverterRealtimeData fetches realtime data for an inverter with displayAccessModel
func (f *Fetcher) FetchInverterRealtimeData(ctx context.Context, deviceDn string) (map[string]interface{}, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	// GET API to /device-realtime-data with displayAccessModel=true
	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var now = Date.now();
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/device/v1/device-realtime-data';
			url += '?deviceDn=%s';
			url += '&displayAccessModel=true';
			url += '&_=' + now;
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			return xhr.responseText;
		})()
	`, strings.ReplaceAll(deviceDn, "=", "%3D"), token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(result, "{") {
		return nil, fmt.Errorf("invalid response")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return response, nil
}

// GetRoarand returns the captured Roarand token
func (f *Fetcher) GetRoarand() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.roarand
}

// SaveJSON saves data to a JSON file in the specified subdirectory
func SaveJSON(data, subdir, filename string) error {
	// Create output directory with subdirectory if not exists
	outputPath := filepath.Join("output", subdir)
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}

	fullPath := filepath.Join(outputPath, filename)

	// Pretty print JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(data), &parsed); err == nil {
		prettyJSON, _ := json.MarshalIndent(parsed, "", "  ")
		return os.WriteFile(fullPath, prettyJSON, 0644)
	}

	return os.WriteFile(fullPath, []byte(data), 0644)
}

// FetchSmartLoggerDetail fetches detailed info (SN, Model, IP, etc.) specifically for SmartLogger
func (f *Fetcher) FetchSmartLoggerDetail(ctx context.Context, deviceDn string) (map[string]interface{}, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	// Try using the exact API provided by the user
	// signals: 10051, 21029, 24001, 50001, 50009, 50010, 50012, 50018, 33595393, 50020, 50022, 14054, 11248
	signals := "10051,21029,24001,50001,50009,50010,50012,50018,33595393,50020,50022,14054,11248"

	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var now = Date.now();
			var url = 'https://intl.fusionsolar.huawei.com/rest/neteco/web/config/device/v1/config/query-moc-config-signal';
			url += '?dn=%s';
			url += '&signals=%s';
			url += '&_=' + now;
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			return xhr.responseText;
		})()
	`, strings.ReplaceAll(deviceDn, "=", "%3D"), signals, token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(result, "{") {
		fmt.Printf("      [DEBUG] Invalid response for %s. Preview: %s\n", deviceDn, result[:min(200, len(result))])
		return nil, fmt.Errorf("invalid response")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return response, nil
}

// FetchDeviceDetail fetches detailed info (SN, Model, IP, etc.) for any device
func (f *Fetcher) FetchDeviceDetail(ctx context.Context, deviceDn string) (map[string]interface{}, error) {
	// Re-use logic of FetchInverterRealtimeData as it uses the same endpoint
	// /device-realtime-data which returns parameters like IP, SN for SmartLoggers
	return f.FetchInverterRealtimeData(ctx, deviceDn)
}

// ChildDevice represents a secondary device connected to SmartLogger
type ChildDevice struct {
	Dn          string                 `json:"dn"`
	Name        string                 `json:"name"`
	ParentName  string                 `json:"parentName"`
	MocTypeName string                 `json:"mocTypeName"`
	Status      string                 `json:"status"`
	ParamValues map[string]interface{} `json:"paramValues"`
}

// FetchSmartLoggerChildren fetches secondary devices for a SmartLogger
func (f *Fetcher) FetchSmartLoggerChildren(ctx context.Context, parentDn string) ([]ChildDevice, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	// GET API to /children-list
	// mocTypes provided by user: 20822,20810,20825,20826,20823,20824,20816,20838,20836,20835,20844,20847,20865
	mocTypes := "20822,20810,20825,20826,20823,20824,20816,20838,20836,20835,20844,20847,20865"

	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var now = Date.now();
			var url = 'https://intl.fusionsolar.huawei.com/rest/neteco/web/config/device/v1/children-list';
			url += '?conditionParams.curPage=0';
			url += '&conditionParams.recordperpage=500';
			url += '&conditionParams.parentDn=%s';
			url += '&conditionParams.monitoringRelation=true';
			url += '&conditionParams.mocTypes=%s';
			url += '&_=' + now;
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('X-Timezone-Offset', '420');
			xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			return xhr.responseText;
		})()
	`, strings.ReplaceAll(parentDn, "=", "%3D"), mocTypes, token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return nil, err
	}

	// Check if result is empty or invalid
	if len(result) == 0 {
		return nil, fmt.Errorf("empty response from browser")
	}

	if !strings.HasPrefix(result, "{") {
		// Log a warning relative to file content length or similar if needed, but here just error
		return nil, fmt.Errorf("invalid response")
	}

	// Parse response
	var response struct {
		Data []ChildDevice `json:"data"`
	}

	if err := json.Unmarshal([]byte(result), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return response.Data, nil
}
