package controller

import (
	"net/http"

	"github.com/Resavin/pr-reviewers/internal/generated"
)

type Server struct{}

// GetTeamGet implements generated.ServerInterface.
func (s *Server) GetTeamGet(w http.ResponseWriter, r *http.Request, params generated.GetTeamGetParams) {
	panic("unimplemented")
}

// PostTeamAdd implements generated.ServerInterface.
func (s *Server) PostTeamAdd(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// GetUsersGetReview implements generated.ServerInterface.
func (s *Server) GetUsersGetReview(w http.ResponseWriter, r *http.Request, params generated.GetUsersGetReviewParams) {
	panic("unimplemented")
}

// PostPullRequestCreate implements generated.ServerInterface.
func (s *Server) PostPullRequestCreate(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// PostPullRequestMerge implements generated.ServerInterface.
func (s *Server) PostPullRequestMerge(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// PostPullRequestReassign implements generated.ServerInterface.
func (s *Server) PostPullRequestReassign(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// PostUsersSetIsActive implements generated.ServerInterface.
func (s *Server) PostUsersSetIsActive(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

func NewServer() *Server {
	return &Server{}
}
