package repository

import (
	"context"
	"elibrary/internal/readmodel"
)

type RoleRepository interface {
	GetAllWithPermissions(ctx context.Context) ([]*readmodel.RoleWithPermissions, error)
	GetAllPermissions(ctx context.Context) ([]*readmodel.Permission, error)
	Create(ctx context.Context, code, name string, permissionCodes []string) error
}

