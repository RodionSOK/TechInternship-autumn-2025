package domain

import (
	"time"
)

type PRStatus string

const (
	PRStatusOpen PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestId string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorId string `json:"author_id"`
	Status PRStatus `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt *time.Time `json:"created_at"`
	MergedAt *time.Time `json:"merged_at"`
}

type PullRequestShort struct {
	PullRequestId string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorId string `json:"author_id"`
	Status PRStatus `json:"status"`
}

func (pr *PullRequest) IsMerged() bool {
	return pr.Status == PRStatusMerged
}

func (pr *PullRequest) CanReassign() bool {
	return !pr.IsMerged()
}