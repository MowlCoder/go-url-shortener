package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringGenerator_GenerateRandom(t *testing.T) {
	stringGenerator := NewStringGenerator()

	t.Run("generate string", func(t *testing.T) {
		generatedString := stringGenerator.GenerateRandom()
		assert.NotEmpty(t, generatedString)
	})
}

func BenchmarkStringGenerator_GenerateRandom(b *testing.B) {
	stringGenerator := NewStringGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stringGenerator.GenerateRandom()
	}
}
