package repository

import (
	"database/sql"
	"pr-reviewer-service/internal/domain"
	"time"
)

type prRepository struct {
	db *sql.DB
}

func NewPRRepository(db *sql.DB) PRRepository {
	return &prRepository{db: db}
}

func (r *prRepository) Create(pr *domain.PullRequest) error {
	query :=
	`
	INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
	VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(query, pr.PullRequestId, pr.PullRequestName, pr.AuthorId, pr.Status, pr.CreatedAt)
	if err != nil {
		return err
	}

	for _, reviewer := range pr.AssignedReviewers {
		reviewerQuery := 
		`
		INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)
		`
		_, err := r.db.Exec(reviewerQuery, pr.PullRequestId, reviewer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *prRepository) GetById(prId string) (*domain.PullRequest, error) {
	prQuery :=
	`
	SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
	FROM pull_requests
	WHERE pull_request_id = $1
	`
	var pr domain.PullRequest
	var createdAt, mergedAt sql.NullTime

	err := r.db.QueryRow(prQuery, prId).Scan(
		&pr.PullRequestId, 
		&pr.PullRequestName, 
		&pr.AuthorId, 
		&pr.Status, 
		&createdAt, 
		&mergedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	if createdAt.Valid {
		pr.CreatedAt = &createdAt.Time
	}
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	reviewersQuery :=
	`
	SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1
	`
	rows, err := r.db.Query(reviewersQuery, prId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var reviewerId string
		if err := rows.Scan(&reviewerId); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerId)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	pr.AssignedReviewers = reviewers
	return &pr, nil
}

func (r *prRepository) GetByReviewerId(reviewerId string) ([]*domain.PullRequestShort, error) {
	query :=
	`
	SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
	FROM pull_requests pr
	JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
	WHERE prr.reviewer_id = $1
	`
	rows, err := r.db.Query(query, reviewerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []*domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(
			&pr.PullRequestId, 
			&pr.PullRequestName, 
			&pr.AuthorId, 
			&pr.Status,
		); err != nil {
			return nil, err
		}
		prs = append(prs, &pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return prs, nil
}

func (r *prRepository) UpdateStatus(prId string, status domain.PRStatus, mergedAt *time.Time) error {
	query :=
	`
	UPDATE pull_requests
	SET status = $1, merged_at = $2
	WHERE pull_request_id = $3
	`
	_, err := r.db.Exec(query, status, mergedAt, prId)
	return err
}

func (r *prRepository) AssignReviewer(prId string, reviewerId string) error {
	query :=
	`
	INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)
	`
	_, err := r.db.Exec(query, prId, reviewerId)
	return err
}

func (r *prRepository) ReplaceReviewer(prId string, oldReviewerId string, newReviewerId string) error {
	deleteQuery :=
	`
	DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2
	`
	_, err := r.db.Exec(deleteQuery, prId, oldReviewerId)
	if err != nil {
		return err
	}

	insertQuery :=
	`
	INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)
	`
	_, err = r.db.Exec(insertQuery, prId, newReviewerId)
	return err
}

func (r *prRepository) RemoveReviewer(prId string, reviewerId string) error {
	query :=
	`
	DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2
	`
	_, err := r.db.Exec(query, prId, reviewerId)
	return err
}

func (r *prRepository) Exists(prId string) (bool, error) {
	query :=
	`
	SELECT EXISTS (SELECT 1 FROM pull_requests WHERE pull_request_id = $1)
	`
	var exists bool
	err := r.db.QueryRow(query, prId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}