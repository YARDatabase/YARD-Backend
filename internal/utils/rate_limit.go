package utils

import (
	"time"
	"yard-backend/internal/config"
)

// waits to respect rate limits by ensuring minimum delay between requests
func RateLimitWait() {
	config.RateLimiterMutex.Lock()
	defer config.RateLimiterMutex.Unlock()
	
	elapsed := time.Since(config.LastRequestTime)
	if elapsed < config.MinRequestDelay {
		sleepTime := config.MinRequestDelay - elapsed
		time.Sleep(sleepTime)
	}
	config.LastRequestTime = time.Now()
}

