package ui

import (
	"testing"
	"time"
)

func TestGenerateDeviceAlerts(t *testing.T) {
	// Fixed time: 12:00 PM (Daytime)
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		inverter InverterNode
		wantAlert bool
		wantMsg  string
	}{
		{
			name: "Normal Operation",
			inverter: InverterNode{
				ID:           "inv1",
				DeviceStatus: "Grid connected",
				POutKw:       10,
				Strings: []StringData{
					{Current: 1, Voltage: 200},
				},
			},
			wantAlert: false,
		},
		{
			name: "Inverter Fault",
			inverter: InverterNode{
				ID:           "inv2",
				DeviceStatus: "Fault",
				Strings: []StringData{
					{Current: 0, Voltage: 0},
				},
			},
			wantAlert: true,
			wantMsg:   "Inverter không hoạt động",
		},
		{
			name: "No Current (Has Voltage)",
			inverter: InverterNode{
				ID:           "inv3",
				DeviceStatus: "Grid connected", // Stat is fine but logic says no current
				Strings: []StringData{
					{Current: 0, Voltage: 200},
				},
			},
			wantAlert: true,
			wantMsg:   "Inverter không có dòng",
		},
		{
			name: "Zero Power at Daytime",
			inverter: InverterNode{
				ID:           "inv4",
				DeviceStatus: "Grid connected",
				POutKw:       0,
				Strings: []StringData{
					{Current: 1, Voltage: 200},
				},
			},
			wantAlert: true,
			wantMsg:   "Công suất đầu ra = 0 kW trong giờ làm việc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sites := []SiteNode{
				{
					Loggers: []LoggerNode{
						{
							Inverters: []InverterNode{tt.inverter},
						},
					},
				},
			}

			alerts := generateDeviceAlerts(sites, fixedTime)

			if tt.wantAlert {
				if len(alerts) == 0 {
					t.Errorf("Expected alert, got none")
				} else {
					// Check message
					found := false
					for _, a := range alerts {
						if a.Message == tt.wantMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected alert message '%s', got %v", tt.wantMsg, alerts)
					}
				}
			} else {
				if len(alerts) > 0 {
					t.Errorf("Expected no alerts, got %v", alerts)
				}
			}
		})
	}
}

func TestGenerateDeviceAlerts_NightMode(t *testing.T) {
	// Fixed time: 22:00 PM (Night)
	fixedTime := time.Date(2023, 1, 1, 22, 0, 0, 0, time.UTC)

	// Inverter with current+voltage but 0 POut at night
	// The only time-sensitive alert is "Zero Power during work hours"
	// At night, this alert should NOT fire.
	inverterActive := InverterNode{
		ID:           "inv-night-3",
		DeviceStatus: "Grid connected",
		POutKw:       0,
		Strings: []StringData{
			{Current: 1, Voltage: 200},
		},
	}

	sites := []SiteNode{
		{
			Loggers: []LoggerNode{
				{
					Inverters: []InverterNode{inverterActive},
				},
			},
		},
	}

	alerts := generateDeviceAlerts(sites, fixedTime)

	// Should NOT have "Công suất đầu ra = 0 kW" alert because it's night
	for _, a := range alerts {
		if a.Message == "Công suất đầu ra = 0 kW trong giờ làm việc" {
			t.Error("Should not alert zero power at night")
		}
	}
}
