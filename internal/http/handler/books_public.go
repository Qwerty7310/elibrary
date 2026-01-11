package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"elibrary/internal/repository"
	"elibrary/internal/service"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BooksPublicHandler struct {
	Service *service.BookService
}

func (h *BooksPublicHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	book, err := h.Service.GetPublicByID(r.Context(), id)
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

func (h *BooksPublicHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseBookFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	books, err := h.Service.GetPublic(r.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"items": []*readmodel.BookPublic{}, "count": 0})
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

func (h *BooksPublicHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		http.Error(w, "q is empty", http.StatusBadRequest)
		return
	}

	filter := repository.BookFilter{
		Query:  &q,
		Limit:  intPtr(parseIntDefault(r.URL.Query().Get("limit"), 20)),
		Offset: intPtr(parseIntDefault(r.URL.Query().Get("offset"), 0)),
	}

	books, err := h.Service.GetPublic(r.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{"items": []*readmodel.BookPublic{}, "count": 0})
			return
		}
		http.Error(w, "failed to search", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": books,
		"count": len(books),
	})
}

func parseBookFilter(r *http.Request) (repository.BookFilter, error) {
	qp := r.URL.Query()

	var f repository.BookFilter

	if s := strings.TrimSpace(qp.Get("id")); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return f, errors.New("invalid id")
		}
		f.ID = &id
	}
	if s := strings.TrimSpace(qp.Get("barcode")); s != "" {
		f.Barcode = &s
	}
	if s := strings.TrimSpace(qp.Get("factory_barcode")); s != "" {
		f.FactoryBarcode = &s
	}
	if s := strings.TrimSpace(qp.Get("q")); s != "" {
		f.Query = &s
	}
	if s := strings.TrimSpace(qp.Get("publisher_id")); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			return f, errors.New("invalid publisher_id")
		}
		f.PublisherID = &id
	}
	if s := strings.TrimSpace(qp.Get("year_from")); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return f, errors.New("invalid year_from")
		}
		f.YearFrom = &v
	}
	if s := strings.TrimSpace(qp.Get("year_to")); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil {
			return f, errors.New("invalid year_to")
		}
		f.YearTo = &v
	}

	limit := parseIntDefault(qp.Get("limit"), 20)
	offset := parseIntDefault(qp.Get("offset"), 0)
	f.Limit = &limit
	f.Offset = &offset

	return f, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseIntDefault(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func intPtr(v int) *int {
	return &v
}
