package role

import (
	"context"
	"errors"
	"testing"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	permissionModel "github.com/davidsugianto/idp-core/internal/model/permission"
	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoleRepo := mocks.NewMockRoleRepository(ctrl)
	mockPermRepo := mocks.NewMockPermissionRepository(ctrl)

	uc := New(Dependencies{
		RoleRepo:       mockRoleRepo,
		PermissionRepo: mockPermRepo,
	})

	tests := []struct {
		name    string
		req     roleModel.CreateRoleRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "successful creation",
			req: roleModel.CreateRoleRequest{
				Name:        "test-role",
				Description: "A test role",
				Scope:       roleModel.ScopeTeam,
			},
			setup: func() {
				mockRoleRepo.EXPECT().
					GetByName(gomock.Any(), "test-role").
					Return(nil, nil)
				mockRoleRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
				mockRoleRepo.EXPECT().
					GetByID(gomock.Any(), gomock.Any()).
					Return(&roleModel.Role{ID: "role-1", Name: "test-role", Scope: roleModel.ScopeTeam}, nil)
			},
			wantErr: false,
		},
		{
			name: "role already exists",
			req: roleModel.CreateRoleRequest{
				Name:  "existing-role",
				Scope: roleModel.ScopeTeam,
			},
			setup: func() {
				mockRoleRepo.EXPECT().
					GetByName(gomock.Any(), "existing-role").
					Return(&roleModel.Role{ID: "role-1", Name: "existing-role"}, nil)
			},
			wantErr: true,
		},
		{
			name: "invalid scope",
			req: roleModel.CreateRoleRequest{
				Name:  "invalid-role",
				Scope: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := uc.Create(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
			}
		})
	}
}

func TestGetRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoleRepo := mocks.NewMockRoleRepository(ctrl)
	mockPermRepo := mocks.NewMockPermissionRepository(ctrl)

	uc := New(Dependencies{
		RoleRepo:       mockRoleRepo,
		PermissionRepo: mockPermRepo,
	})

	tests := []struct {
		name    string
		id      string
		setup   func()
		wantErr bool
	}{
		{
			name: "get role successfully",
			id:   "role-1",
			setup: func() {
				mockRoleRepo.EXPECT().
					GetByID(gomock.Any(), "role-1").
					Return(&roleModel.Role{ID: "role-1", Name: "Test Role"}, nil)
			},
			wantErr: false,
		},
		{
			name: "role not found",
			id:   "nonexistent",
			setup: func() {
				mockRoleRepo.EXPECT().
					GetByID(gomock.Any(), "nonexistent").
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:    "invalid role id",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
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

func TestListRoles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoleRepo := mocks.NewMockRoleRepository(ctrl)
	mockPermRepo := mocks.NewMockPermissionRepository(ctrl)

	uc := New(Dependencies{
		RoleRepo:       mockRoleRepo,
		PermissionRepo: mockPermRepo,
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
			name:   "list roles successfully",
			limit:  20,
			offset: 0,
			setup: func() {
				mockRoleRepo.EXPECT().
					List(gomock.Any(), 20, 0).
					Return([]roleModel.Role{
						{ID: "role-1", Name: "Role 1"},
						{ID: "role-2", Name: "Role 2"},
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
				mockRoleRepo.EXPECT().
					List(gomock.Any(), 20, 0).
					Return([]roleModel.Role{}, int64(0), nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:   "db error",
			limit:  20,
			offset: 0,
			setup: func() {
				mockRoleRepo.EXPECT().
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
				assert.Len(t, result.Roles, tt.wantLen)
			}
		})
	}
}

func TestAssignRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoleRepo := mocks.NewMockRoleRepository(ctrl)
	mockPermRepo := mocks.NewMockPermissionRepository(ctrl)

	uc := New(Dependencies{
		RoleRepo:       mockRoleRepo,
		PermissionRepo: mockPermRepo,
	})

	tests := []struct {
		name      string
		req       roleModel.AssignRoleRequest
		grantedBy string
		setup     func()
		wantErr   bool
	}{
		{
			name: "assign team role successfully",
			req: roleModel.AssignRoleRequest{
				UserID: "user-1",
				RoleID: "role-1",
				TeamID: "team-1",
			},
			grantedBy: "admin-1",
			setup: func() {
				mockRoleRepo.EXPECT().
					GetByID(gomock.Any(), "role-1").
					Return(&roleModel.Role{ID: "role-1", Scope: roleModel.ScopeTeam}, nil)
				mockRoleRepo.EXPECT().
					AssignRole(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "role not found",
			req: roleModel.AssignRoleRequest{
				UserID: "user-1",
				RoleID: "nonexistent",
				TeamID: "team-1",
			},
			grantedBy: "admin-1",
			setup: func() {
				mockRoleRepo.EXPECT().
					GetByID(gomock.Any(), "nonexistent").
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "team-scoped role without team ID",
			req: roleModel.AssignRoleRequest{
				UserID: "user-1",
				RoleID: "role-1",
			},
			grantedBy: "admin-1",
			setup: func() {
				mockRoleRepo.EXPECT().
					GetByID(gomock.Any(), "role-1").
					Return(&roleModel.Role{ID: "role-1", Scope: roleModel.ScopeTeam}, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := uc.AssignRole(context.Background(), tt.req, tt.grantedBy)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.req.UserID, result.UserID)
				assert.Equal(t, tt.req.RoleID, result.RoleID)
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRoleRepo := mocks.NewMockRoleRepository(ctrl)
	mockPermRepo := mocks.NewMockPermissionRepository(ctrl)

	uc := New(Dependencies{
		RoleRepo:       mockRoleRepo,
		PermissionRepo: mockPermRepo,
	})

	tests := []struct {
		name     string
		userID   string
		resource string
		action   string
		setup    func()
		wantHas  bool
		wantErr  bool
	}{
		{
			name:     "has permission",
			userID:   "user-1",
			resource: permissionModel.ResourceEnvironment,
			action:   permissionModel.ActionRead,
			setup: func() {
				mockRoleRepo.EXPECT().
					HasPermission(gomock.Any(), "user-1", permissionModel.ResourceEnvironment, permissionModel.ActionRead).
					Return(true, nil)
			},
			wantHas: true,
			wantErr: false,
		},
		{
			name:     "no permission",
			userID:   "user-1",
			resource: permissionModel.ResourceEnvironment,
			action:   permissionModel.ActionDelete,
			setup: func() {
				mockRoleRepo.EXPECT().
					HasPermission(gomock.Any(), "user-1", permissionModel.ResourceEnvironment, permissionModel.ActionDelete).
					Return(false, nil)
			},
			wantHas: false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			has, err := uc.HasPermission(context.Background(), tt.userID, tt.resource, tt.action)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantHas, has)
			}
		})
	}
}
