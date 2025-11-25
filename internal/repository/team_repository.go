package repository

import (
	"database/sql"
	"pr-reviewer-service/internal/domain"
)

type teamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) Create(team *domain.Team) error {
	query := 
	`
	INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING
	`
	_, err := r.db.Exec(query, team.TeamName)
	if err != nil {
		return err
	}

	for _, member := range team.Members {
		userQuery := 
		`
		INSERT INTO users (user_id, username, team_name, is_active) VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id)
		DO UPDATE SET username = $2, team_name = $3, is_active = $4
		`

		_, err := r.db.Exec(userQuery, member.UserId, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *teamRepository) GetByName(teamName string) (*domain.Team, error) {
	teamQuery := 
	`
	SELECT team_name FROM teams WHERE team_name = $1
	`
	var team domain.Team
	err := r.db.QueryRow(teamQuery, teamName).Scan(&team.TeamName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	membersQuery := 
	`
	SELECT user_id, username, is_active
	FROM users
	WHERE team_name = $1
	`

	rows, err := r.db.Query(membersQuery, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(
			&member.UserId, 
			&member.Username, 
			&member.IsActive,
		); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	team.Members = members
	return &team, nil
}

func (r *teamRepository) Exists(teamName string) (bool, error) {
	query := 
	`
	SELECT EXISTS (SELECT 1 FROM teams WHERE team_name = $1)
	`
	var exists bool
	err := r.db.QueryRow(query, teamName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}