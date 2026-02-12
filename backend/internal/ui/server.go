package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"fusion/internal/database"
	"fusion/internal/platform/utils"
)

// Cache Mechanism
var (
	apiCache      DashboardData
	cacheMutex    sync.RWMutex
	baseSites     []SiteNode
	baseSitesLock sync.RWMutex
)

// StartServer starts the API server
func StartServer() {
	// Start background data aggregator
	go startBackgroundUpdater()

	http.HandleFunc("/api/dashboard", handleDashboard)
	http.HandleFunc("/api/production-monthly", handleMonthlyProduction)
	http.HandleFunc("/api/rename", handleRename)
	http.HandleFunc("/api/inverter/dc-power", handleInverterDCPower)
	http.HandleFunc("/healthz", handleHealthz) // Logic from low priority tasks

	port := ":5039"
	fmt.Printf("[READY] Starting UI Backend Server on %s...\n", port)
	if err := http.ListenAndServe(port, corsMiddleware(http.DefaultServeMux)); err != nil {
		utils.LogError("[ERROR] Server failed: %v", err)
		os.Exit(1)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	cacheMutex.RLock()
	data := apiCache
	cacheMutex.RUnlock()

	// Disable Caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	bytes, _ := json.Marshal(data)
	w.Write(bytes)
}

func handleMonthlyProduction(w http.ResponseWriter, r *http.Request) {
	data := fetchMonthlyProductionData()

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	bytes, _ := json.Marshal(data)
	w.Write(bytes)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" && r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EntityType string `json:"entityType"` // "site", "logger", "device"
		ID         string `json:"id"`         // DB ID
		NewName    string `json:"newName"`
		StringSet  string `json:"stringSet"`  // Option 2
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.EntityType == "" || req.ID == "" {
		http.Error(w, "Missing required fields: entityType, id", http.StatusBadRequest)
		return
	}

	// 1. Update Name (if provided or empty string means reset)
	// For Option 2: rename might not be the only action. 
	// But current UI sends newName. If we want to update ONLY stringSet, frontend should send current name? 
	// Or we make UpdateNameChange optional?
	// Let's assume frontend sends everything.
	if err := database.UpdateNameChange(req.EntityType, req.ID, req.NewName); err != nil {
		utils.LogError("[ERROR] Rename failed: %v", err)
		http.Error(w, fmt.Sprintf("[ERROR] Failed to rename: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Update StringSet (only for device)
	if req.EntityType == "device" {
		// Even if empty, we update it (to allow clearing)
		// But if the field is missing in JSON (nil), we wouldn't know. 
		// Since we use struct, zero value is "". 
		// Frontend should send valid stringSet or "" to clear.
		if err := database.UpdateDeviceStringSet(req.ID, req.StringSet); err != nil {
			utils.LogError("[ERROR] Failed to update string set: %v", err)
			// Don't fail the whole request, but log it.
		}
	}

	utils.LogInfo("[INFO] Updated %s %s: Name='%s', StringSet='%s'", req.EntityType, req.ID, req.NewName, req.StringSet)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "newName": req.NewName})
}

func handleInverterDCPower(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device")
	if deviceID == "" {
		http.Error(w, "Missing 'device' query parameter", http.StatusBadRequest)
		return
	}

	data := fetchInverterPowerData(deviceID)

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "application/json")

	type PowerResponse struct {
		Device string               `json:"device"`
		Data   []InverterPowerPoint `json:"data"`
	}

	resp := PowerResponse{
		Device: deviceID,
		Data:   data,
	}

	bytes, _ := json.Marshal(resp)
	w.Write(bytes)
}

func startBackgroundUpdater() {
	// Initial population
	fmt.Println("[WAITING] Initializing site structure (Disk Scan)...")
	scanSitesStructure()
	fmt.Println("[WAITING] Initializing metrics (VM Fetch)...")
	updateMetrics()
	fmt.Println("[SUCCESS] Background updater started.")

	// Metric Loop (5s)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for range ticker.C {
			updateMetrics()
		}
	}()

	// File Watcher for Structure Updates
	go watchOutputDirectory("output")
}

func watchOutputDirectory(rootDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		utils.LogError("[ERROR] Failed to create watcher: %v", err)
		return
	}
	defer watcher.Close()

	// Debounce mechanism
	var debounceTimer *time.Timer
	debounceDuration := 2 * time.Second

	triggerScan := func() {
		if debounceTimer != nil {
			debounceTimer.Stop()
		}
		debounceTimer = time.AfterFunc(debounceDuration, func() {
			utils.LogInfo("[COLLECT] Detected changes in output directory. Re-scanning structure...")
			scanSitesStructure()
		})
	}

	// Recursive helper to add paths
	addWatchers := func(path string) error {
		return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return watcher.Add(p)
			}
			return nil
		})
	}

	// Initial add
	if err := addWatchers(rootDir); err != nil {
		utils.LogError("[ERROR] Error adding watchers: %v", err)
	}

	utils.LogInfo("[INFO] Started watching %s for structure changes...", rootDir)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// If new directory created, add watcher
			if event.Op&fsnotify.Create == fsnotify.Create {
				info, err := os.Stat(event.Name)
				if err == nil && info.IsDir() {
					watcher.Add(event.Name)
				}
			}

			// Trigger scan on Create, Write, Remove, Rename
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				// Only care about JSON files or Directories
				if strings.HasSuffix(event.Name, ".json") || !strings.Contains(filepath.Base(event.Name), ".") {
					triggerScan()
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			utils.LogError("Watcher error: %v", err)
		}
	}
}

func scanSitesStructure() {
	// Scan output dir manually to build tree STRUCTURE only
	// This is I/O intensive, so run less frequently
	sites := fetchSitesFromOutput("output")

	baseSitesLock.Lock()
	baseSites = sites
	baseSitesLock.Unlock()
}

func updateMetrics() {
	// 1. Fetch KPI from VM (Global)
	kpi := fetchKPIFromVM()

	// 2. Clone Sites from Base (Deep Copy to avoid race conditions)
	baseSitesLock.RLock()
	sites := deepCopySites(baseSites)
	baseSitesLock.RUnlock()

	// 2a. Enrich with Custom Names from DB
	enrichSitesWithCustomNames(&sites)

	// 3. Enrich Sites with Real-time Data from VM (Power, Strings, Status)
	enrichSitesWithVMData(&sites)

	// 4. Enrich Sites with KPI Data (Breakdown)
	enrichSitesWithKPI(&sites)

	// 5. Fetch Chart Data from VM
	chartData := fetchChartDataFromVM()
	siteData := SiteDataMap{
		All:   chartData,
		SiteA: []ChartPoint{},
		SiteB: []ChartPoint{},
	}

	// 6. Fallback: Aggregate KPI from Sites if VM query returned 0
	if kpi.DailyEnergy == 0 || kpi.RatedPower == 0 {
		for _, site := range sites {
			kpi.DailyEnergy += site.KPI.DailyEnergy
			kpi.DailyIncome += site.KPI.DailyIncome
			kpi.TotalEnergy += site.KPI.TotalEnergy
			kpi.RatedPower += site.KPI.RatedPower
			kpi.GridSupplyToday += site.KPI.GridSupplyToday
			kpi.StandardCoalSaved += site.KPI.StandardCoalSaved
			kpi.CO2Reduction += site.KPI.CO2Reduction
			kpi.TreesPlanted += site.KPI.TreesPlanted
		}
	}

	// 7. Fetch Production Data for bar chart
	productionData := fetchProductionDataFromVM()

	// 8. Generate Alerts (Backend Logic)
	baseAlerts := []AlertMessage{
		{ID: "1", Timestamp: time.Now().UnixMilli(), Level: "info", Message: "System Online (VM Backend)", Source: "Backend"},
	}
	deviceAlerts := generateDeviceAlerts(sites, time.Now())

	// Merge alerts
	allAlerts := append(baseAlerts, deviceAlerts...)

	newData := DashboardData{
		Alerts:         allAlerts,
		Sites:          sites,
		KPI:            kpi,
		Sensors:        []SensorData{},
		Meters:         []MeterData{},
		ChartData:      chartData,
		SiteData:       siteData,
		ProductionData: productionData,
	}

	cacheMutex.Lock()
	apiCache = newData
	cacheMutex.Unlock()
}
