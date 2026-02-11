package ui

import (
	"fmt"
	"strings"
	"time"
)

// generateDeviceAlerts generates alerts based on device status
func generateDeviceAlerts(sites []SiteNode, now time.Time) []AlertMessage {
	var alerts []AlertMessage
	ts := now.UnixMilli()
	hour := now.Hour()

	for _, site := range sites {
		for _, logger := range site.Loggers {
			for _, inverter := range logger.Inverters {
				status := strings.ToLower(inverter.DeviceStatus)
				
				hasStrings := len(inverter.Strings) > 0
				hasVoltage := false
				hasCurrent := false
				
				for _, s := range inverter.Strings {
					if s.Voltage > 0 {
						hasVoltage = true
					}
					if s.Current > 0 {
						hasCurrent = true
					}
				}

				var shouldAlert bool
				var msg string
				var level string = "warning"

				if !hasStrings || (!hasCurrent && !hasVoltage) {
					// Case 1: No strings data or V=0, I=0 -> Not Working
					shouldAlert = true
					msg = "Inverter không hoạt động"
					if status == "fault" || strings.Contains(status, "error") {
						level = "error"
					}
				} else if !hasCurrent && hasVoltage {
					// Case 2: Has V, No I -> No Current
					shouldAlert = true
					msg = "Inverter không có dòng"
				} else if status != "grid connected" && status != "" {
					// Case 3: Status is bad
					shouldAlert = true
					if inverter.DeviceStatus != "" {
						msg = fmt.Sprintf("Trạng thái: %s", inverter.DeviceStatus)
					} else {
						msg = "Trạng thái: Không xác định"
					}
					
					if status == "fault" || strings.Contains(status, "error") {
						level = "error"
					}
				}

				if shouldAlert {
					alerts = append(alerts, AlertMessage{
						ID:        fmt.Sprintf("inv-%s-%d", inverter.ID, ts),
						Timestamp: ts,
						Level:     level,
						Message:   msg,
						Source:    fmt.Sprintf("%s - %s", logger.Name, inverter.Name),
					})
				}

				// Check for zero power when should be producing (6am - 6pm)
				if hour >= 6 && hour <= 18 {
					if inverter.POutKw == 0 && status == "grid connected" {
						alerts = append(alerts, AlertMessage{
							ID:        fmt.Sprintf("inv-nopower-%s-%d", inverter.ID, ts),
							Timestamp: ts,
							Level:     "warning",
							Message:   "Công suất đầu ra = 0 kW trong giờ làm việc",
							Source:    fmt.Sprintf("%s - %s", logger.Name, inverter.Name),
						})
					}
				}
			}
		}
	}
	
	// Sort by level (error=0, warning=1, info=2)
	// Stability sort
	// Using generic sort not easy in Go < 1.21 without slices package, but I can implement simple logic if needed.
	// Or just leave unsorted, frontend can sort. But backend ideally provides sorted data.
	
	return alerts
}
