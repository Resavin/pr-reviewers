package service

import (
	"context"

	"github.com/Resavin/pr-reviewers/internal/domain"
	"github.com/Resavin/pr-reviewers/internal/repository"
)

type TeamService interface {
	CreateOrReplaceTeam(ctx context.Context, teamName string, members []domain.User) (domain.Team, []domain.User, error)
	GetTeam(ctx context.Context, teamName string) (domain.Team, []domain.User, error)
}

type teamService struct {
	teamRepo repository.TeamRepository
}

func NewTeamService(teamRepo repository.TeamRepository) TeamService {
	return &teamService{teamRepo: teamRepo}
}

func (s *teamService) CreateOrReplaceTeam(
	ctx context.Context,
	teamName string,
	members []domain.User,
) (domain.Team, []domain.User, error) {
	if err := s.teamRepo.CreateTeam(ctx, teamName); err != nil {
		if err == repository.ErrTeamExists {
			return domain.Team{}, nil, err
		}
		return domain.Team{}, nil, err
	}

	if err := s.teamRepo.UpsertMembers(ctx, teamName, members); err != nil {
		return domain.Team{}, nil, err
	}

	team, users, err := s.teamRepo.GetTeamWithMembers(ctx, teamName)
	if err != nil {
		return domain.Team{}, nil, err
	}

	return team, users, nil
}

func (s *teamService) GetTeam(ctx context.Context, teamName string) (domain.Team, []domain.User, error) {
	return s.teamRepo.GetTeamWithMembers(ctx, teamName)
}
