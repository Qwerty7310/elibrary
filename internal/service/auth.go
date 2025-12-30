package service

import (
	"context"
	"elibrary/internal/repository"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	users repository.UserRepository
	jwt   *JWTManager
}

func NewAuthService(users repository.UserRepository, jwt *JWTManager) *AuthService {
	return &AuthService{users: users, jwt: jwt}
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.users.GetByLogin(ctx, login)
	if err != nil || !user.IsActive {
		return "", ErrInvalidCredentials
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", ErrInvalidCredentials
	}

	return s.jwt.Generate(user.ID)
}
