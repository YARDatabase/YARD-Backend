package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"yard-backend/internal/config"
	"yard-backend/internal/models"
	"yard-backend/internal/services"
)

// sets cors headers to allow cross origin requests from configured origin
func EnableCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if config.AllowedOrigin == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		allowedOrigins := strings.Split(config.AllowedOrigin, ",")
		for _, allowed := range allowedOrigins {
			if strings.TrimSpace(allowed) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "3600")
}

// handles health check requests and returns server status
func HandleHealth(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w, r)
	w.Header().Set("Content-Type", "application/json")
	response := models.HealthResponse{
		Status:  "ok",
		Message: "YARD Backend is running",
		Time:    time.Now(),
	}
	json.NewEncoder(w).Encode(response)
}

// handles requests for all reforge stones fetching them from redis and returning json
func HandleReforgeStones(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w, r)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if config.RDB == nil {
		http.Error(w, "Redis client not initialized", http.StatusInternalServerError)
		return
	}

	ids, err := config.RDB.SMembers(config.Ctx, "reforge_stones:ids").Result()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching reforge stone IDs: %v", err), http.StatusInternalServerError)
		return
	}

	var reforgeStones []models.Item
	for _, id := range ids {
		key := fmt.Sprintf("reforge_stone:%s", id)
		stoneJSON, err := config.RDB.Get(config.Ctx, key).Result()
		if err != nil {
			log.Printf("Error fetching stone %s: %v", id, err)
			continue
		}

		var stone models.Item
		if err := json.Unmarshal([]byte(stoneJSON), &stone); err != nil {
			log.Printf("Error unmarshaling stone %s: %v", id, err)
			continue
		}

		reforgeStones = append(reforgeStones, stone)
	}

	lastUpdatedStr, _ := config.RDB.Get(config.Ctx, "reforge_stones:last_updated").Result()
	var lastUpdated time.Time
	if lastUpdatedStr != "" {
		if timestamp, err := strconv.ParseInt(lastUpdatedStr, 10, 64); err == nil {
			lastUpdated = time.Unix(timestamp/1000, 0)
		}
	}

	response := models.ReforgeStonesResponse{
		Success:       true,
		Count:         len(reforgeStones),
		LastUpdated:   lastUpdated,
		ReforgeStones: reforgeStones,
	}

	json.NewEncoder(w).Encode(response)
}

// handles requests for item images by id upscaling textures and returning png data
func HandleItemImage(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w, r)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	itemID := vars["itemId"]
	if itemID == "" {
		http.Error(w, "Item ID is required", http.StatusBadRequest)
		return
	}

	normalizedID := strings.ToUpper(strings.ReplaceAll(itemID, " ", "_"))
	
	if texturePath, ok := GetItemTexturePath(normalizedID); ok {
		imageData, err := UpscaleTexture(texturePath, 256)
		if err != nil {
			log.Printf("Error upscaling texture file %s: %v", texturePath, err)
			http.Error(w, "Texture processing error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		w.Write(imageData)
		return
	}

	http.Error(w, "Item texture not found", http.StatusNotFound)
}

// handles requests for all reforges returning merged data from reforges.json and reforgestones.json
func HandleReforges(w http.ResponseWriter, r *http.Request) {
	EnableCORS(w, r)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	reforges := services.GetAllReforges()
	
	// sort reforges alphabetically by name
	sort.Slice(reforges, func(i, j int) bool {
		return reforges[i].ReforgeName < reforges[j].ReforgeName
	})

	// get last updated time from redis if available
	var lastUpdated time.Time
	if config.RDB != nil {
		lastUpdatedStr, err := config.RDB.Get(config.Ctx, "reforge_stones:last_updated").Result()
		if err == nil && lastUpdatedStr != "" {
			if timestamp, err := strconv.ParseInt(lastUpdatedStr, 10, 64); err == nil {
				lastUpdated = time.Unix(timestamp/1000, 0)
			}
		}
	}
	
	if lastUpdated.IsZero() {
		lastUpdated = time.Now()
	}

	response := models.ReforgesResponse{
		Success:     true,
		Count:       len(reforges),
		LastUpdated: lastUpdated,
		Reforges:    reforges,
	}

	json.NewEncoder(w).Encode(response)
}

