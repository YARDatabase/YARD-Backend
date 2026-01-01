package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yard-backend/internal/config"
)

func TestLoadNEUReforgeStones_WhenFileExists_LoadsData(t *testing.T) {
	// Arrange
	originalNEURepoPath := config.NEURepoPath
	originalNEUReforgeStones := config.NEUReforgeStones
	defer func() {
		config.NEURepoPath = originalNEURepoPath
		config.NEUReforgeStones = originalNEUReforgeStones
	}()

	tempDir := t.TempDir()
	config.NEURepoPath = tempDir
	config.NEUReforgeStones = make(map[string]interface{})

	path := filepath.Join(tempDir, "constants")
	require.NoError(t, os.MkdirAll(path, 0755))

	data := map[string]interface{}{
		"STONE1": map[string]interface{}{
			"reforgeName": "Test Reforge",
		},
	}
	jsonData, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(path, "reforgestones.json"), jsonData, 0644))

	// Act
	err = LoadNEUReforgeStones()

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, config.NEUReforgeStones)
}

func TestLoadNEUReforgeStones_WhenFileNotFound_ReturnsError(t *testing.T) {
	// Arrange
	originalNEURepoPath := config.NEURepoPath
	defer func() { config.NEURepoPath = originalNEURepoPath }()

	tempDir := t.TempDir()
	config.NEURepoPath = tempDir

	// Act
	err := LoadNEUReforgeStones()

	// Assert
	assert.Error(t, err)
}

func TestGetReforgeEffectForStone_WhenStoneExists_ReturnsEffect(t *testing.T) {
	// Arrange
	originalNEUReforgeStones := config.NEUReforgeStones
	defer func() { config.NEUReforgeStones = originalNEUReforgeStones }()

	config.NEUReforgeStones = map[string]interface{}{
		"TEST_STONE": map[string]interface{}{
			"reforgeName": "Test Reforge",
			"itemTypes":   "SWORD",
		},
	}

	// Act
	effect := GetReforgeEffectForStone("TEST_STONE")

	// Assert
	assert.NotNil(t, effect)
	assert.Equal(t, "Test Reforge", effect.ReforgeName)
}

func TestGetReforgeEffectForStone_WhenStoneNotFound_ReturnsNil(t *testing.T) {
	// Arrange
	originalNEUReforgeStones := config.NEUReforgeStones
	defer func() { config.NEUReforgeStones = originalNEUReforgeStones }()

	config.NEUReforgeStones = make(map[string]interface{})

	// Act
	effect := GetReforgeEffectForStone("NONEXISTENT")

	// Assert
	assert.Nil(t, effect)
}
