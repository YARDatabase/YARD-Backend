package services

import "time"

// starts the scheduler to fetch and store reforge stones every 15 minutes
func StartScheduler() {
	FetchAndStoreReforgeStones(false)

	ticker := time.NewTicker(15 * time.Minute)
	go func() {
		for range ticker.C {
			FetchAndStoreReforgeStones(false)
		}
	}()
}

