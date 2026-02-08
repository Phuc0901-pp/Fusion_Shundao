package browser

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"
)

// NewHeadless creates a headless Chrome browser context
func NewHeadless(timeout time.Duration) (context.Context, context.CancelFunc) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(1400, 900),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, ctxCancel := chromedp.NewContext(allocCtx)

	ctx, timeoutCancel := context.WithTimeout(ctx, timeout)

	// Combined cancel function
	cancel := func() {
		timeoutCancel()
		ctxCancel()
		allocCancel()
	}

	return ctx, cancel
}
