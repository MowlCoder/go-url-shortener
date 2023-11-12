package services

import (
	"math/rand"

	"github.com/MowlCoder/go-url-shortener/pkg/base62encode"
)

type StringGenerator struct {
}

func NewStringGenerator() *StringGenerator {
	return &StringGenerator{}
}

func (sg *StringGenerator) GenerateRandom() string {
	return base62encode.Base62Encode(rand.Uint64())
}
