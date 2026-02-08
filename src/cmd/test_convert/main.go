package main

import (
	"fmt"
	"strings"

	"fusion/src/victoriametrics"
)

func main() {
	// Test inverter
	path := "output/SHUNDAO_1/Smartlogger_Station_1/HF1_Inverter_1/data.json"
	metrics, err := victoriametrics.ConvertToPrometheus(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	lines := strings.Split(metrics, "\n")
	fmt.Println("=== Sample Inverter Metric (Full) ===")
	if len(lines) > 0 {
		fmt.Println(lines[0])
	}
}
