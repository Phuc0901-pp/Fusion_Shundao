package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig represents the structure of app.json
type AppConfig struct {
	Browser   BrowserConfig   `json:"browser"`
	API       APIConfig       `json:"api"`
	Selectors SelectorConfig  `json:"selectors"`
	System    SystemConfig    `json:"system"`
	Database  DatabaseConfig  `json:"database"`
	Sites     []SiteConfig    `json:"sites"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type SiteConfig struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type BrowserConfig struct {
	HeadlessTimeout string `json:"headless_timeout"` // e.g., "168h"
	UserAgent       string `json:"user_agent"`
}

type APIConfig struct {
	BaseURL   string            `json:"base_url"`
	Endpoints map[string]string `json:"endpoints"`
	Params    APIParams         `json:"params"`
}

type APIParams struct {
	SubNodeTypeIDs string `json:"sub_node_type_ids"`
	TypeIDInclude  string `json:"type_id_include"`
}

type SelectorConfig struct {
	UsernameInput string `json:"username_input"`
	PasswordInput string `json:"password_input"`
	LoginButton   string `json:"login_button"`
	ErrorMessage  string `json:"error_message"`
}

type SystemConfig struct {
	VMEndpoint           string `json:"vm_endpoint"`
	BatchSizeInverter    int    `json:"batch_size_inverter"`
	BatchSizeMeter       int    `json:"batch_size_meter"`
	FetchIntervalMinutes int    `json:"fetch_interval_minutes"`
}

// SignalsConfig represents the structure of signals.json
type SignalsConfig struct {
	Inverter        map[string]string `json:"inverter"`
	UnifiedInverter map[string]string `json:"unified_inverter"`
	Meter           map[string]string `json:"meter"`
	Sensor          map[string]string `json:"sensor"`
	SmartLogger     map[string]string `json:"smart_logger"`
}

var (
	App     AppConfig
	Signals SignalsConfig
)

// LoadConfig loads both app and signals configurations
func LoadConfig() error {
	// 1. Load App Config
	if err := loadJSON("configs/app.json", &App); err != nil {
		return fmt.Errorf("failed to load app.json: %v", err)
	}

	// 2. Load Signals Config
	if err := loadJSON("configs/signals.json", &Signals); err != nil {
		return fmt.Errorf("failed to load signals.json: %v", err)
	}

	return nil
}

func loadJSON(path string, target interface{}) error {
	// Try to find file relative to current binary or source
	// For simplicity, we assume running from root or we check existence
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Try going up one level if we are in src/
		path = filepath.Join("..", path)
	}
	
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, target)
}
