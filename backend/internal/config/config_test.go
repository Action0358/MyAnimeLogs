package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_DefaultValues(t *testing.T) {
	// デフォルト値テストでは環境変数を「未設定」にする必要がある
	// t.Setenv("", "") は空文字列を設定してしまうため os.Unsetenv を使用
	envKeys := []string{
		"DATABASE_URL", "PORT", "ENVIRONMENT", "JWT_SECRET",
	}
	for _, key := range envKeys {
		os.Unsetenv(key)
	}

	config := New()

	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "super-secret-key-change-me", config.JWTSecret)
	assert.Contains(t, config.DatabaseURL, "postgres://")
}

func TestNew_CustomEnvironmentVariables(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("JWT_SECRET", "custom-jwt-secret")

	config := New()

	assert.Equal(t, "3000", config.Port)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "custom-jwt-secret", config.JWTSecret)
}

func TestGetEnv(t *testing.T) {
	t.Setenv("TEST_VAR", "test_value")
	result := getEnv("TEST_VAR", "default_value")
	assert.Equal(t, "test_value", result)

	// Non-existing variable
	result = getEnv("NON_EXISTING_VAR_FOR_TEST", "default_value")
	assert.Equal(t, "default_value", result)
}
