package middlewares

import (
	"compress/gzip"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipCompress(t *testing.T) {
	t.Run("is compressible", func(t *testing.T) {
		w := httptest.NewRecorder()
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)

		require.NoError(t, err)

		compress := newGzipWriter(w, gz)
		compress.Close()

		compress.Header().Set("Content-Type", "application/json;charset=utf-8")
		assert.Equal(t, true, compress.isCompressible())
	})

	t.Run("is compressible (invalid)", func(t *testing.T) {
		w := httptest.NewRecorder()
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)

		require.NoError(t, err)

		compress := newGzipWriter(w, gz)
		compress.Close()

		compress.Header().Set("Content-Type", "text/plain")
		assert.Equal(t, false, compress.isCompressible())
	})

	t.Run("write header", func(t *testing.T) {
		w := httptest.NewRecorder()
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)

		require.NoError(t, err)

		compress := newGzipWriter(w, gz)
		compress.Close()

		compress.Header().Set("Content-Type", "application/json")
		compress.WriteHeader(201)

		assert.Equal(t, "gzip", compress.Header().Get("Content-Encoding"))
	})
}
