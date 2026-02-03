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
// PowerMeterSignalMap maps meter signal IDs to field names
// PowerMeterSignalMap maps meter signal IDs to field names
var PowerMeterSignalMap = map[string]string{
	"10001": "status",
	"10004": "active_power_kw",     // Converted from W to kW if needed? Spec says _kw, value 0. Raw was -7921736 W.
	"10005": "reactive_power_kvar", // var -> kvar
	"10006": "power_factor",
	//"10025": "total_apparent_power", // Not in sample
	"10008": "total_positive_active_energy_kwh",     // 10008
	"10009": "total_negative_active_energy_kwh",     // 10009 Reverse active
	"10023": "total_positive_reactive_energy_kvarh", // 10023
	"10024": "total_negative_reactive_energy_kvarh", // 10024

	// Phase Voltages
	"10002": "phase_a_voltage_v",
	"10010": "phase_b_voltage_v",
	"10011": "phase_c_voltage_v",

	// Line Voltages
	"10016": "line_ab_voltage_v",
	"10017": "line_bc_voltage_v",
	"10018": "line_ca_voltage_v",

	// Phase Currents
	"10003": "phase_a_current_a",
	"10012": "phase_b_current_a",
	"10013": "phase_c_current_a",

	// Phase Active Power
	"10019": "phase_a_active_power_kw",
	"10020": "phase_b_active_power_kw",
	"10021": "phase_c_active_power_kw",
}

// GetSpringPVField generates field name for PV strings (e.g., pv1_voltage)
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

// SensorFieldMap maps display names to standardized field names
var SensorFieldMap = map[string]string{
	"Wind speed":                "wind_speed_ms",
	"Wind direction":            "wind_direction_deg",
	"PV Temperature":            "pv_module_temperature_c",
	"Ambient temperature":       "ambient_temperature_c",
	"Irradiance":                "total_irradiance_wm2",
	"Daily irradiation":         "daily_irradiation1_mjm2",
	"Daily irradiation(Energy)": "daily_irradiation1_kwhm2",
	"Total irradiation":         "total_irradiance2_wm2", // Mapping assumes Total->Total2 based on user example, check carefully
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
