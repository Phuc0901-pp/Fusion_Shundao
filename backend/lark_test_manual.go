//go:build ignore

package main

import (
	"fmt"
	"fusion/internal/notify"
	"fusion/internal/platform/config"
	"time"
)

func main() {
	// 1. Load config to get credentials
	err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Config error: %v\n", err)
		return
	}

	fmt.Printf("App ID: %s\n", config.App.System.LarkAppID)
	fmt.Printf("App Token: %s\n", config.App.System.LarkBitableAppToken)

	// 2. Try recording a test alert
	record := notify.AlertRecord{
		Message:     "[TEST] Lỗi mô phỏng thử nghiệm kết nối Lark",
		Inverter:    "Inverter 99",
		SmartLogger: "Logger Test",
		Site:        "SHUNDAO TEST",
	}

	fmt.Println("Đang gọi API Lark Bitable...")
	timeStart := time.Now()
	err = notify.SendLarkBitableAlert(record)

	if err != nil {
		fmt.Printf("LỖI GỬI LARK: %v\n", err)
	} else {
		fmt.Printf("THÀNH CÔNG! Đã phản hồi sau %v\n", time.Since(timeStart))
	}
}
