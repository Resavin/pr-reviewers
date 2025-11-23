package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/Resavin/pr-reviewers/internal/domain"
	"github.com/Resavin/pr-reviewers/internal/generated"
	"github.com/Resavin/pr-reviewers/internal/repository"
)

func (s *Server) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	var req generated.PostTeamAddJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "invalid request body")
		return
	}
	if req.TeamName == "" {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "team_name is required")
		return
	}

	members := make([]domain.User, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, domain.User{
			UserID:   m.UserId,
			Username: m.Username,
			IsActive: m.IsActive,
			TeamName: req.TeamName,
		})
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	team, users, err := s.teamSvc.CreateOrReplaceTeam(ctx, req.TeamName, members)
	if err != nil {
		if err == repository.ErrTeamExists {
			writeError(w, http.StatusBadRequest, generated.TEAMEXISTS, "team_name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
		return
	}

	resp := struct {
		Team generated.Team `json:"team"`
	}{
		Team: toGeneratedTeam(team, users),
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) GetTeamGet(w http.ResponseWriter, r *http.Request, params generated.GetTeamGetParams) {
	if params.TeamName == "" {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "team_name is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	team, users, err := s.teamSvc.GetTeam(ctx, params.TeamName)
	if err != nil {
		if err == repository.ErrTeamNotFound {
			writeError(w, http.StatusNotFound, generated.NOTFOUND, "team not found")
			return
		}
		writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, toGeneratedTeam(team, users))
}

func toGeneratedTeam(t domain.Team, users []domain.User) generated.Team {
	members := make([]generated.TeamMember, 0, len(users))
	for _, u := range users {
		members = append(members, generated.TeamMember{
			UserId:   u.UserID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}
	return generated.Team{
		TeamName: t.TeamName,
		Members:  members,
	}
}

func (s *Server) PostTeamDeactivate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FromTeamName string `json:"from_team_name"`
		ToTeamName   string `json:"to_team_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil ||
		body.FromTeamName == "" || body.ToTeamName == "" {
		writeError(w, http.StatusBadRequest, generated.NOTFOUND, "from_team_name and to_team_name are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 100*time.Millisecond)
	defer cancel()

	from, to, deactivated, reassigned, err := s.teamSvc.DeactivateTeamAndReassign(ctx, body.FromTeamName, body.ToTeamName)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrTeamNotFound):
			writeError(w, http.StatusNotFound, generated.NOTFOUND, "team not found")
			return
		case errors.Is(err, repository.ErrNoCandidatesInNewTeam):
			writeError(w, http.StatusConflict, generated.NOCANDIDATE, "no active candidates in new team")
			return
		default:
			writeError(w, http.StatusInternalServerError, generated.NOTFOUND, "internal error")
			return
		}
	}

	resp := struct {
		FromTeamName        string   `json:"from_team_name"`
		ToTeamName          string   `json:"to_team_name"`
		DeactivatedUsers    []string `json:"deactivated_users"`
		ReassignedReviewers int      `json:"reassigned_reviewers"`
	}{
		FromTeamName:        from,
		ToTeamName:          to,
		DeactivatedUsers:    deactivated,
		ReassignedReviewers: reassigned,
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeError(w http.ResponseWriter, status int, code generated.ErrorResponseErrorCode, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(generated.ErrorResponse{
		Error: struct {
			Code    generated.ErrorResponseErrorCode `json:"code"`
			Message string                           `json:"message"`
		}{
			Code:    code,
			Message: msg,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
