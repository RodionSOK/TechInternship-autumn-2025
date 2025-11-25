package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPRHandler_CreatePR(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.MockPRService)
		expectedStatus int
		expectedError  *domain.APIError
	}{
		{
			name: "successful creation",
			requestBody: map[string]interface{}{
				"pull_request_id":   "pr-1",
				"pull_request_name": "Fix bug",
				"author_id":         "u1",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
				pr := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"u2", "u3"},
					CreatedAt:         &now,
				}
				mockService.On("CreatePR", "pr-1", "Fix bug", "u1").Return(pr, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedError:  nil,
		},
		{
			name: "PR already exists",
			requestBody: map[string]interface{}{
				"pull_request_id":   "pr-1",
				"pull_request_name": "Fix bug",
				"author_id":         "u1",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
				mockService.On("CreatePR", "pr-1", "Fix bug", "u1").Return(nil, domain.NewPRExistsError("pr-1")).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.NewPRExistsError("pr-1"),
		},
		{
			name: "missing pull_request_id",
			requestBody: map[string]interface{}{
				"pull_request_name": "Fix bug",
				"author_id":         "u1",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "missing pull_request_name",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1",
				"author_id":       "u1",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "missing author_id",
			requestBody: map[string]interface{}{
				"pull_request_id":   "pr-1",
				"pull_request_name": "Fix bug",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "invalid request body",
			requestBody: "invalid json",
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "wrong HTTP method",
			requestBody: map[string]interface{}{
				"pull_request_id":   "pr-1",
				"pull_request_name": "Fix bug",
				"author_id":         "u1",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockPRService(t)
			tt.setupMocks(mockService)

			handler := NewPRHandler(mockService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			method := http.MethodPost
			if tt.name == "wrong HTTP method" {
				method = http.MethodGet
			}

			req := httptest.NewRequest(method, "/pullRequest/create", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreatePR(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				errorObj, ok := response["error"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, string(tt.expectedError.Code), errorObj["code"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestPRHandler_MergePR(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.MockPRService)
		expectedStatus int
		expectedError  *domain.APIError
	}{
		{
			name: "successful merge",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
				pr := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusMerged,
					AssignedReviewers: []string{"u2"},
					CreatedAt:         &now,
					MergedAt:          &now,
				}
				mockService.On("MergePR", "pr-1").Return(pr, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name: "PR not found",
			requestBody: map[string]interface{}{
				"pull_request_id": "nonexistent",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
				mockService.On("MergePR", "nonexistent").Return(nil, domain.NewNotFoundError("PR")).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.NewNotFoundError("PR"),
		},
		{
			name: "missing pull_request_id",
			requestBody: map[string]interface{}{},
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "invalid request body",
			requestBody: "invalid json",
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "wrong HTTP method",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockPRService(t)
			tt.setupMocks(mockService)

			handler := NewPRHandler(mockService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			method := http.MethodPost
			if tt.name == "wrong HTTP method" {
				method = http.MethodGet
			}

			req := httptest.NewRequest(method, "/pullRequest/merge", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.MergePR(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				errorObj, ok := response["error"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, string(tt.expectedError.Code), errorObj["code"])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestPRHandler_ReassignReviewer(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.MockPRService)
		expectedStatus int
		expectedError  *domain.APIError
	}{
		{
			name: "successful reassignment",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1",
				"old_user_id":     "u2",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
				pr := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"u3", "u4"},
					CreatedAt:         &now,
				}
				mockService.On("ReassignReviewer", "pr-1", "u2").Return(pr, "u4", nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name: "PR is merged",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1",
				"old_user_id":     "u2",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
				mockService.On("ReassignReviewer", "pr-1", "u2").Return(nil, "", domain.NewPRMergedError()).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.NewPRMergedError(),
		},
		{
			name: "reviewer not assigned",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1",
				"old_user_id":     "u5",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
				mockService.On("ReassignReviewer", "pr-1", "u5").Return(nil, "", domain.NewNotAssignedError()).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.NewNotAssignedError(),
		},
		{
			name: "no candidate found",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1",
				"old_user_id":     "u2",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
				mockService.On("ReassignReviewer", "pr-1", "u2").Return(nil, "", domain.NewNoCandidateError()).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.NewNoCandidateError(),
		},
		{
			name: "PR not found",
			requestBody: map[string]interface{}{
				"pull_request_id": "nonexistent",
				"old_user_id":     "u2",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
				mockService.On("ReassignReviewer", "nonexistent", "u2").Return(nil, "", domain.NewNotFoundError("PR")).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.NewNotFoundError("PR"),
		},
		{
			name: "missing pull_request_id",
			requestBody: map[string]interface{}{
				"old_user_id": "u2",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "missing old_user_id",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "invalid request body",
			requestBody: "invalid json",
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "wrong HTTP method",
			requestBody: map[string]interface{}{
				"pull_request_id": "pr-1",
				"old_user_id":     "u2",
			},
			setupMocks: func(mockService *mocks.MockPRService) {
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockPRService(t)
			tt.setupMocks(mockService)

			handler := NewPRHandler(mockService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			method := http.MethodPost
			if tt.name == "wrong HTTP method" {
				method = http.MethodGet
			}

			req := httptest.NewRequest(method, "/pullRequest/reassign", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ReassignReviewer(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				
				errorObj, ok := response["error"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, string(tt.expectedError.Code), errorObj["code"])
			}

			mockService.AssertExpectations(t)
		})
	}
}