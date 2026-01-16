package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/service"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type LocationHandler struct {
	Service *service.LocationService
}

func NewLocationHandler(service *service.LocationService) *LocationHandler {
	return &LocationHandler{Service: service}
}

type createLocationRequest struct {
	ParentID    *uuid.UUID          `json:"parent_id,omitempty"`
	Type        domain.LocationType `json:"type"`
	Name        string              `json:"name"`
	Barcode     string              `json:"barcode"`
	Address     *string             `json:"address,omitempty"`
	Description *string             `json:"description,omitempty"`
}

//func (h *LocationHandler) Create(w http.ResponseWriter, r *http.Request) {
//
//}

func (h *LocationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing ID %s: %v", idStr, err)
	}

	location, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Printf("error getting location %s: %v", idStr, err)
		http.Error(w, "error getting location", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, location)
}

func (h *LocationHandler) GetByParentID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing ID %s: %v", idStr, err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	typeStr := chi.URLParam(r, "type")

	locType, err := domain.ParseLocationType(typeStr)
	if err != nil {
		log.Printf("invalid location type %s: %v", typeStr, err)
		http.Error(w, "invalid location type", http.StatusBadRequest)
		return
	}

	locations, err := h.Service.GetByTypeParentID(r.Context(), locType, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		log.Printf("error getting locations %s: %v", idStr, err)
		http.Error(w, "error getting locations", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, locations)
}
