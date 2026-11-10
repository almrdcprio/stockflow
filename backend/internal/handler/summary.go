package handler

import (
	"log"
	"net/http"

	"stockflow/internal/repository"
)

type SummaryHandler struct {
	repo *repository.SummaryRepository
}

func NewSummaryHandler(repo *repository.SummaryRepository) *SummaryHandler {
	return &SummaryHandler{
		repo: repo,
	}
}

func (h *SummaryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.repo.GetSummary(r.Context())
	if err != nil {
		log.Printf("failed to get summary: %v", err)
		respondError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"Internal server error",
		)
		return
	}

	respondJSON(w, http.StatusOK, summary)
}
