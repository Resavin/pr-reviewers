package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Resavin/pr-reviewers/internal/domain"
	"github.com/Resavin/pr-reviewers/internal/generated"
	"github.com/Resavin/pr-reviewers/internal/repository"
)

func (s *Server) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	var body generated.PostUsersSetIsActiveJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "invalid request body")
		return
	}
	if body.UserId == "" {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "user_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	user, err := s.userSvc.SetIsActive(ctx, body.UserId, body.IsActive)
	if err != nil {
		if err == repository.ErrUserNotFound {
			writeError(w, http.StatusNotFound, generated.NOTFOUND, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
		return
	}

	resp := struct {
		User generated.User `json:"user"`
	}{
		User: toGeneratedUser(user),
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) GetUsersGetReview(
	w http.ResponseWriter,
	r *http.Request,
	params generated.GetUsersGetReviewParams,
) {
	if params.UserId == "" {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "user_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if _, err := s.userSvc.GetUser(ctx, params.UserId); err != nil {
		if err == repository.ErrUserNotFound {
			writeError(w, http.StatusNotFound, generated.NOTFOUND, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
		return
	}

	prs, err := s.prSvc.ListForReviewer(ctx, params.UserId)
	if err != nil {
		writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
		return
	}

	short := make([]generated.PullRequestShort, 0, len(prs))
	for _, p := range prs {
		short = append(short, toGeneratedPullRequestShort(p))
	}

	resp := struct {
		UserID       string                       `json:"user_id"`
		PullRequests []generated.PullRequestShort `json:"pull_requests"`
	}{
		UserID:       params.UserId,
		PullRequests: short,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) GetUsersStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	users, err := s.userSvc.UsersStats(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
		return
	}

	resp := []struct {
		UserID string `json:"user_id"`
		Count  int64  `json:"count"`
	}{}

	for _, u := range users {
		resp = append(resp, struct {
			UserID string `json:"user_id"`
			Count  int64  `json:"count"`
		}{
			UserID: u.UserID,
			Count:  u.Count,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func toGeneratedUser(u domain.User) generated.User {
	return generated.User{
		UserId:   u.UserID,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

func toGeneratedPullRequestShort(pr domain.PullRequest) generated.PullRequestShort {
	return generated.PullRequestShort{
		PullRequestId:   pr.ID,
		PullRequestName: pr.Name,
		AuthorId:        pr.AuthorID,
		Status:          generated.PullRequestShortStatus(pr.Status),
	}
}
