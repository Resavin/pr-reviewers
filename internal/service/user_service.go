package service

import (
	"context"

	"github.com/Resavin/pr-reviewers/internal/domain"
	"github.com/Resavin/pr-reviewers/internal/repository"
)

type UserService interface {
	GetUser(ctx context.Context, userID string) (domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) GetUser(ctx context.Context, userID string) (domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *userService) SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	return s.userRepo.SetIsActive(ctx, userID, isActive)
}
