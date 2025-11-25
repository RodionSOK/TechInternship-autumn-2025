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
	"github.com/stretchr/testify/mock"
)

func TestTeamHandler_CreateTeam(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mocks.MockTeamService)
		expectedStatus int
		expectedError  *domain.APIError
	}{
		{
			name: "successful creation",
			requestBody: map[string]interface{}{
				"team_name": "backend",
				"members": []map[string]interface{}{
					{
						"user_id":   "u1",
						"username":  "Alice",
						"is_active": true,
					},
				},
			},
			setupMocks: func(mockService *mocks.MockTeamService) {
				team := &domain.Team{
					TeamName: "backend",
					Members: []domain.TeamMember{
						{UserId: "u1", Username: "Alice", IsActive: true},
					},
				}
				mockService.On("CreateTeam", mock.AnythingOfType("*domain.Team")).Return(nil).Once()
				mockService.On("GetTeam", "backend").Return(team, nil).Once()
			},
			expectedStatus: http.StatusCreated,
			expectedError:  nil,
		},
		{
			name: "team already exists",
			requestBody: map[string]interface{}{
				"team_name": "backend",
				"members":   []map[string]interface{}{},
			},
			setupMocks: func(mockService *mocks.MockTeamService) {
				mockService.On("CreateTeam", mock.AnythingOfType("*domain.Team")).Return(domain.NewTeamExistsError("backend")).Once()
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  domain.NewTeamExistsError("backend"),
		},
		{
			name: "invalid request body",
			requestBody: "invalid json",
			setupMocks: func(mockService *mocks.MockTeamService) {
				// Мок не вызывается при невалидном JSON
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "missing team_name",
			requestBody: map[string]interface{}{
				"members": []map[string]interface{}{},
			},
			setupMocks: func(mockService *mocks.MockTeamService) {
				// Мок не вызывается при отсутствии team_name
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name: "team not found after creation",
			requestBody: map[string]interface{}{
				"team_name": "backend",
				"members":   []map[string]interface{}{},
			},
			setupMocks: func(mockService *mocks.MockTeamService) {
				mockService.On("CreateTeam", mock.AnythingOfType("*domain.Team")).Return(nil).Once()
				mockService.On("GetTeam", "backend").Return(nil, domain.NewNotFoundError("team")).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.NewNotFoundError("team"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockTeamService(t)
			tt.setupMocks(mockService)

			handler := NewTeamHandler(mockService)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.CreateTeam(w, req)

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

func TestTeamHandler_GetTeam(t *testing.T) {
	tests := []struct {
		name           string
		teamName       string
		setupMocks     func(*mocks.MockTeamService)
		expectedStatus int
		expectedError  *domain.APIError
	}{
		{
			name:     "successful get",
			teamName: "backend",
			setupMocks: func(mockService *mocks.MockTeamService) {
				team := &domain.Team{
					TeamName: "backend",
					Members: []domain.TeamMember{
						{UserId: "u1", Username: "Alice", IsActive: true},
					},
				}
				mockService.On("GetTeam", "backend").Return(team, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name:     "team not found",
			teamName: "nonexistent",
			setupMocks: func(mockService *mocks.MockTeamService) {
				mockService.On("GetTeam", "nonexistent").Return(nil, domain.NewNotFoundError("team")).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  domain.NewNotFoundError("team"),
		},
		{
			name:     "missing team_name parameter",
			teamName: "",
			setupMocks: func(mockService *mocks.MockTeamService) {
				// Мок не вызывается при отсутствии параметра
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  nil,
		},
		{
			name:     "wrong HTTP method",
			teamName: "backend",
			setupMocks: func(mockService *mocks.MockTeamService) {
				// Мок не вызывается при неправильном методе
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockTeamService(t)
			tt.setupMocks(mockService)

			handler := NewTeamHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/team/get", nil)
			if tt.name == "wrong HTTP method" {
				req = httptest.NewRequest(http.MethodPost, "/team/get", nil)
			}
			if tt.teamName != "" {
				q := req.URL.Query()
				q.Add("team_name", tt.teamName)
				req.URL.RawQuery = q.Encode()
			}

			w := httptest.NewRecorder()

			handler.GetTeam(w, req)

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