package handler

import (
	"elibrary/internal/service"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ScanHandler struct {
	BookService *service.BookService
}

func (h *ScanHandler) Scan(w http.ResponseWriter, r *http.Request) {
	value := chi.URLParam(r, "value")

	book, err := h.BookService.FindByScan(r.Context(), value)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidBarcode):
			http.Error(w, "Invalid Barcode", http.StatusBadRequest)
		case errors.Is(err, service.ErrNotFound):
			http.Error(w, "Not Found", http.StatusNotFound)
		default:
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(book)
}
