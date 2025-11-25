package api

import (
	"AvitoTestTask/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	pruc "AvitoTestTask/internal/usecases/pullrequest"
	teamuc "AvitoTestTask/internal/usecases/team"
	useruc "AvitoTestTask/internal/usecases/user"
)

type Server struct {
	teamSvc teamuc.Service
	userSvc useruc.Service
	prSvc   pruc.Service

	r   *chi.Mux
	srv *http.Server
}

func NewServer(teamSvc teamuc.Service, userSvc useruc.Service, prSvc pruc.Service) *Server {
	r := chi.NewRouter()
	s := &Server{teamSvc: teamSvc, userSvc: userSvc, prSvc: prSvc, r: r}
	r.Post("/team/add", s.handleTeamAdd)
	r.Post("/pullRequest/create", s.handlePRCreate)
	r.Post("/pullRequest/reassign", s.handlePRReassign)
	r.Post("/pullRequest/merge", s.handlePRMerge)
	r.Get("/reviewer/{reviewer_id}/pullRequests", s.handleReviewerPRs)

	r.Post("/user/create", s.handleUserCreate)
	r.Get("/user/{user_id}", s.handleUserGet)
	r.Put("/user/update", s.handleUserUpdate)
	r.Delete("/user/{user_id}", s.handleUserDelete)
	r.Put("/team/update", s.handleTeamUpdate)
	r.Delete("/team/{team_name}", s.handleTeamDelete)

	s.srv = &http.Server{
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return s
}

func (s *Server) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.srv.Serve(ln)
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func decodeStrict(r *http.Request, v interface{}) error {
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body: %v", err)
		}
	}()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, httpCode int, code, message string) {
	writeJSON(w, httpCode, ErrorResponse{Error: ErrorObject{Code: code, Message: message}})
}

func (s *Server) handleTeamAdd(w http.ResponseWriter, r *http.Request) {
	var req TeamAddRequest
	if err := decodeStrict(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
		return
	}
	_, err := s.teamSvc.CreateTeam(r.Context(), req.TeamName)
	if err != nil {
		writeError(w, http.StatusBadRequest, "TEAM_CREATE_FAILED", err.Error())
		return
	}

	for _, uid := range req.Users {
		_, err := s.userSvc.GetUser(r.Context(), uid)
		if err != nil {
			writeError(w, http.StatusBadRequest, "USER_NOT_FOUND", "user not found: "+uid)
			return
		}
		if err := s.userSvc.UpdateUser(r.Context(), domain.User{ID: uid, TeamName: &req.TeamName}); err != nil {
			writeError(w, http.StatusInternalServerError, "USER_UPDATE_FAILED", err.Error())
			return
		}
	}
	writeJSON(w, http.StatusCreated, map[string]string{"team_name": req.TeamName})
}

func (s *Server) handlePRCreate(w http.ResponseWriter, r *http.Request) {
	var req PullRequestCreateRequest
	if err := decodeStrict(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
		return
	}
	pr, err := s.prSvc.CreatePRWithAssignments(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		writeError(w, http.StatusConflict, "PR_CREATE_FAILED", err.Error())
		return
	}
	resp := PullRequestResponse{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
		Reviewers:       pr.AssignedReviewers,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (s *Server) handlePRReassign(w http.ResponseWriter, r *http.Request) {
	var req PullRequestReassignRequest
	if err := decodeStrict(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
		return
	}
	newID, pr, err := s.prSvc.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPRMerged):
			writeError(w, http.StatusConflict, "PR_MERGED", err.Error())
		case errors.Is(err, domain.ErrReviewerNotAssigned):
			writeError(w, http.StatusConflict, "NOT_ASSIGNED", err.Error())
		case errors.Is(err, domain.ErrNoCandidate):
			writeError(w, http.StatusConflict, "NO_CANDIDATE", err.Error())
		default:
			writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		}
		return
	}
	resp := PullRequestResponse{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
		Reviewers:       pr.AssignedReviewers,
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"replaced_by": newID, "pr": resp})
}

func (s *Server) handlePRMerge(w http.ResponseWriter, r *http.Request) {
	var req PullRequestMergeRequest
	if err := decodeStrict(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
		return
	}
	pr, err := s.prSvc.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	resp := PullRequestResponse{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
		Reviewers:       pr.AssignedReviewers,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleReviewerPRs(w http.ResponseWriter, r *http.Request) {
	reviewerID := chi.URLParam(r, "reviewer_id")
	prs, err := s.prSvc.GetPRsForReviewer(r.Context(), reviewerID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	out := ReviewerPullRequestsResponse{}
	for _, p := range prs {
		out.PullRequests = append(out.PullRequests, PullRequestResponse{
			PullRequestID:   p.ID,
			PullRequestName: p.Name,
			AuthorID:        p.AuthorID,
			Status:          string(p.Status),
			Reviewers:       p.AssignedReviewers,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := decodeStrict(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	u := domain.User{
		ID:       req.UserID,
		Username: req.Username,
		TeamName: req.TeamID,
		IsActive: isActive,
	}
	if err := s.userSvc.CreateUser(r.Context(), u); err != nil {
		writeError(w, http.StatusBadRequest, "USER_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"user_id": u.ID})
}

func (s *Server) handleUserGet(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	u, err := s.userSvc.GetUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	var resp GetUserResponse
	resp.User.UserID = u.ID
	resp.User.Username = u.Username
	resp.User.TeamID = u.TeamName
	resp.User.IsActive = u.IsActive
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleUserUpdate(w http.ResponseWriter, r *http.Request) {
	var req UpdateUserRequest
	if err := decodeStrict(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
		return
	}
	existing, err := s.userSvc.GetUser(r.Context(), req.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	if req.Username != nil {
		existing.Username = *req.Username
	}
	if req.TeamID != nil {
		existing.TeamName = req.TeamID
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if err := s.userSvc.UpdateUser(r.Context(), *existing); err != nil {
		writeError(w, http.StatusBadRequest, "USER_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (s *Server) handleUserDelete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if err := s.userSvc.DeleteUser(r.Context(), userID); err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (s *Server) handleTeamUpdate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamName    string `json:"team_name"`
		NewTeamName string `json:"new_team_name"`
	}
	if err := decodeStrict(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request")
		return
	}
	if err := s.teamSvc.UpdateTeam(r.Context(), req.TeamName, req.NewTeamName); err != nil {
		writeError(w, http.StatusBadRequest, "TEAM_UPDATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"team_name": req.NewTeamName})
}

func (s *Server) handleTeamDelete(w http.ResponseWriter, r *http.Request) {
	teamName := chi.URLParam(r, "team_name")
	if err := s.teamSvc.DeleteTeam(r.Context(), teamName); err != nil {
		writeError(w, http.StatusBadRequest, "TEAM_DELETE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
