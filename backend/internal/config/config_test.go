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

func TestNew_RateLimitPerMinute(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", "test-secret")

	t.Run("未設定なら既定値(1分あたり600)", func(t *testing.T) {
		os.Unsetenv("RATE_LIMIT_PER_MINUTE")
		os.Unsetenv("RATE_LIMIT")

		cfg, err := New()

		assert.NoError(t, err)
		assert.Equal(t, 600, cfg.RateLimitPerMinute)
	})

	t.Run("RATE_LIMIT_PER_MINUTE で上書きできる", func(t *testing.T) {
		t.Setenv("RATE_LIMIT_PER_MINUTE", "120")

		cfg, err := New()

		assert.NoError(t, err)
		assert.Equal(t, 120, cfg.RateLimitPerMinute)
	})

	// 旧 RATE_LIMIT は単位が曖昧（コードは毎秒、文書は毎分）だったため読まない。
	// 設定済みの値（例: 60）が毎分60回として突然効き、通常の利用でも429になるのを防ぐ
	t.Run("旧 RATE_LIMIT は無視する", func(t *testing.T) {
		os.Unsetenv("RATE_LIMIT_PER_MINUTE")
		t.Setenv("RATE_LIMIT", "5")

		cfg, err := New()

		assert.NoError(t, err)
		assert.Equal(t, 600, cfg.RateLimitPerMinute)
	})
}
