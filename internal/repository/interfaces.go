package repository

import (
	"pr-reviewer-service/internal/domain"
	"time"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetById(userId string) (*domain.User, error)
	GetByTeamName(teamName string) ([]*domain.User, error)
	UpdateIsActive(userId string, isActive bool) error
	Exists(userId string) (bool, error)
}

type TeamRepository interface {
	Create(team *domain.Team) error
	GetByName(teamName string) (*domain.Team, error)
	Exists(teamName string) (bool, error)
}

type PRRepository interface {
	Create(pr *domain.PullRequest) error
	GetById(prId string) (*domain.PullRequest, error)
	GetByReviewerId(reviewerId string) ([]*domain.PullRequestShort, error)
	UpdateStatus(prId string, status domain.PRStatus, mergedAt *time.Time) error
	AssignReviewer(prId string, reviewerId string) error
	ReplaceReviewer(prId string, oldReviewerId string, newReviewerId string) error
	RemoveReviewer(prId string, reviewerId string) error
	Exists(prId string) (bool, error)
}