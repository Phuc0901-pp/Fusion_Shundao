package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fusion/config/site"
	"fusion/src/api"
	"fusion/src/browser"
	"fusion/src/formatter"
	"fusion/src/login"
)

func main() {
	fmt.Println("=== FusionSolar Site Data Fetcher (Formatted) ===")
	fmt.Println("Đang khởi động (headless mode)...")

	// Create headless browser with 10 minute timeout
	ctx, cancel := browser.NewHeadless(600 * time.Second)
	defer cancel()

	// Setup API fetcher
	fetcher := api.NewFetcher()
	fetcher.SetupNetworkListener(ctx)

	// Enable network monitoring
	if err := fetcher.EnableNetwork(ctx); err != nil {
		log.Fatalf("Lỗi enable network: %v", err)
	}

	// Perform login
	fmt.Println("\n--- Đăng nhập ---")
	if err := login.PerformLogin(ctx); err != nil {
		log.Fatalf("Lỗi đăng nhập: %v", err)
	}

	// Fetch initial site data (to capture Roarand token)
	fmt.Println("\n--- Lấy dữ liệu trạm ---")
	_, err := fetcher.WaitAndFetchSiteData(ctx)
	if err != nil {
		log.Printf("Lỗi lấy thông tin sites: %v", err)
	}

	// --- PROCESS EACH TARGET SITE ---
	for _, s := range site.TargetSites {
		siteDisplay := strings.ReplaceAll(s.Name, " ", "_") // Clean name for folder
		fmt.Printf("\n>>> ĐANG XỬ LÝ TRẠM: %s [%s] <<<\n", s.Name, s.ID)

		// 1. Station Overview (KPI + Social)
		var kpi *api.StationKPI
		var social *api.SocialContribution

		fmt.Print("   + Lấy KPI Trạm... ")
		if k, err := fetcher.FetchStationKPI(ctx, s.ID); err == nil {
			kpi = k
			// Fix missing name in API response
			kpi.StationName = s.Name
			fmt.Println("OK")
		} else {
			fmt.Printf("Lỗi (%v)\n", err)
		}

		fmt.Print("   + Lấy Dữ liệu Môi trường... ")
		if sc, err := fetcher.FetchSocialContribution(ctx, s.ID); err == nil {
			social = sc
			fmt.Println("OK")
		} else {
			fmt.Printf("Lỗi (%v)\n", err)
		}

		if kpi != nil || social != nil {
			overviewData := formatter.FormatStationOverview(kpi, social)
			saveFormattedData(overviewData, siteDisplay, "Station", "overview.json")
		}

		// 2. Fetch Devices (SmartLoggers & Inverters)
		fmt.Println("\n   + Quét thiết bị...")
		smartLoggers, err := fetcher.FetchSmartLoggers(ctx, s.ID) // Use s.ID as parent
		if err != nil {
			fmt.Printf("     ⚠️ Lỗi lấy SmartLogger: %v\n", err)
			continue
		}

		for _, sl := range smartLoggers {
			fmt.Printf("     -> SmartLogger: %s\n", sl.NodeName)

			// Get devices under SmartLogger
			devices, err := fetcher.FetchDevicesForSmartLogger(ctx, sl.ElementDn)
			if err != nil {
				fmt.Printf("        ⚠️ Lỗi lấy thiết bị con: %v\n", err)
				continue
			}

			// Create SmartLogger output folder name
			slFolder := strings.ReplaceAll(sl.NodeName, " ", "_")
			slFolder = strings.ReplaceAll(slFolder, "/", "_")

			// --- Fetch SmartLogger Self-Info ---
			fmt.Printf("     [Info] Fetching SmartLogger info: %s\n", sl.NodeName)
			slInfoData, err := fetcher.FetchSmartLoggerDetail(ctx, sl.ElementDn)

			// --- Fetch SmartLogger Children (Secondary Devices) ---
			var slChildren []api.ChildDevice
			fmt.Printf("     [Info] Fetching SmartLogger Children: %s\n", sl.NodeName)
			children, errChild := fetcher.FetchSmartLoggerChildren(ctx, sl.ElementDn)
			if errChild == nil {
				slChildren = children
				fmt.Printf("        ✓ Found %d child devices\n", len(children))
			} else {
				fmt.Printf("        ⚠️ Error fetching child devices: %v\n", errChild)
			}

			if err == nil && slInfoData != nil {
				// Format SmartLogger Data
				fmtSlData := formatter.FormatSmartLoggerData(slInfoData, sl.NodeName, sl.ElementDn, slChildren)
				saveFormattedData(fmtSlData, siteDisplay, slFolder, "smartLogger_data.json")
			} else {
				fmt.Printf("        ⚠️ Lỗi lấy thông tin SmartLogger: %v\n", err)
			}

			for _, device := range devices {
				// Clean device name
				deviceName := strings.ReplaceAll(device.NodeName, " ", "_")
				deviceName = strings.ReplaceAll(deviceName, "/", "_")

				// Output Path: {SiteName}/{SmartLoggerName}/{DeviceName}
				devicePath := filepath.Join(slFolder, deviceName)

				// --- 1. INVERTER ---
				if device.TypeId == 23022 || strings.Contains(strings.ToLower(device.NodeName), "inverter") {
					fmt.Printf("        * Inverter: %s\n", deviceName)

					// Prepare data containers
					var rtData map[string]interface{}
					var strData map[string]interface{}
					var model, sn string

					// a. Look up Static Info (Model, SN) from slChildren
					for _, child := range slChildren {
						if child.Dn == device.ElementDn {
							// Found matching child
							// Extract model and SN from ParamValues
							// Use existing logic from FormatSmartLoggerData or helper
							// Param IDs: 50009 (Model), 50012 (SN)
							if v, ok := child.ParamValues["50009"].(string); ok {
								model = v
							}
							if v, ok := child.ParamValues["50012"].(string); ok {
								sn = v
							}
							break
						}
					}

					// b. Fetch Running Data
					rt, errRt := fetcher.FetchInverterRealtimeData(ctx, device.ElementDn)
					if errRt == nil && rt != nil {
						rtData = rt
					} else {
						// log error?
					}

					// c. Fetch String Data
					sd, errSd := fetcher.FetchInverterStringData(ctx, device.ElementDn)
					if errSd == nil && sd != nil {
						strData = sd
					}

					// d. Merge and Format
					if rtData != nil { // Require at least IO data? Or can proceed with partial?
						staticInfo := map[string]string{
							"model": model,
							"sn":    sn,
						}
						siteInfo := map[string]string{
							"name": s.Name,
							"id":   s.ID,
						}

						unifiedData := formatter.FormatUnifiedInverterData(rtData, strData, staticInfo, siteInfo, device.NodeName, device.ElementDn)
						saveFormattedData(unifiedData, siteDisplay, devicePath, "data.json")
					}

					continue // Done with this device
				}

				// --- 2. POWER METER ---
				// --- 2. POWER METER ---
				if strings.Contains(strings.ToLower(device.NodeName), "powermeter") || strings.Contains(strings.ToLower(device.NodeName), "meter") {
					fmt.Printf("        * Meter: %s\n", deviceName)

					// Find Static Info
					var model, sn string
					for _, child := range slChildren {
						if child.Dn == device.ElementDn {
							if v, ok := child.ParamValues["50009"].(string); ok {
								model = v
							}
							if v, ok := child.ParamValues["50012"].(string); ok {
								sn = v
							}
							break
						}
					}

					pmData, err := fetcher.FetchEMICData(ctx, device.ElementDn)
					if err == nil && pmData != nil {
						// fmtData := formatter.FormatPowerMeterData(pmData.Data, device.NodeName, device.ElementDn)

						staticInfo := map[string]string{
							"name":  device.NodeName,
							"model": model,
							"sn":    sn,
						}
						siteInfo := map[string]string{
							"name": s.Name,
							"id":   s.ID,
						}

						formatted := formatter.FormatUnifiedPowerMeterData(pmData.Data, staticInfo, siteInfo, device.NodeName, device.ElementDn)
						saveFormattedData(formatted, siteDisplay, devicePath, "data.json")
					}
					continue
				}

				// --- 3. SENSOR (EMIC/EMI/Weather Station) ---
				if strings.Contains(strings.ToLower(device.NodeName), "emic") ||
					strings.Contains(strings.ToLower(device.NodeName), "sensor") ||
					strings.Contains(strings.ToLower(device.NodeName), "emi") ||
					strings.Contains(strings.ToLower(device.NodeName), "weather") {
					fmt.Printf("        * Sensor: %s\n", deviceName)

					// Find Static Info
					var model, sn string
					for _, child := range slChildren {
						if child.Dn == device.ElementDn {
							if v, ok := child.ParamValues["50009"].(string); ok {
								model = v
							}
							if v, ok := child.ParamValues["50012"].(string); ok {
								sn = v
							}
							break
						}
					}

					sensorData, err := fetcher.FetchEMICData(ctx, device.ElementDn)
					if err == nil && sensorData != nil {
						// fmtData := formatter.FormatSensorData(sensorData.Data, device.NodeName, device.ElementDn)

						staticInfo := map[string]string{
							"name":  device.NodeName, // Default name
							"model": model,
							"sn":    sn,
						}
						siteInfo := map[string]string{
							"name": s.Name,
							"id":   s.ID,
						}

						formatted := formatter.FormatUnifiedSensorData(sensorData.Data, staticInfo, siteInfo, device.NodeName, device.ElementDn)
						saveFormattedData(formatted, siteDisplay, devicePath, "data.json")
					}
					continue
				}
			}
		}

		// Remove independent loops for PowerMeter and Sensor as requested "group by tree"
		// If devices are not found in tree, they won't be processed, which is correct for "tree based" structure.
	}

	fmt.Println("\n========================================")
	fmt.Println("           ✓ HOÀN THÀNH!")
	fmt.Println("========================================")
}

// saveFormattedData saves the formatted struct to the target path
func saveFormattedData(data interface{}, siteName, subPath, fileName string) {
	// Root output folder
	rootDir := "output"

	// Full directory: output/{SiteName}/{SubPath}
	fullDir := filepath.Join(rootDir, siteName, subPath)

	if err := os.MkdirAll(fullDir, 0755); err != nil {
		log.Printf("Error creating dir %s: %v", fullDir, err)
		return
	}

	filePath := filepath.Join(fullDir, fileName)

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("Error marshalling json: %v", err)
		return
	}

	if err := os.WriteFile(filePath, bytes, 0644); err != nil {
		log.Printf("Error writing file %s: %v", filePath, err)
	} else {
		// fmt.Printf("Saved: %s\n", filePath) // Quiet success
	}
}
