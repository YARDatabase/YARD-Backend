package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"yard-backend/internal/config"
	"yard-backend/internal/models"
)

func TestEnableCORS_WhenWildcardOrigin_SetsWildcardHeader(t *testing.T) {
	// Arrange
	originalAllowedOrigin := config.AllowedOrigin
	config.AllowedOrigin = "*"
	defer func() { config.AllowedOrigin = originalAllowedOrigin }()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	// Act
	EnableCORS(rr, req)

	// Assert
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestHandleHealth_WhenRequested_ReturnsOkStatus(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()

	// Act
	HandleHealth(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response models.HealthResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	assert.Equal(t, "ok", response.Status)
}

func TestHandleReforgeStones_WhenRedisNotInitialized_ReturnsInternalServerError(t *testing.T) {
	// Arrange
	originalRDB := config.RDB
	config.RDB = nil
	defer func() { config.RDB = originalRDB }()

	req, err := http.NewRequest("GET", "/api/reforge-stones", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()

	// Act
	HandleReforgeStones(rr, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleItemImage_WhenItemNotFound_ReturnsNotFound(t *testing.T) {
	// Arrange
	req, err := http.NewRequest("GET", "/api/item/NONEXISTENT", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	
	// Set up mux vars manually since we're not using router
	req = mux.SetURLVars(req, map[string]string{"itemId": "NONEXISTENT"})

	// Act
	HandleItemImage(rr, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
