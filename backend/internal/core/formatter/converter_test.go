package formatter

import (
	"fusion/internal/platform/config"
	"testing"
)

func TestInitMapper(t *testing.T) {
	// Mock config
	config.Signals.Inverter = map[string]string{
		"100": "active_power",
	}

	InitMapper()

	if InverterSignalMap["100"] != "active_power" {
		t.Errorf("InitMapper failed to populate InverterSignalMap")
	}
}

func TestGetStringPVField(t *testing.T) {
	result := GetStringPVField(1, "voltage")
	expected := "pv01_voltage"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestGetUnifiedPVField(t *testing.T) {
	result := GetUnifiedPVField(5, "volt_v")
	expected := "pv05_volt_v"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
