package services

import "github.com/google/uuid"

// UserService layer to interact with user.
type UserService struct {
}

// NewUserService is constructor function to create UserService.
func NewUserService() *UserService {
	return &UserService{}
}

// GenerateUniqueID generates unique id. Its use uuid to generate id.
func (service *UserService) GenerateUniqueID() string {
	return uuid.NewString()
}
