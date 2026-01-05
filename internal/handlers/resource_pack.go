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

// stores model data from json files
type ModelData struct {
	Parent   string            `json:"parent"`
	Textures map[string]string `json:"textures"`
}

// maps item ids to their texture file paths
var textureRegistry = make(map[string]string)

// base paths for the hypixelplus resource pack
const (
	resourcePackRoot      = "resources/HypixelPlus"
	modelsBasePath        = "resources/HypixelPlus/assets/hplus/models/skyblock"
	texturesBasePath      = "resources/HypixelPlus/assets/hplus/textures"
	minecraftTexturesPath = "resources/HypixelPlus/assets/minecraft/textures"
)

// scans the resource pack and builds a mapping of item ids to texture paths
func LoadResourcePack() {
	log.Println("scanning hypixelplus resource pack...")

	if _, err := os.Stat(modelsBasePath); os.IsNotExist(err) {
		log.Printf("warning: resource pack models not found at %s", modelsBasePath)
		return
	}

	loadedCount := 0
	err := filepath.WalkDir(modelsBasePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("failed to read model %s: %v", path, err)
			return nil
		}

		var model ModelData
		if err := json.Unmarshal(data, &model); err != nil {
			return nil
		}

		// extract item name from filename
		baseName := strings.TrimSuffix(filepath.Base(path), ".json")
		itemIdUpper := strings.ToUpper(baseName)
		itemIdLower := strings.ToLower(baseName)

		// resolve the texture path from model data
		texturePath, namespace := resolveTexturePathWithNamespace(model)
		if texturePath == "" {
			return nil
		}

		// determine which texture base path to use based on namespace
		var fullTexturePath string
		if namespace == "minecraft" {
			fullTexturePath = filepath.Join(minecraftTexturesPath, texturePath+".png")
		} else {
			fullTexturePath = filepath.Join(texturesBasePath, texturePath+".png")
		}

		// verify the texture file exists
		if _, err := os.Stat(fullTexturePath); err == nil {
			textureRegistry[itemIdUpper] = fullTexturePath
			textureRegistry[itemIdLower] = fullTexturePath
			textureRegistry[baseName] = fullTexturePath
			loadedCount++
		}

		return nil
	})

	if err != nil {
		log.Printf("error scanning resource pack: %v", err)
	}

	log.Printf("loaded %d item textures from resource pack", loadedCount)
}

// extracts and converts texture reference to actual file path, returns path and namespace
func resolveTexturePathWithNamespace(model ModelData) (string, string) {
	// prefer layer0 texture, fall back to any available
	var textureRef string
	if layer0, ok := model.Textures["layer0"]; ok {
		textureRef = layer0
	} else {
		for _, tex := range model.Textures {
			textureRef = tex
			break
		}
	}

	if textureRef == "" {
		return "", ""
	}

	// handle hplus:skyblock/... format -> skyblock/...
	if strings.HasPrefix(textureRef, "hplus:") {
		path := strings.TrimPrefix(textureRef, "hplus:")
		return path, "hplus"
	}

	// handle minecraft:item/... format
	if strings.HasPrefix(textureRef, "minecraft:") {
		path := strings.TrimPrefix(textureRef, "minecraft:")
		return path, "minecraft"
	}

	// handle cittofirmgenerated:item/... format (older packs)
	if strings.HasPrefix(textureRef, "cittofirmgenerated:item/") {
		path := strings.TrimPrefix(textureRef, "cittofirmgenerated:item/")
		return path, "hplus"
	}

	// handle other namespace:path formats
	if idx := strings.Index(textureRef, ":"); idx != -1 {
		namespace := textureRef[:idx]
		path := textureRef[idx+1:]
		return path, namespace
	}

	return textureRef, "hplus"
}

// extracts and converts texture reference to actual file path (kept for backward compatibility)
func resolveTexturePath(model ModelData) string {
	path, _ := resolveTexturePathWithNamespace(model)
	return path
}

// looks up texture path for an item, tries multiple id formats
func GetItemTexturePath(itemID string) (string, bool) {
	// try uppercase with underscores
	normalized := strings.ToUpper(strings.ReplaceAll(itemID, " ", "_"))
	if path, ok := textureRegistry[normalized]; ok {
		return path, true
	}

	// try lowercase
	normalized = strings.ToLower(normalized)
	if path, ok := textureRegistry[normalized]; ok {
		return path, true
	}

	// try original with underscores
	normalized = strings.ReplaceAll(itemID, " ", "_")
	if path, ok := textureRegistry[normalized]; ok {
		return path, true
	}

	return "", false
}

// decodes minecraft skin data to extract texture hash for player heads
func ExtractTextureHashFromSkin(skinValue interface{}) (string, error) {
	skinStr, ok := skinValue.(string)
	if !ok {
		return "", fmt.Errorf("skin value is not a string")
	}

	// try different base64 encoding variants
	encodings := []*base64.Encoding{
		base64.RawStdEncoding,
		base64.StdEncoding,
		base64.RawURLEncoding,
		base64.URLEncoding,
	}

	var decoded []byte
	var decodeErr error

	for _, enc := range encodings {
		decoded, decodeErr = enc.DecodeString(skinStr)
		if decodeErr == nil {
			break
		}
	}

	if decoded == nil {
		return "", fmt.Errorf("could not decode base64 skin data")
	}

	// parse the skin json to get texture url
	var skinData struct {
		Textures struct {
			SKIN struct {
				URL string `json:"url"`
			} `json:"SKIN"`
		} `json:"textures"`
	}

	if err := json.Unmarshal(decoded, &skinData); err != nil {
		return "", err
	}

	url := skinData.Textures.SKIN.URL
	if url == "" {
		return "", fmt.Errorf("no skin texture url found")
	}

	// extract hash from url (last path segment)
	segments := strings.Split(url, "/")
	if len(segments) == 0 {
		return "", fmt.Errorf("invalid texture url format")
	}

	return segments[len(segments)-1], nil
}
