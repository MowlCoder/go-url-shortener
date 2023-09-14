package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_GenerateUniqueID(t *testing.T) {
	userService := NewUserService()

	t.Run("generate unique id", func(t *testing.T) {
		firstUserID := userService.GenerateUniqueID()
		assert.NotEmpty(t, firstUserID)

		secondUserID := userService.GenerateUniqueID()
		assert.NotEmpty(t, secondUserID)

		assert.NotEqual(t, firstUserID, secondUserID)
	})
}
