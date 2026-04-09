package utils

import (
	"fmt"
	"time"

	"github.com/beevik/ntp"
)

var (
	// Location for Vietnam time (UTC+7)
	VnLocation *time.Location

	// Time offset: Internet Time - System Time
	// If positive, system time is behind.
	// If negative, system time is ahead.
	timeOffset time.Duration
)

// InitTime initializes the time module by syncing with NTP server
func InitTime() {
	var err error

	// 1. Load Vietnam Location
	// Note: On Windows, "Asia/Ho_Chi_Minh" might not be available in standard installation without zoneinfo.zip
	// We can construct it manually if needed, but let's try LoadLocation first.
	// Alternatively, use FixedZone for simplicity and reliability.
	VnLocation = time.FixedZone("UTC+7", 7*60*60)

	// 2. Get Internet Time from NTP
	// Default pool: pool.ntp.org.
	// You can also use: time.google.com, vn.pool.ntp.org
	ntpServer := "pool.ntp.org"

	fmt.Printf("Syncing time with %s...\n", ntpServer)

	options := ntp.QueryOptions{Timeout: 5 * time.Second}
	response, err := ntp.QueryWithOptions(ntpServer, options)

	if err != nil {
		fmt.Printf("⚠️ Failed to sync with NTP server: %v. Using system time.\n", err)
		// Try fallback to Google NTP
		response, err = ntp.QueryWithOptions("time.google.com", options)
		if err != nil {
			fmt.Printf("⚠️ Failed to sync with Google NTP: %v. Using system time.\n", err)
			return
		}
	}

	// Calculate offset
	// Offset = NTP Time - System Time (at reception)
	// We can use response.ClockOffset directly provided by the library
	timeOffset = response.ClockOffset

	fmt.Printf("✅ Time synced successfully!\n")
	fmt.Printf("   System Time:   %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("   Internet Time: %s\n", time.Now().Add(timeOffset).Format(time.RFC3339))
	fmt.Printf("   Offset:        %v\n", timeOffset)
	fmt.Printf("   Vietnam Time:  %s\n", Now().Format(time.RFC3339))
}

// Now returns the current time adjusted by the NTP offset and in Vietnam timezone
func Now() time.Time {
	return time.Now().Add(timeOffset).In(VnLocation)
}

// GetLocation returns the configured location (UTC+7)
func GetLocation() *time.Location {
	return VnLocation
}
