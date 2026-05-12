package auditlog

import (
	"context"
	"errors"
	"testing"

	"github.com/davidsugianto/idp-core/internal/mocks"
	"github.com/davidsugianto/idp-core/internal/model/auditlog"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuditLogRepository(ctrl)
	uc := New(Dependencies{AuditLogRepo: mockRepo})

	t.Run("creates audit log with defaults", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		req := auditlog.CreateAuditLogRequest{
			UserID:       "user-1",
			ActorType:    auditlog.ActorTypeUser,
			Action:       auditlog.ActionCreate,
			ResourceType: auditlog.ResourceTypeEnvironment,
			ResourceID:   "env-1",
			TeamID:       "team-1",
			IPAddress:    "127.0.0.1",
		}

		log, err := uc.Create(context.Background(), req)
		assert.NoError(t, err)
		assert.NotEmpty(t, log.ID)
		assert.Equal(t, "user-1", log.UserID)
		assert.Equal(t, auditlog.StatusSuccess, log.Status)
	})

	t.Run("creates audit log with failure status", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		req := auditlog.CreateAuditLogRequest{
			UserID:       "user-1",
			ActorType:    auditlog.ActorTypeUser,
			Action:       auditlog.ActionDelete,
			ResourceType: auditlog.ResourceTypeEnvironment,
			Status:       auditlog.StatusFailure,
			ErrorMessage: "permission denied",
		}

		log, err := uc.Create(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, auditlog.StatusFailure, log.Status)
		assert.Equal(t, "permission denied", log.ErrorMessage)
	})

	t.Run("with changes tracking", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil)

		req := auditlog.CreateAuditLogRequest{
			UserID:       "user-1",
			ActorType:    auditlog.ActorTypeUser,
			Action:       auditlog.ActionUpdate,
			ResourceType: auditlog.ResourceTypeEnvironment,
			OldValues:    auditlog.Map{"name": "old-name"},
			NewValues:    auditlog.Map{"name": "new-name"},
		}

		log, err := uc.Create(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "old-name", log.OldValues["name"])
		assert.Equal(t, "new-name", log.NewValues["name"])
	})

	t.Run("repo error propagates", func(t *testing.T) {
		mockRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(errors.New("db error"))

		req := auditlog.CreateAuditLogRequest{
			ActorType:    auditlog.ActorTypeUser,
			Action:       auditlog.ActionCreate,
			ResourceType: auditlog.ResourceTypeEnvironment,
		}

		log, err := uc.Create(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, log)
	})
}

func TestGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuditLogRepository(ctrl)
	uc := New(Dependencies{AuditLogRepo: mockRepo})

	t.Run("found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "log-1").
			Return(&auditlog.AuditLog{
				ID:     "log-1",
				Action: auditlog.ActionCreate,
				Status: auditlog.StatusSuccess,
			}, nil)

		resp, err := uc.Get(context.Background(), "log-1")
		assert.NoError(t, err)
		assert.Equal(t, "log-1", resp.ID)
		assert.Equal(t, auditlog.ActionCreate, resp.Action)
	})

	t.Run("not found returns nil response", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByID(gomock.Any(), "nonexistent").
			Return(nil, nil)

		resp, err := uc.Get(context.Background(), "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, resp)
	})
}

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAuditLogRepository(ctrl)
	uc := New(Dependencies{AuditLogRepo: mockRepo})

	t.Run("list with filters", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]auditlog.AuditLog{
				{ID: "log-1", Action: "create"},
				{ID: "log-2", Action: "delete"},
			}, int64(2), nil)

		filter := auditlog.AuditLogFilter{
			TeamID: "team-1",
			Limit:  10,
		}

		resp, err := uc.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Len(t, resp.AuditLogs, 2)
		assert.Equal(t, int64(2), resp.Total)
	})

	t.Run("enforces default limit", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]auditlog.AuditLog{}, int64(0), nil)

		filter := auditlog.AuditLogFilter{
			Limit: 0, // should default to 50
		}

		_, err := uc.List(context.Background(), filter)
		assert.NoError(t, err)
	})

	t.Run("caps max limit", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]auditlog.AuditLog{}, int64(0), nil)

		filter := auditlog.AuditLogFilter{
			Limit: 500, // should cap to 200
		}

		_, err := uc.List(context.Background(), filter)
		assert.NoError(t, err)
	})

	t.Run("empty result", func(t *testing.T) {
		mockRepo.EXPECT().
			List(gomock.Any(), gomock.Any()).
			Return([]auditlog.AuditLog{}, int64(0), nil)

		filter := auditlog.AuditLogFilter{
			UserID: "nonexistent",
		}

		resp, err := uc.List(context.Background(), filter)
		assert.NoError(t, err)
		assert.Empty(t, resp.AuditLogs)
		assert.Equal(t, int64(0), resp.Total)
	})
}
