package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ItemModel struct {
	Textures map[string]string `json:"textures"`
	Parent   string            `json:"parent"`
}

var itemTextureMap = make(map[string]string)
var textureBasePath = "FurfSky/assets/cittofirmgenerated/textures/item"

// initializes the texture loader by scanning furfsky model files and building texture mappings
func InitTextureLoader() {
	log.Println("Loading FurfSky texture mappings...")
	
	modelsPath := "FurfSky/assets/firmskyblock/models/item"
	if _, err := os.Stat(modelsPath); os.IsNotExist(err) {
		log.Printf("Warning: FurfSky models directory not found at %s", modelsPath)
		return
	}

	count := 0
	err := filepath.WalkDir(modelsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Error reading %s: %v", path, err)
			return nil
		}

		var model ItemModel
		if err := json.Unmarshal(data, &model); err != nil {
			return nil
		}

		fileName := strings.TrimSuffix(filepath.Base(path), ".json")
		itemIDUpper := strings.ToUpper(fileName)
		itemIDLower := strings.ToLower(fileName)

		var texturePath string
		if layer0, ok := model.Textures["layer0"]; ok {
			texturePath = extractTexturePath(layer0)
		} else {
			for _, tex := range model.Textures {
				texturePath = extractTexturePath(tex)
				break
			}
		}

		if texturePath != "" {
			fullPath := filepath.Join(textureBasePath, texturePath+".png")
			if _, err := os.Stat(fullPath); err == nil {
				itemTextureMap[itemIDUpper] = fullPath
				itemTextureMap[itemIDLower] = fullPath
				itemTextureMap[fileName] = fullPath
				count++
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Error walking texture directory: %v", err)
	}

	log.Printf("Loaded %d item texture mappings", count)
}

// extracts the texture file path from a texture reference string
func extractTexturePath(textureRef string) string {
	if strings.HasPrefix(textureRef, "cittofirmgenerated:item/") {
		return strings.TrimPrefix(textureRef, "cittofirmgenerated:item/")
	}
	if strings.Contains(textureRef, ":") {
		parts := strings.Split(textureRef, ":")
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return ""
}

// gets the texture file path for an item id trying different normalization methods
func GetItemTexturePath(itemID string) (string, bool) {
	normalizedID := strings.ToUpper(strings.ReplaceAll(itemID, " ", "_"))
	if path, ok := itemTextureMap[normalizedID]; ok {
		return path, true
	}

	normalizedID = strings.ToLower(normalizedID)
	if path, ok := itemTextureMap[normalizedID]; ok {
		return path, true
	}

	normalizedID = strings.ReplaceAll(itemID, " ", "_")
	if path, ok := itemTextureMap[normalizedID]; ok {
		return path, true
	}

	return "", false
}

// extracts the texture hash from a minecraft skin value by decoding base64 and parsing json
func ExtractTextureHashFromSkin(skinValue interface{}) (string, error) {
	skinStr, ok := skinValue.(string)
	if !ok {
		return "", fmt.Errorf("skin value is not a string")
	}

	var data []byte
	var err error

	encodings := []*base64.Encoding{
		base64.RawStdEncoding,
		base64.StdEncoding,
		base64.RawURLEncoding,
		base64.URLEncoding,
	}

	for _, encoding := range encodings {
		data, err = encoding.DecodeString(skinStr)
		if err == nil {
			break
		}
	}

	if data == nil {
		return "", fmt.Errorf("failed to decode base64 skin value")
	}

	var skinData struct {
		Textures struct {
			SKIN struct {
				URL string `json:"url"`
			} `json:"SKIN"`
		} `json:"textures"`
	}

	if err := json.Unmarshal(data, &skinData); err != nil {
		return "", err
	}

	url := skinData.Textures.SKIN.URL
	if url == "" {
		return "", fmt.Errorf("no SKIN URL found")
	}

	parts := strings.Split(url, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid texture URL")
	}

	return parts[len(parts)-1], nil
}

