package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	t.Run("generate token", func(t *testing.T) {
		token, err := GenerateToken("123")
		require.NoError(t, err)

		assert.NotEmpty(t, token)
	})
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateToken("123")
	}
}

func TestParseToken(t *testing.T) {
	t.Run("parse token", func(t *testing.T) {
		userID := "123"

		token, err := GenerateToken(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := ParseToken(token)
		require.NoError(t, err)

		assert.Equal(t, claims.UserID, userID)
	})
}

func BenchmarkParseToken(b *testing.B) {
	token, _ := GenerateToken("123")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ParseToken(token)
	}
}
