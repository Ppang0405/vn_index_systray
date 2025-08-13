package main

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/getlantern/systray"
)

var currentIndex = "VNINDEX"

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("VN-Index")
	systray.SetTooltip("Vietnam Stock Indices")

	mVNIndex := systray.AddMenuItem("VN-Index", "Show VN-Index")
	mVN30 := systray.AddMenuItem("VN30", "Show VN30")
	mHNX := systray.AddMenuItem("HNX-Index", "Show HNX-Index")
	systray.AddSeparator()
	mRefresh := systray.AddMenuItem("Refresh", "Refresh Data")
	mQuit := systray.AddMenuItem("Quit", "Quit the application")

	go func() {
		for {
			select {
			case <-mVNIndex.ClickedCh:
				currentIndex = "VNINDEX"
				fetchAndUpdateIndex()
			case <-mVN30.ClickedCh:
				currentIndex = "VN30"
				fetchAndUpdateIndex()
			case <-mHNX.ClickedCh:
				currentIndex = "HNX"
				fetchAndUpdateIndex()
			case <-mRefresh.ClickedCh:
				fetchAndUpdateIndex()
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()

	fetchAndUpdateIndex()
	go func() {
		for range time.Tick(60 * time.Second) {
			fetchAndUpdateIndex()
		}
	}()
}

func onExit() {
	// Cleanup code here
}

func fetchAndUpdateIndex() {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// run task list
	var res string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://xpower.vixs.vn/priceboard`),
		chromedp.WaitVisible(`#charts-wrapper > div > div > div:nth-child(1) > div.chart-info > div.chart-info-detail > span`, chromedp.ByQuery),
		chromedp.Text(`#charts-wrapper > div > div > div:nth-child(1) > div.chart-info > div.chart-info-detail > span`, &res, chromedp.ByQuery),
	)
	if err != nil {
		log.Printf("Error crawling page: %v", err)
		systray.SetTitle(fmt.Sprintf("%s: Error", currentIndex))
		return
	}

	// Parse the extracted text
	// Example: "1,234.56 🔺0.12 (0.01%)" or "1,234.56 🔻0.12 (0.01%)"
	re := regexp.MustCompile(`([\d,\.]+) ([\p{Sm}\p{So}])([\d\.]+) \(([\d\.]+)%\)`)
	matches := re.FindStringSubmatch(res)

	if len(matches) < 5 {
		log.Printf("Could not parse data: %s", res)
		systray.SetTitle(fmt.Sprintf("%s: Parse Error", currentIndex))
		return
	}

	valueStr := strings.ReplaceAll(matches[1], ",", "")
	latestValue, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Printf("Error parsing latest value: %v", err)
		systray.SetTitle(fmt.Sprintf("%s: Parse Error", currentIndex))
		return
	}

	indicator := matches[2]
	changeStr := matches[3]
	changePctStr := matches[4]

	change, err := strconv.ParseFloat(changeStr, 64)
	if err != nil {
		log.Printf("Error parsing change: %v", err)
		systray.SetTitle(fmt.Sprintf("%s: Parse Error", currentIndex))
		return
	}

	changePct, err := strconv.ParseFloat(changePctStr, 64)
	if err != nil {
		log.Printf("Error parsing change percentage: %v", err)
		systray.SetTitle(fmt.Sprintf("%s: Parse Error", currentIndex))
		return
	}

	// Adjust change and changePct based on indicator
	if indicator == "🔻" {
		change = -change
		changePct = -changePct
	}

	// Now update the systray title
	finalIndicator := "🔺"
	if change < 0 {
		finalIndicator = "🔻"
	}

	systray.SetTitle(fmt.Sprintf("%s: %.2f %s%.2f%%", currentIndex, latestValue, finalIndicator, abs(changePct)))
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
