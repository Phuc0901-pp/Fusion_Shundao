package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"fusion/internal/platform/utils"
)

type SiteMetadata struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StationID string `json:"station_id"` // Original Station Code
}

func saveSiteMetadata(siteName string, stationID string, dbID string) {
	siteDisplay := strings.ReplaceAll(siteName, " ", "_")
	rootDir := "output"
	fullDir := filepath.Join(rootDir, siteDisplay)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		utils.LogError("Failed to create site dir: %v", err)
		return
	}

	meta := SiteMetadata{
		ID:        dbID,
		Name:      siteName,
		StationID: stationID,
	}

	filePath := filepath.Join(fullDir, "site_meta.json")
	bytes, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filePath, bytes, 0644); err != nil {
		utils.LogError("Failed to save site metadata: %v", err)
	}
}
