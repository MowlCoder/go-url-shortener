package services

import (
	"math/rand"

	"github.com/MowlCoder/go-url-shortener/internal/app/util"
)

type StringGenerator struct {
}

func NewStringGenerator() *StringGenerator {
	return &StringGenerator{}
}

func (sg *StringGenerator) GenerateRandom() string {
	return util.Base62Encode(rand.Uint64())[:6] + util.Base62Encode(rand.Uint64())[:6]
}
