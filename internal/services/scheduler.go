package services

import (
	"log"
	"time"
)

// starts the schedulers for hypixel data (5h) and coflnet prices (5m)
func StartScheduler() {
	// initial fetch of stone list from hypixel + prices
	FetchAndStoreReforgeStones(false)

	// hypixel scheduler: check every hour but only fetch if > 5 hours old
	hypixelTicker := time.NewTicker(1 * time.Hour)
	go func() {
		for range hypixelTicker.C {
			FetchAndStoreReforgeStones(false)
		}
	}()

	// price scheduler: refresh prices every 5 minutes
	priceTicker := time.NewTicker(5 * time.Minute)
	go func() {
		for range priceTicker.C {
			log.Println("Starting scheduled price refresh...")
			RefreshPrices()
		}
	}()
}
