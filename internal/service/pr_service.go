package service

import (
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repository"
	"time"
)

type prService struct {
	prRepo repository.PRRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
}

func NewPRService(prRepo repository.PRRepository, userRepo repository.UserRepository, teamRepo repository.TeamRepository) PRService {
	return &prService{
		prRepo: prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

func (s *prService) isReviewerAssigned(assignedReviewers []string, reviewerId string) bool {
	for _, reviewer := range assignedReviewers {
		if reviewer == reviewerId {
			return true
		}
	}
	return false
}

func (s *prService) selectReviewers(team *domain.Team, authorId string, excludedReviewers []string, maxCount int) []string {
	var reviewers []string
	excludedMap := make(map[string]bool)

	excludedMap[authorId] = true

	for _, reviewer := range excludedReviewers {
		excludedMap[reviewer] = true
	}

	for _, member := range team.Members {
		if excludedMap[member.UserId] {
			continue
		}
		if member.IsActive {
			reviewers = append(reviewers, member.UserId)
			if len(reviewers) >= maxCount {
				break
			}
		}
	}
	return reviewers
}

func (s *prService) CreatePR(prId, prName, authorId string) (*domain.PullRequest, error) {
	exists, err := s.prRepo.Exists(prId)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.NewPRExistsError(prId)
	}

	author, err := s.userRepo.GetById(authorId)
	if err != nil {
		return nil, err
	}
	if author == nil {
		return nil, domain.NewNotFoundError("user")
	}

	team, err := s.teamRepo.GetByName(author.TeamName)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, domain.NewNotFoundError("team")
	}

	reviewers := s.selectReviewers(team, authorId, []string{}, 2)

	now := time.Now()
	pr := &domain.PullRequest{
		PullRequestId: prId,
		PullRequestName: prName,
		AuthorId: authorId,
		Status: domain.PRStatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt: &now,
	}

	err = s.prRepo.Create(pr)
	if err != nil {
		return nil, err
	}

	createdPR, err := s.prRepo.GetById(prId)
	if err != nil {
		return nil, err
	}

	return createdPR, nil
}

func (s *prService) MergePR(prId string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetById(prId)
	if err != nil {
		return nil, err
	}
	if pr == nil {
		return nil, domain.NewNotFoundError("PR")
	}

	if pr.IsMerged() {
		return pr, nil
	}

	for _, reviewerId := range pr.AssignedReviewers {
		reviewer, err := s.userRepo.GetById(reviewerId)
		if err != nil {
			continue
		}
		if reviewer != nil && !reviewer.IsActive {
			err = s.prRepo.RemoveReviewer(prId, reviewerId)
			if err != nil {
				continue
			}
		}
	}

	now := time.Now()
	err = s.prRepo.UpdateStatus(prId, domain.PRStatusMerged, &now)
	if err != nil {
		return nil, err
	}

	mergedPR, err := s.prRepo.GetById(prId)
	if err != nil {
		return nil, err
	}
	return mergedPR, nil
}

func (s *prService) ReassignReviewer(prId, oldReviewerId string) (*domain.PullRequest, string, error) {
	pr, err := s.prRepo.GetById(prId)
	if err != nil {
		return nil, "", err
	}
	if pr == nil {
		return nil, "", domain.NewNotFoundError("PR")
	}

	if pr.IsMerged() {
		return nil, "", domain.NewPRMergedError()
	}

	if !s.isReviewerAssigned(pr.AssignedReviewers, oldReviewerId) {
		return nil, "", domain.NewNotAssignedError()
	}
	oldReviewer, err := s.userRepo.GetById(oldReviewerId)
	if err != nil {
		return nil, "", err
	}
	if oldReviewer == nil {
		return nil, "", domain.NewNotFoundError("user")
	}

	team, err := s.teamRepo.GetByName(oldReviewer.TeamName)
	if err != nil {
		return nil, "", err
	}
	if team == nil {
		return nil, "", domain.NewNotFoundError("team")
	}

	excludedReviewers := pr.AssignedReviewers
	newReviewers := s.selectReviewers(team, pr.AuthorId, excludedReviewers, 1)
	if len(newReviewers) == 0 {
		return nil, "", domain.NewNoCandidateError()
	}

	newReviewerId := newReviewers[0]
	err = s.prRepo.ReplaceReviewer(prId, oldReviewerId, newReviewerId)
	if err != nil {
		return nil, "", err
	}

	updatedPR, err := s.prRepo.GetById(prId)
	if err != nil {
		return nil, "", err
	}

	for _, reviewerId := range updatedPR.AssignedReviewers {
		reviewer, err := s.userRepo.GetById(reviewerId)
		if err != nil {
			continue
		}
		if reviewer != nil && !reviewer.IsActive {
			inactiveReviewerTeam, err := s.teamRepo.GetByName(reviewer.TeamName)
			if err != nil {
				continue
			}
			if inactiveReviewerTeam == nil {
				continue
			}

			excludedForReplacement := updatedPR.AssignedReviewers
			replacementCandidates := s.selectReviewers(inactiveReviewerTeam, pr.AuthorId, excludedForReplacement, 1)

			if len(replacementCandidates) > 0 {
				err = s.prRepo.ReplaceReviewer(prId, reviewerId, replacementCandidates[0])
				if err != nil {
					continue
				}
				updatedPR, err = s.prRepo.GetById(prId)
				if err != nil {
					return nil, "", err
				}
			} else {
				err = s.prRepo.RemoveReviewer(prId, reviewerId)
				if err != nil {
					continue
				}
				updatedPR, err = s.prRepo.GetById(prId)
				if err != nil {
					return nil, "", err
				}
			}
		}
	}

	finalPR, err := s.prRepo.GetById(prId)
	if err != nil {
		return nil, "", err
	}

	if len(finalPR.AssignedReviewers) < 2 {
		author, err := s.userRepo.GetById(pr.AuthorId)
		if err == nil && author != nil {
			authorTeam, err := s.teamRepo.GetByName(author.TeamName)
			if err == nil && authorTeam != nil {
				excludedForAddition := finalPR.AssignedReviewers
				additionalReviewers := s.selectReviewers(authorTeam, pr.AuthorId, excludedForAddition, 2-len(finalPR.AssignedReviewers))
				for _, additionalReviewerId := range additionalReviewers {
					err = s.prRepo.AssignReviewer(prId, additionalReviewerId)
					if err != nil {
						continue
					}
				}
				finalPR, err = s.prRepo.GetById(prId)
				if err != nil {
					return nil, "", err
				}
			}
		}
	}

	return finalPR, newReviewerId, nil
}