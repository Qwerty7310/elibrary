package handler

import (
	"elibrary/internal/service"
	"encoding/json"
	"log"
	"net/http"
)

type PublisherHandler struct {
	Service *service.PublisherService
}

func NewPublisherHandler(service *service.PublisherService) *PublisherHandler {
	return &PublisherHandler{Service: service}
}

func (h *PublisherHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	publishers, err := h.Service.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting all publishers: %v", err)
		http.Error(w, "failed to get publishers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(publishers)
}
