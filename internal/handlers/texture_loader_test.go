package handlers

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTexturePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "cittofirmgenerated prefix",
			input:    "cittofirmgenerated:item/test_item",
			expected: "test_item",
		},
		{
			name:     "other namespace",
			input:    "minecraft:item/diamond",
			expected: "item/diamond",
		},
		{
			name:     "no namespace",
			input:    "test_item",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTexturePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetItemTexturePath(t *testing.T) {
	itemTextureMap["TEST_ITEM"] = "/path/to/test_item.png"
	itemTextureMap["test_item"] = "/path/to/test_item_lower.png"
	itemTextureMap["Test_Item"] = "/path/to/test_item_mixed.png"

	tests := []struct {
		name     string
		itemID   string
		expected string
		found    bool
	}{
		{
			name:     "uppercase match",
			itemID:   "TEST_ITEM",
			expected: "/path/to/test_item.png",
			found:    true,
		},
		{
			name:     "lowercase match",
			itemID:   "test_item",
			expected: "/path/to/test_item.png",
			found:    true,
		},
		{
			name:     "mixed case match",
			itemID:   "Test_Item",
			expected: "/path/to/test_item.png",
			found:    true,
		},
		{
			name:     "with spaces",
			itemID:   "TEST ITEM",
			expected: "/path/to/test_item.png",
			found:    true,
		},
		{
			name:     "not found",
			itemID:   "NONEXISTENT",
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, found := GetItemTexturePath(tt.itemID)
			assert.Equal(t, tt.found, found)
			if found {
				assert.Equal(t, tt.expected, path)
			}
		})
	}

	delete(itemTextureMap, "TEST_ITEM")
	delete(itemTextureMap, "test_item")
	delete(itemTextureMap, "Test_Item")
}

func TestExtractTextureHashFromSkin(t *testing.T) {
	skinData := map[string]interface{}{
		"textures": map[string]interface{}{
			"SKIN": map[string]interface{}{
				"url": "http://textures.minecraft.net/texture/abc123def456",
			},
		},
	}

	jsonData, err := json.Marshal(skinData)
	require.NoError(t, err)

	encodings := []struct {
		name     string
		encoding *base64.Encoding
	}{
		{"StdEncoding", base64.StdEncoding},
		{"RawStdEncoding", base64.RawStdEncoding},
		{"URLEncoding", base64.URLEncoding},
		{"RawURLEncoding", base64.RawURLEncoding},
	}

	for _, enc := range encodings {
		t.Run(enc.name, func(t *testing.T) {
			encoded := enc.encoding.EncodeToString(jsonData)
			hash, err := ExtractTextureHashFromSkin(encoded)
			assert.NoError(t, err)
			assert.Equal(t, "abc123def456", hash)
		})
	}

	tests := []struct {
		name     string
		skinValue interface{}
		expectError bool
	}{
		{
			name:        "not a string",
			skinValue:   123,
			expectError: true,
		},
		{
			name:        "invalid base64",
			skinValue:   "not base64!!!",
			expectError: true,
		},
		{
			name:        "invalid json",
			skinValue:   base64.StdEncoding.EncodeToString([]byte("invalid json")),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := ExtractTextureHashFromSkin(tt.skinValue)
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hash)
			}
		})
	}
}

func TestInitTextureLoader(t *testing.T) {
	tempDir := t.TempDir()
	modelsPath := filepath.Join(tempDir, "FurfSky", "assets", "firmskyblock", "models", "item")
	texturePath := filepath.Join(tempDir, "FurfSky", "assets", "cittofirmgenerated", "textures", "item", "test_item.png")

	require.NoError(t, os.MkdirAll(modelsPath, 0755))
	require.NoError(t, os.MkdirAll(filepath.Dir(texturePath), 0755))

	model := ItemModel{
		Textures: map[string]string{
			"layer0": "cittofirmgenerated:item/test_item",
		},
	}
	modelJSON, err := json.Marshal(model)
	require.NoError(t, err)

	modelFile := filepath.Join(modelsPath, "test_item.json")
	require.NoError(t, os.WriteFile(modelFile, modelJSON, 0644))

	require.NoError(t, os.WriteFile(texturePath, []byte("fake png data"), 0644))

	originalTextureBasePath := textureBasePath
	originalItemTextureMap := make(map[string]string)
	for k, v := range itemTextureMap {
		originalItemTextureMap[k] = v
	}

	textureBasePath = filepath.Join(tempDir, "FurfSky", "assets", "cittofirmgenerated", "textures", "item")
	
	itemTextureMap = make(map[string]string)

	count := 0
	err = filepath.WalkDir(modelsPath, func(path string, d os.DirEntry, err error) error {
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
			return err
		}

		var model ItemModel
		if err := json.Unmarshal(data, &model); err != nil {
			return nil
		}

		fileName := strings.TrimSuffix(filepath.Base(path), ".json")
		itemIDUpper := strings.ToUpper(fileName)

		var texturePath string
		if layer0, ok := model.Textures["layer0"]; ok {
			texturePath = extractTexturePath(layer0)
		}

		if texturePath != "" {
			fullPath := filepath.Join(textureBasePath, texturePath+".png")
			if _, err := os.Stat(fullPath); err == nil {
				itemTextureMap[itemIDUpper] = fullPath
				count++
			}
		}

		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Contains(t, itemTextureMap, "TEST_ITEM")

	textureBasePath = originalTextureBasePath
	itemTextureMap = originalItemTextureMap
}

