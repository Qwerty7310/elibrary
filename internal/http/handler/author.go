package handler

import (
	"elibrary/internal/service"
	"encoding/json"
	"log"
	"net/http"
)

type AuthorHandler struct {
	Service *service.AuthorService
}

func NewAuthorHandler(service *service.AuthorService) *AuthorHandler {
	return &AuthorHandler{Service: service}
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
