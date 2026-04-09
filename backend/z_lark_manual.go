//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"fusion/internal/platform/config"
)

func getLarkToken() (string, error) {
	appID := config.App.System.LarkAppID
	appSecret := config.App.System.LarkAppSecret
	payload := map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(
		"https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Token string `json:"tenant_access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Token, nil
}

func main() {
	config.LoadConfig()
	token, _ := getLarkToken()

	appToken := config.App.System.LarkBitableAppToken
	tableID := config.App.System.LarkBitableTableID

	url := fmt.Sprintf("https://open.larksuite.com/open-apis/bitable/v1/apps/%s/tables/%s/records/search?page_size=10", appToken, tableID)
	
	reqBody := map[string]interface{}{
		"sort": []map[string]interface{}{
			{
				"field_name": "Created At",
				"desc":       false,
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	
	bytesOut, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(bytesOut))
}
