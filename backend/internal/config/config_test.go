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

	_, err := New()
	assert.Error(t, err, "SUPABASE_JWT_SECRET is required but not set")
}

func TestNew_CustomEnvironmentVariables(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("SUPABASE_JWT_SECRET", "custom-jwt-secret")

	cfg, err := New()

	assert.NoError(t, err)
	assert.Equal(t, "3000", cfg.Port)
	assert.Equal(t, "production", cfg.Environment)
	assert.Equal(t, "custom-jwt-secret", cfg.SupabaseJWTSecret)
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_VAR", "test_value")
	result := getEnv("TEST_VAR", "default_value")
	assert.Equal(t, "test_value", result)

	result = getEnv("NON_EXISTING_VAR_FOR_TEST", "default_value")
	assert.Equal(t, "default_value", result)
}
