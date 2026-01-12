package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"elibrary/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookAdminHandler struct {
	Service *service.BookService
}

type createBookRequest struct {
	Book  domain.Book                `json:"book"`
	Works []repository.BookWorkInput `json:"works,omitempty"`
}

func (h *BookAdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Book.Title) == "" {
		http.Error(w, "book title is required", http.StatusBadRequest)
		return
	}

	created, err := h.Service.Create(r.Context(), req.Book, req.Works)
	if err != nil {
		if errors.Is(err, domain.ErrBarcodeExists) {
			http.Error(w, "barcode already exists", http.StatusConflict)
			return
		}
		http.Error(w, "failed to create book", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, created)
}

type updateBookRequest = service.UpdateBookRequest

func (h *BookAdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.Service.Update(r.Context(), id, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, domain.ErrBarcodeExists) {
			http.Error(w, "barcode already exists", http.StatusConflict)
			return
		}
		http.Error(w, "failed to update book", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
