package services

import (
	"context"
	"errors"
	"fmt"
	"proofofpeacemaking/internal/core/domain"
	"proofofpeacemaking/internal/core/ports"

	"strings"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userService struct {
	userRepo ports.UserRepository
	validate *validator.Validate
}

func NewUserService(userRepo ports.UserRepository) ports.UserService {
	return &userService{
		userRepo: userRepo,
		validate: validator.New(),
	}
}

func (s *userService) GetUserByAddress(ctx context.Context, address string) (*domain.User, error) {
	return s.userRepo.GetByAddress(ctx, address)
}

func (s *userService) Create(ctx context.Context, user *domain.User) error {
	// Validate user fields
	if err := s.validate.Struct(user); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			// Create a more user-friendly error message
			var errMsgs []string
			for _, e := range validationErrors {
				switch e.Field() {
				case "Username":
					errMsgs = append(errMsgs, "username must be between 3 and 30 characters")
				case "Email":
					errMsgs = append(errMsgs, "email must be a valid email address")
				}
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errMsgs, ", "))
		}
		return fmt.Errorf("validation failed: %w", err)
	}

	return s.userRepo.Create(ctx, user)
}

func (s *userService) Update(ctx context.Context, user *domain.User) error {
	return s.userRepo.Update(ctx, user)
}

func (s *userService) UpdateNonce(ctx context.Context, id primitive.ObjectID, nonce int) error {
	return s.userRepo.UpdateNonce(ctx, id, nonce)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.userRepo.GetByEmail(ctx, email)
}

func (s *userService) ConnectWallet(ctx context.Context, userID primitive.ObjectID, address string) error {
	return s.userRepo.ConnectWallet(ctx, userID, address)
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.userRepo.Delete(ctx, id)
}
