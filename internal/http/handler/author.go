package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/service"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuthorHandler struct {
	Service *service.AuthorService
}

func NewAuthorHandler(service *service.AuthorService) *AuthorHandler {
	return &AuthorHandler{Service: service}
}

type createAuthorRequest struct {
	LastName   string     `json:"last_name"`
	FirstName  *string    `json:"first_name,omitempty"`
	MiddleName *string    `json:"middle_name,omitempty"`
	BirthDate  *time.Time `json:"birth_date,omitempty"`
	DeathDate  *time.Time `json:"death_date,omitempty"`
	Bio        *string    `json:"bio,omitempty"`
	PhotoURL   *string    `json:"photo_url,omitempty"`
}

func (h *AuthorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error decoding create author request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.LastName) == "" {
		http.Error(w, "last name must not be empty", http.StatusBadRequest)
		return
	}

	author := domain.Author{
		LastName:   req.LastName,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		BirthDate:  req.BirthDate,
		DeathDate:  req.DeathDate,
		Bio:        req.Bio,
		PhotoURL:   req.PhotoURL,
	}

	created, err := h.Service.Create(r.Context(), author)
	if err != nil {
		log.Printf("error creating author: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

type updateAuthorRequest = service.UpdateAuthorRequest

func (h *AuthorHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing id from request: %v", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updateAuthorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("failed to decode update author request: %s", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.Service.Update(r.Context(), id, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "author not found", http.StatusNotFound)
			return
		}
		log.Printf("failed to update author: %v", err)
		http.Error(w, "failed to update author", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing id from request: %v", err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), id); err != nil {
		log.Printf("failed to delete author: %v", err)
		http.Error(w, "failed to delete author", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("Error parsing author ID %s: %v", idStr, err)
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	author, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("author %s not found: %v", idStr, err)
			http.Error(w, "author not found", http.StatusNotFound)
			return
		}
		log.Printf("error getting author %s: %v", idStr, err)
		http.Error(w, "error getting author", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, author)
}

func (h *AuthorHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	authors, err := h.Service.GetAll(r.Context())
	if err != nil {
		log.Printf("AuthorHandler.GetAll: %v", err)
		http.Error(w, "failed to get authors", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(authors)
}
