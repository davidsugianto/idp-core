package user

import (
	"context"
	"errors"

	"github.com/davidsugianto/idp-core/internal/model/user"
	userRepo "github.com/davidsugianto/idp-core/internal/repository/user"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserID     = errors.New("invalid user id")
)

// Usecase defines the interface for user business logic
type Usecase interface {
	Create(ctx context.Context, req user.CreateUserRequest) (*user.User, error)
	Get(ctx context.Context, id string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	List(ctx context.Context, limit, offset int) (*user.UserListResponse, error)
	Update(ctx context.Context, id string, req user.UpdateUserRequest) (*user.User, error)
	UpdateStatus(ctx context.Context, id, status string) error
	Delete(ctx context.Context, id string) error
}

type usecase struct {
	userRepo userRepo.Repository
}

// Dependencies holds usecase dependencies
type Dependencies struct {
	UserRepo userRepo.Repository
}

// New creates a new user usecase
func New(deps Dependencies) Usecase {
	return &usecase{
		userRepo: deps.UserRepo,
	}
}
