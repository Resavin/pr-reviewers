package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Resavin/pr-reviewers/internal/domain"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type StatsByUser struct {
	UserID string
	Count  int64
}

type UserRepository interface {
	GetByID(ctx context.Context, userID string) (domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	ListByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	UsersStats(ctx context.Context) ([]StatsByUser, error)
	DeactivateByTeam(ctx context.Context, teamName string) ([]string, error)
	ActiveByTeamExcept(ctx context.Context, teamName string, exclude []string) ([]string, error)
}

type userRepo struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) GetByID(ctx context.Context, userID string) (domain.User, error) {
	var u domain.User

	err := r.db.QueryRow(ctx,
		`SELECT user_id, username, is_active, team_name
         FROM users
         WHERE user_id = $1`,
		userID,
	).Scan(&u.UserID, &u.Username, &u.IsActive, &u.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, err
	}

	return u, nil
}

func (r *userRepo) SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	var u domain.User

	err := r.db.QueryRow(ctx,
		`UPDATE users
         SET is_active = $2
         WHERE user_id = $1
         RETURNING user_id, username, is_active, team_name`,
		userID,
		isActive,
	).Scan(&u.UserID, &u.Username, &u.IsActive, &u.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrUserNotFound
		}
		return domain.User{}, err
	}

	return u, nil
}

func (r *userRepo) ListByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT user_id, username, is_active, team_name
         FROM users
         WHERE team_name = $1
         ORDER BY user_id`,
		teamName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.IsActive, &u.TeamName); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
func (r *userRepo) UsersStats(ctx context.Context) ([]StatsByUser, error) {
	userRows, err := r.db.Query(ctx,
		`SELECT reviewer_id, COUNT(*) 
         FROM pull_request_reviewers 
         GROUP BY reviewer_id 
         ORDER BY reviewer_id`,
	)
	if err != nil {
		return nil, err
	}
	defer userRows.Close()

	var users []StatsByUser
	for userRows.Next() {
		var s StatsByUser
		if err := userRows.Scan(&s.UserID, &s.Count); err != nil {
			return nil, err
		}
		users = append(users, s)
	}
	if err := userRows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepo) DeactivateByTeam(ctx context.Context, teamName string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`UPDATE users
         SET is_active = FALSE
         WHERE team_name = $1 AND is_active = TRUE
         RETURNING user_id`,
		teamName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *userRepo) ActiveByTeamExcept(ctx context.Context, teamName string, exclude []string) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT user_id
         FROM users
         WHERE team_name = $1 AND is_active = TRUE
           AND user_id <> ALL($2)
         ORDER BY user_id`,
		teamName,
		exclude,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
