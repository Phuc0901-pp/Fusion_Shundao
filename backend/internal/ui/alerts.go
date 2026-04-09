package ui

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"fusion/internal/notify"
	"fusion/internal/platform/utils"
)

// ─── Constants – mirrors frontend ALERT_CONFIG ────────────────────────────
const (
	alertThresholdPercent    = 0.80  // String below 80% average → suspect
	alertMinCurrentThreshold = 0.5   // A – below this is noise / dark
	alertDebounceMinutes     = 15    // minutes to confirm a fault
	alertSolarStartHour      = 6     // 06:00 start
	alertSolarEndHour        = 18    // 18:00 end
)

// ─── In-Memory State (reset at midnight) ─────────────────────────────────
type pendingFault struct {
	firstSeenAt time.Time
	faultType   string // "dead" | "noI" | "noV" | "thresh"
	message     string // human-readable description
	inverterId  string // needed for frontend routing
	inverterName string
	loggerName   string
	siteName     string
}

var (
	pendingFaults   sync.Map // key: string → *pendingFault
	confirmedFaults sync.Map // key: string → struct{} (already notified)
	lastResetDay    int      = -1
	resetMu         sync.Mutex
)

// maybeResetAtMidnight clears all state once per calendar day.
func maybeResetAtMidnight(now time.Time) {
	resetMu.Lock()
	defer resetMu.Unlock()
	if now.Day() != lastResetDay {
		lastResetDay = now.Day()
		pendingFaults.Range(func(k, _ interface{}) bool { pendingFaults.Delete(k); return true })
		confirmedFaults.Range(func(k, _ interface{}) bool { confirmedFaults.Delete(k); return true })
		utils.LogInfo("[SMART-ALERT] Midnight reset – cleared all pending/confirmed fault state.")
		
		// Prune Lark Bitable to prevent infinite growth (> 15000 records -> delete older than 5 days)
		go notify.PruneLarkBaseIfNeeded(15000, 5)
	}
}

// ─── Main Entry Point ─────────────────────────────────────────────────────

// generateSmartAlerts replaces the old generateDeviceAlerts function.
// It uses the same stateful-debounce algorithm that previously lived in the
// React hook useSmartAlerts.ts, now running 24/7 inside the Go backend.
func generateSmartAlerts(sites []SiteNode, now time.Time) []AlertMessage {
	maybeResetAtMidnight(now)

	var alerts []AlertMessage
	ts := now.UnixMilli()

	hour := now.Hour()
	isWorking := hour >= alertSolarStartHour && hour < alertSolarEndHour

	if !isWorking {
		// Outside solar hours – nothing to evaluate, cleanup stale state
		return alerts
	}

	currentCycleKeys := make(map[string]struct{})

	// ── PASS 1: Detect faults in current snapshot ─────────────────────
	for _, site := range sites {
		for _, logger := range site.Loggers {
			for _, inverter := range logger.Inverters {

				allStrings := inverter.Strings
				if len(allStrings) == 0 {
					continue
				}

				// Parse setupCount and excludedIndices (same as frontend)
				setupCount := 0
				fmt.Sscanf(inverter.NumberStringSet, "%d", &setupCount)

				excludedIdx := make(map[int]struct{})
				for _, part := range strings.Split(inverter.ExcludedStrings, ",") {
					part = strings.TrimSpace(part)
					if part == "" {
						continue
					}
					var n int
					if _, err := fmt.Sscan(part, &n); err == nil {
						excludedIdx[n] = struct{}{}
					}
				}

				// Filter valid strings
				var validStrings []StringData
				for _, s := range allStrings {
					idx := 0
					fmt.Sscanf(strings.TrimPrefix(s.ID, "PV"), "%d", &idx)
					if setupCount > 0 && idx > setupCount {
						continue
					}
					if _, excluded := excludedIdx[idx]; excluded {
						continue
					}
					validStrings = append(validStrings, s)
				}

				if len(validStrings) == 0 {
					continue
				}

				// Compute averages from healthy strings only
				var activeStrings []StringData
				for _, s := range validStrings {
					if s.Voltage > 10 && s.Current > alertMinCurrentThreshold {
						activeStrings = append(activeStrings, s)
					}
				}

				var avgV, avgI float64
				for _, s := range activeStrings {
					avgV += s.Voltage
					avgI += s.Current
				}
				if len(activeStrings) > 0 {
					avgV /= float64(len(activeStrings))
					avgI /= float64(len(activeStrings))
				}
				threshV := avgV * alertThresholdPercent
				threshI := avgI * alertThresholdPercent

				// Evaluate each string
				for _, s := range validStrings {
					hasV := s.Voltage > 10
					hasI := s.Current > alertMinCurrentThreshold
					baseKey := fmt.Sprintf("%s-%s-%s-%s", site.ID, logger.ID, inverter.ID, s.ID)

					var faultType, faultMsg string

					if !hasV && !hasI {
						faultType = "dead"
						faultMsg = fmt.Sprintf("Mất dòng & mất điện áp (%.2fA / %.1fV)",
							s.Current, s.Voltage)
					} else if !hasI && hasV {
						faultType = "noI"
						faultMsg = fmt.Sprintf("Mất dòng điện (%.2fA)", s.Current)
					} else if hasI && !hasV {
						faultType = "noV"
						faultMsg = fmt.Sprintf("Mất điện áp (%.1fV)", s.Voltage)
					} else if len(activeStrings) >= 2 {
						var warns []string
						if avgI > 0 && s.Current < threshI {
							warns = append(warns, fmt.Sprintf("%.2fA < %.2fA avg", s.Current, avgI))
						}
						if avgV > 0 && s.Voltage < threshV {
							warns = append(warns, fmt.Sprintf("%.1fV < %.1fV avg", s.Voltage, avgV))
						}
						if len(warns) > 0 {
							faultType = "thresh"
							faultMsg = fmt.Sprintf("Vượt ngưỡng 80%% (%s)", strings.Join(warns, " | "))
						}
					}

					if faultType == "" {
						continue
					}

					faultKey := fmt.Sprintf("%s-%s", baseKey, faultType)
					currentCycleKeys[faultKey] = struct{}{}

					// Register if not already pending
					if _, exists := pendingFaults.Load(faultKey); !exists {
						pendingFaults.Store(faultKey, &pendingFault{
							firstSeenAt:  now,
							faultType:    faultType,
							message:      fmt.Sprintf("%s → %s: %s", inverter.Name, s.ID, faultMsg),
							inverterId:   inverter.ID,
							inverterName: inverter.Name,
							loggerName:   logger.Name,
							siteName:     site.Name,
						})
					}
				}
			}
		}
	}

	// ── PASS 2: Self-healing – remove faults not seen this cycle ──────
	pendingFaults.Range(func(k, _ interface{}) bool {
		key := k.(string)
		if _, alive := currentCycleKeys[key]; !alive {
			pendingFaults.Delete(key)
			confirmedFaults.Delete(key)
		}
		return true
	})

	// ── PASS 3: Evaluate age & confirm ────────────────────────────────
	pendingFaults.Range(func(k, v interface{}) bool {
		key := k.(string)
		fault := v.(*pendingFault)
		ageMs := now.Sub(fault.firstSeenAt).Milliseconds()
		debounceMs := int64(alertDebounceMinutes * 60 * 1000)

		if ageMs >= debounceMs {
			// CONFIRMED – append to alert list
			ageMin := int(ageMs / 60000)
			alerts = append(alerts, AlertMessage{
				ID:         fmt.Sprintf("smart-%s", key),
				Timestamp:  ts,
				Level:      "error",
				Message:    fmt.Sprintf("🔴 NGUY HIỂM [%dp] %s - %s", ageMin, fault.loggerName, fault.message),
				Source:     fmt.Sprintf("%s - %s", fault.loggerName, fault.inverterName),
				DeviceType: "inverter",
				DeviceID:   fault.inverterId,
			})

			// First time confirmed → write to Lark Bitable
			if _, already := confirmedFaults.Load(key); !already {
				confirmedFaults.Store(key, struct{}{})
				record := notify.AlertRecord{
					Message:     fmt.Sprintf("[%s] %s", strings.ToUpper(fault.faultType), fault.message),
					Inverter:    fault.inverterName,
					SmartLogger: fault.loggerName,
					Site:        fault.siteName,
				}
				go func(r notify.AlertRecord) {
					if err := notify.SendLarkBitableAlert(r); err != nil {
						utils.LogWarn("[SMART-ALERT] Lark notify failed: %v", err)
					}
				}(record)
			}
		}
		return true
	})

	return alerts
}
