package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"yard-backend/internal/config"
)

func TestExtractObtainingFromLore_WhenLoreContainsObtained_ReturnsObtainingInfo(t *testing.T) {
	// Arrange
	lore := []string{
		"§7Obtained from:",
		"§7Mining",
	}

	// Act
	result := ExtractObtainingFromLore(lore)

	// Assert
	assert.Equal(t, "Obtained from:", result)
}

func TestExtractObtainingFromLore_WhenLoreEmpty_ReturnsEmpty(t *testing.T) {
	// Arrange
	lore := []string{}

	// Act
	result := ExtractObtainingFromLore(lore)

	// Assert
	assert.Empty(t, result)
}

func TestExtractMiningLevelFromLore_WhenLoreContainsMiningLevel_ReturnsLevel(t *testing.T) {
	// Arrange
	lore := []string{
		"§7Mining Skill Level §a50!",
	}

	// Act
	result := ExtractMiningLevelFromLore(lore)

	// Assert
	assert.Contains(t, result, "Mining Skill Level")
	assert.Contains(t, result, "50")
}

func TestRateLimitWait_WhenElapsedTimeLessThanMinDelay_Waits(t *testing.T) {
	// Arrange
	originalLastRequestTime := config.LastRequestTime
	originalMinRequestDelay := config.MinRequestDelay
	defer func() {
		config.LastRequestTime = originalLastRequestTime
		config.MinRequestDelay = originalMinRequestDelay
	}()

	config.MinRequestDelay = 50 * time.Millisecond
	config.RateLimiterMutex.Lock()
	config.LastRequestTime = time.Now().Add(-10 * time.Millisecond)
	config.RateLimiterMutex.Unlock()

	// Act
	start := time.Now()
	RateLimitWait()
	elapsed := time.Since(start)

	// Assert
	assert.GreaterOrEqual(t, elapsed, 35*time.Millisecond)
}
