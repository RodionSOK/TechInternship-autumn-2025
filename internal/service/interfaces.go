package service

import (
	"pr-reviewer-service/internal/domain"
)

type TeamService interface {
	CreateTeam(team *domain.Team) error
	GetTeam(teamName string) (*domain.Team, error)
}

type UserService interface {
	SetIsActive(userId string, isActive bool) (*domain.User, error)
	GetUserReviews(userId string) ([]*domain.PullRequestShort, error)
}

type PRService interface {
	CreatePR(prId, prName, authorId string) (*domain.PullRequest, error)
	MergePR(prId string) (*domain.PullRequest, error)
	ReassignReviewer(prId, oldReviewerId string) (*domain.PullRequest, string, error)
}