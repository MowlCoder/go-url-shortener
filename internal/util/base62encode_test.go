package util

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase62Encode(t *testing.T) {
	t.Run("explicit number", func(t *testing.T) {
		number := uint64(10326702548241138877)
		encodedString := Base62Encode(number)

		assert.Equal(t, "9e4Ljypz0sm", encodedString)
	})

	t.Run("random number", func(t *testing.T) {
		randomNumber := rand.Uint64()
		encodedString := Base62Encode(randomNumber)

		assert.NotEmpty(t, encodedString)
	})
}

func BenchmarkBase62Encode(b *testing.B) {
	number := uint64(10326702548241138877)

	for i := 0; i < b.N; i++ {
		Base62Encode(number)
	}
}
