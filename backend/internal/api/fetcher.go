package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"fusion/internal/platform/config"
	"fusion/internal/platform/utils"
)

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

// GetRoarand returns the captured Roarand token
func (f *Fetcher) GetRoarand() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.roarand
}

// HasValidToken checks if the fetcher has a token and verifies it
// It returns true if token exists and API returns valid JSON
// It returns false if token is missing, expired, or server returns HTML (login page)
func (f *Fetcher) HasValidToken(ctx context.Context) bool {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return false
	}

	// Try a lightweight API validation
	// Return both status and first 50 chars of response body
	js := fmt.Sprintf(`
		(function() {
			var xhr = new XMLHttpRequest();
			var url = 'https://intl.fusionsolar.huawei.com/rest/pvms/web/station/v1/overview/station-kpi-data';
			url += '?stationDn=NE=00000000'; 
			url += '&_=' + Date.now();
			
			xhr.open('GET', url, false);
			xhr.setRequestHeader('Accept', 'application/json');
			xhr.setRequestHeader('Roarand', '%s');
			
			xhr.send();
			
			var bodyStart = (xhr.responseText || '').substring(0, 50).trim();
			return JSON.stringify({status: xhr.status, body: bodyStart});
		})()
	`, token)

	var result string
	err := chromedp.Run(ctx, chromedp.Evaluate(js, &result))
	if err != nil {
		return false
	}

	// Parse result
	var resp struct {
		Status int    `json:"status"`
		Body   string `json:"body"`
	}
	if json.Unmarshal([]byte(result), &resp) != nil {
		return false
	}

	// Check HTTP status
	if resp.Status == 401 || resp.Status == 403 {
		return false
	}

	// CRITICAL: Check if response body starts with HTML (login page = session expired)
	// Even with 200 OK, if body starts with "<" it means we got HTML (login redirect)
	bodyStart := strings.TrimSpace(resp.Body)
	if strings.HasPrefix(bodyStart, "<") || strings.HasPrefix(strings.ToLower(bodyStart), "<!doctype") {
		// Server returned HTML instead of JSON = session expired
		return false
	}

	// Valid if body starts with JSON object or array
	if strings.HasPrefix(bodyStart, "{") || strings.HasPrefix(bodyStart, "[") {
		return true
	}

	// Unknown response format, treat as invalid to be safe
	return false
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
						utils.LogInfo(">>> Bắt được Roarand từ request: %s...", f.roarand[:min(30, len(f.roarand))])
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
						utils.LogInfo(">>> Bắt được Roarand từ response: %s...", f.roarand[:min(30, len(f.roarand))])
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
		utils.LogInfo("[Chờ] Đang load trang (15 giây)...")
		time.Sleep(15 * time.Second)

		// Check if token captured
		f.mu.Lock()
		if f.roarand == "" {
			f.mu.Unlock()
			utils.LogWarn("⚠️ Chưa bắt được Token. Thử reload trang...")
			if err := chromedp.Reload().Do(ctx); err != nil {
				utils.LogError("Lỗi reload: %v", err)
			}
			time.Sleep(10 * time.Second)
		} else {
			f.mu.Unlock()
		}

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
				utils.LogInfo(">>> Lấy được Roarand từ browser: %s...", token[:min(30, len(token))])
			} else {
				utils.LogWarn(">>> Không tìm thấy Roarand token trong browser")
			}
		}

		// Try to get response from captured requests
		f.mu.Lock()
		ids := make([]network.RequestID, len(f.requestIDs))
		copy(ids, f.requestIDs)
		f.mu.Unlock()

		utils.LogInfo("      Số API locate-tree đã phát hiện: %d", len(ids))

		for i, reqID := range ids {
			utils.LogDebug("      Thử lấy body từ request %d...", i+1)

			body, err := network.GetResponseBody(reqID).Do(ctx)
			if err != nil {
				continue
			}

			if len(body) > 0 && body[0] == '{' {
				siteData = string(body)
				utils.LogInfo("      ✓ Lấy được response! Length: %d bytes", len(body))
				break
			}
		}

		// If no data, try calling API directly with Roarand
		if siteData == "" {
			f.mu.Lock()
			token := f.roarand
			f.mu.Unlock()

			if token != "" {
				utils.LogInfo("      Thử gọi API trực tiếp...")
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
