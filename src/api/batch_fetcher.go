package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// BatchResult represents the result for a single device in a batch
type BatchResult struct {
	Dn      string                 `json:"dn"`
	Data    map[string]interface{} `json:"data"`
	Success bool                   `json:"success"`
	Error   string                 `json:"error"`
}

// FetchBatchRealtimeData fetches realtime data for multiple devices in parallel
// isInverter: if true, adds displayAccessModel=true param
func (f *Fetcher) FetchBatchRealtimeData(ctx context.Context, deviceDns []string, isInverter bool) ([]BatchResult, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	if len(deviceDns) == 0 {
		return []BatchResult{}, nil
	}

	dnsJson, _ := json.Marshal(deviceDns)
	extraParams := ""
	if isInverter {
		extraParams = "&displayAccessModel=true"
	}

	// Use a random variable name to avoid collision
	resVar := fmt.Sprintf("_res_%d", time.Now().UnixNano())

	js := fmt.Sprintf(`
		(function() {
			window["%s"] = null; // Reset
			var deviceDns = %s;
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/device/v1/device-realtime-data';
			var extraParams = '%s';
			var token = '%s';
			
			var promises = deviceDns.map(function(dn) {
				return new Promise(function(resolve, reject) {
					var xhr = new XMLHttpRequest();
					var fullUrl = url + '?deviceDn=' + encodeURIComponent(dn) + extraParams + '&_=' + Date.now();
					xhr.open('GET', fullUrl, true); // Asynchronous
					xhr.setRequestHeader('Accept', 'application/json');
					xhr.setRequestHeader('X-Timezone-Offset', '420');
					xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
					xhr.setRequestHeader('Roarand', token);
					
					xhr.onreadystatechange = function() {
						if (xhr.readyState === 4) {
							if (xhr.status === 200) {
								try {
									var json = JSON.parse(xhr.responseText);
									resolve({dn: dn, data: json, success: true, error: ""});
								} catch (e) {
									resolve({dn: dn, data: null, success: false, error: "Parse error"});
								}
							} else {
								resolve({dn: dn, data: null, success: false, error: "Http " + xhr.status});
							}
						}
					};
					xhr.send();
				});
			});

			Promise.all(promises).then(function(results) {
				window["%s"] = JSON.stringify(results);
			});
		})()
	`, resVar, dnsJson, extraParams, token, resVar)

	// Execute JS initiation
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, nil)); err != nil {
		return nil, fmt.Errorf("batch script error: %v", err)
	}

	// Poll for result
	var resultStr string
	for i := 0; i < 40; i++ { // Wait up to 20s (500ms interval)
		time.Sleep(500 * time.Millisecond)
		err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(`window["%s"]`, resVar), &resultStr))
		if err == nil && resultStr != "" && resultStr != "null" {
			break
		}
	}

	if resultStr == "" || resultStr == "null" {
		return nil, fmt.Errorf("batch fetch timeout")
	}

	var results []BatchResult
	if err := json.Unmarshal([]byte(resultStr), &results); err != nil {
		return nil, fmt.Errorf("batch parse error: %v", err)
	}

	return results, nil
}

// FetchBatchInverterStringData fetches string data for multiple inverters in parallel
func (f *Fetcher) FetchBatchInverterStringData(ctx context.Context, deviceDns []string) ([]BatchResult, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return nil, fmt.Errorf("no Roarand token available")
	}

	if len(deviceDns) == 0 {
		return []BatchResult{}, nil
	}

	dnsJson, _ := json.Marshal(deviceDns)

	// Build signalIds query string
	signalIdsStr := ""
	for i, id := range StringDataSignalIds {
		if i > 0 {
			signalIdsStr += "&signalIds=" + fmt.Sprint(id)
		} else {
			signalIdsStr += fmt.Sprint(id)
		}
	}

	resVar := fmt.Sprintf("_str_res_%d", time.Now().UnixNano())

	js := fmt.Sprintf(`
		(function() {
			window["%s"] = null;
			var deviceDns = %s;
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/device/v1/device-real-kpi';
			var signalParams = '%s';
			var token = '%s';
			
			var promises = deviceDns.map(function(dn) {
				return new Promise(function(resolve, reject) {
					var xhr = new XMLHttpRequest();
					var fullUrl = url + '?signalIds=' + signalParams + '&deviceDn=' + encodeURIComponent(dn) + '&_=' + Date.now();
					xhr.open('GET', fullUrl, true); // Asynchronous
					xhr.setRequestHeader('Accept', 'application/json');
					xhr.setRequestHeader('X-Timezone-Offset', '420');
					xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
					xhr.setRequestHeader('Roarand', token);
					
					xhr.onreadystatechange = function() {
						if (xhr.readyState === 4) {
							if (xhr.status === 200) {
								try {
									var json = JSON.parse(xhr.responseText);
									resolve({dn: dn, data: json, success: true, error: ""});
								} catch (e) {
									resolve({dn: dn, data: null, success: false, error: "Parse error"});
								}
							} else {
								resolve({dn: dn, data: null, success: false, error: "Http " + xhr.status});
							}
						}
					};
					xhr.send();
				});
			});

			Promise.all(promises).then(function(results) {
				window["%s"] = JSON.stringify(results);
			});
		})()
	`, resVar, dnsJson, signalIdsStr, token, resVar)

	// Execute JS initiation
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, nil)); err != nil {
		return nil, fmt.Errorf("batch string script error: %v", err)
	}

	// Poll for result
	var resultStr string
	for i := 0; i < 40; i++ { // Wait up to 20s
		time.Sleep(500 * time.Millisecond)
		err := chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf(`window["%s"]`, resVar), &resultStr))
		if err == nil && resultStr != "" && resultStr != "null" {
			break
		}
	}

	if resultStr == "" || resultStr == "null" {
		// Just return empty if timeout, maybe partial fail
		return nil, fmt.Errorf("batch string fetch timeout")
	}

	var results []BatchResult
	if err := json.Unmarshal([]byte(resultStr), &results); err != nil {
		return nil, fmt.Errorf("batch parse error: %v", err)
	}

	return results, nil
}
