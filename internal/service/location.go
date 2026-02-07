package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
)

type LocationService struct {
	locRepo    repository.LocationRepository
	BarcodeSvc *BarcodeService
}

var (
	ErrParentNotFound           = errors.New("parent not found")
	ErrLocationCannotHaveParent = errors.New("location can not have parent")
	ErrInvalidLocationType      = errors.New("invalid location type")
	ErrLocationHasChildren      = errors.New("location has children")
)

func NewLocationService(locRepo repository.LocationRepository, barcodeSvc *BarcodeService) *LocationService {
	return &LocationService{
		locRepo:    locRepo,
		BarcodeSvc: barcodeSvc,
	}
}

func (s *LocationService) Create(ctx context.Context, location domain.Location) (*domain.Location, error) {
	ean13, err := s.BarcodeSvc.GenerateEAN13(ctx, domain.BarcodeTypeLocation)
	if err != nil {
		return nil, err
	}

	location.ID = uuid.New()
	location.Barcode = ean13

	if strings.TrimSpace(location.Name) == "" {
		return nil, errors.New("name is required")
	}

	if location.Type == domain.LocationTypeBuilding && (location.Address == nil || strings.TrimSpace(*location.Address) == "") {
		return nil, errors.New("address is required")
	}

	if err := s.locRepo.Create(ctx, location); err != nil {
		log.Printf("Error creating location %s: %v", location.Name, err)
		return nil, err
	}

	return &location, nil
}

type UpdateLocationRequest struct {
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        *string    `json:"name,omitempty"`
	Address     *string    `json:"address,omitempty"`
	Description *string    `json:"description,omitempty"`
}

func (s *LocationService) Update(ctx context.Context, id uuid.UUID, updates UpdateLocationRequest) error {
	location, err := s.locRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if updates.ParentID != nil {
		if location.Type == domain.LocationTypeBuilding {
			return ErrLocationCannotHaveParent
		}

		parent, err := s.locRepo.GetByID(ctx, *updates.ParentID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrParentNotFound
			}
			return err
		}

		if !location.Type.IsChildOf(parent.Type) {
			return ErrInvalidLocationType
		}

		location.ParentID = updates.ParentID
	}

	if updates.Name != nil {
		if strings.TrimSpace(*updates.Name) == "" {
			return errors.New("name is required")
		}
		location.Name = *updates.Name
	}

	if updates.Address != nil {
		if strings.TrimSpace(*updates.Address) == "" {
			return errors.New("address is required")
		}
		location.Address = updates.Address
	}

	if updates.Description != nil {
		if strings.TrimSpace(*updates.Description) == "" {
			return errors.New("description is required")
		}
		location.Description = updates.Description
	}

	return s.locRepo.Update(ctx, *location)
}

func (s *LocationService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Location, error) {
	location, err := s.locRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return location, nil
}

func (s *LocationService) GetByType(ctx context.Context, locType domain.LocationType) ([]*domain.Location, error) {
	locations, err := s.locRepo.GetByType(ctx, locType)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return locations, nil
}

func (s *LocationService) GetByTypeParentID(ctx context.Context, locType domain.LocationType, parentID uuid.UUID) ([]*domain.Location, error) {
	parent, err := s.locRepo.GetByID(ctx, parentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrParentNotFound
		}
		return nil, err
	}

	if !locType.IsChildOf(parent.Type) {
		return nil, ErrInvalidLocationType
	}

	locations, err := s.locRepo.GetByTypeParentID(ctx, locType, parentID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return locations, nil
}

func (s *LocationService) GetByBarcode(ctx context.Context, barcode string) (*domain.Location, error) {
	if !s.BarcodeSvc.ValidateEAN13(barcode) {
		return nil, domain.ErrInvalidBarcode
	}

	location, err := s.locRepo.GetByBarcode(ctx, barcode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return location, nil
}

func (s *LocationService) Delete(ctx context.Context, id uuid.UUID) error {
	hasChildren, err := s.locRepo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return ErrLocationHasChildren
	}

	err = s.locRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	return nil
}
