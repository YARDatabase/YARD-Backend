package services

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"yard-backend/internal/config"
	"yard-backend/internal/models"

	"github.com/redis/go-redis/v9"
)

// stores reforge stones metadata in redis (without prices)
func StoreReforgeStones(reforgeStones []models.Item) error {
	if config.RDB == nil {
		return fmt.Errorf("redis client not initialized")
	}

	existingIDs, err := config.RDB.SMembers(config.Ctx, "reforge_stones:ids").Result()
	if err != nil && err != redis.Nil {
		return err
	}

	existingMap := make(map[string]bool)
	for _, id := range existingIDs {
		existingMap[id] = true
	}

	newCount := 0
	for i := range reforgeStones {
		stone := &reforgeStones[i]

		// add reforge effect from NEU data
		reforgeEffect := GetReforgeEffectForStone(stone.ID)
		if reforgeEffect != nil {
			stone.ReforgeEffect = reforgeEffect
		}

		stoneJSON, err := json.Marshal(stone)
		if err != nil {
			log.Printf("Error marshaling stone %s: %v", stone.ID, err)
			continue
		}

		key := fmt.Sprintf("reforge_stone:%s", stone.ID)
		err = config.RDB.Set(config.Ctx, key, stoneJSON, 0).Err()
		if err != nil {
			log.Printf("Error storing stone %s: %v", stone.ID, err)
			continue
		}

		if !existingMap[stone.ID] {
			newCount++
			log.Printf("New reforge stone found: %s (%s)", stone.Name, stone.ID)
		}

		config.RDB.SAdd(config.Ctx, "reforge_stones:ids", stone.ID)
	}

	if newCount > 0 {
		log.Printf("Stored %d new reforge stones (total: %d)", newCount, len(reforgeStones))
	} else {
		log.Printf("No new reforge stones found (total: %d)", len(reforgeStones))
	}

	// track when hypixel data was last fetched
	currentTime := time.Now().UnixMilli()
	config.RDB.Set(config.Ctx, "reforge_stones:hypixel_updated", currentTime, 0)
	config.RDB.Set(config.Ctx, "reforge_stones:count", len(reforgeStones), 0)

	return nil
}

// refreshes prices from coflnet for all cached stones
func RefreshPrices() {
	if config.RDB == nil {
		log.Println("Redis not initialized, skipping price refresh")
		return
	}

	ids, err := config.RDB.SMembers(config.Ctx, "reforge_stones:ids").Result()
	if err != nil || len(ids) == 0 {
		log.Println("No stones cached, skipping price refresh")
		return
	}

	log.Printf("Refreshing prices for %d stones from Coflnet...", len(ids))
	startTime := time.Now()
	updatedCount := 0

	for _, stoneID := range ids {
		key := fmt.Sprintf("reforge_stone:%s", stoneID)
		stoneJSON, err := config.RDB.Get(config.Ctx, key).Result()
		if err != nil {
			continue
		}

		var stone models.Item
		if err := json.Unmarshal([]byte(stoneJSON), &stone); err != nil {
			continue
		}

		// fetch fresh prices from coflnet
		auctionPrice := FetchAuctionPrice(stone.ID)
		if auctionPrice != nil {
			stone.AuctionPrice = auctionPrice
		}

		buyPrice, sellPrice, buyOrders, sellOrders := FetchBazaarPrice(stone.ID)
		if buyPrice != nil {
			stone.BazaarBuyPrice = buyPrice
		}
		if sellPrice != nil {
			stone.BazaarSellPrice = sellPrice
		}
		if len(buyOrders) > 0 {
			stone.BazaarBuyOrders = buyOrders
		}
		if len(sellOrders) > 0 {
			stone.BazaarSellOrders = sellOrders
		}

		// save updated stone back to redis
		updatedJSON, err := json.Marshal(stone)
		if err != nil {
			continue
		}

		if err := config.RDB.Set(config.Ctx, key, updatedJSON, 0).Err(); err != nil {
			continue
		}

		updatedCount++
	}

	elapsed := time.Since(startTime)
	config.RDB.Set(config.Ctx, "reforge_stones:prices_updated", time.Now().UnixMilli(), 0)
	log.Printf("Price refresh complete: %d/%d stones updated in %v", updatedCount, len(ids), elapsed.Round(time.Second))
}

// fetches reforge stone list from hypixel api (runs every 5 hours)
func FetchAndStoreReforgeStones(force bool) {
	if config.RDB == nil {
		log.Println("Redis not initialized, skipping fetch")
		return
	}

	if !force {
		lastUpdatedStr, err := config.RDB.Get(config.Ctx, "reforge_stones:hypixel_updated").Result()
		if err == nil && lastUpdatedStr != "" {
			if timestamp, err := strconv.ParseInt(lastUpdatedStr, 10, 64); err == nil {
				lastUpdatedTime := time.Unix(timestamp/1000, 0)
				timeSinceUpdate := time.Since(lastUpdatedTime)

				// hypixel data only needs refresh every 5 hours
				if timeSinceUpdate < 5*time.Hour {
					ids, _ := config.RDB.SMembers(config.Ctx, "reforge_stones:ids").Result()
					if len(ids) > 0 {
						log.Printf("Hypixel data is fresh (updated %v ago, %d stones cached). Skipping fetch.",
							timeSinceUpdate.Round(time.Minute), len(ids))
						return
					}
				} else {
					log.Printf("Hypixel data is stale (updated %v ago). Fetching new stone list...", timeSinceUpdate.Round(time.Hour))
				}
			}
		} else {
			log.Println("No existing data found. Fetching initial data from Hypixel...")
		}
	} else {
		log.Println("Force fetch requested. Fetching new data from Hypixel...")
	}

	log.Println("Fetching reforge stones from Hypixel API...")
	reforgeStones, _, err := FetchReforgeStones()
	if err != nil {
		log.Printf("Error fetching reforge stones: %v", err)
		return
	}

	if err := StoreReforgeStones(reforgeStones); err != nil {
		log.Printf("Error storing reforge stones: %v", err)
		return
	}

	log.Printf("Successfully stored %d reforge stones from Hypixel", len(reforgeStones))

	RefreshPrices()
}
