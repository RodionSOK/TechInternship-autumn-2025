package service

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repository"
)

type userService struct {
	userRepo repository.UserRepository
	prRepo repository.PRRepository
}

func NewUserService(userRepo repository.UserRepository, prRepo repository.PRRepository) UserService {
	return &userService {
		userRepo: userRepo,
		prRepo: prRepo,
	}
}

func (s *userService) SetIsActive(userId string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.GetById(userId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, domain.NewNotFoundError("user")
	}

	err = s.userRepo.UpdateIsActive(userId, isActive)
	if err != nil {
		return nil, err
	}
	updatedUser, err := s.userRepo.GetById(userId)
	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *userService) GetUserReviews(userId string) ([]*domain.PullRequestShort, error) {
	exists, err := s.userRepo.Exists(userId)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.NewNotFoundError("user")
	}

	prs, err := s.prRepo.GetByReviewerId(userId)
	if err != nil {
		return nil, err
	}

	return prs, nil
}