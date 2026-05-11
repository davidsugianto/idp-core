package user

import (
	"context"
	"strings"

	"github.com/davidsugianto/idp-core/internal/model/user"
)

// Create creates a new user
func (u *usecase) Create(ctx context.Context, req user.CreateUserRequest) (*user.User, error) {
	// Check if user already exists
	existing, err := u.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// Create new user
	newUser := &user.User{
		Email:     strings.ToLower(req.Email),
		Name:      req.Name,
		Provider:  req.Provider,
		AvatarURL: req.AvatarURL,
	}

	if newUser.Provider == "" {
		newUser.Provider = "local"
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	return newUser, nil
}

// Get retrieves a user by ID
func (u *usecase) Get(ctx context.Context, id string) (*user.User, error) {
	if id == "" {
		return nil, ErrInvalidUserID
	}

	usr, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if usr == nil {
		return nil, ErrUserNotFound
	}

	return usr, nil
}

// GetByEmail retrieves a user by email
func (u *usecase) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	usr, err := u.userRepo.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil, err
	}
	if usr == nil {
		return nil, ErrUserNotFound
	}
	return usr, nil
}

// List retrieves a paginated list of users
func (u *usecase) List(ctx context.Context, limit, offset int) (*user.UserListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := u.userRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return user.ToUserListResponse(users, total), nil
}

// Update updates a user
func (u *usecase) Update(ctx context.Context, id string, req user.UpdateUserRequest) (*user.User, error) {
	usr, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		usr.Name = req.Name
	}
	if req.AvatarURL != "" {
		usr.AvatarURL = req.AvatarURL
	}
	if req.Status != "" {
		usr.Status = req.Status
	}

	if err := u.userRepo.Update(ctx, usr); err != nil {
		return nil, err
	}

	return usr, nil
}

// UpdateStatus updates a user's status
func (u *usecase) UpdateStatus(ctx context.Context, id, status string) error {
	if id == "" {
		return ErrInvalidUserID
	}

	usr, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if usr == nil {
		return ErrUserNotFound
	}

	return u.userRepo.UpdateStatus(ctx, id, status)
}

// Delete soft deletes a user
func (u *usecase) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidUserID
	}

	usr, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if usr == nil {
		return ErrUserNotFound
	}

	return u.userRepo.SoftDelete(ctx, id)
}
