package config

import (
	"os"
	"testing"
	"time"

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

func TestMaintenanceConfig(t *testing.T) {
	t.Setenv("SUPABASE_JWT_SECRET", "test-secret")

	t.Run("未設定ならメンテナンスモードはオフ", func(t *testing.T) {
		os.Unsetenv("MAINTENANCE_MODE")
		os.Unsetenv("MAINTENANCE_ENDS_AT")

		cfg, err := New()

		assert.NoError(t, err)
		assert.False(t, cfg.MaintenanceMode)
		assert.Nil(t, cfg.MaintenanceEndsAt)
	})

	t.Run("MAINTENANCE_MODE=true で有効、終了予定はRFC3339で読む", func(t *testing.T) {
		t.Setenv("MAINTENANCE_MODE", "true")
		t.Setenv("MAINTENANCE_ENDS_AT", "2026-10-25T15:00:00+09:00")

		cfg, err := New()

		assert.NoError(t, err)
		assert.True(t, cfg.MaintenanceMode)
		if assert.NotNil(t, cfg.MaintenanceEndsAt) {
			assert.True(t, cfg.MaintenanceEndsAt.Equal(time.Date(2026, 10, 25, 6, 0, 0, 0, time.UTC)))
		}
	})

	t.Run("終了予定が不正でも起動は止めない", func(t *testing.T) {
		t.Setenv("MAINTENANCE_MODE", "true")
		t.Setenv("MAINTENANCE_ENDS_AT", "tomorrow")

		cfg, err := New()

		assert.NoError(t, err)
		assert.True(t, cfg.MaintenanceMode)
		assert.Nil(t, cfg.MaintenanceEndsAt)
	})
}
