package services

import (
	"context"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userService struct {
	userRepo ports.UserRepository
}

func NewUserService(userRepo ports.UserRepository) ports.UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUserByAddress(ctx context.Context, address string) (*domain.User, error) {
	return s.userRepo.FindByAddress(ctx, address)
}

func (s *userService) Create(ctx context.Context, user *domain.User) error {
	return s.userRepo.Create(ctx, user)
}

func (s *userService) Update(ctx context.Context, user *domain.User) error {
	return s.userRepo.Update(ctx, user)
}

func (s *userService) UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error {
	return s.userRepo.UpdateNonce(ctx, id, nonce)
}
