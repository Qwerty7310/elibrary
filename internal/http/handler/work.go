package handler

import (
	"elibrary/internal/service"
	"encoding/json"
	"log"
	"net/http"
)

type WorkHandler struct {
	Service *service.WorkService
}

func NewWorkHandler(service *service.WorkService) *WorkHandler {
	return &WorkHandler{Service: service}
}

func (h *WorkHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	works, err := h.Service.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting all works: %v", err)
		http.Error(w, "failed to get works", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(works)
}
