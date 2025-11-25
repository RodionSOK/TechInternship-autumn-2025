package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserHandler_SetIsActive(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.MockUserService)
		expectedStatus int
		expectedError  *domain.APIError
	}{
		{
			name: "successful activation",
			requestBody: map[string]interface{}{
				"user_id":   "u1",
				"is_active": true,
			},
			setupMocks: func(mockService *mocks.MockUserService) {
				user := &domain.User{
					UserId:   "u1",
					Username: "Alice",
					TeamName: "backend",
					IsActive: true,
				}
				mockService.On("SetIsActive", "u1", true).Return(user, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name: "successful deactivation",
			requestBody: map[string]interface{}{
				"user_id":   "u1",
				"is_active": false,
			},
			setupMocks: func(mockService *mocks.MockUserService) {
				user := &domain.User{
					UserId:   "u1",
					Username: "Alice",
					TeamName: "backend",
					IsActive: false,
				}
				mockService.On("SetIsActive", "u1", false).Return(user, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name: "user not found",
			requestBody: map[string]interface{}{
				"user_id":   "nonexistent",
				"is_active": true,
			},
			setupMocks: func(mockService *mocks.MockUserService) {
				mockService.On("SetIsActive", "nonexistent", true).Return(nil, domain.NewNotFoundError("user")).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.NewNotFoundError("user"),
		},
		{
			name: "invalid request body",
			requestBody: "invalid json",
			setupMocks: func(mockService *mocks.MockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "missing user_id",
			requestBody: map[string]interface{}{
				"is_active": true,
			},
			setupMocks: func(mockService *mocks.MockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "wrong HTTP method",
			requestBody: map[string]interface{}{
				"user_id":   "u1",
				"is_active": true,
			},
			setupMocks: func(mockService *mocks.MockUserService) {
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockUserService(t)
			tt.setupMocks(mockService)

			handler := NewUserHandler(mockService)

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

			req := httptest.NewRequest(method, "/users/setIsActive", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SetIsActive(w, req)

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

func TestUserHandler_GetUserReviews(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMocks     func(*mocks.MockUserService)
		expectedStatus int
		expectedError  *domain.APIError
	}{
		{
			name:   "successful get reviews",
			userID: "u1",
			setupMocks: func(mockService *mocks.MockUserService) {
				prs := []*domain.PullRequestShort{
					{
						PullRequestId:   "pr-1",
						PullRequestName: "Fix bug",
						AuthorId:        "u2",
						Status:          domain.PRStatusOpen,
					},
					{
						PullRequestId:   "pr-2",
						PullRequestName: "Add feature",
						AuthorId:        "u3",
						Status:          domain.PRStatusOpen,
					},
				}
				mockService.On("GetUserReviews", "u1").Return(prs, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name:   "empty reviews list",
			userID: "u1",
			setupMocks: func(mockService *mocks.MockUserService) {
				mockService.On("GetUserReviews", "u1").Return([]*domain.PullRequestShort{}, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMocks: func(mockService *mocks.MockUserService) {
				mockService.On("GetUserReviews", "nonexistent").Return(nil, domain.NewNotFoundError("user")).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.NewNotFoundError("user"),
		},
		{
			name:   "missing user_id parameter",
			userID: "",
			setupMocks: func(mockService *mocks.MockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name:   "wrong HTTP method",
			userID: "u1",
			setupMocks: func(mockService *mocks.MockUserService) {
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockUserService(t)
			tt.setupMocks(mockService)

			handler := NewUserHandler(mockService)

			method := http.MethodGet
			if tt.name == "wrong HTTP method" {
				method = http.MethodPost
			}

			req := httptest.NewRequest(method, "/users/getReview", nil)
			if tt.userID != "" {
				q := req.URL.Query()
				q.Add("user_id", tt.userID)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()

			handler.GetUserReviews(w, req)

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