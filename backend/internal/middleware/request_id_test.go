package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRequestID(t *testing.T) {
	id := generateRequestID()

	assert.NotEmpty(t, id)
	// 16バイト → 32文字のhex文字列
	assert.Len(t, id, 32)
}

func TestGenerateRequestID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateRequestID()
		assert.False(t, ids[id], "重複するリクエストIDが生成された: %s", id)
		ids[id] = true
	}
}
