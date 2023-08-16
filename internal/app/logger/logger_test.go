package logger

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewLogger(t *testing.T) {
	t.Run("Test new logger (success)", func(t *testing.T) {
		l, err := NewLogger(Options{
			Level:        LogInfo,
			IsProduction: false,
		})

		require.NoError(t, err)
		require.NotNil(t, l)
	})

	t.Run("Test new logger (fail)", func(t *testing.T) {
		l, err := NewLogger(Options{
			Level: "random",
		})

		require.Error(t, err)
		require.Nil(t, l)
	})
}
