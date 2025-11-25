package service

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/mocks"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/assert"
)

func TestTeamService_CreateTeam(t *testing.T) {
	tests := []struct {
		name string
		team *domain.Team
		setupMocks func(*mocks.MockTeamRepository, *mocks.MockUserRepository)
		expectedError error
	}{
		{
			name: "successful creation",
			team: &domain.Team{
				TeamName: "backend",
				Members: []domain.TeamMember{
					{UserId: "u1", Username: "User1", IsActive: true},
				},
			},
			setupMocks: func(mockTeamRepo *mocks.MockTeamRepository, mockUserRepo *mocks.MockUserRepository) {
				mockTeamRepo.On("Exists", "backend").Return(false, nil)
				mockTeamRepo.On("Create", mock.AnythingOfType("*domain.Team")).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "team already exists",
			team: &domain.Team{
				TeamName: "backend",
				Members: []domain.TeamMember{},
			},
			setupMocks: func(mockTeamRepo *mocks.MockTeamRepository, mockUserRepo *mocks.MockUserRepository) {
				mockTeamRepo.On("Exists", "backend").Return(true, nil)
			},
			expectedError: domain.NewTeamExistsError("backend"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTeamRepo := mocks.NewMockTeamRepository(t)
			mockUserRepo := mocks.NewMockUserRepository(t)
			tt.setupMocks(mockTeamRepo, mockUserRepo)

			service := NewTeamService(mockTeamRepo, mockUserRepo)
			err := service.CreateTeam(tt.team)

			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				apiErr, ok := err.(*domain.APIError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError.(*domain.APIError).Code, apiErr.Code)
			}

			mockTeamRepo.AssertExpectations(t)
		})
	}
}

func TestTeamService_GetTeam(t *testing.T) {
	tests := []struct {
		name          string
		teamName      string
		setupMocks    func(*mocks.MockTeamRepository, *mocks.MockUserRepository)
		expectedTeam  *domain.Team
		expectedError error
	}{
		{
			name:     "team found",
			teamName: "backend",
			setupMocks: func(mockTeamRepo *mocks.MockTeamRepository, mockUserRepo *mocks.MockUserRepository) {
				team := &domain.Team{
					TeamName: "backend",
					Members: []domain.TeamMember{
						{UserId: "u1", Username: "User1", IsActive: true},
					},
				}
				mockTeamRepo.On("GetByName", "backend").Return(team, nil)
			},
			expectedTeam: &domain.Team{
				TeamName: "backend",
				Members: []domain.TeamMember{
					{UserId: "u1", Username: "User1", IsActive: true},
				},
			},
			expectedError: nil,
		},
		{
			name:     "team not found",
			teamName: "nonexistent",
			setupMocks: func(mockTeamRepo *mocks.MockTeamRepository, mockUserRepo *mocks.MockUserRepository) {
				mockTeamRepo.On("GetByName", "nonexistent").Return(nil, nil)
			},
			expectedTeam:  nil,
			expectedError: domain.NewNotFoundError("team"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTeamRepo := mocks.NewMockTeamRepository(t)
			mockUserRepo := mocks.NewMockUserRepository(t)
			tt.setupMocks(mockTeamRepo, mockUserRepo)

			service := NewTeamService(mockTeamRepo, mockUserRepo)
			team, err := service.GetTeam(tt.teamName)

			if tt.expectedError == nil {
				assert.NoError(t, err)
				assert.NotNil(t, team)
				assert.Equal(t, tt.expectedTeam.TeamName, team.TeamName)
			} else {
				assert.Error(t, err)
				apiErr, ok := err.(*domain.APIError)
				assert.True(t, ok)
				assert.Equal(t, tt.expectedError.(*domain.APIError).Code, apiErr.Code)
			}

			mockTeamRepo.AssertExpectations(t)
		})
	}
}