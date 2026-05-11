package user

import (
	"context"
	"errors"
	"testing"

	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/model/user"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := New(Dependencies{
		UserRepo: mockUserRepo,
	})

	tests := []struct {
		name    string
		req     user.CreateUserRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "successful creation",
			req: user.CreateUserRequest{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setup: func() {
				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(nil, nil)
				mockUserRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user already exists",
			req: user.CreateUserRequest{
				Email: "existing@example.com",
				Name:  "Existing User",
			},
			setup: func() {
				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), "existing@example.com").
					Return(&user.User{ID: "user-1", Email: "existing@example.com"}, nil)
			},
			wantErr: true,
		},
		{
			name: "db error on create",
			req: user.CreateUserRequest{
				Email: "test@example.com",
				Name:  "Test User",
			},
			setup: func() {
				mockUserRepo.EXPECT().
					GetByEmail(gomock.Any(), "test@example.com").
					Return(nil, nil)
				mockUserRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.Create(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Email, result.Email)
				assert.Equal(t, tt.req.Name, result.Name)
			}
		})
	}
}

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := New(Dependencies{
		UserRepo: mockUserRepo,
	})

	tests := []struct {
		name    string
		id      string
		setup   func()
		wantErr bool
	}{
		{
			name: "get user successfully",
			id:   "user-1",
			setup: func() {
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), "user-1").
					Return(&user.User{ID: "user-1", Email: "test@example.com", Name: "Test User"}, nil)
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   "nonexistent",
			setup: func() {
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), "nonexistent").
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "invalid user id",
			id:   "",
			setup: func() {
				// No repo call expected
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.Get(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
			}
		})
	}
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := New(Dependencies{
		UserRepo: mockUserRepo,
	})

	tests := []struct {
		name    string
		limit   int
		offset  int
		setup   func()
		wantErr bool
		wantLen int
	}{
		{
			name:   "list users successfully",
			limit:  20,
			offset: 0,
			setup: func() {
				mockUserRepo.EXPECT().
					List(gomock.Any(), 20, 0).
					Return([]user.User{
						{ID: "user-1", Email: "user1@example.com"},
						{ID: "user-2", Email: "user2@example.com"},
					}, int64(2), nil)
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:   "empty list",
			limit:  20,
			offset: 0,
			setup: func() {
				mockUserRepo.EXPECT().
					List(gomock.Any(), 20, 0).
					Return([]user.User{}, int64(0), nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:   "db error",
			limit:  20,
			offset: 0,
			setup: func() {
				mockUserRepo.EXPECT().
					List(gomock.Any(), 20, 0).
					Return(nil, int64(0), errors.New("db error"))
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.List(context.Background(), tt.limit, tt.offset)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Users, tt.wantLen)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := New(Dependencies{
		UserRepo: mockUserRepo,
	})

	tests := []struct {
		name    string
		id      string
		req     user.UpdateUserRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "update user successfully",
			id:   "user-1",
			req: user.UpdateUserRequest{
				Name:   "Updated Name",
				Status: "active",
			},
			setup: func() {
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), "user-1").
					Return(&user.User{ID: "user-1", Email: "test@example.com", Name: "Test User"}, nil)
				mockUserRepo.EXPECT().
					Update(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   "nonexistent",
			req:  user.UpdateUserRequest{Name: "Updated Name"},
			setup: func() {
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), "nonexistent").
					Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			result, err := uc.Update(context.Background(), tt.id, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := New(Dependencies{
		UserRepo: mockUserRepo,
	})

	tests := []struct {
		name    string
		id      string
		setup   func()
		wantErr bool
	}{
		{
			name: "delete user successfully",
			id:   "user-1",
			setup: func() {
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), "user-1").
					Return(&user.User{ID: "user-1", Email: "test@example.com"}, nil)
				mockUserRepo.EXPECT().
					SoftDelete(gomock.Any(), "user-1").
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "user not found",
			id:   "nonexistent",
			setup: func() {
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), "nonexistent").
					Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := uc.Delete(context.Background(), tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
