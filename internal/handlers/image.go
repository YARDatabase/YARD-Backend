package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/image/draw"
	"yard-backend/internal/config"
	"yard-backend/internal/models"
)

// upscales a texture image to the target size using nearest neighbor scaling
func UpscaleTexture(texturePath string, targetSize int) ([]byte, error) {
	file, err := os.Open(texturePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	srcImg, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	bounds := srcImg.Bounds()
	dstImg := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))
	draw.NearestNeighbor.Scale(dstImg, dstImg.Bounds(), srcImg, bounds, draw.Over, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dstImg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// handles requests for item images by data fetching from redis and rendering textures or skins
func HandleItemImageByData(w http.ResponseWriter, r *http.Request) {
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

	if config.RDB == nil {
		http.Error(w, "Redis client not initialized", http.StatusInternalServerError)
		return
	}

	ids, err := config.RDB.SMembers(config.Ctx, "reforge_stones:ids").Result()
	if err != nil {
		http.Error(w, "Error fetching reforge stones", http.StatusInternalServerError)
		return
	}

	var targetItem *models.Item
	for _, id := range ids {
		if id == itemID {
			key := fmt.Sprintf("reforge_stone:%s", id)
			stoneJSON, err := config.RDB.Get(config.Ctx, key).Result()
			if err != nil {
				continue
			}

			var stone models.Item
			if err := json.Unmarshal([]byte(stoneJSON), &stone); err != nil {
				continue
			}

			targetItem = &stone
			break
		}
	}

	if targetItem == nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	normalizedID := strings.ToUpper(strings.ReplaceAll(itemID, " ", "_"))
	
	if texturePath, ok := GetItemTexturePath(normalizedID); ok {
		imageData, err := UpscaleTexture(texturePath, 256)
		if err == nil {
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			w.Write(imageData)
			return
		}
	}

	if targetItem.Skin != nil {
		if skinValue, ok := targetItem.Skin["value"]; ok {
			textureHash, err := ExtractTextureHashFromSkin(skinValue)
			if err == nil && textureHash != "" {
				textureURL := fmt.Sprintf("https://textures.minecraft.net/texture/%s", textureHash)
				resp, err := http.Get(textureURL)
				if err == nil && resp.StatusCode == http.StatusOK {
					defer resp.Body.Close()
					
					srcImg, err := png.Decode(resp.Body)
					if err == nil {
						bounds := srcImg.Bounds()
						
						targetSize := 256
						dstImg := image.NewRGBA(image.Rect(0, 0, targetSize, targetSize))
						draw.NearestNeighbor.Scale(dstImg, dstImg.Bounds(), srcImg, bounds, draw.Over, nil)
						
						var buf bytes.Buffer
						if err := png.Encode(&buf, dstImg); err == nil {
							w.Header().Set("Content-Type", "image/png")
							w.Header().Set("Cache-Control", "public, max-age=31536000")
							w.Write(buf.Bytes())
							return
						}
					}
				}
			}
		}
	}

	http.Error(w, "Item texture not found", http.StatusNotFound)
}

