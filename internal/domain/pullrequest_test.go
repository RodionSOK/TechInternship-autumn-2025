package domain

import (
	"testing"
	"time"
)

func TestPullRequest_IsMerged(t *testing.T) {
	tests := []struct {
		name string
		pr *PullRequest
		expected bool
	}{
		{
			name: "open PR",
			pr: &PullRequest{
				Status: PRStatusOpen,
			},
			expected: false,
		},
		{
			name: "merged PR",
			pr: &PullRequest{
				Status: PRStatusMerged,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pr.IsMerged(); got != tt.expected {
				t.Errorf("IsMerged() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPullRequest_CanReassign(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		pr *PullRequest
		expected bool
	}{
		{
			name: "open PR can reassign",
			pr: &PullRequest{
				Status: PRStatusOpen,
				CreatedAt: &now,
			},
			expected: true,
		},
		{
			name: "merged PR cannot reassign",
			pr: &PullRequest{
				Status: PRStatusMerged,
				CreatedAt: &now,
				MergedAt: &now,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pr.CanReassign(); got != tt.expected {
				t.Errorf("CanReassign() = %v, want %v", got, tt.expected)
			}
		})
	}
}