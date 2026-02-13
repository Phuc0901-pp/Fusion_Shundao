package utils

import (
	"fmt"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

// TimeProvider handles NTP synchronization and time calculations
type TimeProvider struct {
	offset   time.Duration
	location *time.Location
	mu       sync.RWMutex
	ready    bool
}

var (
	GlobalTimeProvider *TimeProvider
	once               sync.Once
)

// InitTimeProvider initializes the global time provider
func InitTimeProvider() {
	once.Do(func() {
		GlobalTimeProvider = &TimeProvider{}
		GlobalTimeProvider.syncTime()
		
		// Periodic sync every 30 minutes
		go func() {
			for {
				time.Sleep(30 * time.Minute)
				GlobalTimeProvider.syncTime()
			}
		}()
	})
}

// syncTime attempts to synchronize time with NTP server
func (tp *TimeProvider) syncTime() {
	// Try Vietnam pool first, then Google as backup
	servers := []string{"vn.pool.ntp.org", "time.google.com", "pool.ntp.org"}
	
	// Load Vietnam location
	loc, err := time.LoadLocation("Asia/Ho_Chi_Minh")
	if err != nil {
		fmt.Printf("ERROR: Failed to load location Asia/Ho_Chi_Minh: %v\n", err)
		// Fallback to FixedZone if LoadLocation fails (e.g. missing zoneinfo)
		loc = time.FixedZone("ICT", 7*60*60)
	}

	tp.mu.Lock()
	tp.location = loc
	tp.mu.Unlock()

	for _, server := range servers {
		response, err := ntp.Query(server)
		if err == nil {
			tp.mu.Lock()
			tp.offset = response.ClockOffset
			tp.ready = true
			tp.mu.Unlock()
			
			// Use fmt.Println solely because logger might depend on this provider (avoid cyclic deps or uninit logger)
			// But since we use slog in logger, we can use it here if initialized. 
			// However, to be safe:
			fmt.Printf("[INFO] NTP Sync Successful with %s. Offset: %v. Location: %v\n", server, response.ClockOffset, loc)
			return
		}
		fmt.Printf("[WARN] Failed to sync with %s: %v\n", server, err)
	}
	
	fmt.Println("[ERROR] All NTP sync attempts failed. Using system time.")
}

// Now returns the current time adjusted by NTP offset and in Vietnam timezone
func (tp *TimeProvider) Now() time.Time {
	if tp == nil {
		return time.Now()
	}

	tp.mu.RLock()
	defer tp.mu.RUnlock()

	// If not ready (failed sync), return system time in target location if possible
	t := time.Now().Add(tp.offset)
	if tp.location != nil {
		return t.In(tp.location)
	}
	return t
}

// GetNow returns the current time from the global provider
// This is a helper function to avoid direct access to GlobalTimeProvider
func GetNow() time.Time {
	if GlobalTimeProvider != nil {
		return GlobalTimeProvider.Now()
	}
	return time.Now()
}
