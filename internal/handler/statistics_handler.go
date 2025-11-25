package handler

import (
	"encoding/json"
	"net/http"
	"pr-reviewer-service/internal/service"
)

type StatisticsHandler struct {
	statisticsService service.StatisticsService
}

func NewStatisticsHandler(statisticsService service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{
		statisticsService: statisticsService,
	}
}

func (h *StatisticsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.statisticsService.GetStatistics()
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

