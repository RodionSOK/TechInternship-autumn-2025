package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatisticsHandler_GetStatistics(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		setupMocks     func(*mocks.MockStatisticsService)
		expectedStatus int
		expectedBody   *domain.Statistics
	}{
		{
			name:   "successful get statistics",
			method: http.MethodGet,
			setupMocks: func(mockService *mocks.MockStatisticsService) {
				stats := &domain.Statistics{
					TotalPRs:  10,
					OpenPRs:   3,
					MergedPRs: 7,
					UserStatistics: []*domain.UserStatistics{
						{
							UserId:       "u1",
							Username:     "Alice",
							TotalReviews: 5,
							OpenReviews:  2,
							MergedReviews: 3,
						},
						{
							UserId:       "u2",
							Username:     "Bob",
							TotalReviews: 3,
							OpenReviews:  1,
							MergedReviews: 2,
						},
					},
				}
				mockService.On("GetStatistics").Return(stats, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: &domain.Statistics{
				TotalPRs:  10,
				OpenPRs:   3,
				MergedPRs: 7,
				UserStatistics: []*domain.UserStatistics{
					{
						UserId:       "u1",
						Username:     "Alice",
						TotalReviews: 5,
						OpenReviews:  2,
						MergedReviews: 3,
					},
					{
						UserId:       "u2",
						Username:     "Bob",
						TotalReviews: 3,
						OpenReviews:  1,
						MergedReviews: 2,
					},
				},
			},
		},
		{
			name:   "empty statistics",
			method: http.MethodGet,
			setupMocks: func(mockService *mocks.MockStatisticsService) {
				stats := &domain.Statistics{
					TotalPRs:       0,
					OpenPRs:        0,
					MergedPRs:      0,
					UserStatistics: []*domain.UserStatistics{},
				}
				mockService.On("GetStatistics").Return(stats, nil).Once()
			},
			expectedStatus: http.StatusOK,
			expectedBody: &domain.Statistics{
				TotalPRs:       0,
				OpenPRs:        0,
				MergedPRs:      0,
				UserStatistics: []*domain.UserStatistics{},
			},
		},
		{
			name:   "method not allowed",
			method: http.MethodPost,
			setupMocks: func(mockService *mocks.MockStatisticsService) {
				// Мок не вызывается при неправильном методе
			},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   nil,
		},
		{
			name:   "service error",
			method: http.MethodGet,
			setupMocks: func(mockService *mocks.MockStatisticsService) {
				mockService.On("GetStatistics").Return(nil, domain.NewNotFoundError("resource")).Once()
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockStatisticsService(t)
			tt.setupMocks(mockService)

			handler := NewStatisticsHandler(mockService)
			req := httptest.NewRequest(tt.method, "/statistics", nil)
			w := httptest.NewRecorder()

			handler.GetStatistics(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response domain.Statistics
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, &response)
			}

			mockService.AssertExpectations(t)
		})
	}
}

