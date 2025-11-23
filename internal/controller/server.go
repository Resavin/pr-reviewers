package controller

import (
	"github.com/Resavin/pr-reviewers/internal/generated"
	"github.com/Resavin/pr-reviewers/internal/service"
	"net/http"
)

type Server struct {
	teamSvc service.TeamService
	userSvc service.UserService
	prSvc   service.PullRequestService
}

var _ generated.ServerInterface = (*Server)(nil)

func NewServer(
	teamSvc service.TeamService,
	userSvc service.UserService,
	prSvc service.PullRequestService,
) *Server {
	return &Server{
		teamSvc: teamSvc,
		userSvc: userSvc,
		prSvc:   prSvc,
	}
}

func (s *Server) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
