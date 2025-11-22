package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Resavin/pr-reviewers/internal/domain"
)

var (
	ErrPRNotFound = errors.New("pull request not found")
)

type PullRequestRepository interface {
	Exists(ctx context.Context, prID string) (bool, error)
	Create(ctx context.Context, pr domain.PullRequest) error
	AddReviewers(ctx context.Context, prID string, reviewerIDs []string) error
	GetWithReviewers(ctx context.Context, prID string) (domain.PullRequest, []string, error)
	SetStatusMerged(ctx context.Context, prID string) (domain.PullRequest, []string, error)
	ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error
	ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
}

type prRepo struct {
	db *pgxpool.Pool
}

func NewPullRequestRepository(db *pgxpool.Pool) PullRequestRepository {
	return &prRepo{db: db}
}

func (r *prRepo) Exists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`,
		prID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *prRepo) Create(ctx context.Context, pr domain.PullRequest) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id)
         VALUES ($1, $2, $3)`,
		pr.ID, pr.Name, pr.AuthorID,
	)
	return err
}

func (r *prRepo) AddReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	if len(reviewerIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	const q = `
INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
VALUES ($1, $2)
ON CONFLICT (pull_request_id, reviewer_id) DO NOTHING
`
	for _, rid := range reviewerIDs {
		if _, err := tx.Exec(ctx, q, prID, rid); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	tx = nil

	return nil
}

func (r *prRepo) GetWithReviewers(ctx context.Context, prID string) (domain.PullRequest, []string, error) {
	var pr domain.PullRequest

	err := r.db.QueryRow(ctx,
		`SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
         FROM pull_requests
         WHERE pull_request_id = $1`,
		prID,
	).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PullRequest{}, nil, ErrPRNotFound
		}
		return domain.PullRequest{}, nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT reviewer_id
         FROM pull_request_reviewers
         WHERE pull_request_id = $1
         ORDER BY reviewer_id`,
		prID,
	)
	if err != nil {
		return domain.PullRequest{}, nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return domain.PullRequest{}, nil, err
		}
		reviewers = append(reviewers, rid)
	}
	if err := rows.Err(); err != nil {
		return domain.PullRequest{}, nil, err
	}

	return pr, reviewers, nil
}

// duplicate calls return current status (idempotent)
func (r *prRepo) SetStatusMerged(ctx context.Context, prID string) (domain.PullRequest, []string, error) {
	var pr domain.PullRequest

	err := r.db.QueryRow(ctx,
		`UPDATE pull_requests
         SET status    = 'MERGED',
             merged_at = COALESCE(merged_at, now())
         WHERE pull_request_id = $1
         RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at`,
		prID,
	).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PullRequest{}, nil, ErrPRNotFound
		}
		return domain.PullRequest{}, nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT reviewer_id
         FROM pull_request_reviewers
         WHERE pull_request_id = $1
         ORDER BY reviewer_id`,
		prID,
	)
	if err != nil {
		return domain.PullRequest{}, nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var rid string
		if err := rows.Scan(&rid); err != nil {
			return domain.PullRequest{}, nil, err
		}
		reviewers = append(reviewers, rid)
	}
	if err := rows.Err(); err != nil {
		return domain.PullRequest{}, nil, err
	}

	return pr, reviewers, nil
}

func (r *prRepo) ReplaceReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	cmdTag, err := r.db.Exec(ctx,
		`UPDATE pull_request_reviewers
         SET reviewer_id = $3
         WHERE pull_request_id = $1 AND reviewer_id = $2`,
		prID, oldReviewerID, newReviewerID,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrPRNotFound
	}
	return nil
}

func (r *prRepo) ListByReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.pull_request_id,
                p.pull_request_name,
                p.author_id,
                p.status,
                p.created_at,
                p.merged_at
         FROM pull_requests p
         JOIN pull_request_reviewers r
           ON p.pull_request_id = r.pull_request_id
         WHERE r.reviewer_id = $1
         ORDER BY p.pull_request_id`,
		reviewerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequest
	for rows.Next() {
		var pr domain.PullRequest
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
