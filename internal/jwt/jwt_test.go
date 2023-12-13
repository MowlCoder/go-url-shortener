package jwt

import (
	"os"
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

	t.Run("invalid token", func(t *testing.T) {
		_, err := ParseToken("invalid token")
		require.Error(t, err)
	})
}

func TestGetJWTSecretKey(t *testing.T) {
	t.Run("get secret from env (exist in env)", func(t *testing.T) {
		secret := "super-secret-jwt-secret"
		err := os.Setenv(envKeyJWTSecret, secret)
		require.NoError(t, err)
		secretFromENV := getJWTSecretKey()

		assert.Equal(t, secretFromENV, secret)
	})

	t.Run("get secret from env (no exist in env)", func(t *testing.T) {
		err := os.Unsetenv(envKeyJWTSecret)
		require.NoError(t, err)
		secretFromENV := getJWTSecretKey()

		assert.Equal(t, secretFromENV, defaultJWTSecret)
	})
}

func BenchmarkParseToken(b *testing.B) {
	token, _ := GenerateToken("123")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ParseToken(token)
	}
}
