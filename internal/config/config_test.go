package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadEnv_WhenEnvVarsSet_LoadsFromEnvironment(t *testing.T) {
	// Arrange
	originalAPIURL := APIURL
	defer func() { APIURL = originalAPIURL }()

	os.Setenv("API_URL", "https://test-api.example.com")

	// Act
	LoadEnv()

	// Assert
	assert.Equal(t, "https://test-api.example.com", APIURL)

	os.Unsetenv("API_URL")
}

func TestLoadEnv_WhenNoEnvVars_KeepsDefaults(t *testing.T) {
	// Arrange
	originalAPIURL := APIURL
	defer func() { APIURL = originalAPIURL }()

	os.Clearenv()

	// Act
	LoadEnv()

	// Assert
	assert.NotEmpty(t, APIURL)
}
