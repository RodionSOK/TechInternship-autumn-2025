package service

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_SetIsActive(t *testing.T) {
	tests := []struct {
		name          string
		userId        string
		isActive      bool
		setupMocks    func(*mocks.MockUserRepository)
		expectedUser  *domain.User
		expectedError error
	}{
		{
			name:     "successful activation",
			userId:   "u1",
			isActive: true,
			setupMocks: func(mockUserRepo *mocks.MockUserRepository) {
				user := &domain.User{
					UserId:   "u1",
					Username: "Alice",
					TeamName: "backend",
					IsActive: false,
				}
				updatedUser := &domain.User{
					UserId:   "u1",
					Username: "Alice",
					TeamName: "backend",
					IsActive: true,
				}
				mockUserRepo.On("GetById", "u1").Return(user, nil).Once()
				mockUserRepo.On("UpdateIsActive", "u1", true).Return(nil).Once()
				mockUserRepo.On("GetById", "u1").Return(updatedUser, nil).Once()
			},
			expectedUser: &domain.User{
				UserId:   "u1",
				Username: "Alice",
				TeamName: "backend",
				IsActive: true,
			},
			expectedError: nil,
		},
		{
			name:     "successful deactivation",
			userId:   "u1",
			isActive: false,
			setupMocks: func(mockUserRepo *mocks.MockUserRepository) {
				user := &domain.User{
					UserId:   "u1",
					Username: "Alice",
					TeamName: "backend",
					IsActive: true,
				}
				updatedUser := &domain.User{
					UserId:   "u1",
					Username: "Alice",
					TeamName: "backend",
					IsActive: false,
				}
				mockUserRepo.On("GetById", "u1").Return(user, nil).Once()
				mockUserRepo.On("UpdateIsActive", "u1", false).Return(nil).Once()
				mockUserRepo.On("GetById", "u1").Return(updatedUser, nil).Once()
			},
			expectedUser: &domain.User{
				UserId:   "u1",
				Username: "Alice",
				TeamName: "backend",
				IsActive: false,
			},
			expectedError: nil,
		},
		{
			name:     "user not found",
			userId:   "nonexistent",
			isActive: true,
			setupMocks: func(mockUserRepo *mocks.MockUserRepository) {
				mockUserRepo.On("GetById", "nonexistent").Return(nil, nil).Once()
			},
			expectedUser:  nil,
			expectedError: domain.NewNotFoundError("user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockPRRepo := mocks.NewMockPRRepository(t)
			tt.setupMocks(mockUserRepo)

			service := NewUserService(mockUserRepo, mockPRRepo)
			user, err := service.SetIsActive(tt.userId, tt.isActive)

			if tt.expectedError == nil {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.UserId, user.UserId)
				assert.Equal(t, tt.expectedUser.IsActive, user.IsActive)
			} else {
				assert.Error(t, err)
				apiErr, ok := err.(*domain.APIError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError.(*domain.APIError).Code, apiErr.Code)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserReviews(t *testing.T) {
	tests := []struct {
		name          string
		userId        string
		setupMocks    func(*mocks.MockUserRepository, *mocks.MockPRRepository)
		expectedPRs   []*domain.PullRequestShort
		expectedError error
	}{
		{
			name:   "successful get reviews",
			userId: "u1",
			setupMocks: func(mockUserRepo *mocks.MockUserRepository, mockPRRepo *mocks.MockPRRepository) {
				mockUserRepo.On("Exists", "u1").Return(true, nil).Once()
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
				mockPRRepo.On("GetByReviewerId", "u1").Return(prs, nil).Once()
			},
			expectedPRs: []*domain.PullRequestShort{
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
			},
			expectedError: nil,
		},
		{
			name:   "empty reviews list",
			userId: "u1",
			setupMocks: func(mockUserRepo *mocks.MockUserRepository, mockPRRepo *mocks.MockPRRepository) {
				mockUserRepo.On("Exists", "u1").Return(true, nil).Once()
				mockPRRepo.On("GetByReviewerId", "u1").Return([]*domain.PullRequestShort{}, nil).Once()
			},
			expectedPRs:   []*domain.PullRequestShort{},
			expectedError: nil,
		},
		{
			name:   "user not found",
			userId: "nonexistent",
			setupMocks: func(mockUserRepo *mocks.MockUserRepository, mockPRRepo *mocks.MockPRRepository) {
				mockUserRepo.On("Exists", "nonexistent").Return(false, nil).Once()
			},
			expectedPRs:   nil,
			expectedError: domain.NewNotFoundError("user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockPRRepo := mocks.NewMockPRRepository(t)
			tt.setupMocks(mockUserRepo, mockPRRepo)

			service := NewUserService(mockUserRepo, mockPRRepo)
			prs, err := service.GetUserReviews(tt.userId)

			if tt.expectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedPRs), len(prs))
				for i, expectedPR := range tt.expectedPRs {
					assert.Equal(t, expectedPR.PullRequestId, prs[i].PullRequestId)
					assert.Equal(t, expectedPR.PullRequestName, prs[i].PullRequestName)
				}
			} else {
				assert.Error(t, err)
				apiErr, ok := err.(*domain.APIError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError.(*domain.APIError).Code, apiErr.Code)
			}

			mockUserRepo.AssertExpectations(t)
			mockPRRepo.AssertExpectations(t)
		})
	}
}