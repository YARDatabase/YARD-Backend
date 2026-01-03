package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"yard-backend/internal/config"
	"yard-backend/internal/handlers"
	"yard-backend/internal/models"
	"yard-backend/internal/utils"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(handlers.HandleHealth)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))

	var response models.HealthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response.Status)
	assert.Equal(t, "YARD Backend is running", response.Message)
}

func TestEnableCORS(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	handlers.EnableCORS(rr, req)

	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type", rr.Header().Get("Access-Control-Allow-Headers"))
}

func TestRateLimitWait(t *testing.T) {
	config.RateLimiterMutex.Lock()
	config.LastRequestTime = time.Now().Add(-100 * time.Millisecond)
	config.RateLimiterMutex.Unlock()

	start := time.Now()
	utils.RateLimitWait()
	elapsed := time.Since(start)

	assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond)
}

func TestExtractObtainingFromLore(t *testing.T) {
	tests := []struct {
		name     string
		lore     []string
		expected string
	}{
		{
			name:     "empty lore",
			lore:     []string{},
			expected: "",
		},
		{
			name: "found in lore",
			lore: []string{
				"§7Obtained from:",
				"§7Mining",
			},
			expected: "Obtained from:",
		},
		{
			name: "drop mentioned",
			lore: []string{
				"§7Drops from:",
				"§7Slayer Bosses",
			},
			expected: "Drops from:",
		},
		{
			name: "no obtaining info",
			lore: []string{
				"§7Some other text",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ExtractObtainingFromLore(tt.lore)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractMiningLevelFromLore(t *testing.T) {
	tests := []struct {
		name     string
		lore     []string
		expected string
	}{
		{
			name:     "empty lore",
			lore:     []string{},
			expected: "",
		},
		{
			name: "mining level found",
			lore: []string{
				"§7Mining Skill Level §a50!",
			},
			expected: "Mining Skill Level 50",
		},
		{
			name: "mining level not found",
			lore: []string{
				"§7Some other text",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ExtractMiningLevelFromLore(tt.lore)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpscaleTexture(t *testing.T) {
	_, err := handlers.UpscaleTexture("/nonexistent/path.png", 256)
	assert.Error(t, err)
}

func TestHandleItemImage(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		expectedStatus int
	}{
		{
			name:           "missing item ID",
			itemID:         "",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "valid item ID",
			itemID:         "TEST_ITEM",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/item/"+tt.itemID, nil)
			require.NoError(t, err)

			router := mux.NewRouter()
			router.HandleFunc("/api/item/{itemId}", handlers.HandleItemImage).Methods("GET")
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestHandleItemImageByData(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/item-data/TEST_ITEM", nil)
	require.NoError(t, err)

	router := mux.NewRouter()
	router.HandleFunc("/api/item-data/{itemId}", handlers.HandleItemImageByData).Methods("GET")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Contains(t, []int{http.StatusNotFound, http.StatusInternalServerError}, rr.Code)
}

func TestReforgeStonesResponse(t *testing.T) {
	response := models.ReforgeStonesResponse{
		Success:       true,
		Count:         5,
		LastUpdated:   time.Now(),
		ReforgeStones: []models.Item{},
	}

	assert.True(t, response.Success)
	assert.Equal(t, 5, response.Count)
	assert.NotNil(t, response.LastUpdated)
}

func TestItemStruct(t *testing.T) {
	item := models.Item{
		Name:       "Test Stone",
		Category:   "REFORGE_STONE",
		Tier:       "LEGENDARY",
		ID:         "TEST_STONE",
		Glowing:    false,
		CanAuction: true,
	}

	assert.Equal(t, "Test Stone", item.Name)
	assert.Equal(t, "REFORGE_STONE", item.Category)
	assert.Equal(t, "LEGENDARY", item.Tier)
	assert.Equal(t, "TEST_STONE", item.ID)
}

func TestBazaarOrder(t *testing.T) {
	order := models.BazaarOrder{
		Amount:       100,
		PricePerUnit: 50.5,
		Orders:       5,
	}

	assert.Equal(t, int64(100), order.Amount)
	assert.Equal(t, 50.5, order.PricePerUnit)
	assert.Equal(t, 5, order.Orders)
}

func TestReforgeStats(t *testing.T) {
	health := 100.0
	defense := 50.0
	stats := models.ReforgeStats{
		Health:  &health,
		Defense: &defense,
	}

	assert.NotNil(t, stats.Health)
	assert.NotNil(t, stats.Defense)
	assert.Equal(t, 100.0, *stats.Health)
	assert.Equal(t, 50.0, *stats.Defense)
}

func TestReforgeEffect(t *testing.T) {
	effect := models.ReforgeEffect{
		ReforgeName:      "Test Reforge",
		ItemTypes:        "SWORD,AXE",
		RequiredRarities: []string{"EPIC", "LEGENDARY"},
		ReforgeCosts: map[string]int{
			"EPIC": 1000,
		},
	}

	assert.Equal(t, "Test Reforge", effect.ReforgeName)
	assert.Equal(t, "SWORD,AXE", effect.ItemTypes)
	assert.Len(t, effect.RequiredRarities, 2)
	assert.Equal(t, 1000, effect.ReforgeCosts["EPIC"])
}
