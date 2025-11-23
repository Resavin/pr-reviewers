package service

import (
	"context"
	"errors"

	"github.com/Resavin/pr-reviewers/internal/domain"
	"github.com/Resavin/pr-reviewers/internal/repository"
)

type TeamService interface {
	CreateOrReplaceTeam(ctx context.Context, teamName string, members []domain.User) (domain.Team, []domain.User, error)
	GetTeam(ctx context.Context, teamName string) (domain.Team, []domain.User, error)
	DeactivateTeamAndReassign(ctx context.Context, fromTeam, toTeam string) (from string, to string, deactivated []string, reassigned int, err error)
}

type teamService struct {
	teamRepo repository.TeamRepository
	userSvc  UserService
	prSvc    PullRequestService
}

func NewTeamService(teamRepo repository.TeamRepository, userSvc UserService, prSvc PullRequestService) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userSvc:  userSvc,
		prSvc:    prSvc,
	}
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

func (s *teamService) DeactivateTeamAndReassign(
	ctx context.Context,
	fromTeam, toTeam string,
) (string, string, []string, int, error) {
	if _, _, err := s.GetTeam(ctx, fromTeam); err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			return "", "", nil, 0, repository.ErrTeamNotFound
		}
		return "", "", nil, 0, err
	}

	if _, _, err := s.GetTeam(ctx, toTeam); err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			return "", "", nil, 0, repository.ErrTeamNotFound
		}
		return "", "", nil, 0, err
	}

	deactivated, err := s.userSvc.DeactivateByTeam(ctx, fromTeam)
	if err != nil {
		return "", "", nil, 0, err
	}
	if len(deactivated) == 0 {
		return fromTeam, toTeam, nil, 0, nil
	}

	assignments, err := s.prSvc.FindOpenAssignmentsForUsers(ctx, deactivated)
	if err != nil {
		return "", "", nil, 0, err
	}
	if len(assignments) == 0 {
		return fromTeam, toTeam, deactivated, 0, nil
	}

	// candidates from NEW team
	candidates, err := s.userSvc.ActiveByTeamExcept(ctx, toTeam, deactivated)
	if err != nil {
		return "", "", nil, 0, err
	}
	if len(candidates) == 0 {
		return "", "", nil, 0, repository.ErrNoCandidatesInNewTeam
	}

	changes := make([]repository.ReassignChange, 0, len(assignments))
	ci := 0

	for _, a := range assignments {
		newReviewer := candidates[ci%len(candidates)]
		ci++

		if newReviewer == a.ReviewerID {
			continue
		}

		changes = append(changes, repository.ReassignChange{
			PullRequestID: a.PullRequestID,
			OldReviewerID: a.ReviewerID,
			NewReviewerID: newReviewer,
		})
	}

	if len(changes) == 0 {
		return fromTeam, toTeam, deactivated, 0, nil
	}

	reassigned, err := s.prSvc.BulkReassignReviewers(ctx, changes)
	if err != nil {
		return "", "", nil, 0, err
	}

	return fromTeam, toTeam, deactivated, reassigned, nil
}
