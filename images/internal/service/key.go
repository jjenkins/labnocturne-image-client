package service

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/jjenkins/labnocturne/images/internal/model"
	"github.com/jjenkins/labnocturne/images/internal/store"
	"github.com/jxskiss/base62"
)

// KeyService handles API key generation business logic
type KeyService struct {
	userStore *store.UserStore
}

// NewKeyService creates a new KeyService
func NewKeyService(userStore *store.UserStore) *KeyService {
	return &KeyService{userStore: userStore}
}

// GenerateTestKey creates a new test API key and user record
func (s *KeyService) GenerateTestKey(ctx context.Context) (*model.User, error) {
	// Generate cryptographically secure random key
	apiKey, err := GenerateAPIKey("ln_test_")
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Create user record
	user := &model.User{
		APIKey:  apiKey,
		KeyType: "test",
		Plan:    "test",
	}

	if err := s.userStore.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GenerateAPIKey creates a cryptographically secure API key with the given prefix
func GenerateAPIKey(prefix string) (string, error) {
	// Generate 32 bytes of randomness (256 bits)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base62 for URL-safe, readable keys
	encoded := base62.EncodeToString(b)

	// Result: ln_test_{32-40 characters}
	return prefix + encoded, nil
}
