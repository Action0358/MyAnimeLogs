package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected Environment
	}{
		{"production", "production", Production},
		{"staging", "staging", Staging},
		{"test", "test", Test},
		{"未設定 → development", "", Development},
		{"不明な値 → development", "unknown", Development},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("ENVIRONMENT", tt.envValue)
			result := GetEnvironment()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnvironment_Methods(t *testing.T) {
	assert.True(t, Development.IsDevelopment())
	assert.False(t, Development.IsProduction())
	assert.False(t, Development.IsStaging())
	assert.False(t, Development.IsTest())

	assert.True(t, Production.IsProduction())
	assert.False(t, Production.IsDevelopment())

	assert.True(t, Staging.IsStaging())
	assert.False(t, Staging.IsDevelopment())

	assert.True(t, Test.IsTest())
	assert.False(t, Test.IsDevelopment())
}
