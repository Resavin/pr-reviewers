package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Resavin/pr-reviewers/internal/domain"
	"github.com/Resavin/pr-reviewers/internal/generated"
	"github.com/Resavin/pr-reviewers/internal/service"
)

func (s *Server) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	var body generated.PostPullRequestCreateJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "invalid request body")
		return
	}
	if body.PullRequestId == "" || body.PullRequestName == "" || body.AuthorId == "" {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "pull_request_id, pull_request_name and author_id are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	pr, reviewers, err := s.prSvc.Create(ctx, body.PullRequestId, body.PullRequestName, body.AuthorId)
	if err != nil {
		switch err {
		case service.ErrPRExists:
			writeError(w, http.StatusConflict, generated.PREXISTS, "PR id already exists")
			return
		case service.ErrAuthorNotFound:
			writeError(w, http.StatusNotFound, generated.NOTFOUND, "author or team not found")
			return
		default:
			writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
			return
		}
	}

	resp := struct {
		PR generated.PullRequest `json:"pr"`
	}{
		PR: toGeneratedPullRequest(pr, reviewers),
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	var body generated.PostPullRequestMergeJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "invalid request body")
		return
	}
	if body.PullRequestId == "" {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "pull_request_id is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	pr, reviewers, err := s.prSvc.Merge(ctx, body.PullRequestId)
	if err != nil {
		if err == service.ErrPRNotFound {
			writeError(w, http.StatusNotFound, generated.NOTFOUND, "PR not found")
			return
		}
		writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
		return
	}

	resp := struct {
		PR generated.PullRequest `json:"pr"`
	}{
		PR: toGeneratedPullRequest(pr, reviewers),
	}

	// Идемпотентность
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	var body generated.PostPullRequestReassignJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "invalid request body")
		return
	}
	if body.PullRequestId == "" || body.OldUserId == "" {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "pull_request_id and old_user_id are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	pr, reviewers, replacedBy, err := s.prSvc.Reassign(ctx, body.PullRequestId, body.OldUserId)
	if err != nil {
		switch err {
		case service.ErrPRNotFound:
			writeError(w, http.StatusNotFound, generated.NOTFOUND, "PR not found")
			return
		case service.ErrPRMerged:
			writeError(w, http.StatusConflict, generated.PRMERGED, "cannot reassign on merged PR")
			return
		case service.ErrReviewerNotAssigned:
			writeError(w, http.StatusConflict, generated.NOTASSIGNED, "reviewer is not assigned to this PR")
			return
		case service.ErrNoCandidate:
			writeError(w, http.StatusConflict, generated.NOCANDIDATE, "no active replacement candidate in team")
			return
		default:
			writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
			return
		}
	}

	resp := struct {
		PR         generated.PullRequest `json:"pr"`
		ReplacedBy string                `json:"replaced_by"`
	}{
		PR:         toGeneratedPullRequest(pr, reviewers),
		ReplacedBy: replacedBy,
	}

	writeJSON(w, http.StatusOK, resp)
}

func toGeneratedPullRequest(pr domain.PullRequest, reviewers []string) generated.PullRequest {
	createdAt := pr.CreatedAt
	return generated.PullRequest{
		PullRequestId:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorId:          pr.AuthorID,
		Status:            generated.PullRequestStatus(pr.Status),
		AssignedReviewers: reviewers,
		CreatedAt:         createdAt,
		MergedAt:          pr.MergedAt,
	}
}
