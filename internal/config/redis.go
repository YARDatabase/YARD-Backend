package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// initializes redis connection using environment variables or defaults with retry logic
func InitRedis() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       0,
	})

	maxRetries := 10
	retryDelay := 2 * time.Second
	
	for i := 0; i < maxRetries; i++ {
		_, err := RDB.Ping(Ctx).Result()
		if err == nil {
			log.Println("Connected to Redis successfully")
			return
		}
		
		if i < maxRetries-1 {
			log.Printf("Failed to connect to Redis (attempt %d/%d): %v. Retrying in %v...", i+1, maxRetries, err, retryDelay)
			time.Sleep(retryDelay)
		} else {
			log.Fatalf("Failed to connect to Redis after %d attempts: %v", maxRetries, err)
		}
	}
}

