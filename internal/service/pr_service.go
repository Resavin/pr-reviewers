package service

import (
	"context"
	"errors"
	"math/rand"
	"slices"

	"github.com/Resavin/pr-reviewers/internal/domain"
	"github.com/Resavin/pr-reviewers/internal/repository"
)

var (
	ErrPRExists            = errors.New("pull request already exists")
	ErrPRNotFound          = repository.ErrPRNotFound
	ErrPRMerged            = errors.New("pull request already merged")
	ErrReviewerNotAssigned = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate         = errors.New("no active replacement candidate in team")
	ErrAuthorNotFound      = repository.ErrUserNotFound
)

type PullRequestService interface {
	Create(ctx context.Context, prID, name, authorID string) (domain.PullRequest, []string, error)
	Merge(ctx context.Context, prID string) (domain.PullRequest, []string, error)
	Reassign(ctx context.Context, prID, oldReviewerID string) (domain.PullRequest, []string, string, error)
	ListForReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
}

type prService struct {
	prRepo   repository.PullRequestRepository
	userRepo repository.UserRepository
}

func NewPullRequestService(
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
) PullRequestService {
	return &prService{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (s *prService) Create(
	ctx context.Context,
	prID, name, authorID string,
) (domain.PullRequest, []string, error) {
	exists, err := s.prRepo.Exists(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, nil, err
	}
	if exists {
		return domain.PullRequest{}, nil, ErrPRExists
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.PullRequest{}, nil, ErrAuthorNotFound
		}
		return domain.PullRequest{}, nil, err
	}

	members, err := s.userRepo.ListByTeam(ctx, author.TeamName)
	if err != nil {
		return domain.PullRequest{}, nil, err
	}

	var candidateIDs []string
	for _, u := range members {
		if !u.IsActive {
			continue
		}
		if u.UserID == author.UserID {
			continue
		}
		candidateIDs = append(candidateIDs, u.UserID)
	}

	assigned := pickRandomUpTo(candidateIDs, 2)

	pr := domain.PullRequest{
		ID:       prID,
		Name:     name,
		AuthorID: authorID,
		Status:   domain.StatusOpen,
		// created_at/merged_at will be set by db
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		return domain.PullRequest{}, nil, err
	}

	if err := s.prRepo.AddReviewers(ctx, prID, assigned); err != nil {
		return domain.PullRequest{}, nil, err
	}

	pr, reviewers, err := s.prRepo.GetWithReviewers(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, nil, err
	}

	return pr, reviewers, nil
}

func (s *prService) Merge(
	ctx context.Context,
	prID string,
) (domain.PullRequest, []string, error) {
	pr, reviewers, err := s.prRepo.SetStatusMerged(ctx, prID)
	if err != nil {
		if errors.Is(err, repository.ErrPRNotFound) {
			return domain.PullRequest{}, nil, ErrPRNotFound
		}
		return domain.PullRequest{}, nil, err
	}
	return pr, reviewers, nil
}

func (s *prService) Reassign(
	ctx context.Context,
	prID, oldReviewerID string,
) (domain.PullRequest, []string, string, error) {
	pr, reviewers, err := s.prRepo.GetWithReviewers(ctx, prID)
	if err != nil {
		if errors.Is(err, repository.ErrPRNotFound) {
			return domain.PullRequest{}, nil, "", ErrPRNotFound
		}
		return domain.PullRequest{}, nil, "", err
	}

	if pr.Status == domain.StatusMerged {
		return domain.PullRequest{}, nil, "", ErrPRMerged
	}

	if !contains(reviewers, oldReviewerID) {
		return domain.PullRequest{}, nil, "", ErrReviewerNotAssigned
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return domain.PullRequest{}, nil, "", ErrReviewerNotAssigned
		}
		return domain.PullRequest{}, nil, "", err
	}

	teamMembers, err := s.userRepo.ListByTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		return domain.PullRequest{}, nil, "", err
	}

	var candidates []string
	for _, u := range teamMembers {
		if !u.IsActive {
			continue
		}
		if u.UserID == oldReviewerID {
			continue
		}
		if contains(reviewers, u.UserID) {
			continue
		}
		candidates = append(candidates, u.UserID)
	}

	if len(candidates) == 0 {
		return domain.PullRequest{}, nil, "", ErrNoCandidate
	}

	newReviewerID := candidates[rand.Intn(len(candidates))]

	if err := s.prRepo.ReplaceReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		return domain.PullRequest{}, nil, "", err
	}

	pr, reviewers, err = s.prRepo.GetWithReviewers(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, nil, "", err
	}

	return pr, reviewers, newReviewerID, nil
}

func (s *prService) ListForReviewer(
	ctx context.Context,
	reviewerID string,
) ([]domain.PullRequest, error) {
	return s.prRepo.ListByReviewer(ctx, reviewerID)
}

func pickRandomUpTo(ids []string, n int) []string {
	if len(ids) <= n {
		out := make([]string, len(ids))
		copy(out, ids)
		return out
	}

	rand.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })

	out := make([]string, n)
	copy(out, ids[:n])
	return out
}

func contains(xs []string, v string) bool {
	return slices.Contains(xs, v)
}
