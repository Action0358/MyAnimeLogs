package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_DefaultValues(t *testing.T) {
	envKeys := []string{
		"DATABASE_URL", "PORT", "ENVIRONMENT", "SUPABASE_JWT_SECRET",
	}
	for _, key := range envKeys {
		os.Unsetenv(key)
	}

	config := New()

	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "", config.SupabaseJWTSecret)
	assert.Contains(t, config.DatabaseURL, "postgres://")
}

func TestNew_CustomEnvironmentVariables(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SUPABASE_JWT_SECRET", "custom-jwt-secret")

	config := New()

	assert.Equal(t, "3000", config.Port)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "custom-jwt-secret", config.SupabaseJWTSecret)
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_VAR", "test_value")
	result := getEnv("TEST_VAR", "default_value")
	assert.Equal(t, "test_value", result)

	result = getEnv("NON_EXISTING_VAR_FOR_TEST", "default_value")
	assert.Equal(t, "default_value", result)
}
