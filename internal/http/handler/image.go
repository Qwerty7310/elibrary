package handler

import (
	"elibrary/internal/service"
	"elibrary/internal/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ImageHandler struct {
	Service *service.ImageService
}

func NewImageHandler(service *service.ImageService) *ImageHandler {
	return &ImageHandler{Service: service}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	entityStr := chi.URLParam(r, "entity")
	idStr := chi.URLParam(r, "id")

	entity := storage.EntityType(entityStr)
	switch entity {
	case storage.Book, storage.Author, storage.Publisher:
	default:
		http.Error(w, "invalid entity", http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	url, err := h.Service.Upload(ctx, entity, entityID, file)
	if err != nil {
		http.Error(w, "upload failed", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"url": url,
	})
}
