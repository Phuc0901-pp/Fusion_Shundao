package victoriametrics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds VictoriaMetrics configuration
type Config struct {
	Endpoint string
}

// Client is the VictoriaMetrics client
type Client struct {
	Config     Config
	HTTPClient *http.Client
}

// NewClient creates a new VictoriaMetrics client
func NewClient(endpoint string) *Client {
	return &Client{
		Config: Config{Endpoint: endpoint},
		HTTPClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

// GenericData represents any device data from JSON
type GenericData struct {
	Timestamp   int64                  `json:"timestamp"`
	SiteName    string                 `json:"sitename"`
	SiteID      string                 `json:"siteid"`
	Name        string                 `json:"name"`
	ID          string                 `json:"id"`
	Model       string                 `json:"model"`
	SN          string                 `json:"sn"`
	Measurement string                 `json:"measurement"`
	Fields      map[string]interface{} `json:"fields"`
}

// PlantData represents plant overview data
type PlantData struct {
	Timestamp   int64                  `json:"timestamp"`
	SiteName    string                 `json:"sitename"`
	SiteID      string                 `json:"siteid"`
	Measurement string                 `json:"measurement"`
	Fields      map[string]interface{} `json:"fields"`
}

// PushMetrics pushes Prometheus format data to VictoriaMetrics
func (c *Client) PushMetrics(data string) error {
	url := c.Config.Endpoint + "/api/v1/import/prometheus"
	resp, err := c.HTTPClient.Post(url, "text/plain", strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to push metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("push failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// DeleteMetrics deletes metrics matching the given pattern from VictoriaMetrics
func (c *Client) DeleteMetrics(matchPattern string) error {
	url := fmt.Sprintf("%s/api/v1/admin/tsdb/delete_series?match[]=%s",
		c.Config.Endpoint, matchPattern)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Deleted old shundao_* metrics")
	return nil
}

// ConvertToPrometheus converts a JSON data file to Prometheus format
func ConvertToPrometheus(jsonPath string) (string, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return "", err
	}

	// Extract station from path
	station := extractStationFromPath(jsonPath)

	// Try to detect type by checking measurement field
	var generic GenericData
	if err := json.Unmarshal(data, &generic); err != nil {
		return "", err
	}

	var lines []string

	switch generic.Measurement {
	case "plant":
		var plant PlantData
		json.Unmarshal(data, &plant)
		lines = convertPlantMetrics(plant)
	case "inverter":
		lines = convertDeviceMetrics("shundao_inverter", generic, station)
	case "zonemeter":
		lines = convertDeviceMetrics("shundao_zonemeter", generic, station)
	case "sensor":
		lines = convertDeviceMetrics("shundao_sensor", generic, station)
	default:
		lines = convertDeviceMetrics("shundao_"+generic.Measurement, generic, station)
	}

	return strings.Join(lines, "\n"), nil
}

// extractStationFromPath extracts station name from file path
func extractStationFromPath(path string) string {
	path = filepath.ToSlash(path)
	parts := strings.Split(path, "/")

	if len(parts) >= 4 {
		// For data.json: output/SITE/STATION/DEVICE/data.json
		if parts[len(parts)-1] == "data.json" {
			stationIdx := len(parts) - 3
			station := parts[stationIdx]
			if strings.Contains(station, "Smartlogger") || strings.Contains(station, "Station") {
				return station
			}
			// Try one level up
			if len(parts) >= 5 {
				return parts[len(parts)-4]
			}
		}
	}
	return "unknown"
}

func convertPlantMetrics(data PlantData) []string {
	var lines []string
	siteName := sanitizeLabel(data.SiteName)
	siteID := data.SiteID

	for fieldName, fieldValue := range data.Fields {
		val, ok := toFloat64(fieldValue)
		if !ok {
			continue
		}
		// Format: shundao_plant{...} value [timestamp]
		// Timestamp in ms from JSON (need to convert to ms string if VM expects it, or s if Prometheus)
		// VM import/prometheus expects timestamp in milliseconds? No, Prometheus text format usually doesn't have timestamp.
		// Wait, VM supports timestamp in import/prometheus?
		// "VictoriaMetrics accepts Prometheus text exposition format... It also accepts lines with timestamp: metric_name{labels} value timestamp"
		// The timestamp must be in milliseconds.
		metric := fmt.Sprintf("shundao_plant{site_name=\"%s\",site_id=\"%s\",name=\"%s\"} %v %d",
			siteName, siteID, sanitizeMetricName(fieldName), val, data.Timestamp)
		lines = append(lines, metric)
	}
	return lines
}

func convertDeviceMetrics(prefix string, data GenericData, station string) []string {
	var lines []string
	siteName := sanitizeLabel(data.SiteName)
	siteID := data.SiteID
	device := sanitizeLabel(data.Name)
	model := sanitizeLabel(data.Model)
	sn := data.SN

	for fieldName, fieldValue := range data.Fields {
		val, ok := toFloat64(fieldValue)
		if !ok {
			continue
		}
		// Label order: site_name, site_id, station, device, model, sn, name
		metric := fmt.Sprintf("%s{site_name=\"%s\",site_id=\"%s\",station=\"%s\",device=\"%s\",model=\"%s\",sn=\"%s\",name=\"%s\"} %v %d",
			prefix, siteName, siteID, station, device, model, sn, sanitizeMetricName(fieldName), val, data.Timestamp)
		lines = append(lines, metric)
	}
	return lines
}

// PushAllFromDirectory reads all data.json files from output directory and pushes to VM
func (c *Client) PushAllFromDirectory(outputDir string) error {
	var allMetrics []string
	var fileCount int
	batchSize := 50 // Push every 50 files to avoid timeout

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == "overview.json" || info.Name() == "data.json" {
			metrics, err := ConvertToPrometheus(path)
			if err != nil {
				fmt.Printf("Warning: failed to convert %s: %v\n", path, err)
				return nil
			}
			if metrics != "" {
				allMetrics = append(allMetrics, metrics)
				fileCount++

				// Check if batch is full
				if len(allMetrics) >= batchSize {
					payload := strings.Join(allMetrics, "\n")
					if err := c.PushMetrics(payload); err != nil {
						fmt.Printf("⚠️  Lỗi push batch (%d files): %v\n", len(allMetrics), err)
					} else {
						fmt.Print("↑") // Visual feedback for batch push
					}
					// Reset batch
					allMetrics = nil
					// Add small sleep to not overwhelm server
					time.Sleep(200 * time.Millisecond)
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Push remaining metrics
	if len(allMetrics) > 0 {
		payload := strings.Join(allMetrics, "\n")
		if err := c.PushMetrics(payload); err != nil {
			return err
		}
		fmt.Print("↑")
	}

	if fileCount == 0 {
		return fmt.Errorf("no metrics found in %s", outputDir)
	}

	fmt.Printf("\nSuccessfully pushed %d files to VictoriaMetrics\n", fileCount)
	return nil
}

// Helper functions

func sanitizeMetricName(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

func sanitizeLabel(s string) string {
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case string:
		if val == "-" || val == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

// PushToVictoriaMetrics is a convenience function to push all data from output directory
// to VictoriaMetrics. Uses default endpoint and output directory.
func PushToVictoriaMetrics() {
	endpoint := "http://100.118.142.45:8428"
	outputDir := "output"

	fmt.Println(">>> Đẩy dữ liệu lên VictoriaMetrics...")
	client := NewClient(endpoint)

	if err := client.PushAllFromDirectory(outputDir); err != nil {
		fmt.Printf("⚠️  Lỗi push VM: %v\n", err)
	} else {
		fmt.Println("✅ Push VictoriaMetrics thành công!")
	}
}
