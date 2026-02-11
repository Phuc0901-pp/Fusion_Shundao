package api_test

import (
	"fusion/internal/api"
	"fusion/internal/platform/config"
	"strings"
	"testing"
)

func TestNewFetcher(t *testing.T) {
	f := api.NewFetcher()
	if f == nil {
		t.Fatal("NewFetcher returned nil")
	}
}

func TestFilterSites(t *testing.T) {
	// Setup mock config
	config.App.Sites = []config.SiteConfig{
		{ID: "site-1", Name: "Site 1"},
	}

	// Sample JSON input based on SiteTree structure
	inputJSON := `{
		"childList": [
			{
				"elementDn": "site-1",
				"nodeName": "Site 1",
				"status": "normal",
				"childList": []
			},
			{
				"elementDn": "site-2",
				"nodeName": "Site 2",
				"status": "alarm",
				"childList": []
			}
		],
		"hasMoreChild": false
	}`

	output, err := api.FilterSites(inputJSON)
	if err != nil {
		t.Fatalf("FilterSites failed: %v", err)
	}

	if !strings.Contains(output, "site-1") {
		t.Error("Output should contain site-1")
	}
	if strings.Contains(output, "site-2") {
		t.Error("Output should NOT contain site-2")
	}
}
