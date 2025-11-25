package handler

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserId string `json:"user_id"`
		IsActive bool `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", "")
		return
	}

	if req.UserId == "" {
		writeErrorResponse(w, http.StatusBadRequest, "user_id is required", "")
		return
	}
	
	user, err := h.userService.SetIsActive(req.UserId, req.IsActive)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"user": user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		writeErrorResponse(w, http.StatusBadRequest, "user_id parameter is required", "")
		return
	}

	prs, err := h.userService.GetUserReviews(userId)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"user_id": userId,
		"pull_requests": prs,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}