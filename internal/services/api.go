package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"yard-backend/internal/config"
	"yard-backend/internal/models"
	"yard-backend/internal/utils"
)

// fetches reforge stones from the hypixel api and filters for reforge stone category
func FetchReforgeStones() ([]models.Item, int64, error) {
	resp, err := http.Get(config.APIURL)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	var apiResponse models.HypixelAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, 0, err
	}

	if !apiResponse.Success {
		return nil, 0, fmt.Errorf("API returned success: false")
	}

	var reforgeStones []models.Item
	for _, item := range apiResponse.Items {
		if item.Category == "REFORGE_STONE" {
			reforgeStones = append(reforgeStones, item)
		}
	}

	return reforgeStones, apiResponse.LastUpdated, nil
}

// fetches the lowest auction price for an item from skycofl api
func FetchAuctionPrice(itemTag string) *int64 {
	url := fmt.Sprintf("%s/api/auctions/tag/%s/active/bin", config.SkyCoflURL, itemTag)
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var auctions []struct {
		StartingBid      int64 `json:"startingBid"`
		HighestBidAmount int64 `json:"highestBidAmount"`
		Bin              bool  `json:"bin"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&auctions); err != nil {
		return nil
	}

	if len(auctions) == 0 {
		return nil
	}

	lowestPrice := auctions[0].StartingBid
	if auctions[0].Bin && auctions[0].StartingBid > 0 {
		lowestPrice = auctions[0].StartingBid
	} else if auctions[0].HighestBidAmount > 0 {
		lowestPrice = auctions[0].HighestBidAmount
	}

	for _, auction := range auctions {
		if auction.Bin && auction.StartingBid > 0 && auction.StartingBid < lowestPrice {
			lowestPrice = auction.StartingBid
		} else if !auction.Bin && auction.HighestBidAmount > 0 && auction.HighestBidAmount < lowestPrice {
			lowestPrice = auction.HighestBidAmount
		}
	}

	if lowestPrice > 0 {
		return &lowestPrice
	}

	return nil
}

// fetches bazaar price data
// retries indefinitely until success when rate limited
func FetchBazaarPriceWithRetry(itemTag string, normalizedTag string) (*http.Response, error) {
	baseDelay := 2 * time.Second
	maxDelay := 60 * time.Second
	attempt := 0
	
	for {
		utils.RateLimitWait()
		
		url := fmt.Sprintf("%s/api/bazaar/%s/snapshot", config.SkyCoflURL, normalizedTag)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		
		if resp.StatusCode == 429 {
			resp.Body.Close()
			attempt++
			
			retryAfter := baseDelay * time.Duration(1<<attempt)
			if retryAfter > maxDelay {
				retryAfter = maxDelay
			}
			
			if resetTimeStr := resp.Header.Get("X-RateLimit-Reset"); resetTimeStr != "" {
				if resetTime, err := time.Parse(time.RFC3339, resetTimeStr); err == nil {
					waitTime := time.Until(resetTime)
					if waitTime > 0 && waitTime < retryAfter {
						retryAfter = waitTime
					}
				}
			}
			
			log.Printf("Rate limited for %s (attempt %d), waiting %v", itemTag, attempt, retryAfter)
			time.Sleep(retryAfter)
			continue
		}
		
		return resp, nil
	}
}

// fetches bazaar buy and sell prices along with top buy and sell orders for an item
func FetchBazaarPrice(itemTag string) (*float64, *float64, []models.BazaarOrder, []models.BazaarOrder) {
	normalizedTag := strings.ToLower(itemTag)
	
	resp, err := FetchBazaarPriceWithRetry(itemTag, normalizedTag)
	if err != nil {
		log.Printf("Error fetching bazaar data for %s (tried %s): %v", itemTag, normalizedTag, err)
		return nil, nil, nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != 429 && normalizedTag != itemTag {
			resp.Body.Close()
			resp2, err2 := FetchBazaarPriceWithRetry(itemTag, itemTag)
			if err2 != nil {
				log.Printf("Bazaar API error for %s (tried both %s and %s): %v", itemTag, normalizedTag, itemTag, err2)
				return nil, nil, nil, nil
			}
			defer resp2.Body.Close()
			if resp2.StatusCode == http.StatusOK {
				resp = resp2
			} else {
				if resp2.StatusCode != 429 {
					log.Printf("Bazaar API returned status %d for %s (tried both %s and %s)", resp2.StatusCode, itemTag, normalizedTag, itemTag)
				}
				return nil, nil, nil, nil
			}
		} else {
			if resp.StatusCode == 429 {
				log.Printf("Rate limited for %s, skipping bazaar data", itemTag)
			} else {
				log.Printf("Bazaar API returned status %d for %s", resp.StatusCode, itemTag)
			}
			return nil, nil, nil, nil
		}
	}

	var snapshot struct {
		BuyPrice   float64 `json:"buyPrice"`
		SellPrice  float64 `json:"sellPrice"`
		BuyOrders  []struct {
			Amount      int64   `json:"amount"`
			PricePerUnit float64 `json:"pricePerUnit"`
			Orders      int     `json:"orders"`
		} `json:"buyOrders"`
		SellOrders []struct {
			Amount      int64   `json:"amount"`
			PricePerUnit float64 `json:"pricePerUnit"`
			Orders      int     `json:"orders"`
		} `json:"sellOrders"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		log.Printf("Error decoding bazaar snapshot for %s: %v", itemTag, err)
		return nil, nil, nil, nil
	}

	var buyPrice, sellPrice *float64
	if snapshot.BuyPrice > 0 {
		buyPrice = &snapshot.BuyPrice
	}
	if snapshot.SellPrice > 0 {
		sellPrice = &snapshot.SellPrice
	}

	var topBuyOrders []models.BazaarOrder
	if len(snapshot.BuyOrders) > 0 {
		buyOrders := make([]models.BazaarOrder, len(snapshot.BuyOrders))
		for i, order := range snapshot.BuyOrders {
			buyOrders[i] = models.BazaarOrder{
				Amount:      order.Amount,
				PricePerUnit: order.PricePerUnit,
				Orders:      order.Orders,
			}
		}
		
		for i := 0; i < len(buyOrders)-1; i++ {
			for j := i + 1; j < len(buyOrders); j++ {
				if buyOrders[i].PricePerUnit < buyOrders[j].PricePerUnit {
					buyOrders[i], buyOrders[j] = buyOrders[j], buyOrders[i]
				}
			}
		}
		
		count := 3
		if len(buyOrders) < count {
			count = len(buyOrders)
		}
		topBuyOrders = buyOrders[:count]
	}

	var topSellOrders []models.BazaarOrder
	if len(snapshot.SellOrders) > 0 {
		sellOrders := make([]models.BazaarOrder, len(snapshot.SellOrders))
		for i, order := range snapshot.SellOrders {
			sellOrders[i] = models.BazaarOrder{
				Amount:      order.Amount,
				PricePerUnit: order.PricePerUnit,
				Orders:      order.Orders,
			}
		}
		
		for i := 0; i < len(sellOrders)-1; i++ {
			for j := i + 1; j < len(sellOrders); j++ {
				if sellOrders[i].PricePerUnit > sellOrders[j].PricePerUnit {
					sellOrders[i], sellOrders[j] = sellOrders[j], sellOrders[i]
				}
			}
		}
		
		count := 3
		if len(sellOrders) < count {
			count = len(sellOrders)
		}
		topSellOrders = sellOrders[:count]
	}

	return buyPrice, sellPrice, topBuyOrders, topSellOrders
}

