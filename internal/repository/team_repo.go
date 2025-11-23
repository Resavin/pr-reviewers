package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Resavin/pr-reviewers/internal/domain"
)

var (
	ErrTeamExists            = errors.New("team already exists")
	ErrTeamNotFound          = errors.New("team not found")
	ErrNoCandidatesInNewTeam = errors.New("no active candidates in new team")
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, name string) error
	UpsertMembers(ctx context.Context, teamName string, users []domain.User) error
	GetTeamWithMembers(ctx context.Context, name string) (domain.Team, []domain.User, error)
}

type teamRepo struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) TeamRepository {
	return &teamRepo{db: db}
}

func (r *teamRepo) CreateTeam(ctx context.Context, name string) error {
	cmdTag, err := r.db.Exec(ctx,
		`INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING`,
		name,
	)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrTeamExists
	}

	return nil
}

func (r *teamRepo) UpsertMembers(ctx context.Context, teamName string, users []domain.User) error {
	if len(users) == 0 {
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
INSERT INTO users (user_id, username, is_active, team_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE
SET username = EXCLUDED.username,
    is_active = EXCLUDED.is_active,
    team_name = EXCLUDED.team_name
`
	for _, u := range users {
		_, err := tx.Exec(ctx, q, u.UserID, u.Username, u.IsActive, teamName)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	tx = nil

	return nil
}

func (r *teamRepo) GetTeamWithMembers(ctx context.Context, name string) (domain.Team, []domain.User, error) {
	var t domain.Team

	err := r.db.QueryRow(ctx,
		`SELECT team_name FROM teams WHERE team_name = $1`,
		name,
	).Scan(&t.TeamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Team{}, nil, ErrTeamNotFound
		}
		return domain.Team{}, nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT user_id, username, is_active, team_name
         FROM users
         WHERE team_name = $1
         ORDER BY user_id`,
		name,
	)
	if err != nil {
		return domain.Team{}, nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.IsActive, &u.TeamName); err != nil {
			return domain.Team{}, nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return domain.Team{}, nil, err
	}

	return t, users, nil
}
