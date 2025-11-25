package domain

import (
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Code: ErrorCodeNotFound,
		Message: "test error message",
	}

	expected := "test error message"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}


func TestNewTeamExistsError(t *testing.T) {
	err := NewTeamExistsError("team1")

	if err.Code != ErrorCodeTeamExists {
		t.Errorf("Expected '%s', got '%s'", ErrorCodeTeamExists, err.Code)
	}

	expectedMessage := "team_name already exists"
	if err.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, err.Message)
	}
}

func TestNewPRExistsError(t *testing.T) {
	err := NewPRExistsError("pr1")

	if err.Code != ErrorCodePRExists {
		t.Errorf("Expected '%s', got '%s'", ErrorCodePRExists, err.Code)
	}

	expectedMessage := "PR id already exists"
	if err.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, err.Message)
	}
}

func TestNewPRMergedError(t *testing.T) {
	err := NewPRMergedError()

	if err.Code != ErrorCodePRMerged {
		t.Errorf("Expected '%s', got '%s'", ErrorCodePRMerged, err.Code)
	}

	expectedMessage := "cannot reassign on merged PR"
	if err.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, err.Message)
	}
}

func TestNewNotAssignedError(t *testing.T) {
	err := NewNotAssignedError()

	if err.Code != ErrorCodeNotAssigned {
		t.Errorf("Expected '%s', got '%s'", ErrorCodeNotAssigned, err.Code)
	}

	expectedMessage := "reviewer is not assigned to this PR"
	if err.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, err.Message)
	}
}

func TestNewNoCandidateError(t *testing.T) {
	err := NewNoCandidateError()

	if err.Code != ErrorCodeNoCandidate {
		t.Errorf("Expected '%s', got '%s'", ErrorCodeNoCandidate, err.Code)
	}

	expectedMessage := "no active replacement candidate in team"
	if err.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, err.Message)
	}
}

func TestNewNotFoundError_User(t *testing.T) {
	err := NewNotFoundError("user")

	if err.Code != ErrorCodeNotFound {
		t.Errorf("Expected '%s', got '%s'", ErrorCodeNotFound, err.Code)
	}

	expectedMessage := "user not found"
	if err.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, err.Message)
	}
}

func TestNewNotFoundError_PR(t *testing.T) {
	err := NewNotFoundError("pr")

	if err.Code != ErrorCodeNotFound {
		t.Errorf("Expected '%s', got '%s'", ErrorCodeNotFound, err.Code)
	}

	expectedMessage := "pr not found"
	if err.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, err.Message)
	}
}

func TestNewNotFoundError_Team(t *testing.T) {
	err := NewNotFoundError("team")

	if err.Code != ErrorCodeNotFound {
		t.Errorf("Expected '%s', got '%s'", ErrorCodeNotFound, err.Code)
	}

	expectedMessage := "team not found"
	if err.Message != expectedMessage {
		t.Errorf("Expected '%s', got '%s'", expectedMessage, err.Message)
	}
}