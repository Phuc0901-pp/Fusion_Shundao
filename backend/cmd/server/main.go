package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fusion/internal/api"
	"fusion/internal/browser"
	"fusion/internal/database"
	"fusion/internal/core/formatter"
	"fusion/internal/login"
	"fusion/internal/platform/config"
	"fusion/internal/platform/utils"
	"fusion/internal/ui"
	"fusion/internal/victoriametrics"
)

// DeviceTask holds all info needed to process a device
type DeviceTask struct {
	Device     api.Device
	SiteName   string
	SiteID     string
	SubPath    string // e.g. "SmartLogger_1/Inverter_1"
	Model      string
	SN         string
	DevicePath string // Full output path relative to site
}

func main() {
	// 1. Initialize System
	if err := config.LoadConfig(); err != nil {
		utils.LogError("[ERROR] Failed to load config: %v", err)
		os.Exit(1)
	}
	utils.InitLogger()
	formatter.InitMapper()

	// 2. Init Database
	if err := database.InitDB(); err != nil {
		utils.LogError("[ERROR] Failed to init database: %v", err)
		// We might want to exit or continue depending on severity. 
		// For now, let's log and maybe exit if strict.
		// os.Exit(1) 
	}

	utils.LogInfo("=== FusionSolar Site Data Fetcher (Continuous 24/7) ===")
	utils.LogInfo("[READY] Đang khởi động (headless check)...")

	// Create headless browser with VERY LONG timeout (e.g., 7 days) to stay alive
	ctx, cancel := browser.NewHeadless(168 * time.Hour)
	defer cancel()

	// Setup API fetcher
	fetcher := api.NewFetcher()
	fetcher.SetupNetworkListener(ctx)

	if err := fetcher.EnableNetwork(ctx); err != nil {
		utils.LogError("[ERROR] Lỗi enable network: %v", err)
		os.Exit(1)
	}

	// First Run Flag
	firstRun := true

	// Start UI Backend Server
	go ui.StartServer()

	// Infinite Loop
	for {
		utils.LogInfo("[INFO] BẮT ĐẦU CHU KỲ MỚI: %s", time.Now().Format("15:04:05 02/01/2006"))

		// 1. Check Session / Login
		if firstRun || !fetcher.HasValidToken(ctx) {
			utils.LogWarn("[WARNING] Phiên làm việc hết hạn hoặc chưa đăng nhập. Đang login lại...")
			fetcher.ClearToken()

			if err := login.PerformLogin(ctx); err != nil {
				utils.LogError("[ERROR] Lỗi đăng nhập: %v. Thử lại sau 1 phút...", err)
				time.Sleep(1 * time.Minute)
				continue
			}

			// Initial fetch to capture token if fresh login
			utils.LogInfo("[INFO] Lấy dữ liệu trạm để bắt Token...")
			_, err := fetcher.WaitAndFetchSiteData(ctx)
			if err != nil {
				utils.LogWarn("[WARNING] Cảnh báo lấy site data: %v", err)
			}
			firstRun = false
		} else {
			utils.LogInfo("[INFO] Phiên làm việc (Roarand) vẫn OK.")
		}

		// 2. Process Data
		startProcess := time.Now()
		processAllSites(ctx, fetcher)
		utils.LogInfo("[INFO] Running Times %.2fs", time.Since(startProcess).Seconds())

		// 3. Push to VictoriaMetrics
		victoriametrics.PushToVictoriaMetrics()

		// 4. Wait 5 Minutes
		utils.LogInfo("[WAITTING] CHỜ 5 PHÚT (TIẾP TỤC LÚC " + time.Now().Add(5*time.Minute).Format("15:04:05") + ")")
		time.Sleep(5 * time.Minute)
	}
}

// processAllSites contains the logic to fetch everything once
func processAllSites(ctx context.Context, fetcher *api.Fetcher) {
	// Lists to hold tasks
	var inverterTasks []DeviceTask
	var meterTasks []DeviceTask
	var sensorTasks []DeviceTask

	// --- PHASE 1: DISCOVERY & KPI ---
	for _, s := range config.App.Sites {
		siteDisplay := strings.ReplaceAll(s.Name, " ", "_")
		utils.LogInfo("\n--- XỬ LÝ TRẠM: %s ---\n", s.Name)

		// 1. Station Overview
		fetchStationOverview(ctx, fetcher, s, siteDisplay)


		// [DB] Save Site
		// Use Generated UUID for Site ID to match JSON output
		siteUUID := utils.GenerateUUID(s.ID)
		created, err := database.UpsertSite(siteUUID, s.Name)
		if err != nil {
			utils.LogError("[ERROR] DB Error (Site): %v", err)
		} else if created {
			utils.LogInfo("[SUCCESS] DB Saved New Site: %s (UUID: %s)", s.Name, siteUUID)
		}

		// Save detailed metadata for enrichment.go to read consistent ID
		saveSiteMetadata(s.Name, s.ID, siteUUID)

		// 2. Scan SmartLoggers & Devices
		utils.LogInfo("[INFO] Quét thiết bị... ")
		smartLoggers, err := fetcher.FetchSmartLoggers(ctx, s.ID)
		if err != nil {
			utils.LogError("[ERROR] Lỗi lấy SL: %v", err)
			continue
		}
		utils.LogInfo("[SUCCESS] OK (%d SL)", len(smartLoggers))

		for _, sl := range smartLoggers {
			slFolder := cleanName(sl.NodeName)

			// Fetch SL Self-Info
			slInfoData, _ := fetcher.FetchSmartLoggerDetail(ctx, sl.ElementDn)

			// Fetch SL Children (for Model/SN)
			children, _ := fetcher.FetchSmartLoggerChildren(ctx, sl.ElementDn)
			var slChildren []api.ChildDevice
			if children != nil {
				slChildren = children
			}

			// Save SmartLogger Data
			if slInfoData != nil {
				fmtSlData := formatter.FormatSmartLoggerData(slInfoData, sl.NodeName, sl.ElementDn, slChildren)
				saveFormattedData(fmtSlData, siteDisplay, slFolder, "smartLogger_data.json")
			}


			// [DB] Save SmartLogger
			// KEEP ID as ElementDn (NE=...) as requested
			// SiteID FK must match the Site's UUID
			created, err := database.UpsertSmartLogger(sl.ElementDn, siteUUID, sl.NodeName)
			if err != nil {
				utils.LogError("[ERROR] DB Error (SmartLogger): %v", err)
			} else if created {
				utils.LogInfo("[SUCCESS] DB Saved New SL: %s", sl.NodeName)
			}

			// Fetch Devices List
			devices, errDev := fetcher.FetchDevicesForSmartLogger(ctx, sl.ElementDn)
			if errDev != nil {
				utils.LogWarn("[WARNING] Lỗi lấy device con: %v", errDev)
				continue
			}

			// Classify Devices
			for _, d := range devices {
				dName := cleanName(d.NodeName)
				dPath := filepath.Join(slFolder, dName)
				model, sn := findStaticInfo(d.ElementDn, slChildren)

				task := DeviceTask{
					Device:     d,
					SiteName:   s.Name,
					SiteID:     s.ID,
					SubPath:    dPath,
					Model:      model,
					SN:         sn,
					DevicePath: dPath,
				}

				// [DB] Save Device
				// Determine Type:
				dTypeForDB := "Unknown"
				if d.TypeId == 23022 || strings.Contains(strings.ToLower(d.NodeName), "inverter") {
					dTypeForDB = "Inverter"
				} else if strings.Contains(strings.ToLower(d.NodeName), "meter") {
					dTypeForDB = "Meter"
				} else {
					dTypeForDB = "Sensor"
				}


				// [DB] Save Device
				// Use Generated UUID for Device ID
				// SmartLoggerID FK remains ElementDn
				deviceUUID := utils.GenerateUUID(d.ElementDn)
				created, err := database.UpsertDevice(deviceUUID, sl.ElementDn, d.NodeName, dTypeForDB, model, sn)
				if err != nil {
					utils.LogError("[ERROR] DB Error (Device): %v", err)
				} else if created {
					// Optional: Log only if needed, or keep it silent for devices as requested "clean log"
					// utils.LogInfo("      [SUCCESS] DB Saved New Device: %s", d.NodeName)
				}

				dType := strings.ToLower(d.NodeName)
				if d.TypeId == 23022 || strings.Contains(dType, "inverter") {
					inverterTasks = append(inverterTasks, task)
				} else if strings.Contains(dType, "meter") {
					meterTasks = append(meterTasks, task)
				} else if strings.Contains(dType, "emic") || strings.Contains(dType, "sensor") || strings.Contains(dType, "weather") || strings.Contains(dType, "emi") {
					sensorTasks = append(sensorTasks, task)
				}
			}
		}
	}

	// --- PHASE 2: BATCH PROCESSING ---
	if len(inverterTasks) > 0 {
		processInverterBatch(ctx, fetcher, inverterTasks)
	}
	if len(meterTasks) > 0 {
		processSimpleDeviceBatch(ctx, fetcher, meterTasks, "METER")
	}
	if len(sensorTasks) > 0 {
		processSimpleDeviceBatch(ctx, fetcher, sensorTasks, "SENSOR")
	}
}

// processInverterBatch processes inverters in chunks
func processInverterBatch(ctx context.Context, f *api.Fetcher, tasks []DeviceTask) {
	batchSize := 15
	total := len(tasks)
	utils.LogInfo("[INFO] Batch Inverters: %d total", total)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		chunk := tasks[i:end]
		randomSleep()

		var dns []string
		taskMap := make(map[string]DeviceTask)
		for _, t := range chunk {
			dns = append(dns, t.Device.ElementDn)
			taskMap[t.Device.ElementDn] = t
		}

		rtResults, _ := f.FetchBatchRealtimeData(ctx, dns, true)
		strResults, _ := f.FetchBatchInverterStringData(ctx, dns)

		rtMap := make(map[string]map[string]interface{})
		for _, r := range rtResults {
			if r.Success && r.Data != nil {
				rtMap[r.Dn] = r.Data
			}
		}
		strMap := make(map[string]map[string]interface{})
		for _, r := range strResults {
			if r.Success && r.Data != nil {
				strMap[r.Dn] = r.Data
			}
		}

		for _, t := range chunk {
			rtData := rtMap[t.Device.ElementDn]
			strData := strMap[t.Device.ElementDn]

			if rtData != nil {
				staticInfo := map[string]string{"model": t.Model, "sn": t.SN}
				siteInfo := map[string]string{"name": t.SiteName, "id": t.SiteID}
				unified := formatter.FormatUnifiedInverterData(rtData, strData, staticInfo, siteInfo, t.Device.NodeName, t.Device.ElementDn)
				saveFormattedData(unified, strings.ReplaceAll(t.SiteName, " ", "_"), t.DevicePath, "data.json")
			}
		}
		utils.LogDebug("[DEBUG] Processed batch %d-%d", i, end)
	}
	utils.LogInfo("[SUCCESS] Batch Inverters Done")
}

// processSimpleDeviceBatch processes meters/sensors
func processSimpleDeviceBatch(ctx context.Context, f *api.Fetcher, tasks []DeviceTask, label string) {
	batchSize := 20
	total := len(tasks)
	utils.LogInfo("[INFO] Batch %s: %d total", label, total)

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		chunk := tasks[i:end]
		randomSleep()

		var dns []string
		taskMap := make(map[string]DeviceTask)
		for _, t := range chunk {
			dns = append(dns, t.Device.ElementDn)
			taskMap[t.Device.ElementDn] = t
		}

		rtResults, _ := f.FetchBatchRealtimeData(ctx, dns, false)

		for _, res := range rtResults {
			if res.Success && res.Data != nil {
				t, ok := taskMap[res.Dn]
				if !ok {
					continue
				}

				staticInfo := map[string]string{"name": t.Device.NodeName, "model": t.Model, "sn": t.SN}
				siteInfo := map[string]string{"name": t.SiteName, "id": t.SiteID}
				siteDisplay := strings.ReplaceAll(t.SiteName, " ", "_")

				if label == "METER" {
					fmtData := formatter.FormatUnifiedPowerMeterData(res.Data, staticInfo, siteInfo, t.Device.NodeName, t.Device.ElementDn)
					saveFormattedData(fmtData, siteDisplay, t.DevicePath, "data.json")
				} else {
					fmtData := formatter.FormatUnifiedSensorData(res.Data, staticInfo, siteInfo, t.Device.NodeName, t.Device.ElementDn)
					saveFormattedData(fmtData, siteDisplay, t.DevicePath, "data.json")
				}
			}
		}
		utils.LogDebug("[DEBUG] Processed batch %d-%d", i, end)
	}
	utils.LogInfo("[SUCCESS] Batch %s Done", label)
}

// Helpers

func fetchStationOverview(ctx context.Context, f *api.Fetcher, s config.SiteConfig, siteDisplay string) {
	kpi, _ := f.FetchStationKPI(ctx, s.ID)
	if kpi != nil {
		kpi.StationName = s.Name
	}
	social, _ := f.FetchSocialContribution(ctx, s.ID)

	if kpi != nil || social != nil {
		overviewData := formatter.FormatStationOverview(kpi, social)
		saveFormattedData(overviewData, siteDisplay, "Station", "overview.json")
	}
}

func findStaticInfo(dn string, children []api.ChildDevice) (string, string) {
	for _, child := range children {
		if child.Dn == dn {
			m := ""
			s := ""
			if v, ok := child.ParamValues["50009"].(string); ok {
				m = v
			}
			if v, ok := child.ParamValues["50012"].(string); ok {
				s = v
			}
			return m, s
		}
	}
	return "", ""
}

func cleanName(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

func randomSleep() {
	// Sleep 200ms to 500ms
	ms := 200 + rand.Intn(300)
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func saveFormattedData(data interface{}, siteName, subPath, fileName string) {
	rootDir := "output"
	fullDir := filepath.Join(rootDir, siteName, subPath)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return
	}
	filePath := filepath.Join(fullDir, fileName)
	bytes, _ := json.MarshalIndent(data, "", "  ")
	os.WriteFile(filePath, bytes, 0644)
}
