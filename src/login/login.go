package login

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"

	"fusion/config/auth"
	"fusion/src/config"
	"fusion/src/utils"
)

// PerformLogin navigates to FusionSolar and logs in
func PerformLogin(ctx context.Context) error {
	username, password, _ := auth.Credentials()
	loginURL := config.App.API.Endpoints["login_page"]
	if loginURL == "" {
		// Fallback for backward compatibility if json missing
		_, _, loginURL = auth.Credentials() 
	}

	return chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitVisible(config.App.Selectors.UsernameInput, chromedp.ByID),
		chromedp.Sleep(2*time.Second),

		// Fill username
		chromedp.ActionFunc(func(ctx context.Context) error {
			utils.LogInfo("[1/4] Điền username...")
			chromedp.Click(config.App.Selectors.UsernameInput, chromedp.ByID).Do(ctx)
			time.Sleep(300 * time.Millisecond)
			for _, char := range username {
				input.InsertText(string(char)).Do(ctx)
				time.Sleep(20 * time.Millisecond)
			}
			return nil
		}),

		// Fill password
		chromedp.ActionFunc(func(ctx context.Context) error {
			utils.LogInfo("[2/4] Điền password...")
			chromedp.Click(config.App.Selectors.PasswordInput, chromedp.ByQuery).Do(ctx)
			time.Sleep(300 * time.Millisecond)
			for _, char := range password {
				input.InsertText(string(char)).Do(ctx)
				time.Sleep(20 * time.Millisecond)
			}
			return nil
		}),

		chromedp.Sleep(500*time.Millisecond),

		// Press Enter to login
		chromedp.ActionFunc(func(ctx context.Context) error {
			utils.LogInfo("[3/4] Nhấn Enter...")
			return chromedp.KeyEvent("\r").Do(ctx)
		}),

		// Wait for login
		chromedp.ActionFunc(func(ctx context.Context) error {
			utils.LogInfo("[4/4] Chờ đăng nhập...")

			for i := 0; i < 30; i++ {
				time.Sleep(1 * time.Second)

				var url string
				chromedp.Location(&url).Do(ctx)

				if !strings.Contains(url, "login") && len(url) > 50 {
					utils.LogInfo("      ✓ Đăng nhập thành công!")
					return nil
				}
			}

			return fmt.Errorf("timeout đăng nhập")
		}),
	)
}
