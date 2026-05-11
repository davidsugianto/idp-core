package team

import (
	"context"
	"errors"
	"testing"

	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/model/team"
	"github.com/davidsugianto/idp-core/internal/model/user"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := New(Dependencies{
		TeamRepo: mockTeamRepo,
		UserRepo: mockUserRepo,
	})

	tests := []struct {
		name    string
		req     team.CreateTeamRequest
		setup   func()
		wantErr bool
	}{
		{
			name: "successful creation",
			req: team.CreateTeamRequest{
				Name:        "Test Team",
				Slug:        "test-team",
				Description: "A test team",
			},
			setup: func() {
				mockTeamRepo.EXPECT().
					GetBySlug(gomock.Any(), "test-team").
					Return(nil, nil)
				mockTeamRepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name: "team already exists",
			req: team.CreateTeamRequest{
				Name: "Existing Team",
				Slug: "existing-team",
			},
			setup: func() {
				mockTeamRepo.EXPECT().
					GetBySlug(gomock.Any(), "existing-team").
					Return(&team.Team{ID: "team-1", Slug: "existing-team"}, nil)
			},
			wantErr: true,
		},
		{
			name: "invalid slug",
			req: team.CreateTeamRequest{
				Name: "Invalid Team",
				Slug: "INVALID_SLUG!",
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

func TestGetTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := New(Dependencies{
		TeamRepo: mockTeamRepo,
		UserRepo: mockUserRepo,
	})

	tests := []struct {
		name    string
		id      string
		setup   func()
		wantErr bool
	}{
		{
			name: "get team successfully",
			id:   "team-1",
			setup: func() {
				mockTeamRepo.EXPECT().
					GetByID(gomock.Any(), "team-1").
					Return(&team.Team{ID: "team-1", Name: "Test Team", Slug: "test-team"}, nil)
			},
			wantErr: false,
		},
		{
			name: "team not found",
			id:   "nonexistent",
			setup: func() {
				mockTeamRepo.EXPECT().
					GetByID(gomock.Any(), "nonexistent").
					Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "invalid team id",
			id:   "",
			setup: func() {},
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

func TestListTeams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := New(Dependencies{
		TeamRepo: mockTeamRepo,
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
			name:   "list teams successfully",
			limit:  20,
			offset: 0,
			setup: func() {
				mockTeamRepo.EXPECT().
					List(gomock.Any(), 20, 0).
					Return([]team.Team{
						{ID: "team-1", Name: "Team 1"},
						{ID: "team-2", Name: "Team 2"},
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
				mockTeamRepo.EXPECT().
					List(gomock.Any(), 20, 0).
					Return([]team.Team{}, int64(0), nil)
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:   "db error",
			limit:  20,
			offset: 0,
			setup: func() {
				mockTeamRepo.EXPECT().
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
				assert.Len(t, result.Teams, tt.wantLen)
			}
		})
	}
}

func TestAddMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	uc := New(Dependencies{
		TeamRepo: mockTeamRepo,
		UserRepo: mockUserRepo,
	})

	tests := []struct {
		name    string
		teamID  string
		req     team.AddTeamMemberRequest
		setup   func()
		wantErr bool
	}{
		{
			name:   "add member successfully",
			teamID: "team-1",
			req: team.AddTeamMemberRequest{
				UserID: "user-1",
				Role:   team.RoleMember,
			},
			setup: func() {
				mockTeamRepo.EXPECT().
					GetByID(gomock.Any(), "team-1").
					Return(&team.Team{ID: "team-1"}, nil)
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), "user-1").
					Return(&user.User{ID: "user-1", Email: "test@example.com"}, nil)
				mockTeamRepo.EXPECT().
					GetMember(gomock.Any(), "team-1", "user-1").
					Return(nil, nil)
				mockTeamRepo.EXPECT().
					AddMember(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "member already exists",
			teamID: "team-1",
			req: team.AddTeamMemberRequest{
				UserID: "user-1",
				Role:   team.RoleMember,
			},
			setup: func() {
				mockTeamRepo.EXPECT().
					GetByID(gomock.Any(), "team-1").
					Return(&team.Team{ID: "team-1"}, nil)
				mockUserRepo.EXPECT().
					GetByID(gomock.Any(), "user-1").
					Return(&user.User{ID: "user-1"}, nil)
				mockTeamRepo.EXPECT().
					GetMember(gomock.Any(), "team-1", "user-1").
					Return(&team.TeamMember{ID: "member-1", TeamID: "team-1", UserID: "user-1"}, nil)
			},
			wantErr: true,
		},
		{
			name:   "user not found",
			teamID: "team-1",
			req: team.AddTeamMemberRequest{
				UserID: "nonexistent",
				Role:   team.RoleMember,
			},
			setup: func() {
				mockTeamRepo.EXPECT().
					GetByID(gomock.Any(), "team-1").
					Return(&team.Team{ID: "team-1"}, nil)
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
			result, err := uc.AddMember(context.Background(), tt.teamID, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.teamID, result.TeamID)
				assert.Equal(t, tt.req.UserID, result.UserID)
			}
		})
	}
}

func TestRemoveMember(t *testing.T) {
	tests := []struct {
		name    string
		teamID  string
		userID  string
		setup   func(ctrl *gomock.Controller) (Usecase, *mocks.MockTeamRepository)
		wantErr bool
	}{
		{
			name:   "remove member successfully",
			teamID: "team-1",
			userID: "user-1",
			setup: func(ctrl *gomock.Controller) (Usecase, *mocks.MockTeamRepository) {
				mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
				mockUserRepo := mocks.NewMockUserRepository(ctrl)
				uc := New(Dependencies{
					TeamRepo: mockTeamRepo,
					UserRepo: mockUserRepo,
				})
				mockTeamRepo.EXPECT().
					GetByID(gomock.Any(), "team-1").
					Return(&team.Team{ID: "team-1"}, nil)
				mockTeamRepo.EXPECT().
					GetMember(gomock.Any(), "team-1", "user-1").
					Return(&team.TeamMember{ID: "member-1", TeamID: "team-1", UserID: "user-1", Role: team.RoleMember}, nil)
				mockTeamRepo.EXPECT().
					RemoveMember(gomock.Any(), "team-1", "user-1").
					Return(nil)
				return uc, mockTeamRepo
			},
			wantErr: false,
		},
		{
			name:   "cannot remove last owner",
			teamID: "team-1",
			userID: "user-1",
			setup: func(ctrl *gomock.Controller) (Usecase, *mocks.MockTeamRepository) {
				mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
				mockUserRepo := mocks.NewMockUserRepository(ctrl)
				uc := New(Dependencies{
					TeamRepo: mockTeamRepo,
					UserRepo: mockUserRepo,
				})
				mockTeamRepo.EXPECT().
					GetByID(gomock.Any(), "team-1").
					Return(&team.Team{ID: "team-1"}, nil)
				mockTeamRepo.EXPECT().
					GetMember(gomock.Any(), "team-1", "user-1").
					Return(&team.TeamMember{ID: "member-1", TeamID: "team-1", UserID: "user-1", Role: team.RoleOwner}, nil)
				mockTeamRepo.EXPECT().
					ListMembers(gomock.Any(), "team-1").
					Return([]team.TeamMember{
						{ID: "member-1", Role: team.RoleOwner},
					}, nil)
				return uc, mockTeamRepo
			},
			wantErr: true,
		},
		{
			name:   "member not found",
			teamID: "team-1",
			userID: "nonexistent",
			setup: func(ctrl *gomock.Controller) (Usecase, *mocks.MockTeamRepository) {
				mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
				mockUserRepo := mocks.NewMockUserRepository(ctrl)
				uc := New(Dependencies{
					TeamRepo: mockTeamRepo,
					UserRepo: mockUserRepo,
				})
				mockTeamRepo.EXPECT().
					GetByID(gomock.Any(), "team-1").
					Return(&team.Team{ID: "team-1"}, nil)
				mockTeamRepo.EXPECT().
					GetMember(gomock.Any(), "team-1", "nonexistent").
					Return(nil, nil)
				return uc, mockTeamRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			uc, _ := tt.setup(ctrl)
			err := uc.RemoveMember(context.Background(), tt.teamID, tt.userID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
