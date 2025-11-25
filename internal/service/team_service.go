package service

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repository"
)

type teamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) TeamService {
	return &teamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *teamService) CreateTeam(team *domain.Team) error {
	exists, err := s.teamRepo.Exists(team.TeamName)
	if err != nil {
		return err
	}
	if exists {
		return domain.NewTeamExistsError(team.TeamName)
	}

	err = s.teamRepo.Create(team)
	if err != nil {
		return err
	}
	return nil
}

func (s *teamService) GetTeam(teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetByName(teamName)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, domain.NewNotFoundError("team")
	}
	return team, nil
}

