package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/apikey"
	"github.com/google/uuid"
)

var (
	ErrAPIKeyNotFound  = errors.New("api key not found")
	ErrAPIKeyExpired   = errors.New("api key has expired")
	ErrAPIKeyInactive  = errors.New("api key is inactive")
	ErrAPIKeyRevoked   = errors.New("api key has been revoked")
	ErrInvalidKey      = errors.New("invalid api key")
	ErrKeyNameRequired = errors.New("api key name is required")
)

const (
	keyPrefix = "idp_"
	keyLength = 40 // length of random part (total: 4 + 40 = 44 chars)
)

// generateKey creates a new API key with prefix and random suffix
func generateKey() (string, error) {
	bytes := make([]byte, keyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return keyPrefix + hex.EncodeToString(bytes), nil
}

// hashKey returns the SHA-256 hash of a key
func hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// Create generates a new API key, hashes it, and stores it
func (u *usecase) Create(ctx context.Context, createdBy string, req apikey.CreateAPIKeyRequest) (*apikey.APIKeyResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrKeyNameRequired
	}

	plainKey, err := generateKey()
	if err != nil {
		return nil, err
	}

	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = apikey.DefaultScopes()
	}

	key := &apikey.APIKey{
		ID:          uuid.New().String(),
		Key:         hashKey(plainKey),
		Name:        req.Name,
		Description: req.Description,
		TeamID:      req.TeamID,
		CreatedBy:   createdBy,
		Scopes:      strings.Join(scopes, ","),
		IsAdmin:     req.IsAdmin,
		IsReadOnly:  req.IsReadOnly,
		RateLimit:   req.RateLimit,
		IsActive:    true,
		ExpiresAt:   req.ExpiresAt,
	}

	if key.RateLimit <= 0 {
		key.RateLimit = 100
	}

	if err := u.apiKeyRepo.Create(ctx, key); err != nil {
		return nil, err
	}

	resp := apikey.ToAPIKeyResponse(key, true)
	resp.Key = plainKey // return the plain key once
	return resp, nil
}

// Get retrieves an API key by ID
func (u *usecase) Get(ctx context.Context, id string) (*apikey.APIKeyResponse, error) {
	key, err := u.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, ErrAPIKeyNotFound
	}
	return apikey.ToAPIKeyResponse(key, false), nil
}

// List lists API keys, optionally filtered by team
func (u *usecase) List(ctx context.Context, teamID string) ([]apikey.APIKeyResponse, error) {
	var keys []apikey.APIKey
	var err error

	if teamID != "" {
		keys, err = u.apiKeyRepo.ListByTeam(ctx, teamID)
	} else {
		keys, err = u.apiKeyRepo.ListActive(ctx)
	}

	if err != nil {
		return nil, err
	}

	responses := make([]apikey.APIKeyResponse, len(keys))
	for i, k := range keys {
		responses[i] = *apikey.ToAPIKeyResponse(&k, false)
	}
	return responses, nil
}

// Update updates an API key
func (u *usecase) Update(ctx context.Context, id string, req apikey.CreateAPIKeyRequest) (*apikey.APIKeyResponse, error) {
	key, err := u.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, ErrAPIKeyNotFound
	}

	if req.Name != "" {
		key.Name = req.Name
	}
	if req.Description != "" {
		key.Description = req.Description
	}
	if len(req.Scopes) > 0 {
		key.Scopes = strings.Join(req.Scopes, ",")
	}
	if req.RateLimit > 0 {
		key.RateLimit = req.RateLimit
	}
	key.IsAdmin = req.IsAdmin
	key.IsReadOnly = req.IsReadOnly
	key.ExpiresAt = req.ExpiresAt

	if err := u.apiKeyRepo.Update(ctx, key); err != nil {
		return nil, err
	}

	return apikey.ToAPIKeyResponse(key, false), nil
}

// Delete soft deletes an API key
func (u *usecase) Delete(ctx context.Context, id string) error {
	key, err := u.apiKeyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if key == nil {
		return ErrAPIKeyNotFound
	}
	return u.apiKeyRepo.Delete(ctx, id)
}

// Validate checks if an API key is valid and returns it
func (u *usecase) Validate(ctx context.Context, key string) (*apikey.APIKey, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	hashed := hashKey(key)
	apiKey, err := u.apiKeyRepo.GetByKey(ctx, hashed)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, ErrInvalidKey
	}
	if !apiKey.IsActive {
		return nil, ErrAPIKeyInactive
	}
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, ErrAPIKeyExpired
	}

	// Update usage tracking (best effort, don't fail auth on tracking error)
	_ = u.apiKeyRepo.UpdateLastUsed(ctx, apiKey.ID)

	return apiKey, nil
}
