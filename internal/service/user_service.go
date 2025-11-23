package service

import (
	"context"

	"github.com/Resavin/pr-reviewers/internal/domain"
	"github.com/Resavin/pr-reviewers/internal/repository"
)

type UserService interface {
	GetUser(ctx context.Context, userID string) (domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	UsersStats(ctx context.Context) ([]repository.StatsByUser, error)
	DeactivateByTeam(ctx context.Context, teamName string) ([]string, error)
	ActiveByTeamExcept(ctx context.Context, teamName string, exclude []string) ([]string, error)
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

func (s *userService) UsersStats(ctx context.Context) ([]repository.StatsByUser, error) {
	users, err := s.userRepo.UsersStats(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userService) DeactivateByTeam(ctx context.Context, teamName string) ([]string, error) {
	return s.userRepo.DeactivateByTeam(ctx, teamName)
}

func (s *userService) ActiveByTeamExcept(ctx context.Context, teamName string, exclude []string) ([]string, error) {
	return s.userRepo.ActiveByTeamExcept(ctx, teamName, exclude)
}
