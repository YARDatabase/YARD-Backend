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

// stores reforge stones in redis after enriching them with prices and reforge effects
func StoreReforgeStones(reforgeStones []models.Item, lastUpdated int64) error {
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

	currentTime := time.Now().UnixMilli()
	config.RDB.Set(config.Ctx, "reforge_stones:last_updated", currentTime, 0)
	config.RDB.Set(config.Ctx, "reforge_stones:count", len(reforgeStones), 0)

	return nil
}

// fetches reforge stones from the api and stores them checking if data is fresh first
func FetchAndStoreReforgeStones(force bool) {
	if config.RDB == nil {
		log.Println("Redis not initialized, skipping fetch")
		return
	}

	if !force {
		lastUpdatedStr, err := config.RDB.Get(config.Ctx, "reforge_stones:last_updated").Result()
		if err == nil && lastUpdatedStr != "" {
			if timestamp, err := strconv.ParseInt(lastUpdatedStr, 10, 64); err == nil {
				lastUpdatedTime := time.Unix(timestamp/1000, 0)
				timeSinceUpdate := time.Since(lastUpdatedTime)
				
				if timeSinceUpdate < 6*time.Hour {
					ids, _ := config.RDB.SMembers(config.Ctx, "reforge_stones:ids").Result()
					if len(ids) > 0 {
						log.Printf("Data is fresh (updated %v ago, %d stones cached). Skipping fetch.", 
							timeSinceUpdate.Round(time.Minute), len(ids))
						return
					}
				} else {
					log.Printf("Data is stale (updated %v ago). Fetching new data...", timeSinceUpdate.Round(time.Minute))
				}
			}
		} else {
			log.Println("No existing data found. Fetching initial data...")
		}
	} else {
		log.Println("Force fetch requested. Fetching new data...")
	}

	log.Println("Fetching reforge stones from Hypixel API...")
	reforgeStones, lastUpdated, err := FetchReforgeStones()
	if err != nil {
		log.Printf("Error fetching reforge stones: %v", err)
		return
	}

	if err := StoreReforgeStones(reforgeStones, lastUpdated); err != nil {
		log.Printf("Error storing reforge stones: %v", err)
		return
	}

	log.Printf("Successfully processed %d reforge stones", len(reforgeStones))
}

