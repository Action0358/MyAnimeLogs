package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLimitQuery(t *testing.T) {
	tests := []struct {
		name         string
		limitStr     string
		defaultLimit int
		maxLimit     int
		expected     int
	}{
		{"正常値", "10", 20, 100, 10},
		{"空文字列 → デフォルト", "", 20, 100, 20},
		{"不正な文字列 → デフォルト", "abc", 20, 100, 20},
		{"0 → デフォルト", "0", 20, 100, 20},
		{"負の値 → デフォルト", "-5", 20, 100, 20},
		{"最大値超過 → 最大値", "200", 20, 100, 100},
		{"最大値ちょうど", "100", 20, 100, 100},
		{"1（最小有効値）", "1", 20, 100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseLimitQuery(tt.limitStr, tt.defaultLimit, tt.maxLimit)
			assert.Equal(t, tt.expected, result)
		})
	}
}
