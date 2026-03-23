package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fusion/internal/platform/config"
	"fusion/internal/platform/utils"
)

type telegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// SendTelegram sends a message to the configured Telegram bot.
// Returns an error if the bot token or chat ID is not configured, or if the request fails.
func SendTelegram(message string) error {
	token := config.App.System.TelegramBotToken
	chatID := config.App.System.TelegramChatID

	if token == "" || chatID == "" {
		return fmt.Errorf("[NOTIFY] Telegram not configured (token or chat_id is empty)")
	}

	payload := telegramMessage{
		ChatID:    chatID,
		Text:      message,
		ParseMode: "HTML",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("[NOTIFY] Failed to marshal message: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("[NOTIFY] Failed to send Telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[NOTIFY] Telegram API returned status %d", resp.StatusCode)
	}

	return nil
}

// AlertLoginFailure sends a Telegram alert when login fails repeatedly.
// It uses the configured Telegram bot to send an alert message.
func AlertLoginFailure(reason string, attempt int) {
	maxRetries := config.App.System.MaxLoginRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	msg := fmt.Sprintf(
		"🚨 <b>[FUSION-SHUNDAO] CẢNH BÁO ĐĂNG NHẬP THẤT BẠI</b>\n\n"+
			"⏰ Thời gian: <code>%s</code>\n"+
			"🔁 Lần thử: <code>%d/%d</code>\n"+
			"❌ Lý do: <code>%s</code>\n\n"+
			"⚠️ Hệ thống đã TẠM DỪNG thu thập dữ liệu.\n"+
			"Vui lòng kiểm tra:\n"+
			"  • Thông tin đăng nhập trong <code>configs/app.json</code>\n"+
			"  • Kết nối mạng tới FusionSolar\n"+
			"  • Tài khoản có bị khóa không?",
		time.Now().Format("15:04:05 02/01/2006"),
		attempt, maxRetries, reason,
	)

	if err := SendTelegram(msg); err != nil {
		utils.LogWarn("[NOTIFY] Không thể gửi cảnh báo Telegram: %v", err)
	} else {
		utils.LogInfo("[NOTIFY] Đã gửi cảnh báo Telegram thành công!")
	}
}

// AlertSystemError sends a Telegram alert for a critical system error.
func AlertSystemError(component, reason string) {
	msg := fmt.Sprintf(
		"🔴 <b>[FUSION-SHUNDAO] LỖI HỆ THỐNG</b>\n\n"+
			"⏰ Thời gian: <code>%s</code>\n"+
			"🔧 Thành phần: <code>%s</code>\n"+
			"❌ Lỗi: <code>%s</code>",
		time.Now().Format("15:04:05 02/01/2006"),
		component, reason,
	)

	if err := SendTelegram(msg); err != nil {
		utils.LogWarn("[NOTIFY] Không thể gửi cảnh báo Telegram: %v", err)
	} else {
		utils.LogInfo("[NOTIFY] Đã gửi cảnh báo Telegram thành công!")
	}
}
