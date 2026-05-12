package apikey

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/model/apikey"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	key, err := generateKey()
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, keyPrefix))
	// prefix (4) + hex(40 bytes) = 4 + 80 = 84
	assert.Len(t, key, 4+80)
}

func TestGenerateKey_Uniqueness(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := generateKey()
		assert.NoError(t, err)
		assert.False(t, keys[key], "duplicate key generated")
		keys[key] = true
	}
}

func TestHashKey(t *testing.T) {
	hash1 := hashKey("test-key")
	hash2 := hashKey("test-key")
	assert.Equal(t, hash1, hash2, "same input should produce same hash")
	assert.Len(t, hash1, 64, "SHA-256 produces 64 hex chars")

	hash3 := hashKey("different-key")
	assert.NotEqual(t, hash1, hash3, "different inputs should produce different hashes")
}

func TestHashKey_Deterministic(t *testing.T) {
	// SHA-256 should be deterministic
	key := "idp_my-secret-api-key"
	assert.Equal(t, hashKey(key), hashKey(key))
}

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockApiKeyRepository(ctrl)
	uc := New(Dependencies{APIKeyRepo: mockRepo})

	t.Run("success with defaults", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		req := apikey.CreateAPIKeyRequest{
			Name: "test-key",
		}

		resp, err := uc.Create(context.Background(), "user-1", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "test-key", resp.Name)
		assert.NotEmpty(t, resp.Key, "plain key should be returned on creation")
		assert.True(t, strings.HasPrefix(resp.Key, keyPrefix))
		assert.True(t, resp.IsActive)
	})

	t.Run("empty name returns error", func(t *testing.T) {
		req := apikey.CreateAPIKeyRequest{
			Name: "",
		}
		resp, err := uc.Create(context.Background(), "user-1", req)
		assert.ErrorIs(t, err, ErrKeyNameRequired)
		assert.Nil(t, resp)
	})

	t.Run("whitespace-only name returns error", func(t *testing.T) {
		req := apikey.CreateAPIKeyRequest{
			Name: "   ",
		}
		resp, err := uc.Create(context.Background(), "user-1", req)
		assert.ErrorIs(t, err, ErrKeyNameRequired)
		assert.Nil(t, resp)
	})

	t.Run("with team and expiry", func(t *testing.T) {
		expiry := time.Now().Add(24 * time.Hour)
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		req := apikey.CreateAPIKeyRequest{
			Name:      "team-key",
			TeamID:    "team-1",
			ExpiresAt: &expiry,
		}

		resp, err := uc.Create(context.Background(), "user-1", req)
		assert.NoError(t, err)
		assert.Equal(t, "team-1", resp.TeamID)
		assert.NotNil(t, resp.ExpiresAt)
	})

	t.Run("repo error propagates", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("db error"))

		req := apikey.CreateAPIKeyRequest{
			Name: "error-key",
		}
		resp, err := uc.Create(context.Background(), "user-1", req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockApiKeyRepository(ctrl)
	uc := New(Dependencies{APIKeyRepo: mockRepo})

	t.Run("found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "key-1").
			Return(&apikey.APIKey{
				ID:   "key-1",
				Name: "my-key",
			}, nil)

		resp, err := uc.Get(context.Background(), "key-1")
		assert.NoError(t, err)
		assert.Equal(t, "key-1", resp.ID)
		assert.Equal(t, "my-key", resp.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, nil)

		resp, err := uc.Get(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
		assert.Nil(t, resp)
	})
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockApiKeyRepository(ctrl)
	uc := New(Dependencies{APIKeyRepo: mockRepo})

	t.Run("list by team", func(t *testing.T) {
		mockRepo.EXPECT().
			ListByTeam(gomock.Any(), "team-1").
			Return([]apikey.APIKey{
				{ID: "key-1", Name: "key-a"},
				{ID: "key-2", Name: "key-b"},
			}, nil)

		resp, err := uc.List(context.Background(), "team-1")
		assert.NoError(t, err)
		assert.Len(t, resp, 2)
	})

	t.Run("list all active", func(t *testing.T) {
		mockRepo.EXPECT().
			ListActive(gomock.Any()).
			Return([]apikey.APIKey{
				{ID: "key-1", Name: "active-key"},
			}, nil)

		resp, err := uc.List(context.Background(), "")
		assert.NoError(t, err)
		assert.Len(t, resp, 1)
	})
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockApiKeyRepository(ctrl)
	uc := New(Dependencies{APIKeyRepo: mockRepo})

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "key-1").
			Return(&apikey.APIKey{ID: "key-1"}, nil)
		mockRepo.EXPECT().
			Delete(gomock.Any(), "key-1").
			Return(nil)

		err := uc.Delete(context.Background(), "key-1")
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, nil)

		err := uc.Delete(context.Background(), "nonexistent")
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
	})
}

func TestValidate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockApiKeyRepository(ctrl)
	uc := New(Dependencies{APIKeyRepo: mockRepo})

	t.Run("valid key", func(t *testing.T) {
		plainKey := "idp_testkey1234"
		hashed := hashKey(plainKey)

		mockRepo.EXPECT().
			GetByKey(gomock.Any(), hashed).
			Return(&apikey.APIKey{
				ID:       "key-1",
				Key:      hashed,
				IsActive: true,
			}, nil)
		mockRepo.EXPECT().
			UpdateLastUsed(gomock.Any(), "key-1").
			Return(nil)

		result, err := uc.Validate(context.Background(), plainKey)
		assert.NoError(t, err)
		assert.Equal(t, "key-1", result.ID)
	})

	t.Run("empty key", func(t *testing.T) {
		result, err := uc.Validate(context.Background(), "")
		assert.ErrorIs(t, err, ErrInvalidKey)
		assert.Nil(t, result)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByKey(gomock.Any(), gomock.Any()).
			Return(nil, nil)

		result, err := uc.Validate(context.Background(), "unknown-key")
		assert.ErrorIs(t, err, ErrInvalidKey)
		assert.Nil(t, result)
	})

	t.Run("inactive key", func(t *testing.T) {
		hashed := hashKey("inactive-key")
		mockRepo.EXPECT().
			GetByKey(gomock.Any(), hashed).
			Return(&apikey.APIKey{
				ID:       "key-2",
				IsActive: false,
			}, nil)

		result, err := uc.Validate(context.Background(), "inactive-key")
		assert.ErrorIs(t, err, ErrAPIKeyInactive)
		assert.Nil(t, result)
	})

	t.Run("expired key", func(t *testing.T) {
		hashed := hashKey("expired-key")
		past := time.Now().Add(-1 * time.Hour)
		mockRepo.EXPECT().
			GetByKey(gomock.Any(), hashed).
			Return(&apikey.APIKey{
				ID:        "key-3",
				IsActive:  true,
				ExpiresAt: &past,
			}, nil)

		result, err := uc.Validate(context.Background(), "expired-key")
		assert.ErrorIs(t, err, ErrAPIKeyExpired)
		assert.Nil(t, result)
	})

	t.Run("update last used failure does not block auth", func(t *testing.T) {
		plainKey := "idp_trackingfail"
		hashed := hashKey(plainKey)

		mockRepo.EXPECT().
			GetByKey(gomock.Any(), hashed).
			Return(&apikey.APIKey{
				ID:       "key-4",
				IsActive: true,
			}, nil)
		mockRepo.EXPECT().
			UpdateLastUsed(gomock.Any(), "key-4").
			Return(errors.New("tracking failed"))

		result, err := uc.Validate(context.Background(), plainKey)
		assert.NoError(t, err, "auth should succeed even if tracking fails")
		assert.Equal(t, "key-4", result.ID)
	})
}

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockApiKeyRepository(ctrl)
	uc := New(Dependencies{APIKeyRepo: mockRepo})

	t.Run("update name", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "key-1").
			Return(&apikey.APIKey{ID: "key-1", Name: "old-name", IsActive: true}, nil)
		mockRepo.EXPECT().
			Update(gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := uc.Update(context.Background(), "key-1", apikey.CreateAPIKeyRequest{
			Name: "new-name",
		})
		assert.NoError(t, err)
		assert.Equal(t, "new-name", resp.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, nil)

		resp, err := uc.Update(context.Background(), "nonexistent", apikey.CreateAPIKeyRequest{
			Name: "new-name",
		})
		assert.ErrorIs(t, err, ErrAPIKeyNotFound)
		assert.Nil(t, resp)
	})
}
