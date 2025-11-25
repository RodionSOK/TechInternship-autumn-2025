package service

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPRService_CreatePR(t *testing.T) {
	tests := []struct {
		name          string
		prId          string
		prName        string
		authorId      string
		setupMocks    func(*mocks.MockPRRepository, *mocks.MockUserRepository, *mocks.MockTeamRepository)
		expectedError error
	}{
		{
			name:     "successful creation with reviewers",
			prId:     "pr-1",
			prName:   "Fix bug",
			authorId: "u1",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository, mockUserRepo *mocks.MockUserRepository, mockTeamRepo *mocks.MockTeamRepository) {
				mockPRRepo.On("Exists", "pr-1").Return(false, nil).Once()
				
				author := &domain.User{
					UserId:   "u1",
					Username: "Alice",
					TeamName: "backend",
					IsActive: true,
				}
				mockUserRepo.On("GetById", "u1").Return(author, nil).Once()

				team := &domain.Team{
					TeamName: "backend",
					Members: []domain.TeamMember{
						{UserId: "u1", Username: "Alice", IsActive: true},
						{UserId: "u2", Username: "Bob", IsActive: true},
						{UserId: "u3", Username: "Charlie", IsActive: true},
					},
				}
				mockTeamRepo.On("GetByName", "backend").Return(team, nil).Once()

				mockPRRepo.On("Create", mock.MatchedBy(func(p *domain.PullRequest) bool {
					return p.PullRequestId == "pr-1" &&
						p.PullRequestName == "Fix bug" &&
						p.AuthorId == "u1" &&
						len(p.AssignedReviewers) == 2
				})).Return(nil).Once()

				createdPR := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"u2", "u3"},
					CreatedAt:         &time.Time{},
				}
				mockPRRepo.On("GetById", "pr-1").Return(createdPR, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:     "PR already exists",
			prId:     "pr-1",
			prName:   "Fix bug",
			authorId: "u1",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository, mockUserRepo *mocks.MockUserRepository, mockTeamRepo *mocks.MockTeamRepository) {
				mockPRRepo.On("Exists", "pr-1").Return(true, nil).Once()
			},
			expectedError: domain.NewPRExistsError("pr-1"),
		},
		{
			name:     "author not found",
			prId:     "pr-1",
			prName:   "Fix bug",
			authorId: "nonexistent",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository, mockUserRepo *mocks.MockUserRepository, mockTeamRepo *mocks.MockTeamRepository) {
				mockPRRepo.On("Exists", "pr-1").Return(false, nil).Once()
				mockUserRepo.On("GetById", "nonexistent").Return(nil, nil).Once()
			},
			expectedError: domain.NewNotFoundError("user"),
		},
		{
			name:     "team not found",
			prId:     "pr-1",
			prName:   "Fix bug",
			authorId: "u1",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository, mockUserRepo *mocks.MockUserRepository, mockTeamRepo *mocks.MockTeamRepository) {
				mockPRRepo.On("Exists", "pr-1").Return(false, nil).Once()
				
				author := &domain.User{
					UserId:   "u1",
					Username: "Alice",
					TeamName: "backend",
					IsActive: true,
				}
				mockUserRepo.On("GetById", "u1").Return(author, nil).Once()
				mockTeamRepo.On("GetByName", "backend").Return(nil, nil).Once()
			},
			expectedError: domain.NewNotFoundError("team"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRRepo := mocks.NewMockPRRepository(t)
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockTeamRepo := mocks.NewMockTeamRepository(t)
			tt.setupMocks(mockPRRepo, mockUserRepo, mockTeamRepo)

			service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)
			pr, err := service.CreatePR(tt.prId, tt.prName, tt.authorId)

			if tt.expectedError == nil {
				assert.NoError(t, err)
				assert.NotNil(t, pr)
				assert.Equal(t, tt.prId, pr.PullRequestId)
			} else {
				assert.Error(t, err)
				apiErr, ok := err.(*domain.APIError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError.(*domain.APIError).Code, apiErr.Code)
			}

			mockPRRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
			mockTeamRepo.AssertExpectations(t)
		})
	}
}

func TestPRService_MergePR(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name          string
		prId          string
		setupMocks    func(*mocks.MockPRRepository)
		expectedError error
	}{
		{
			name: "successful merge",
			prId: "pr-1",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository) {
				pr := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"u2"},
					CreatedAt:         &now,
				}
				mockPRRepo.On("GetById", "pr-1").Return(pr, nil).Once()
				mockPRRepo.On("UpdateStatus", "pr-1", domain.PRStatusMerged, mock.AnythingOfType("*time.Time")).Return(nil).Once()
				
				mergedPR := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusMerged,
					AssignedReviewers: []string{"u2"},
					CreatedAt:         &now,
					MergedAt:          &now,
				}
				mockPRRepo.On("GetById", "pr-1").Return(mergedPR, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "PR already merged",
			prId: "pr-1",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository) {
				pr := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusMerged,
					AssignedReviewers: []string{"u2"},
					CreatedAt:         &now,
					MergedAt:          &now,
				}
				mockPRRepo.On("GetById", "pr-1").Return(pr, nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "PR not found",
			prId: "nonexistent",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository) {
				mockPRRepo.On("GetById", "nonexistent").Return(nil, nil).Once()
			},
			expectedError: domain.NewNotFoundError("PR"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRRepo := mocks.NewMockPRRepository(t)
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockTeamRepo := mocks.NewMockTeamRepository(t)
			tt.setupMocks(mockPRRepo)

			if tt.name == "successful merge" {
				reviewer := &domain.User{
					UserId:   "u2",
					Username: "Bob",
					TeamName: "backend",
					IsActive: true,
				}
				mockUserRepo.On("GetById", "u2").Return(reviewer, nil).Once()
			}

			service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)
			pr, err := service.MergePR(tt.prId)

			if tt.expectedError == nil {
				assert.NoError(t, err)
				assert.NotNil(t, pr)
				assert.Equal(t, domain.PRStatusMerged, pr.Status)
			} else {
				assert.Error(t, err)
				apiErr, ok := err.(*domain.APIError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError.(*domain.APIError).Code, apiErr.Code)
			}

			mockPRRepo.AssertExpectations(t)
			if tt.name == "successful merge" {
				mockUserRepo.AssertExpectations(t)
			}
		})
	}
}

func TestPRService_ReassignReviewer(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		prId            string
		oldReviewerId   string
		setupMocks      func(*mocks.MockPRRepository, *mocks.MockUserRepository, *mocks.MockTeamRepository)
		expectedNewId   string
		expectedError   error
	}{
		{
			name:          "successful reassignment",
			prId:          "pr-1",
			oldReviewerId: "u2",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository, mockUserRepo *mocks.MockUserRepository, mockTeamRepo *mocks.MockTeamRepository) {
				pr := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"u2", "u3"},
					CreatedAt:         &now,
				}
				mockPRRepo.On("GetById", "pr-1").Return(pr, nil).Once()

				oldReviewer := &domain.User{
					UserId:   "u2",
					Username: "Bob",
					TeamName: "backend",
					IsActive: true,
				}
				mockUserRepo.On("GetById", "u2").Return(oldReviewer, nil).Once()

				team := &domain.Team{
					TeamName: "backend",
					Members: []domain.TeamMember{
						{UserId: "u1", Username: "Alice", IsActive: true},
						{UserId: "u2", Username: "Bob", IsActive: true},
						{UserId: "u3", Username: "Charlie", IsActive: true},
						{UserId: "u4", Username: "David", IsActive: true},
					},
				}
				mockTeamRepo.On("GetByName", "backend").Return(team, nil).Once()

				mockPRRepo.On("ReplaceReviewer", "pr-1", "u2", "u4").Return(nil).Once()

				updatedPR := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"u3", "u4"},
					CreatedAt:         &now,
				}
				mockPRRepo.On("GetById", "pr-1").Return(updatedPR, nil).Times(2)

				reviewer3 := &domain.User{
					UserId:   "u3",
					Username: "Charlie",
					TeamName: "backend",
					IsActive: true,
				}
				mockUserRepo.On("GetById", "u3").Return(reviewer3, nil).Once()

				reviewer4 := &domain.User{
					UserId:   "u4",
					Username: "David",
					TeamName: "backend",
					IsActive: true,
				}
				mockUserRepo.On("GetById", "u4").Return(reviewer4, nil).Once()
			},
			expectedNewId: "u4",
			expectedError: nil,
		},
		{
			name:          "PR is merged",
			prId:          "pr-1",
			oldReviewerId: "u2",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository, mockUserRepo *mocks.MockUserRepository, mockTeamRepo *mocks.MockTeamRepository) {
				pr := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusMerged,
					AssignedReviewers: []string{"u2"},
					CreatedAt:         &now,
					MergedAt:          &now,
				}
				mockPRRepo.On("GetById", "pr-1").Return(pr, nil).Once()
			},
			expectedNewId: "",
			expectedError: domain.NewPRMergedError(),
		},
		{
			name:          "reviewer not assigned",
			prId:          "pr-1",
			oldReviewerId: "u5",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository, mockUserRepo *mocks.MockUserRepository, mockTeamRepo *mocks.MockTeamRepository) {
				pr := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"u2", "u3"},
					CreatedAt:         &now,
				}
				mockPRRepo.On("GetById", "pr-1").Return(pr, nil).Once()
			},
			expectedNewId: "",
			expectedError: domain.NewNotAssignedError(),
		},
		{
			name:          "no candidate found",
			prId:          "pr-1",
			oldReviewerId: "u2",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository, mockUserRepo *mocks.MockUserRepository, mockTeamRepo *mocks.MockTeamRepository) {
				pr := &domain.PullRequest{
					PullRequestId:     "pr-1",
					PullRequestName:   "Fix bug",
					AuthorId:          "u1",
					Status:            domain.PRStatusOpen,
					AssignedReviewers: []string{"u2", "u3"},
					CreatedAt:         &now,
				}
				mockPRRepo.On("GetById", "pr-1").Return(pr, nil).Once()

				oldReviewer := &domain.User{
					UserId:   "u2",
					Username: "Bob",
					TeamName: "backend",
					IsActive: true,
				}
				mockUserRepo.On("GetById", "u2").Return(oldReviewer, nil).Once()

				team := &domain.Team{
					TeamName: "backend",
					Members: []domain.TeamMember{
						{UserId: "u1", Username: "Alice", IsActive: true},
						{UserId: "u2", Username: "Bob", IsActive: true},
						{UserId: "u3", Username: "Charlie", IsActive: true},
					},
				}
				mockTeamRepo.On("GetByName", "backend").Return(team, nil).Once()
			},
			expectedNewId: "",
			expectedError: domain.NewNoCandidateError(),
		},
		{
			name:          "PR not found",
			prId:          "nonexistent",
			oldReviewerId: "u2",
			setupMocks: func(mockPRRepo *mocks.MockPRRepository, mockUserRepo *mocks.MockUserRepository, mockTeamRepo *mocks.MockTeamRepository) {
				mockPRRepo.On("GetById", "nonexistent").Return(nil, nil).Once()
			},
			expectedNewId: "",
			expectedError: domain.NewNotFoundError("PR"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPRRepo := mocks.NewMockPRRepository(t)
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockTeamRepo := mocks.NewMockTeamRepository(t)
			tt.setupMocks(mockPRRepo, mockUserRepo, mockTeamRepo)

			service := NewPRService(mockPRRepo, mockUserRepo, mockTeamRepo)
			pr, newReviewerId, err := service.ReassignReviewer(tt.prId, tt.oldReviewerId)

			if tt.expectedError == nil {
				assert.NoError(t, err)
				assert.NotNil(t, pr)
				assert.Equal(t, tt.expectedNewId, newReviewerId)
			} else {
				assert.Error(t, err)
				apiErr, ok := err.(*domain.APIError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError.(*domain.APIError).Code, apiErr.Code)
			}

			mockPRRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
			mockTeamRepo.AssertExpectations(t)
		})
	}
}