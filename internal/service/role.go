package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/readmodel"
	"elibrary/internal/repository"
	"errors"
	"strings"
)

type RoleService struct {
	roleRepo repository.RoleRepository
}

func NewRoleService(roleRepo repository.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}

func (s *RoleService) GetAllWithPermissions(ctx context.Context) ([]*readmodel.RoleWithPermissions, error) {
	roles, err := s.roleRepo.GetAllWithPermissions(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return roles, nil
}

func (s *RoleService) GetAllPermissions(ctx context.Context) ([]*readmodel.Permission, error) {
	perms, err := s.roleRepo.GetAllPermissions(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return perms, nil
}

func (s *RoleService) Create(ctx context.Context, code, name string, permissionCodes []string) error {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return errors.New("code and name are required")
	}
	for i := range permissionCodes {
		permissionCodes[i] = strings.TrimSpace(permissionCodes[i])
	}
	return s.roleRepo.Create(ctx, code, name, permissionCodes)
}

