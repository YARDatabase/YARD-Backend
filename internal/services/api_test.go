package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"yard-backend/internal/config"
	"yard-backend/internal/models"
)

func TestFetchReforgeStones_WhenApiReturnsSuccess_ReturnsStones(t *testing.T) {
	// Arrange
	originalAPIURL := config.APIURL
	defer func() { config.APIURL = originalAPIURL }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := models.HypixelAPIResponse{
			Success:     true,
			LastUpdated: 1234567890,
			Items: []models.Item{
				{ID: "STONE1", Category: "REFORGE_STONE", Name: "Stone 1"},
				{ID: "OTHER", Category: "OTHER", Name: "Other"},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config.APIURL = server.URL

	// Act
	stones, _, err := FetchReforgeStones()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, stones)
	assert.Equal(t, 1, len(stones))
	assert.Equal(t, "STONE1", stones[0].ID)
}

func TestFetchReforgeStones_WhenApiReturnsError_ReturnsError(t *testing.T) {
	// Arrange
	originalAPIURL := config.APIURL
	defer func() { config.APIURL = originalAPIURL }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config.APIURL = server.URL

	// Act
	stones, _, err := FetchReforgeStones()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, stones)
}

func TestFetchAuctionPrice_WhenAuctionsExist_ReturnsLowestPrice(t *testing.T) {
	// Arrange
	originalSkyCoflURL := config.SkyCoflURL
	defer func() { config.SkyCoflURL = originalSkyCoflURL }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auctions := []map[string]interface{}{
			{"startingBid": int64(2000), "highestBidAmount": int64(0), "bin": true},
			{"startingBid": int64(1000), "highestBidAmount": int64(0), "bin": true},
		}
		json.NewEncoder(w).Encode(auctions)
	}))
	defer server.Close()

	config.SkyCoflURL = server.URL

	// Act
	price := FetchAuctionPrice("TEST_ITEM")

	// Assert
	assert.NotNil(t, price)
	assert.Equal(t, int64(1000), *price)
}
