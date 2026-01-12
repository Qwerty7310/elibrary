package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/service"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookInternalHandler struct {
	Service *service.BookService
}

func NewBookInternalHandler(service *service.BookService) *BookInternalHandler {
	return &BookInternalHandler{Service: service}
}

func (h *BookInternalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	book, err := h.Service.GetInternalByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get book", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, book)
}

func (h *BookInternalHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseBookFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	books, err := h.Service.GetInternal(r.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"items": []any{}, "count": 0})
			return
		}
		http.Error(w, "failed to get books", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": books,
		"count": len(books),
	})
}
