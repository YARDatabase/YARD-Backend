package config

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var (
	RDB           *redis.Client
	Ctx           = context.Background()
	APIURL        = "https://api.hypixel.net/v2/resources/skyblock/items"
	SkyCoflURL    = "https://sky.coflnet.com"
	NEURepoPath   = "NotEnoughUpdates-REPO"
	AllowedOrigin = "*"

	RateLimiterMutex sync.Mutex
	LastRequestTime  time.Time
	// coflnet rate limit: 30 req/10s = 333ms
	MinRequestDelay = 350 * time.Millisecond

	NEUReforgeStones      map[string]interface{}
	NEUReforgeStonesMutex sync.RWMutex

	NEUReforges      map[string]interface{}
	NEUReforgesMutex sync.RWMutex

	APIRateLimitMutex       sync.Mutex
	APIRateLimitMap         = make(map[string]time.Time)
	APIRateLimitWindow      = 1 * time.Minute
	APIRateLimitMaxRequests = 60

	MetricsEnabled     = false
	MetricsIPWhitelist = ""
)

// reads env vars from file or system with defaults
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	if apiURL := os.Getenv("API_URL"); apiURL != "" {
		APIURL = apiURL
	}

	if skyCoflURL := os.Getenv("SKYCOFL_URL"); skyCoflURL != "" {
		SkyCoflURL = skyCoflURL
	}

	if neuRepoPath := os.Getenv("NEU_REPO_PATH"); neuRepoPath != "" {
		NEURepoPath = neuRepoPath
	}

	if allowedOrigin := os.Getenv("ALLOWED_ORIGIN"); allowedOrigin != "" {
		AllowedOrigin = allowedOrigin
	}

	if metricsEnabled := os.Getenv("METRICS_ENABLED"); metricsEnabled == "true" || metricsEnabled == "1" {
		MetricsEnabled = true
	}

	if metricsIPWhitelist := os.Getenv("METRICS_IP_WHITELIST"); metricsIPWhitelist != "" {
		MetricsIPWhitelist = metricsIPWhitelist
	}
}
