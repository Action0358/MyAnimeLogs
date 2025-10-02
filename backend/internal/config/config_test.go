package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_DefaultValues(t *testing.T) {
	// Clear environment variables
	envKeys := []string{
		"DATABASE_URL", "PORT", "ENVIRONMENT", "JWT_SECRET",
	}
	for _, key := range envKeys {
		os.Unsetenv(key)
	}

	config := New()

	// Test critical default values
	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "super-secret-key-change-me", config.JWTSecret)
	assert.Contains(t, config.DatabaseURL, "postgres://")
}

func TestNew_CustomEnvironmentVariables(t *testing.T) {
	// Set custom environment variables
	os.Setenv("PORT", "3000")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("JWT_SECRET", "custom-jwt-secret")

	config := New()

	assert.Equal(t, "3000", config.Port)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "custom-jwt-secret", config.JWTSecret)

	// Clean up
	os.Unsetenv("PORT")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("JWT_SECRET")
}

func TestGetEnv(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_VAR", "test_value")
	result := getEnv("TEST_VAR", "default_value")
	assert.Equal(t, "test_value", result)

	// Test with non-existing environment variable
	result = getEnv("NON_EXISTING_VAR", "default_value")
	assert.Equal(t, "default_value", result)

	// Clean up
	os.Unsetenv("TEST_VAR")
}
