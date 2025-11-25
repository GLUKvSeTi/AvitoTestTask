package api

type TeamAddRequest struct {
	TeamName string   `json:"team_name"`
	Users    []string `json:"users"`
}

type CreateUserRequest struct {
	UserID   string  `json:"user_id"`
	Username string  `json:"username"`
	TeamID   *string `json:"team_id,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type UpdateUserRequest struct {
	UserID   string  `json:"user_id"`
	Username *string `json:"username,omitempty"`
	TeamID   *string `json:"team_id,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type GetUserResponse struct {
	User struct {
		UserID   string  `json:"user_id"`
		Username string  `json:"username"`
		TeamID   *string `json:"team_id,omitempty"`
		IsActive bool    `json:"is_active"`
	} `json:"user"`
}

type PullRequestCreateRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type PullRequestReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

type PullRequestMergeRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type PullRequestResponse struct {
	PullRequestID   string   `json:"pull_request_id"`
	PullRequestName string   `json:"pull_request_name"`
	AuthorID        string   `json:"author_id"`
	Status          string   `json:"status"`
	Reviewers       []string `json:"reviewers"`
}

type ReviewerPullRequestsResponse struct {
	PullRequests []PullRequestResponse `json:"pull_requests"`
}

type ErrorObject struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorObject `json:"error"`
}
