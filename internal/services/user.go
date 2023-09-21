package services

import "github.com/google/uuid"

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (service *UserService) GenerateUniqueID() string {
	return uuid.NewString()
}
