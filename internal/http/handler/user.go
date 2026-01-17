package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/service"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	Service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{Service: service}
}

type createUserRequest struct {
	Login      string        `json:"login"`
	FirstName  string        `json:"first_name"`
	LastName   *string       `json:"last_name,omitempty"`
	MiddleName *string       `json:"middle_name,omitempty"`
	Email      *string       `json:"email,omitempty"`
	Password   string        `json:"password"`
	Roles      []domain.Role `json:"roles"`
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error decoding create user request: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Login) == "" {
		http.Error(w, "login is empty", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.FirstName) == "" {
		http.Error(w, "first_name is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Password) == "" {
		http.Error(w, "password is empty", http.StatusBadRequest)
		return
	}
	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error generating password_hash", http.StatusInternalServerError)
		return
	}

	user := domain.User{
		Login:        req.Login,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		MiddleName:   req.MiddleName,
		Email:        req.Email,
		PasswordHash: string(password),
		Roles:        req.Roles,
	}

	created, err := h.Service.Create(r.Context(), user)
	if err != nil {
		log.Printf("error creating user: %v", err)
		http.Error(w, "error creating user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

type updateUserRequest = service.UpdateUserRequest

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing user id %s: %v", idStr, err)
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error decoding update user request: %v", err)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.Service.Update(r.Context(), id, req); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("user %s not found: %v", id, err)
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		log.Printf("error updating user %s: %v", id, err)
		http.Error(w, "error updating user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing user id %s: %v", idStr, err)
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("user %s not found: %v", id, err)
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		log.Printf("error deleting user %s: %v", id, err)
		http.Error(w, "error deleting user", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("error parsing user id %s: %v", idStr, err)
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.Service.GetByIDWithRoles(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			log.Printf("user %s not found: %v", id, err)
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		log.Printf("error getting user %s: %v", id, err)
		http.Error(w, "error getting user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, user)
}
