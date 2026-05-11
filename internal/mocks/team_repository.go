package mocks

import (
	"context"
	"reflect"

	"github.com/davidsugianto/idp-core/internal/model/team"
	"github.com/golang/mock/gomock"
)

// MockTeamRepository is a mock implementation of the Team repository
type MockTeamRepository struct {
	ctrl     *gomock.Controller
	recorder *MockTeamRepositoryMockRecorder
}

type MockTeamRepositoryMockRecorder struct {
	mock *MockTeamRepository
}

func NewMockTeamRepository(ctrl *gomock.Controller) *MockTeamRepository {
	mock := &MockTeamRepository{ctrl: ctrl}
	mock.recorder = &MockTeamRepositoryMockRecorder{mock}
	return mock
}

func (m *MockTeamRepository) EXPECT() *MockTeamRepositoryMockRecorder {
	return m.recorder
}

// Team CRUD
func (m *MockTeamRepository) Create(ctx context.Context, t *team.Team) error {
	ret := m.ctrl.Call(m, "Create", ctx, t)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTeamRepositoryMockRecorder) Create(ctx, t interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockTeamRepository)(nil).Create), ctx, t)
}

func (m *MockTeamRepository) GetByID(ctx context.Context, id string) (*team.Team, error) {
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*team.Team)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTeamRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockTeamRepository)(nil).GetByID), ctx, id)
}

func (m *MockTeamRepository) GetBySlug(ctx context.Context, slug string) (*team.Team, error) {
	ret := m.ctrl.Call(m, "GetBySlug", ctx, slug)
	ret0, _ := ret[0].(*team.Team)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTeamRepositoryMockRecorder) GetBySlug(ctx, slug interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBySlug", reflect.TypeOf((*MockTeamRepository)(nil).GetBySlug), ctx, slug)
}

func (m *MockTeamRepository) List(ctx context.Context, limit, offset int) ([]team.Team, int64, error) {
	ret := m.ctrl.Call(m, "List", ctx, limit, offset)
	ret0, _ := ret[0].([]team.Team)
	ret1, _ := ret[1].(int64)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockTeamRepositoryMockRecorder) List(ctx, limit, offset interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockTeamRepository)(nil).List), ctx, limit, offset)
}

func (m *MockTeamRepository) ListByStatus(ctx context.Context, status string) ([]team.Team, error) {
	ret := m.ctrl.Call(m, "ListByStatus", ctx, status)
	ret0, _ := ret[0].([]team.Team)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTeamRepositoryMockRecorder) ListByStatus(ctx, status interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByStatus", reflect.TypeOf((*MockTeamRepository)(nil).ListByStatus), ctx, status)
}

func (m *MockTeamRepository) Update(ctx context.Context, t *team.Team) error {
	ret := m.ctrl.Call(m, "Update", ctx, t)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTeamRepositoryMockRecorder) Update(ctx, t interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockTeamRepository)(nil).Update), ctx, t)
}

func (m *MockTeamRepository) SoftDelete(ctx context.Context, id string) error {
	ret := m.ctrl.Call(m, "SoftDelete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTeamRepositoryMockRecorder) SoftDelete(ctx, id interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SoftDelete", reflect.TypeOf((*MockTeamRepository)(nil).SoftDelete), ctx, id)
}

// Team Member operations
func (m *MockTeamRepository) AddMember(ctx context.Context, member *team.TeamMember) error {
	ret := m.ctrl.Call(m, "AddMember", ctx, member)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTeamRepositoryMockRecorder) AddMember(ctx, member interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMember", reflect.TypeOf((*MockTeamRepository)(nil).AddMember), ctx, member)
}

func (m *MockTeamRepository) GetMember(ctx context.Context, teamID, userID string) (*team.TeamMember, error) {
	ret := m.ctrl.Call(m, "GetMember", ctx, teamID, userID)
	ret0, _ := ret[0].(*team.TeamMember)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTeamRepositoryMockRecorder) GetMember(ctx, teamID, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMember", reflect.TypeOf((*MockTeamRepository)(nil).GetMember), ctx, teamID, userID)
}

func (m *MockTeamRepository) ListMembers(ctx context.Context, teamID string) ([]team.TeamMember, error) {
	ret := m.ctrl.Call(m, "ListMembers", ctx, teamID)
	ret0, _ := ret[0].([]team.TeamMember)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTeamRepositoryMockRecorder) ListMembers(ctx, teamID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMembers", reflect.TypeOf((*MockTeamRepository)(nil).ListMembers), ctx, teamID)
}

func (m *MockTeamRepository) ListTeamsByUser(ctx context.Context, userID string) ([]team.TeamMember, error) {
	ret := m.ctrl.Call(m, "ListTeamsByUser", ctx, userID)
	ret0, _ := ret[0].([]team.TeamMember)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockTeamRepositoryMockRecorder) ListTeamsByUser(ctx, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListTeamsByUser", reflect.TypeOf((*MockTeamRepository)(nil).ListTeamsByUser), ctx, userID)
}

func (m *MockTeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID, role string) error {
	ret := m.ctrl.Call(m, "UpdateMemberRole", ctx, teamID, userID, role)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTeamRepositoryMockRecorder) UpdateMemberRole(ctx, teamID, userID, role interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateMemberRole", reflect.TypeOf((*MockTeamRepository)(nil).UpdateMemberRole), ctx, teamID, userID, role)
}

func (m *MockTeamRepository) RemoveMember(ctx context.Context, teamID, userID string) error {
	ret := m.ctrl.Call(m, "RemoveMember", ctx, teamID, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockTeamRepositoryMockRecorder) RemoveMember(ctx, teamID, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveMember", reflect.TypeOf((*MockTeamRepository)(nil).RemoveMember), ctx, teamID, userID)
}

func (m *MockTeamRepository) IsTeamMember(ctx context.Context, teamID, userID string) (bool, string, error) {
	ret := m.ctrl.Call(m, "IsTeamMember", ctx, teamID, userID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockTeamRepositoryMockRecorder) IsTeamMember(ctx, teamID, userID interface{}) *gomock.Call {
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsTeamMember", reflect.TypeOf((*MockTeamRepository)(nil).IsTeamMember), ctx, teamID, userID)
}
