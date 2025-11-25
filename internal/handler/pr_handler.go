package handler

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/service"
)

type PRHandler struct {
	prService service.PRService
}

func NewPRHandler(prService service.PRService) *PRHandler {
	return &PRHandler{
		prService: prService,
	}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", "")
		return
	}

	if req.PullRequestID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "pull_request_id is required", "")
		return
	}

	if req.PullRequestName == "" {
		writeErrorResponse(w, http.StatusBadRequest, "pull_request_name is required", "")
		return
	}
	
	if req.AuthorID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "author_id is required", "")
		return
	}

	pr, err := h.prService.CreatePR(req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response := map[string]interface{}{
		"pr": pr,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}


	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", "")
		return
	}

	if req.PullRequestID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "pull_request_id is required", "")
		return
	}

	pr, err := h.prService.MergePR(req.PullRequestID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	
	response := map[string]interface{}{
		"pr": pr,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "invalid request body", "")
		return
	}

	if req.PullRequestID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "pull_request_id is required", "")
		return
	}
	
	if req.OldUserID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "old_user_id is required", "")
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(req.PullRequestID, req.OldUserID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	
	response := map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}