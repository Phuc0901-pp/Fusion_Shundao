package formatter

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

// OrderedDataMap is a map that marshals with specific key ordering
type OrderedDataMap map[string]interface{}

// FusionFormattedData represents the standard output format for all devices
type FusionFormattedData struct {
	Timestamp  int64          `json:"timestamp"`
	DeviceName string         `json:"device_name"`
	DeviceID   string         `json:"device_id,omitempty"`
	Data       OrderedDataMap `json:"data"`
}

// StationFormattedData represents the station overview data
type StationFormattedData struct {
	Timestamp   int64          `json:"timestamp"`
	SiteName    string         `json:"sitename"`
	SiteID      string         `json:"siteid,omitempty"`
	Measurement string         `json:"measurement"`
	Fields      OrderedDataMap `json:"fields"`
}

// UnifiedInverterData represents the combined inverter data format
type UnifiedInverterData struct {
	Timestamp   int64          `json:"timestamp"`
	SiteName    string         `json:"sitename"`
	SiteID      string         `json:"siteid"`
	Name        string         `json:"name"`
	ID          string         `json:"id"`
	Model       string         `json:"model"`
	SN          string         `json:"sn"`
	Measurement string         `json:"measurement"`
	Fields      OrderedDataMap `json:"fields"`
}

// UnifiedSensorData represents the combined sensor data format
type UnifiedSensorData struct {
	Timestamp   int64          `json:"timestamp"`
	SiteName    string         `json:"sitename"`
	SiteID      string         `json:"siteid"`
	Name        string         `json:"name"`
	ID          string         `json:"id"`
	Model       string         `json:"model"`
	SN          string         `json:"sn"`
	Measurement string         `json:"measurement"` // "sensor"
	Fields      OrderedDataMap `json:"fields"`
}

// UnifiedPowerMeterData represents the combined power meter data format
type UnifiedPowerMeterData struct {
	Timestamp   int64          `json:"timestamp"`
	SiteName    string         `json:"sitename"`
	SiteID      string         `json:"siteid"`
	Name        string         `json:"name"`
	ID          string         `json:"id"`
	Model       string         `json:"model"`
	SN          string         `json:"sn"`
	Measurement string         `json:"measurement"` // "zonemeter"
	Fields      OrderedDataMap `json:"fields"`
}

// Implement MarshalJSON to sort keys: pvXX_status < pvXX_voltage < pvXX_current
func (m OrderedDataMap) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		ki, kj := keys[i], keys[j]

		// Check if both are PV keys (start with pv followed by digit)
		// Simple check: starts with "pv"
		isPVi := strings.HasPrefix(ki, "pv")
		isPVj := strings.HasPrefix(kj, "pv")

		// Case 1: One is PV, one is not. Non-PV comes first.
		if isPVi != isPVj {
			return !isPVi // if i is PV (true), i should be > j (false) -> return false. If i not PV (false), return true.
		}

		// Case 2: Both are PV keys
		if isPVi && isPVj {
			// Split into parts: pvXX and suffix
			partsI := strings.SplitN(ki, "_", 2)
			partsJ := strings.SplitN(kj, "_", 2)

			if len(partsI) == 2 && len(partsJ) == 2 {
				prefixI, suffixI := partsI[0], partsI[1]
				prefixJ, suffixJ := partsJ[0], partsJ[1]

				// Compare prefix (pv01 vs pv02)
				if prefixI != prefixJ {
					return prefixI < prefixJ
				}

				// Same prefix, compare suffix with custom order
				// Order: status < voltage < current
				weight := func(s string) int {
					switch s {
					case "status":
						return 1
					case "voltage", "volt_v":
						return 2
					case "current", "amp_a":
						return 3
					default:
						return 4 // Other suffixes sorted alphabetically later
					}
				}

				wI, wJ := weight(suffixI), weight(suffixJ)
				if wI != wJ {
					return wI < wJ
				}
				return suffixI < suffixJ
			}
		}

		// Case 3: Both are non-PV keys (or malformed PV keys) -> Alphabetical
		return ki < kj
	})

	// Build JSON object manually
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		// Marshal key
		keyBytes, _ := json.Marshal(k)
		buf.Write(keyBytes)
		buf.WriteByte(':')
		// Marshal value
		valBytes, err := json.Marshal(m[k])
		if err != nil {
			return nil, err
		}
		buf.Write(valBytes)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
