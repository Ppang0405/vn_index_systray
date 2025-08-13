package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

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
	today := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -14).Format("2006-01-02")

	url := fmt.Sprintf("https://finfo-api.vndirect.com.vn/v4/stock_prices?symbols=%s&fromDate=%s&toDate=%s", currentIndex, startDate, today)
	resp, err := http.Get(url)
	if err != nil {
		systray.SetTitle(fmt.Sprintf("%s: Error", currentIndex))
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		systray.SetTitle(fmt.Sprintf("%s: Error", currentIndex))
		return
	}

	var result struct {
		Data []struct {
			Close float64 `json:"close"`
			Open  float64 `json:"open"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		systray.SetTitle(fmt.Sprintf("%s: Error", currentIndex))
		return
	}

	if len(result.Data) < 2 {
		systray.SetTitle(fmt.Sprintf("%s: No Data", currentIndex))
		return
	}

	latestValue := result.Data[len(result.Data)-1].Close
	prevClose := result.Data[len(result.Data)-2].Close
	change := latestValue - prevClose
	changePct := (change / prevClose) * 100

	var indicator string
	if change >= 0 {
		indicator = "🔺"
	} else {
		indicator = "🔻"
	}

	systray.SetTitle(fmt.Sprintf("%s: %.2f %s%.2f%%", currentIndex, latestValue, indicator, abs(changePct)))
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
