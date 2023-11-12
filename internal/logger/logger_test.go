package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
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

	t.Run("Test new logger (success, production)", func(t *testing.T) {
		l, err := NewLogger(Options{
			Level:        LogInfo,
			IsProduction: true,
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

func TestLogger_Info(t *testing.T) {
	l, err := NewLogger(Options{
		Level:        LogInfo,
		IsProduction: false,
	})
	require.NoError(t, err)

	l.Info("test message")
}
