package formatter

import (
	"fusion/internal/platform/config"
	"fmt"
)

// InverterSignalMap maps generic signal IDs to FusionSolar style field names
var InverterSignalMap map[string]string

// PowerMeterSignalMap maps meter signal IDs to field names
var PowerMeterSignalMap map[string]string

// SensorFieldMap maps display names to standardized field names
var SensorFieldMap map[string]string

// SmartLoggerFieldMap maps display names to standardized field names for SmartLogger info
var SmartLoggerFieldMap map[string]string

// UnifiedInverterSignalMap maps signals for the Unified data format
var UnifiedInverterSignalMap map[string]string

// InitMapper initializes the maps from configuration
func InitMapper() {
	InverterSignalMap = config.Signals.Inverter
	UnifiedInverterSignalMap = config.Signals.UnifiedInverter
	PowerMeterSignalMap = config.Signals.Meter
	SensorFieldMap = config.Signals.Sensor
	SmartLoggerFieldMap = config.Signals.SmartLogger
}

// GetStringPVField generates field name for PV strings (e.g., pv1_voltage)
// index: 1-based index of the string
// signalType: "voltage" or "current"
func GetStringPVField(index int, signalType string) string {
	return fmt.Sprintf("pv%02d_%s", index, signalType)
}

// GetUnifiedPVField generates field name for unified format (e.g., pv01_volt_v)
func GetUnifiedPVField(index int, signalType string) string {
	// signalType should be "volt_v" or "amp_a"
	return fmt.Sprintf("pv%02d_%s", index, signalType)
}

