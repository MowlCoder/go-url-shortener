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
