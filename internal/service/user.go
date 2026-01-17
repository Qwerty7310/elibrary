package service

import (
	"context"
	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Create(ctx context.Context, user domain.User) (*domain.User, error) {
	if strings.TrimSpace(user.Login) == "" {
		return nil, errors.New("user login is required")
	}

	exist, err := s.userRepo.GetByLogin(ctx, user.Login)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	if exist != nil {
		return nil, domain.ErrLoginExists
	}

	user.ID = uuid.New()
	user.IsActive = true

	if strings.TrimSpace(user.FirstName) == "" {
		return nil, errors.New("user first name is required")
	}
	if strings.TrimSpace(user.PasswordHash) == "" {
		return nil, errors.New("invalid password hash")
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &user, nil
}

type UpdateUserRequest struct {
	Login      *string        `json:"login"`
	FirstName  *string        `json:"first_name"`
	LastName   *string        `json:"last_name,omitempty"`
	MiddleName *string        `json:"middle_name,omitempty"`
	Email      *string        `json:"email,omitempty"`
	Password   *string        `json:"password,omitempty"`
	IsActive   *bool          `json:"is_active"`
	Roles      *[]domain.Role `json:"roles"`
}

func (s *UserService) Update(ctx context.Context, id uuid.UUID, updates UpdateUserRequest) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}

	if updates.Login != nil {
		login := strings.TrimSpace(*updates.Login)
		if login == "" {
			return errors.New("login is empty")
		}

		exist, err := s.userRepo.GetByLogin(ctx, login)
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return err
		}
		if exist != nil && exist.ID != user.ID {
			return domain.ErrLoginExists
		}

		user.Login = login
	}
	if updates.FirstName != nil {
		if strings.TrimSpace(*updates.FirstName) == "" {
			return errors.New("user first name is required")
		}
		user.FirstName = *updates.FirstName
	}
	if updates.LastName != nil {
		user.LastName = updates.LastName
	}
	if updates.MiddleName != nil {
		user.MiddleName = updates.MiddleName
	}
	if updates.Email != nil {
		user.Email = updates.Email
	}
	if updates.Password != nil {
		if strings.TrimSpace(*updates.Password) == "" {
			return errors.New("password is empty")
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(*updates.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.PasswordHash = string(hash)
	}
	if updates.IsActive != nil {
		user.IsActive = *updates.IsActive
	}
	if updates.Roles != nil {
		user.Roles = *updates.Roles
	}

	return s.userRepo.Update(ctx, *user)
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.userRepo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (s *UserService) GetByIDWithRoles(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByIDWithRoles(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}
