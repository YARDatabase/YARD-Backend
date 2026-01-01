package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"yard-backend/internal/config"
)

func TestRateLimitMiddleware_WhenWithinLimit_AllowsRequest(t *testing.T) {
	// Arrange
	originalMaxRequests := config.APIRateLimitMaxRequests
	config.APIRateLimitMaxRequests = 3
	defer func() {
		config.APIRateLimitMaxRequests = originalMaxRequests
		rateLimitMutex.Lock()
		rateLimitMap = make(map[string]*rateLimitEntry)
		rateLimitMutex.Unlock()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := RateLimitMiddleware(handler)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:8080"

	// Act
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRateLimitMiddleware_WhenExceedsLimit_ReturnsTooManyRequests(t *testing.T) {
	// Arrange
	originalMaxRequests := config.APIRateLimitMaxRequests
	config.APIRateLimitMaxRequests = 2
	defer func() {
		config.APIRateLimitMaxRequests = originalMaxRequests
		rateLimitMutex.Lock()
		rateLimitMap = make(map[string]*rateLimitEntry)
		rateLimitMutex.Unlock()
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := RateLimitMiddleware(handler)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.2:8080"

	// Act - make requests up to limit
	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
	}
	
	// Act - exceed limit
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}
