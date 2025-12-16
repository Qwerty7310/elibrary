package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/service"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type BookHandler struct {
	Service *service.BookService
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var book domain.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	created, err := h.Service.Create(r.Context(), book)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/books/")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest
		return)
	}

	book, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(book)
}
