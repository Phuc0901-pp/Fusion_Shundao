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

	"fusion/internal/database"
	"fusion/internal/platform/utils"
	"fusion/internal/victoriametrics"

	"github.com/fsnotify/fsnotify"
)

// Cache Mechanism
var (
	apiCache      DashboardData
	cacheMutex    sync.RWMutex
	baseSites     []SiteNode
	baseSitesLock sync.RWMutex

	// Entity Config Cache
	entityConfigCache map[string]database.EntityConfig
	configCacheLock   sync.RWMutex
)

// === SSE Hub ===
// Keeps track of all active SSE clients and broadcasts updates.
type sseHub struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

var hub = &sseHub{clients: make(map[chan []byte]struct{})}

func (h *sseHub) subscribe() chan []byte {
	ch := make(chan []byte, 4)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *sseHub) unsubscribe(ch chan []byte) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}

func (h *sseHub) broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- data:
		default: // slow client – skip, don't block broadcaster
		}
	}
}

// StartServer starts the API server
func StartServer() {
	// Start background data aggregator
	initEntityConfigCache()
	go startBackgroundUpdater()

	http.HandleFunc("/api/dashboard", handleDashboard)
	http.HandleFunc("/api/static", handleStaticData)
	http.HandleFunc("/api/stream/dashboard", handleSSEDashboard)
	http.HandleFunc("/api/production-monthly", handleMonthlyProduction)
	http.HandleFunc("/api/rename", handleRename)
	http.HandleFunc("/api/inverter/dc-power", handleInverterDCPower)
	http.HandleFunc("/healthz", handleHealthz)

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
	// Parse optional ?month=YYYY-MM query param
	var selectedMonth time.Time
	if m := r.URL.Query().Get("month"); m != "" {
		// Expected format: "2026-02"
		if t, err := time.Parse("2006-01", m); err == nil {
			selectedMonth = t
		}
	}

	data := fetchMonthlyProductionData(selectedMonth)

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

// handleSSEDashboard – long-lived SSE connection.
// Browser connects once; server pushes fresh DashboardData whenever updateMetrics runs.
func handleSSEDashboard(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Send current snapshot immediately so the UI renders without waiting
	cacheMutex.RLock()
	snapshot, _ := json.Marshal(apiCache)
	cacheMutex.RUnlock()
	fmt.Fprintf(w, "data: %s\n\n", snapshot)
	flusher.Flush()

	// Subscribe to future updates
	ch := hub.subscribe()
	defer hub.unsubscribe(ch)

	ticker := time.NewTicker(30 * time.Second) // heartbeat to keep connection alive
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			// Heartbeat comment – keeps proxy connections alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// handleStaticData returns site/logger structure only (infrequently changes).
// Frontend fetches this once on load, not on every poll cycle.
func handleStaticData(w http.ResponseWriter, r *http.Request) {
	baseSitesLock.RLock()
	sites := baseSites
	baseSitesLock.RUnlock()

	w.Header().Set("Cache-Control", "max-age=300") // Cache-able for 5 minutes
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sites)
}



func handleRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" && r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EntityType      string `json:"entityType"`      // "site", "logger", "device"
		ID              string `json:"id"`              // DB ID
		NewName         string `json:"newName"`
		StringSet       string `json:"stringSet"`       // Total string count
		ExcludedStrings string `json:"excludedStrings"` // Comma-separated excluded indices
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
		if err := database.UpdateDeviceStringSet(req.ID, req.StringSet); err != nil {
			utils.LogError("[ERROR] Failed to update string set: %v", err)
		}
		if err := database.UpdateDeviceExcludedStrings(req.ID, req.ExcludedStrings); err != nil {
			utils.LogError("[ERROR] Failed to update excluded strings: %v", err)
		}
	}

	// 3. Update Cache
	configCacheLock.Lock()
	if entityConfigCache == nil {
		entityConfigCache = make(map[string]database.EntityConfig)
	}

	cfg := entityConfigCache[req.ID]
	cfg.Name = req.NewName
	if req.EntityType == "device" {
		if req.StringSet != "" {
			cfg.StringSet = req.StringSet
		}
		cfg.ExcludedStrings = req.ExcludedStrings
	}
	entityConfigCache[req.ID] = cfg
	configCacheLock.Unlock()

	utils.LogInfo("[INFO] Updated %s %s: Name='%s', StringSet='%s', ExcludedStrings='%s'", req.EntityType, req.ID, req.NewName, req.StringSet, req.ExcludedStrings)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "newName": req.NewName})
}

func initEntityConfigCache() {
	configs, err := database.GetAllEntityConfigs()
	if err != nil {
		utils.LogError("[ERROR] Failed to initialize entity config cache: %v", err)
		return
	}

	configCacheLock.Lock()
	entityConfigCache = configs
	configCacheLock.Unlock()
	utils.LogInfo("[INFO] Entity config cache initialized with %d entries", len(configs))
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

	// VM Push Loop (5 minutes)
	// Đẩy lại toàn bộ data hiện tại lên VictoriaMetrics với timestamp mới
	// để tạo continuous time-series cho biểu đồ tổng hợp.
	go func() {
		// Push ngay lần đầu
		victoriametrics.PushToVictoriaMetrics()
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			victoriametrics.PushToVictoriaMetrics()
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

	// Push fresh data to all SSE subscribers immediately
	if payload, err := json.Marshal(newData); err == nil {
		go hub.broadcast(payload)
	}
}
