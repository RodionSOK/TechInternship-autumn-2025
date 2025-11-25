package service

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatisticsService_GetStatistics(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockPRRepository)
		expectedStats *domain.Statistics
		expectedError error
	}{
		{
			name: "successful get statistics",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository) {
				mockPRRepo.On("GetPRStatistics").Return(10, 3, 7, nil)
				mockPRRepo.On("GetUserReviewStatistics").Return([]*domain.UserStatistics{
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
				}, nil)
			},
			expectedStats: &domain.Statistics{
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
			expectedError: nil,
		},
		{
			name: "empty statistics",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository) {
				mockPRRepo.On("GetPRStatistics").Return(0, 0, 0, nil)
				mockPRRepo.On("GetUserReviewStatistics").Return([]*domain.UserStatistics{}, nil)
			},
			expectedStats: &domain.Statistics{
				TotalPRs:       0,
				OpenPRs:        0,
				MergedPRs:      0,
				UserStatistics: []*domain.UserStatistics{},
			},
			expectedError: nil,
		},
		{
			name: "error getting PR statistics",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository) {
				mockPRRepo.On("GetPRStatistics").Return(0, 0, 0, assert.AnError)
			},
			expectedStats: nil,
			expectedError: assert.AnError,
		},
		{
			name: "error getting user review statistics",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository) {
				mockPRRepo.On("GetPRStatistics").Return(10, 3, 7, nil)
				mockPRRepo.On("GetUserReviewStatistics").Return(nil, assert.AnError)
			},
			expectedStats: nil,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRRepo := mocks.NewMockPRRepository(t)
			tt.setupMocks(mockPRRepo)

			service := NewStatisticsService(mockPRRepo)
			stats, err := service.GetStatistics()

			if tt.expectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStats, stats)
			} else {
				assert.Error(t, err)
				assert.Nil(t, stats)
			}

			mockPRRepo.AssertExpectations(t)
		})
	}
}

