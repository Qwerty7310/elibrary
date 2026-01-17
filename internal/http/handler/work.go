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

type WorkHandler struct {
	Service *service.WorkService
}

func NewWorkHandler(service *service.WorkService) *WorkHandler {
	return &WorkHandler{Service: service}
}

type createWorkRequest struct {
	Work    domain.Work `json:"work"`
	Authors []uuid.UUID `json:"authors,omitempty"`
}

func (h *WorkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error decoding create work request: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Work.Title) == "" {
		http.Error(w, "work title is required", http.StatusBadRequest)
		return
	}

	created, err := h.Service.Create(r.Context(), req.Work, req.Authors)
	if err != nil {
		log.Printf("failed to create work: %s", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

type updateWorkRequest = service.UpdateWorkRequest

func (h *WorkHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("failed to parse work id: %s", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updateWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode request body: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.Service.Update(r.Context(), id, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("work %s not found: %v", idStr, err)
			http.Error(w, "work not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to update work: %v", err)
		http.Error(w, "failed to update work", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("failed to parse id: %v", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("work %s not found: %v", idStr, err)
			http.Error(w, "work not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to delete work: %v", err)
		http.Error(w, "failed to delete work", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing ID %s: %v", idStr, err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	work, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Printf("error getting work %s: %v", idStr, err)
		http.Error(w, "failed to get work", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, work)
}

func (h *WorkHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	works, err := h.Service.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting all works: %v", err)
		http.Error(w, "failed to get works", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, works)
}
