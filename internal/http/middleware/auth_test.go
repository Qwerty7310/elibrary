package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"elibrary/internal/domain"
	"elibrary/internal/repository"
	"elibrary/internal/service"

	"github.com/google/uuid"
)

type stubUserRepo struct {
	getByIDWithRoles func(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

func (s stubUserRepo) Create(ctx context.Context, user domain.User) error { return nil }
func (s stubUserRepo) Update(ctx context.Context, user domain.User) error { return nil }
func (s stubUserRepo) Delete(ctx context.Context, id uuid.UUID) error     { return nil }
func (s stubUserRepo) GetByLogin(ctx context.Context, login string) (*domain.User, error) {
	return nil, nil
}
func (s stubUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, nil
}
func (s stubUserRepo) GetByIDWithRoles(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.getByIDWithRoles(ctx, id)
}
func (s stubUserRepo) GetAllWithRoles(ctx context.Context) ([]*domain.User, error) { return nil, nil }

var _ repository.UserRepository = stubUserRepo{}

func TestAuthRejectsMissingBearerHeader(t *testing.T) {
	t.Parallel()

	jwt := &service.JWTManager{Secret: []byte("secret"), TTL: time.Minute}
	handler := Auth(jwt, stubUserRepo{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	jwt := &service.JWTManager{Secret: []byte("secret"), TTL: time.Minute}
	handler := Auth(jwt, stubUserRepo{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthRejectsInactiveUser(t *testing.T) {
	t.Parallel()

	jwt := &service.JWTManager{Secret: []byte("secret"), TTL: time.Minute}
	userID := uuid.New()
	token, err := jwt.Generate(userID)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	handler := Auth(jwt, stubUserRepo{
		getByIDWithRoles: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			if id != userID {
				t.Fatalf("GetByIDWithRoles() id = %v, want %v", id, userID)
			}
			return &domain.User{ID: id, IsActive: false}, nil
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthAddsUserToContext(t *testing.T) {
	t.Parallel()

	jwt := &service.JWTManager{Secret: []byte("secret"), TTL: time.Minute}
	userID := uuid.New()
	token, err := jwt.Generate(userID)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	wantUser := &domain.User{ID: userID, IsActive: true}
	called := false
	handler := Auth(jwt, stubUserRepo{
		getByIDWithRoles: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			return wantUser, nil
		},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		gotUser, ok := UserFromContext(r.Context())
		if !ok {
			t.Fatal("UserFromContext() ok = false, want true")
		}
		if gotUser != wantUser {
			t.Fatalf("user in context = %#v, want %#v", gotUser, wantUser)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("next handler was not called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireRole(t *testing.T) {
	t.Parallel()

	makeHandler := func(ctx context.Context) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		handler := RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		handler.ServeHTTP(rec, req)
		return rec
	}

	rec := makeHandler(context.Background())
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status without user = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	forbiddenCtx := context.WithValue(context.Background(), userKey, &domain.User{
		Roles: []domain.Role{{Code: "user"}},
	})
	rec = makeHandler(forbiddenCtx)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status without role = %d, want %d", rec.Code, http.StatusForbidden)
	}

	allowedCtx := context.WithValue(context.Background(), userKey, &domain.User{
		Roles: []domain.Role{{Code: "admin"}},
	})
	rec = makeHandler(allowedCtx)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status with role = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
