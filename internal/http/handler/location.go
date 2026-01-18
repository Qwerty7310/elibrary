package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/service"
	"encoding/json"
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
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Type        string     `json:"type"`
	Name        string     `json:"name"`
	Address     *string    `json:"address,omitempty"`
	Description *string    `json:"description,omitempty"`
}

func (h *LocationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createLocationRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error decoding create location request: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	locType, err := domain.ParseLocationType(req.Type)
	if err != nil {
		log.Printf("invalid location type: %v", err)
		http.Error(w, "invalid location type", http.StatusBadRequest)
		return
	}

	location := domain.Location{
		ParentID:    req.ParentID,
		Type:        locType,
		Name:        req.Name,
		Address:     req.Address,
		Description: req.Description,
	}

	created, err := h.Service.Create(r.Context(), location)
	if err != nil {
		if errors.Is(err, domain.ErrBarcodeExists) {
			log.Printf("location barcode already exists: %v", err)
			http.Error(w, "barcode already exists", http.StatusConflict)
			return
		}
		log.Printf("error creating location: %v", err)
		http.Error(w, "error creating location", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

type updateLocationRequest = service.UpdateLocationRequest

func (h *LocationHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("invalid id %s: %v", idStr, err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error decoding update location request: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.Service.Update(r.Context(), id, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("location %s not found: %v", idStr, err)
			http.Error(w, "location not found", http.StatusNotFound)
			return
		}
		log.Printf("error updating location: %v", err)
		http.Error(w, "error updating location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LocationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing ID %s: %v", idStr, err)
		http.Error(w, "invalid ID", http.StatusBadRequest)
		return
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

func (h *LocationHandler) GetByType(w http.ResponseWriter, r *http.Request) {
	typeStr := chi.URLParam(r, "type")
	locType, err := domain.ParseLocationType(typeStr)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLocationType) {
			log.Printf("invalid location type %s: %v", typeStr, err)
			http.Error(w, "invalid location type", http.StatusBadRequest)
			return
		}
		log.Printf("failed to parse location type %s: %v", typeStr, err)
		http.Error(w, "failed to parse location type", http.StatusInternalServerError)
		return
	}

	locations, err := h.Service.GetByType(r.Context(), locType)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("location %s not found: %v", typeStr, err)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInvalidLocationType) {
			log.Printf("invalid location type %s: %v", locType, err)
			http.Error(w, "invalid location type", http.StatusBadRequest)
			return
		}
		log.Printf("error getting locations %s: %v", typeStr, err)
		http.Error(w, "error getting locations", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, locations)
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
		if errors.Is(err, service.ErrInvalidLocationType) {
			log.Printf("invalid locatrion type %s: %v", typeStr, err)
			http.Error(w, "invalid locatrion type", http.StatusBadRequest)
			return
		}
		log.Printf("failed to parse location type %s: %v", typeStr, err)
		http.Error(w, "failed to parse location type", http.StatusInternalServerError)
		return
	}

	locations, err := h.Service.GetByTypeParentID(r.Context(), locType, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("location %s not found: %v", idStr, err)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrParentNotFound) {
			log.Printf("parent %s not found: %v", idStr, err)
			http.Error(w, "parent not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInvalidLocationType) {
			log.Printf("invalid location type %s: %v", locType, err)
			http.Error(w, "invalid location type", http.StatusBadRequest)
			return
		}
		log.Printf("error getting locations %s: %v", idStr, err)
		http.Error(w, "error getting locations", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, locations)
}

func (h *LocationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing ID %s: %v", idStr, err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("location %s not found: %v", idStr, err)
			http.Error(w, "location not found", http.StatusNotFound)
			return
		}
		log.Printf("error deleting location %s: %v", idStr, err)
		http.Error(w, "error deleting location", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
