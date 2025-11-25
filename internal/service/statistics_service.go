package service

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repository"
)

type statisticsService struct {
	prRepo repository.PRRepository
}

func NewStatisticsService(prRepo repository.PRRepository) StatisticsService {
	return &statisticsService{
		prRepo: prRepo,
	}
}

func (s *statisticsService) GetStatistics() (*domain.Statistics, error) {
	total, open, merged, err := s.prRepo.GetPRStatistics()
	if err != nil {
		return nil, err
	}

	userStats, err := s.prRepo.GetUserReviewStatistics()
	if err != nil {
		return nil, err
	}

	return &domain.Statistics{
		TotalPRs:       total,
		OpenPRs:        open,
		MergedPRs:      merged,
		UserStatistics: userStats,
	}, nil
}

