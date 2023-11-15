package services

import (
	"math/rand"

	"github.com/MowlCoder/go-url-shortener/pkg/base62encode"
)

// StringGenerator responsible for generate different strings.
// For example random string.
type StringGenerator struct {
}

// NewStringGenerator is constructor function to create StringGenerator.
func NewStringGenerator() *StringGenerator {
	return &StringGenerator{}
}

// GenerateRandom generate random string. It is convert random uint64 to string using base 62 encoding.
func (sg *StringGenerator) GenerateRandom() string {
	return base62encode.Base62Encode(rand.Uint64())
}
