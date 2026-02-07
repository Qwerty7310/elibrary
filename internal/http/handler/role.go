package handler

import (
	"elibrary/internal/domain"
	"elibrary/internal/service"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type RoleHandler struct {
	Service *service.RoleService
}

func NewRoleHandler(service *service.RoleService) *RoleHandler {
	return &RoleHandler{Service: service}
}

func (h *RoleHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	roles, err := h.Service.GetAllWithPermissions(r.Context())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusOK, []any{})
			return
		}
		log.Printf("error getting roles: %v", err)
		http.Error(w, "error getting roles", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, roles)
}

func (h *RoleHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.Service.GetAllPermissions(r.Context())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeJSON(w, http.StatusOK, []any{})
			return
		}
		log.Printf("error getting permissions: %v", err)
		http.Error(w, "error getting permissions", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, perms)
}

type createRoleRequest struct {
	Code            string   `json:"code"`
	Name            string   `json:"name"`
	PermissionCodes []string `json:"permission_codes"`
}

func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := h.Service.Create(r.Context(), req.Code, req.Name, req.PermissionCodes); err != nil {
		if errors.Is(err, domain.ErrRoleExists) {
			http.Error(w, "role already exists", http.StatusConflict)
			return
		}
		http.Error(w, "failed to create role", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

