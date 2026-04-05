package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateRefreshTokenPair(t *testing.T) {
	token, hash, err := GenerateRefreshTokenPair()
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, token, hash, "トークンとハッシュは異なるべき")
}

func TestGenerateRefreshTokenPair_Uniqueness(t *testing.T) {
	token1, hash1, err1 := GenerateRefreshTokenPair()
	require.NoError(t, err1)

	token2, hash2, err2 := GenerateRefreshTokenPair()
	require.NoError(t, err2)

	assert.NotEqual(t, token1, token2, "生成されるトークンは毎回異なるべき")
	assert.NotEqual(t, hash1, hash2, "生成されるハッシュは毎回異なるべき")
}

func TestHashRefreshToken(t *testing.T) {
	token := "test-refresh-token"

	hash1 := HashRefreshToken(token)
	hash2 := HashRefreshToken(token)

	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2, "同じトークンからは同じハッシュが生成されるべき")
}

func TestHashRefreshToken_DifferentInputs(t *testing.T) {
	hash1 := HashRefreshToken("token-a")
	hash2 := HashRefreshToken("token-b")

	assert.NotEqual(t, hash1, hash2, "異なるトークンからは異なるハッシュが生成されるべき")
}
