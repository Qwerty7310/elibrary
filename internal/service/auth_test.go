package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"elibrary/internal/domain"
	"elibrary/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type stubAuthUserRepo struct {
	getByLogin func(ctx context.Context, login string) (*domain.User, error)
	getByID    func(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func (s stubAuthUserRepo) Create(ctx context.Context, user domain.User) error { return nil }
func (s stubAuthUserRepo) Update(ctx context.Context, user domain.User) error { return nil }
func (s stubAuthUserRepo) Delete(ctx context.Context, id uuid.UUID) error     { return nil }
func (s stubAuthUserRepo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	return s.getByLogin(ctx, login)
}
func (s stubAuthUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.getByID(ctx, id)
}
func (s stubAuthUserRepo) GetByIDWithRoles(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (s stubAuthUserRepo) GetAllWithRoles(ctx context.Context) ([]*domain.User, error) {
	return nil, nil
}

var _ repository.UserRepository = stubAuthUserRepo{}

func TestAuthServiceLogin(t *testing.T) {
	t.Parallel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	userID := uuid.New()
	svc := NewAuthService(stubAuthUserRepo{
		getByLogin: func(ctx context.Context, login string) (*domain.User, error) {
			if login != "reader" {
				t.Fatalf("GetByLogin() login = %q, want %q", login, "reader")
			}
			return &domain.User{
				ID:           userID,
				Login:        login,
				PasswordHash: string(passwordHash),
				IsActive:     true,
			}, nil
		},
	}, &JWTManager{Secret: []byte("secret"), TTL: time.Minute})

	token, err := svc.Login(context.Background(), "reader", "secret")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token == "" {
		t.Fatal("Login() returned empty token")
	}
}

func TestAuthServiceLoginRejectsInvalidCredentials(t *testing.T) {
	t.Parallel()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}

	tests := []struct {
		name string
		repo stubAuthUserRepo
		pass string
	}{
		{
			name: "repo error",
			repo: stubAuthUserRepo{
				getByLogin: func(ctx context.Context, login string) (*domain.User, error) {
					return nil, errors.New("db error")
				},
			},
			pass: "secret",
		},
		{
			name: "inactive user",
			repo: stubAuthUserRepo{
				getByLogin: func(ctx context.Context, login string) (*domain.User, error) {
					return &domain.User{PasswordHash: string(passwordHash), IsActive: false}, nil
				},
			},
			pass: "secret",
		},
		{
			name: "wrong password",
			repo: stubAuthUserRepo{
				getByLogin: func(ctx context.Context, login string) (*domain.User, error) {
					return &domain.User{PasswordHash: string(passwordHash), IsActive: true}, nil
				},
			},
			pass: "wrong",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewAuthService(tt.repo, &JWTManager{Secret: []byte("secret"), TTL: time.Minute})
			_, err := svc.Login(context.Background(), "reader", tt.pass)
			if !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
			}
		})
	}
}

func TestAuthServiceMe(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	wantUser := &domain.User{ID: userID, IsActive: true}
	svc := NewAuthService(stubAuthUserRepo{
		getByLogin: nil,
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			if id != userID {
				t.Fatalf("GetByID() id = %v, want %v", id, userID)
			}
			return wantUser, nil
		},
	}, &JWTManager{})

	got, err := svc.Me(context.Background(), userID)
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}
	if got != wantUser {
		t.Fatalf("Me() = %#v, want %#v", got, wantUser)
	}
}

func TestAuthServiceMeNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		repo stubAuthUserRepo
	}{
		{
			name: "repo error",
			repo: stubAuthUserRepo{
				getByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					return nil, errors.New("db error")
				},
			},
		},
		{
			name: "inactive",
			repo: stubAuthUserRepo{
				getByID: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
					return &domain.User{ID: id, IsActive: false}, nil
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewAuthService(tt.repo, &JWTManager{})
			_, err := svc.Me(context.Background(), uuid.New())
			if !errors.Is(err, domain.ErrNotFound) {
				t.Fatalf("Me() error = %v, want %v", err, domain.ErrNotFound)
			}
		})
	}
}
