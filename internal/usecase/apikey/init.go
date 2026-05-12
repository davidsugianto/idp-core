package apikey

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/apikey"
	apikeyRepo "github.com/davidsugianto/idp-core/internal/repository/apikey"
)

// Usecase defines the interface for API key business logic
type Usecase interface {
	Create(ctx context.Context, createdBy string, req apikey.CreateAPIKeyRequest) (*apikey.APIKeyResponse, error)
	Get(ctx context.Context, id string) (*apikey.APIKeyResponse, error)
	List(ctx context.Context, teamID string) ([]apikey.APIKeyResponse, error)
	Update(ctx context.Context, id string, req apikey.CreateAPIKeyRequest) (*apikey.APIKeyResponse, error)
	Delete(ctx context.Context, id string) error
	Validate(ctx context.Context, key string) (*apikey.APIKey, error)
}

type usecase struct {
	apiKeyRepo apikeyRepo.Repository
}

type Dependencies struct {
	APIKeyRepo apikeyRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{apiKeyRepo: deps.APIKeyRepo}
}
