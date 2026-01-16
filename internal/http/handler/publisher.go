package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/service"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PublisherHandler struct {
	Service *service.PublisherService
}

func NewPublisherHandler(service *service.PublisherService) *PublisherHandler {
	return &PublisherHandler{Service: service}
}

type createPublisherRequest struct {
	Name    string  `json:"name"`
	LogoURL *string `json:"logo_url,omitempty"`
	WebURL  *string `json:"web_url,omitempty"`
}

func (h *PublisherHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createPublisherRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode request body: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "publisher name is required", http.StatusBadRequest)
		return
	}

	publisher := domain.Publisher{
		Name:    req.Name,
		LogoURL: req.LogoURL,
		WebURL:  req.WebURL,
	}

	created, err := h.Service.Create(r.Context(), publisher)
	if err != nil {
		log.Printf("failed to create publisher: %v", err)
		http.Error(w, "failed to create publisher", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

type updatePublisherRequest = service.UpdatePublisherRequest

func (h *PublisherHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("failed to parse publisher id %s: %v", idStr, err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updatePublisherRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode request body: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.Service.Update(r.Context(), id, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "publisher not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to update publisher: %v", err)
		http.Error(w, "failed to update publisher", http.StatusInternalServerError)
		return
	}
}

func (h *PublisherHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("failed to parse publisher id %s: %v", idStr, err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), id); err != nil {
		log.Printf("failed to delete publisher: %v", err)
		http.Error(w, "failed to delete publisher", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PublisherHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("Error parsing publisher ID %s: %v", idStr, err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	publisher, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		log.Printf("Error getting publisher %s: %v", idStr, err)
		http.Error(w, "publisher not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, publisher)
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
