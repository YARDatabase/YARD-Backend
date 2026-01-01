package services

import "time"

// starts the scheduler to fetch and store reforge stones every 6 hours
func StartScheduler() {
	FetchAndStoreReforgeStones(false)

	ticker := time.NewTicker(6 * time.Hour)
	go func() {
		for range ticker.C {
			FetchAndStoreReforgeStones(false)
		}
	}()
}

