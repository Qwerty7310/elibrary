package service

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTManagerGenerateAndParse(t *testing.T) {
	t.Parallel()

	manager := &JWTManager{
		Secret: []byte("test-secret"),
		TTL:    time.Minute,
	}
	userID := uuid.New()

	token, err := manager.Generate(userID)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	gotUserID, err := manager.Parse(token)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if gotUserID != userID {
		t.Fatalf("Parse() userID = %v, want %v", gotUserID, userID)
	}
}

func TestJWTManagerParseRejectsDifferentSecret(t *testing.T) {
	t.Parallel()

	issuer := &JWTManager{
		Secret: []byte("issuer-secret"),
		TTL:    time.Minute,
	}
	validator := &JWTManager{
		Secret: []byte("validator-secret"),
		TTL:    time.Minute,
	}

	token, err := issuer.Generate(uuid.New())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := validator.Parse(token); err == nil {
		t.Fatal("Parse() error = nil, want signature validation error")
	}
}

func TestJWTManagerParseRejectsTamperedToken(t *testing.T) {
	t.Parallel()

	manager := &JWTManager{
		Secret: []byte("test-secret"),
		TTL:    time.Minute,
	}

	token, err := manager.Generate(uuid.New())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	tampered := token[:len(token)-1] + "x"
	if tampered == token {
		t.Fatal("failed to tamper token")
	}

	if _, err := manager.Parse(tampered); err == nil {
		t.Fatal("Parse() error = nil, want parse error")
	}
}

func TestJWTManagerParseRejectsExpiredToken(t *testing.T) {
	t.Parallel()

	manager := &JWTManager{
		Secret: []byte("test-secret"),
		TTL:    -time.Second,
	}

	token, err := manager.Generate(uuid.New())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	_, err = manager.Parse(token)
	if err == nil {
		t.Fatal("Parse() error = nil, want expiration error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "expired") {
		t.Fatalf("Parse() error = %q, want expiration-related error", err.Error())
	}
}
