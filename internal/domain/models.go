package domain

import "time"

type Team struct {
	TeamName string
}

type User struct {
	UserID   string
	Username string
	IsActive bool
	TeamName string
}

type PullRequestStatus string

const (
	StatusOpen   PullRequestStatus = "OPEN"
	StatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID        string
	Name      string
	AuthorID  string
	Status    PullRequestStatus
	CreatedAt *time.Time
	MergedAt  *time.Time
}
