package services

import (
	"testing"

	"yard-backend/internal/config"
	"yard-backend/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestStoreReforgeStones_WhenRedisNotInitialized_ReturnsError(t *testing.T) {
	// Arrange
	originalRDB := config.RDB
	config.RDB = nil
	defer func() { config.RDB = originalRDB }()

	// Act
	err := StoreReforgeStones([]models.Item{})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis client not initialized")
}
