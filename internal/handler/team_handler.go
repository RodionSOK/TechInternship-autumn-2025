package handler

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/service"
	"pr-reviewer-service/internal/domain"
)

type TeamHandler struct {
	teamService service.TeamService
}

func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		TeamName string `json:"team_name"`
		Members []domain.TeamMember `json:"members"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", "")
		return
	}

	if req.TeamName == "" {
		writeErrorResponse(w, http.StatusBadRequest, "team_name is required", "")
		return
	}

	team := &domain.Team{
		TeamName: req.TeamName, 
		Members: req.Members,
	}

	err := h.teamService.CreateTeam(team)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	createdTeam, err := h.teamService.GetTeam(req.TeamName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"team": createdTeam,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeErrorResponse(w, http.StatusBadRequest, "team_name is required", "")
		return
	}

	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(team)
}

func handleServiceError(w http.ResponseWriter, err error) {
	apiErr, ok := err.(*domain.APIError)
	if !ok {
		writeErrorResponse(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	var statusCode int

	switch apiErr.Code {
	case domain.ErrorCodeTeamExists:
		statusCode = http.StatusBadRequest
	case domain.ErrorCodePRExists:
		statusCode = http.StatusBadRequest
	case domain.ErrorCodePRMerged:
		statusCode = http.StatusBadRequest
	case domain.ErrorCodeNotAssigned:
		statusCode = http.StatusBadRequest
	case domain.ErrorCodeNoCandidate:
		statusCode = http.StatusBadRequest
	case domain.ErrorCodeNotFound:
		statusCode = http.StatusNotFound
	default:
		statusCode = http.StatusInternalServerError
	}

	writeErrorResponse(w, statusCode, apiErr.Message, string(apiErr.Code))
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message, code string) {
	response := map[string]interface{}{
		"error": map[string]string{
			"code": code,
			"message": message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}