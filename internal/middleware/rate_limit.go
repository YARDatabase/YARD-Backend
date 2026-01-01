package middleware

import (
	"net/http"
	"sync"
	"time"

	"yard-backend/internal/config"
)

type rateLimitEntry struct {
	count     int
	firstSeen time.Time
}

var (
	rateLimitMap = make(map[string]*rateLimitEntry)
	rateLimitMutex sync.RWMutex
	cleanupTicker *time.Ticker
	cleanupOnce sync.Once
)

func init() {
	cleanupOnce.Do(func() {
		cleanupTicker = time.NewTicker(5 * time.Minute)
		go func() {
			for range cleanupTicker.C {
				cleanupOldEntries()
			}
		}()
	})
}

// cleans up old rate limit entries to prevent memory leaks
func cleanupOldEntries() {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()
	
	now := time.Now()
	for key, entry := range rateLimitMap {
		if now.Sub(entry.firstSeen) > config.APIRateLimitWindow*2 {
			delete(rateLimitMap, key)
		}
	}
}

// gets the client identifier from the request
func getClientID(r *http.Request) string {
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	return ip
}

// rate limit middleware that limits requests per client per time window
func RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID := getClientID(r)
		now := time.Now()
		
		rateLimitMutex.Lock()
		entry, exists := rateLimitMap[clientID]
		
		if !exists {
			entry = &rateLimitEntry{
				count:     1,
				firstSeen: now,
			}
			rateLimitMap[clientID] = entry
			rateLimitMutex.Unlock()
			next(w, r)
			return
		}
		
		elapsed := now.Sub(entry.firstSeen)
		if elapsed > config.APIRateLimitWindow {
			entry.count = 1
			entry.firstSeen = now
			rateLimitMutex.Unlock()
			next(w, r)
			return
		}
		
		if entry.count >= config.APIRateLimitMaxRequests {
			rateLimitMutex.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded","message":"too many requests please try again later"}`))
			return
		}
		
		entry.count++
		rateLimitMutex.Unlock()
		next(w, r)
	}
}

