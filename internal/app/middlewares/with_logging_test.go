package middlewares

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingResponseWriter(t *testing.T) {
	t.Run("Test write status", func(t *testing.T) {
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: httptest.NewRecorder(),
			responseData:   responseData,
		}

		lw.WriteHeader(200)

		assert.Equal(t, 200, lw.responseData.status)
	})

	t.Run("Test write size", func(t *testing.T) {
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: httptest.NewRecorder(),
			responseData:   responseData,
		}

		size, err := lw.Write([]byte{1, 2, 3})

		require.NoError(t, err)

		assert.Equal(t, 3, size)
		assert.Equal(t, 3, lw.responseData.size)
	})
}
