package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// BatchResult represents the result for a single device in a batch
type BatchResult struct {
	Dn      string                 `json:"dn"`
	Data    map[string]interface{} `json:"data"`
	Success bool                   `json:"success"`
	Error   string                 `json:"error"`
}

// FetchBatchRealtimeData fetches realtime data for multiple devices in parallel.
// Uses native JS async/await inside chromedp.Evaluate so Go does NOT need a
// poll loop – the Goroutine simply waits for the JS Promise to settle.
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

	// === ASYNC JS – chromedp awaits the promise natively, no Go polling needed ===
	js := fmt.Sprintf(`
		(async function() {
			var deviceDns = %s;
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/device/v1/device-realtime-data';
			var extraParams = '%s';
			var token = '%s';

			var promises = deviceDns.map(function(dn) {
				return new Promise(function(resolve) {
					var xhr = new XMLHttpRequest();
					var fullUrl = url + '?deviceDn=' + encodeURIComponent(dn) + extraParams + '&_=' + Date.now();
					xhr.open('GET', fullUrl, true);
					xhr.setRequestHeader('Accept', 'application/json');
					xhr.setRequestHeader('X-Timezone-Offset', '420');
					xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
					xhr.setRequestHeader('Roarand', token);
					xhr.onreadystatechange = function() {
						if (xhr.readyState === 4) {
							if (xhr.status === 200) {
								try {
									resolve({dn: dn, data: JSON.parse(xhr.responseText), success: true, error: ""});
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

			var results = await Promise.all(promises);
			return JSON.stringify(results);
		})()
	`, dnsJson, extraParams, token)

	var resultStr string
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &resultStr, func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) })); err != nil {
		return nil, fmt.Errorf("batch script error: %v", err)
	}

	if resultStr == "" || resultStr == "null" {
		return nil, fmt.Errorf("batch fetch returned empty result")
	}

	var results []BatchResult
	if err := json.Unmarshal([]byte(resultStr), &results); err != nil {
		return nil, fmt.Errorf("batch parse error: %v", err)
	}

	return results, nil
}

// FetchBatchInverterStringData fetches string data for multiple inverters in parallel.
// Uses native JS async/await – no Go poll loop.
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

	signalIdsStr := ""
	for i, id := range StringDataSignalIds {
		if i > 0 {
			signalIdsStr += "&signalIds=" + fmt.Sprint(id)
		} else {
			signalIdsStr += fmt.Sprint(id)
		}
	}

	js := fmt.Sprintf(`
		(async function() {
			var deviceDns = %s;
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/device/v1/device-real-kpi';
			var signalParams = '%s';
			var token = '%s';

			var promises = deviceDns.map(function(dn) {
				return new Promise(function(resolve) {
					var xhr = new XMLHttpRequest();
					var fullUrl = url + '?signalIds=' + signalParams + '&deviceDn=' + encodeURIComponent(dn) + '&_=' + Date.now();
					xhr.open('GET', fullUrl, true);
					xhr.setRequestHeader('Accept', 'application/json');
					xhr.setRequestHeader('X-Timezone-Offset', '420');
					xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
					xhr.setRequestHeader('Roarand', token);
					xhr.onreadystatechange = function() {
						if (xhr.readyState === 4) {
							if (xhr.status === 200) {
								try {
									resolve({dn: dn, data: JSON.parse(xhr.responseText), success: true, error: ""});
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

			var results = await Promise.all(promises);
			return JSON.stringify(results);
		})()
	`, dnsJson, signalIdsStr, token)

	var resultStr string
	if err := chromedp.Run(ctx, chromedp.Evaluate(js, &resultStr, func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) })); err != nil {
		return nil, fmt.Errorf("batch string script error: %v", err)
	}

	if resultStr == "" || resultStr == "null" {
		return nil, fmt.Errorf("batch string fetch returned empty result")
	}

	var results []BatchResult
	if err := json.Unmarshal([]byte(resultStr), &results); err != nil {
		return nil, fmt.Errorf("batch parse error: %v", err)
	}

	return results, nil
}

// keepTime keeps the time import alive if needed elsewhere
var _ = time.Now
