package delivery_target

import (
	"context"
	"testing"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	environmentMovementModel "github.com/davidsugianto/idp-core/internal/model/environment_movement"
	deliveryTargetRepo "github.com/davidsugianto/idp-core/internal/repository/delivery_target"
	environmentRepo "github.com/davidsugianto/idp-core/internal/repository/environment"
	environmentMovementRepo "github.com/davidsugianto/idp-core/internal/repository/environment_movement"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type targetTestDeliveryTargetRepo struct {
	deliveryTargetRepo.Repository
	target    *deliveryTargetModel.DeliveryTarget
	deletedID string
}

func (r *targetTestDeliveryTargetRepo) GetByID(ctx context.Context, id string) (*deliveryTargetModel.DeliveryTarget, error) {
	if r.target != nil && r.target.ID == id {
		copy := *r.target
		return &copy, nil
	}
	return nil, nil
}

func (r *targetTestDeliveryTargetRepo) Delete(ctx context.Context, id string) error {
	r.deletedID = id
	return nil
}

type targetTestEnvironmentRepo struct {
	environmentRepo.Repository
	count int64
}

func (r *targetTestEnvironmentRepo) CountByDeliveryTarget(ctx context.Context, deliveryTargetID string) (int64, error) {
	return r.count, nil
}

type targetTestMovementRepo struct {
	environmentMovementRepo.Repository
	movements []environmentMovementModel.EnvironmentMovement
}

func (r *targetTestMovementRepo) ListActiveByTarget(ctx context.Context, targetID string) ([]environmentMovementModel.EnvironmentMovement, error) {
	return r.movements, nil
}

func TestDeleteRejectsTargetReferencedByEnvironment(t *testing.T) {
	targetRepo := &targetTestDeliveryTargetRepo{target: &deliveryTargetModel.DeliveryTarget{ID: "target-1"}}
	uc := New(Dependencies{
		DeliveryTargetRepo:      targetRepo,
		EnvironmentRepo:         &targetTestEnvironmentRepo{count: 1},
		EnvironmentMovementRepo: &targetTestMovementRepo{},
	})

	err := uc.Delete(context.Background(), "target-1")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDeliveryTargetInUse)
	assert.Empty(t, targetRepo.deletedID)
}

func TestDeleteRejectsTargetReferencedByActiveMovement(t *testing.T) {
	targetRepo := &targetTestDeliveryTargetRepo{target: &deliveryTargetModel.DeliveryTarget{ID: "target-1"}}
	uc := New(Dependencies{
		DeliveryTargetRepo: targetRepo,
		EnvironmentRepo:    &targetTestEnvironmentRepo{},
		EnvironmentMovementRepo: &targetTestMovementRepo{movements: []environmentMovementModel.EnvironmentMovement{
			{ID: "move-1", DestinationTargetID: "target-1", Status: environmentMovementModel.StatusRunning},
		}},
	})

	err := uc.Delete(context.Background(), "target-1")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDeliveryTargetInUse)
	assert.Empty(t, targetRepo.deletedID)
}

func TestDeleteAllowsUnusedTarget(t *testing.T) {
	targetRepo := &targetTestDeliveryTargetRepo{target: &deliveryTargetModel.DeliveryTarget{ID: "target-1"}}
	uc := New(Dependencies{
		DeliveryTargetRepo:      targetRepo,
		EnvironmentRepo:         &targetTestEnvironmentRepo{},
		EnvironmentMovementRepo: &targetTestMovementRepo{},
	})

	err := uc.Delete(context.Background(), "target-1")
	require.NoError(t, err)
	assert.Equal(t, "target-1", targetRepo.deletedID)
}
