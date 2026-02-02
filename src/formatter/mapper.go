package formatter

import "fmt"

// InverterSignalMap maps generic signal IDs to FusionSolar style field names
var InverterSignalMap = map[string]string{
	"10025": "status",
	"10018": "active_power",
	"10019": "reactive_power",
	"10020": "power_factor",
	"10021": "grid_frequency",
	"10006": "efficiency",
	"10041": "internal_temperature",
	"10043": "insulation_resistance",
	"10032": "daily_yield",
	"10029": "total_yield",
	"10011": "phase_a_voltage",
	"10012": "phase_b_voltage",
	"10013": "phase_c_voltage",
	"10014": "phase_a_current",
	"10015": "phase_b_current",
	"10016": "phase_c_current",
}

// PowerMeterSignalMap maps meter signal IDs to field names
var PowerMeterSignalMap = map[string]string{
	"10001": "status",
	"10004": "active_power",
	"10005": "reactive_power",
	"10006": "power_factor",
	"10025": "total_apparent_power",
	"10009": "active_energy_import", // Reverse Active Energy
	"10008": "active_energy_export", // Positive Active Energy
	"10016": "voltage_a",            // Or Line AB depending on meter type
	"10017": "voltage_b",
	"10002": "voltage_c", // Note: Check actual ID mapping for Phase C
}

// GetSpringPVField generates field name for PV strings (e.g., pv1_voltage)
// index: 1-based index of the string
// signalType: "voltage" or "current"
func GetStringPVField(index int, signalType string) string {
	return fmt.Sprintf("pv%02d_%s", index, signalType)
}

// SensorFieldMap maps display names to standardized field names
var SensorFieldMap = map[string]string{
	"Wind speed":                "wind_speed",
	"Wind direction":            "wind_direction",
	"PV Temperature":            "pv_temperature",
	"Ambient temperature":       "ambient_temperature",
	"Irradiance":                "irradiance",
	"Daily irradiation":         "daily_irradiation",
	"Daily irradiation(Energy)": "daily_irradiation_energy",
	"Total irradiation":         "total_irradiation",
	"Status":                    "status",
}

// SmartLoggerFieldMap maps display names to standardized field names for SmartLogger info
var SmartLoggerFieldMap = map[string]string{
	"Device model":        "device_model",
	"Software Version":    "software_version",
	"SN":                  "serial_number",
	"Device IP address":   "ip_address",
	"Device name":         "device_name",
	"Device Name":         "device_name",
	"Maximum azimuth (°)": "maximum_azimuth",
}
