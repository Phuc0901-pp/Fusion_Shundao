package api

import (
	"context"
	"encoding/json"
	"fmt"

	"fusion/internal/platform/config"
)

// SiteNode represents a node in the site tree (full data)
type SiteNode struct {
	ChildList    []SiteNode `json:"childList"`
	ElementDn    string     `json:"elementDn"`
	ElementId    string     `json:"elementId"`
	HasMoreChild bool       `json:"hasMoreChild"`
	IsParent     bool       `json:"isParent"`
	MocId        int        `json:"mocId"`
	NodeIcon     string     `json:"nodeIcon"`
	NodeName     string     `json:"nodeName"`
	ParentDn     string     `json:"parentDn"`
	Status       string     `json:"status"`
	TypeId       int        `json:"typeId"`
}

// SiteTree represents the root structure (full data)
type SiteTree struct {
	ChildList    []SiteNode `json:"childList"`
	HasMoreChild bool       `json:"hasMoreChild"`
}

// SimpleSite represents simplified site data for output
type SimpleSite struct {
	ElementDn string `json:"elementDn"`
	NodeName  string `json:"nodeName"`
	ParentDn  string `json:"parentDn"`
	Status    string `json:"status"`
}

// SimpleOutput represents the simplified output structure
type SimpleOutput struct {
	ChildList []SimpleSite `json:"childList"`
}

// FilterSites filters the site data to only include target sites with simplified fields
func FilterSites(jsonData string) (string, error) {
	var tree SiteTree
	if err := json.Unmarshal([]byte(jsonData), &tree); err != nil {
		return "", err
	}

	// Get target site IDs from config
	targetIDs := make(map[string]bool)
	for _, s := range config.App.Sites {
		targetIDs[s.ID] = true
	}

	// Filter to only target sites
	var filtered []SiteNode
	filterNodes(tree.ChildList, targetIDs, &filtered)

	// Convert to simplified output
	var simplified []SimpleSite
	for _, node := range filtered {
		simplified = append(simplified, SimpleSite{
			ElementDn: node.ElementDn,
			NodeName:  node.NodeName,
			ParentDn:  node.ParentDn,
			Status:    node.Status,
		})
	}

	result := SimpleOutput{
		ChildList: simplified,
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// filterNodes recursively searches for target sites
func filterNodes(nodes []SiteNode, targetIDs map[string]bool, result *[]SiteNode) {
	for _, node := range nodes {
		if targetIDs[node.ElementDn] {
			*result = append(*result, node)
		}
		// Also search in children
		if len(node.ChildList) > 0 {
			filterNodes(node.ChildList, targetIDs, result)
		}
	}
}

// FetchSiteData fetches data for a specific site
func (f *Fetcher) FetchSiteData(ctx context.Context, siteID string) (string, error) {
	f.mu.Lock()
	token := f.roarand
	f.mu.Unlock()

	if token == "" {
		return "", fmt.Errorf("no Roarand token available")
	}

	return f.callAPIDirectly(ctx, token, siteID), nil
}
