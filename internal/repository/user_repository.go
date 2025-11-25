package repository

import (
	"database/sql"
	"pr-reviewer-service/internal/domain"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	query := 
	`
	INSERT INTO users (user_id, username, team_name, is_active)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (user_id)
	DO UPDATE SET username = $2, team_name = $3, is_active = $4
	`

	_, err := r.db.Exec(query, user.UserId, user.Username, user.TeamName, user.IsActive)
	return err
}

func (r *userRepository) GetById(userId string) (*domain.User, error) {
	query := 
	`
	SELECT user_id, username, team_name, is_active
	FROM users
	WHERE user_id = $1
	`
	var user domain.User
	err := r.db.QueryRow(query, userId).Scan(
		&user.UserId, 
		&user.Username, 
		&user.TeamName, 
		&user.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}	

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByTeamName(teamName string) ([]*domain.User, error) {
	query := 
	`
	SELECT user_id, username, team_name, is_active
	FROM users
	WHERE team_name = $1
	`
	rows, err := r.db.Query(query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.UserId, 
			&user.Username, 
			&user.TeamName, 
			&user.IsActive,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil	
}

func (r *userRepository) UpdateIsActive(userId string, IsActive bool) error {
	query :=
	`
	UPDATE users
	SET is_active = $1, updated_at = CURRENT_TIMESTAMP
	WHERE user_id = $2
	`
	_, err := r.db.Exec(query, IsActive, userId)
	return err
}

func (r *userRepository) GetAll() ([]*domain.User, error) {
	query := 
	`
	SELECT user_id, username, team_name, is_active
	FROM users
	ORDER BY username
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.UserId, 
			&user.Username, 
			&user.TeamName, 
			&user.IsActive,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) Exists(userId string) (bool, error) {
	query := 
	`
	SELECT EXISTS (SELECT 1 FROM users WHERE user_id = $1)
	`
	var exists bool
	err := r.db.QueryRow(query, userId).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}