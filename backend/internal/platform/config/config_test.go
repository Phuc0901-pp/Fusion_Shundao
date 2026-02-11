package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary directory for dummy config
	tmpDir, err := os.MkdirTemp("", "fusion_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// dummy app.json
	appConfig := AppConfig{
		System: SystemConfig{
			BatchSizeInverter: 10,
		},
	}
	appContent, _ := json.Marshal(appConfig)

	// dummy signals.json
	signalsConfig := SignalsConfig{
		Inverter: map[string]string{
			"1": "voltage",
		},
	}
	signalsContent, _ := json.Marshal(signalsConfig)

	// Write files
	configDir := filepath.Join(tmpDir, "configs")
	os.Mkdir(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "app.json"), appContent, 0644)
	os.WriteFile(filepath.Join(configDir, "signals.json"), signalsContent, 0644)

	// Change working directory to tmpDir so LoadConfig can find config/
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	// Run LoadConfig
	if err := LoadConfig(); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify App Config
	if App.System.BatchSizeInverter != 10 {
		t.Errorf("Expected BatchSizeInverter 10, got %d", App.System.BatchSizeInverter)
	}

	// Verify Signals Config
	if Signals.Inverter["1"] != "voltage" {
		t.Errorf("Expected Inverter signal '1' -> 'voltage', got %v", Signals.Inverter["1"])
	}
}
