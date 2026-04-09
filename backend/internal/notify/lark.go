package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fusion/internal/platform/config"
	"fusion/internal/platform/utils"
)

// getLarkToken obtains a valid tenant_access_token from Lark Open API.
func getLarkToken() (string, error) {
	appID := config.App.System.LarkAppID
	appSecret := config.App.System.LarkAppSecret
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("[LARK] App ID or App Secret not configured")
	}

	payload := map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		"https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", fmt.Errorf("[LARK] Failed to get token: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code   int    `json:"code"`
		Msg    string `json:"msg"`
		Token  string `json:"tenant_access_token"`
		Expire int    `json:"expire"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("[LARK] Failed to decode token response: %w", err)
	}
	if result.Code != 0 {
		return "", fmt.Errorf("[LARK] Token API error (code=%d): %s", result.Code, result.Msg)
	}
	return result.Token, nil
}

// AlertRecord represents a single fault alert to be logged into Lark Bitable.
type AlertRecord struct {
	Message     string // Alert code / description
	Inverter    string // Inverter name
	SmartLogger string // SmartLogger name
	Site        string // Site name
}

// SendLarkBitableAlert writes one alert record as a new row into the configured Lark Bitable table.
func SendLarkBitableAlert(record AlertRecord) error {
	appToken := config.App.System.LarkBitableAppToken
	tableID := config.App.System.LarkBitableTableID
	if appToken == "" || tableID == "" {
		return fmt.Errorf("[LARK] Bitable App Token or Table ID not configured")
	}

	token, err := getLarkToken()
	if err != nil {
		return err
	}

	now := time.Now()
	// Lark Bitable datetime fields expect Unix milliseconds (int64)
	createdAtMs := now.UnixMilli()

	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"Message":     record.Message,
			"Inverter":    record.Inverter,
			"SmartLogger": record.SmartLogger,
			"Site":        record.Site,
			"Created At":  createdAtMs,
		},
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf(
		"https://open.larksuite.com/open-apis/bitable/v1/apps/%s/tables/%s/records",
		appToken, tableID,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("[LARK] Failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("[LARK] HTTP POST failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("[LARK] Failed to decode response: %w", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("[LARK] Bitable API error (code=%d): %s", result.Code, result.Msg)
	}

	utils.LogInfo("[LARK] ✅ Alert recorded → Site=%s | Logger=%s | Inverter=%s | %s",
		record.Site, record.SmartLogger, record.Inverter, record.Message)
	return nil
}

// PruneLarkBaseIfNeeded checks if the Bitable table exceeds maxRecords.
// If it does, it scans for and batch-deletes all records older than daysToKeep.
func PruneLarkBaseIfNeeded(maxRecords int, daysToKeep int) {
	appToken := config.App.System.LarkBitableAppToken
	tableID := config.App.System.LarkBitableTableID
	if appToken == "" || tableID == "" {
		return
	}

	token, err := getLarkToken()
	if err != nil {
		utils.LogError("[LARK-PRUNE] Failed to get token: %v", err)
		return
	}

	// 1. Check total records
	searchUrl := fmt.Sprintf("https://open.larksuite.com/open-apis/bitable/v1/apps/%s/tables/%s/records/search", appToken, tableID)
	reqCheck, _ := http.NewRequest("POST", searchUrl+"?page_size=1", bytes.NewReader([]byte(`{}`)))
	reqCheck.Header.Set("Authorization", "Bearer "+token)
	reqCheck.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	respCheck, err := client.Do(reqCheck)
	if err != nil {
		utils.LogError("[LARK-PRUNE] HTTP error getting total: %v", err)
		return
	}

	var resCheck struct {
		Code int `json:"code"`
		Data struct {
			Total int `json:"total"`
		} `json:"data"`
	}
	json.NewDecoder(respCheck.Body).Decode(&resCheck)
	respCheck.Body.Close()

	if resCheck.Code != 0 || resCheck.Data.Total <= maxRecords {
		return // Do nothing if under limit
	}

	utils.LogWarn("[LARK-PRUNE] Triggered! Total records (%d) exceeds limit (%d). Deleting data older than %d days...", 
		resCheck.Data.Total, maxRecords, daysToKeep)

	cutoffMs := time.Now().Add(-time.Duration(daysToKeep) * 24 * time.Hour).UnixMilli()
	var toDelete []string

	// 2. Fetch sorted older records
	var pageToken string
	hasMore := true

	for hasMore {
		reqBody := map[string]interface{}{
			"sort": []map[string]interface{}{
				{"field_name": "Created At", "desc": false},
			},
		}
		if pageToken != "" {
			reqBody["page_token"] = pageToken
		}
		bodyBytes, _ := json.Marshal(reqBody)
		reqFetch, _ := http.NewRequest("POST", searchUrl+"?page_size=500", bytes.NewReader(bodyBytes))
		reqFetch.Header.Set("Authorization", "Bearer "+token)
		reqFetch.Header.Set("Content-Type", "application/json")

		respFetch, err := client.Do(reqFetch)
		if err != nil {
			utils.LogError("[LARK-PRUNE] HTTP error fetching records: %v", err)
			break
		}

		var resFetch struct {
			Code int `json:"code"`
			Data struct {
				HasMore   bool   `json:"has_more"`
				PageToken string `json:"page_token"`
				Items     []struct {
					RecordID string                 `json:"record_id"`
					Fields   map[string]interface{} `json:"fields"`
				} `json:"items"`
			} `json:"data"`
		}
		if err := json.NewDecoder(respFetch.Body).Decode(&resFetch); err != nil {
			respFetch.Body.Close()
			break
		}
		respFetch.Body.Close()

		if resFetch.Code != 0 {
			break
		}

		stopFetching := false
		for _, item := range resFetch.Data.Items {
			// Created At is returned as float64 by encoding/json unmarshaling into interface{}
			createdAtField, ok := item.Fields["Created At"]
			if !ok {
				continue
			}
			createdAtFloat, ok := createdAtField.(float64)
			if ok && int64(createdAtFloat) < cutoffMs {
				toDelete = append(toDelete, item.RecordID)
			} else {
				// Record is newer than cutoff! Since it's sorted ascending, all subsequent records are even newer.
				stopFetching = true
				break
			}
		}

		if stopFetching || !resFetch.Data.HasMore {
			break
		}
		pageToken = resFetch.Data.PageToken
		time.Sleep(100 * time.Millisecond) // Be kind to Rate Limits
	}

	// 3. Batch delete in chunks of 500
	if len(toDelete) == 0 {
		return
	}

	deleteUrl := fmt.Sprintf("https://open.larksuite.com/open-apis/bitable/v1/apps/%s/tables/%s/records/batch_delete", appToken, tableID)
	
	chunks := 0
	for i := 0; i < len(toDelete); i += 500 {
		end := i + 500
		if end > len(toDelete) {
			end = len(toDelete)
		}
		batch := toDelete[i:end]

		payload := map[string]interface{}{
			"records": batch,
		}
		bodyBytes, _ := json.Marshal(payload)
		
		reqDel, _ := http.NewRequest("POST", deleteUrl, bytes.NewReader(bodyBytes))
		reqDel.Header.Set("Authorization", "Bearer "+token)
		reqDel.Header.Set("Content-Type", "application/json")

		respDel, err := client.Do(reqDel)
		if err == nil {
			respDel.Body.Close()
			chunks++
		}
		time.Sleep(200 * time.Millisecond) // Rate Limit
	}

	utils.LogInfo("[LARK-PRUNE] ✅ Completed. Deleted %d old records across %d batches.", len(toDelete), chunks)
}
