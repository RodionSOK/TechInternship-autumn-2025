package domain

type ErrorCode string

const (
	ErrorCodeTeamExists ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound ErrorCode = "NOT_FOUND"
)

type APIError struct {
	Code ErrorCode
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func NewTeamExistsError(teamName string) *APIError {
	return &APIError{
		Code: ErrorCodeTeamExists,
		Message: "team_name already exists",
	}
}

func NewPRExistsError(prID string) *APIError {
	return &APIError{
		Code: ErrorCodePRExists,
		Message: "PR id already exists",
	}
}

func NewPRMergedError() *APIError {
	return &APIError{
		Code: ErrorCodePRMerged,
		Message: "cannot reassign on merged PR",
	}
}

func NewNotAssignedError() *APIError {
	return &APIError{
		Code: ErrorCodeNotAssigned,
		Message: "reviewer is not assigned to this PR",
	}
}

func NewNoCandidateError() *APIError {
	return &APIError{
		Code: ErrorCodeNoCandidate,
		Message: "no active replacement candidate in team",
	}
}

func NewNotFoundError(resource string) *APIError {
	return &APIError{
		Code: ErrorCodeNotFound,
		Message: resource + " not found",
	}
}