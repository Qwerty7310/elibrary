package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BookHandler struct {
	Service        *service.BookService
	BarcodeService *service.BarcodeService
}

func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var book domain.Book
	if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	created, barcodeImage, err := h.Service.CreateBook(r.Context(), book)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBarcodeExists):
			http.Error(w, "barcode already exists", http.StatusConflict)
		case errors.Is(err, service.ErrInvalidBarcode):
			http.Error(w, "invalid barcode", http.StatusBadRequest)
		default:
			http.Error(w, "failed to create book", http.StatusInternalServerError)
		}
	}

	responce := map[string]interface{}{
		"book":    created,
		"message": "Book created successfully",
	}

	if barcodeImage != nil {
		responce["barcode_image"] = barcodeImage
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responce)
}

func (h *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	book, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get book", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(book)
}

func (h *BookHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if strings.TrimSpace(query) == "" {
		http.Error(w, "query is required", http.StatusBadRequest)
		return
	}

	books, err := h.Service.Search(r.Context(), query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responce := map[string]interface{}{
		"books": books,
		"count": len(books),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(responce)
}

func (h *BookHandler) Barcode(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	book, err := h.Service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "book not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get book", http.StatusInternalServerError)
		return
	}

	barcodeImage, err := h.BarcodeService.GenerateBarcodeImage(book.Barcode)
	if err != nil {
		http.Error(w, "failed to generate barcode image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(barcodeImage)
}
