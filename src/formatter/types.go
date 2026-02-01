package formatter

// FusionFormattedData represents the standard output format for all devices
type FusionFormattedData struct {
	Timestamp   int64                  `json:"timestamp"`
	DeviceName  string                 `json:"device_name"`
	DeviceID    string                 `json:"device_id,omitempty"`
	Data        map[string]interface{} `json:"data"`
}

// StationFormattedData represents the station overview data
type StationFormattedData struct {
	Timestamp   int64                  `json:"timestamp"`
	SiteName    string                 `json:"sitename"`
	SiteID      string                 `json:"siteid,omitempty"`
	Data        map[string]interface{} `json:"data"`
}
