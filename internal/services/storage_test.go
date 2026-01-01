package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"yard-backend/internal/config"
	"yard-backend/internal/models"
)

func TestStoreReforgeStones_WhenRedisNotInitialized_ReturnsError(t *testing.T) {
	// Arrange
	originalRDB := config.RDB
	config.RDB = nil
	defer func() { config.RDB = originalRDB }()

	// Act
	err := StoreReforgeStones([]models.Item{}, 0)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis client not initialized")
}
