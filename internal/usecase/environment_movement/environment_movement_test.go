package environment_movement

import (
	"context"
	"testing"
	"time"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	environmentModel "github.com/davidsugianto/idp-core/internal/model/environment"
	environmentMovementModel "github.com/davidsugianto/idp-core/internal/model/environment_movement"
	deliveryTargetRepo "github.com/davidsugianto/idp-core/internal/repository/delivery_target"
	environmentRepo "github.com/davidsugianto/idp-core/internal/repository/environment"
	environmentMovementRepo "github.com/davidsugianto/idp-core/internal/repository/environment_movement"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type movementTestEnvironmentRepo struct {
	environmentRepo.Repository
	env                  *environmentModel.Environment
	updatedTargetID      string
	updatedClusterName   string
	updatedClusterServer string
}

func (r *movementTestEnvironmentRepo) GetByIDAndTeam(ctx context.Context, id, teamID string) (*environmentModel.Environment, error) {
	if r.env != nil && r.env.ID == id && r.env.TeamID == teamID {
		copy := *r.env
		return &copy, nil
	}
	return nil, nil
}

func (r *movementTestEnvironmentRepo) UpdateDeliveryTarget(ctx context.Context, id, teamID, deliveryTargetID, clusterName, clusterServer string) error {
	r.updatedTargetID = deliveryTargetID
	r.updatedClusterName = clusterName
	r.updatedClusterServer = clusterServer
	if r.env != nil && r.env.ID == id && r.env.TeamID == teamID {
		r.env.DeliveryTargetID = deliveryTargetID
		r.env.ClusterName = clusterName
		r.env.ClusterServer = clusterServer
	}
	return nil
}

type movementTestDeliveryTargetRepo struct {
	deliveryTargetRepo.Repository
	target *deliveryTargetModel.DeliveryTarget
}

func (r *movementTestDeliveryTargetRepo) GetByID(ctx context.Context, id string) (*deliveryTargetModel.DeliveryTarget, error) {
	if r.target != nil && r.target.ID == id {
		copy := *r.target
		return &copy, nil
	}
	return nil, nil
}

type movementTestMovementRepo struct {
	environmentMovementRepo.Repository
	movement *environmentMovementModel.EnvironmentMovement
	updated  *environmentMovementModel.EnvironmentMovement
}

func (r *movementTestMovementRepo) GetByID(ctx context.Context, id string) (*environmentMovementModel.EnvironmentMovement, error) {
	if r.movement != nil && r.movement.ID == id {
		copy := *r.movement
		return &copy, nil
	}
	return nil, nil
}

func (r *movementTestMovementRepo) Update(ctx context.Context, movement *environmentMovementModel.EnvironmentMovement) error {
	copy := *movement
	r.updated = &copy
	r.movement = &copy
	return nil
}

func TestUpdateStatusPreservesResolvedTargetPlacement(t *testing.T) {
	envRepo := &movementTestEnvironmentRepo{
		env: &environmentModel.Environment{
			ID:               "env-1",
			TeamID:           "team-1",
			DeliveryTargetID: "target-old",
			ClusterName:      "cluster-old",
			ClusterServer:    "https://old.example",
		},
	}
	targetRepo := &movementTestDeliveryTargetRepo{
		target: &deliveryTargetModel.DeliveryTarget{
			ID:            "target-new",
			ClusterName:   "cluster-new",
			ClusterServer: "https://new.example",
		},
	}
	movementRepo := &movementTestMovementRepo{
		movement: &environmentMovementModel.EnvironmentMovement{
			ID:                  "move-1",
			EnvironmentID:       "env-1",
			DestinationTargetID: "target-new",
			Status:              environmentMovementModel.StatusRunning,
		},
	}

	uc := New(Dependencies{
		EnvironmentRepo:         envRepo,
		DeliveryTargetRepo:      targetRepo,
		EnvironmentMovementRepo: movementRepo,
	})

	result, err := uc.UpdateStatus(context.Background(), "team-1", "env-1", "move-1", environmentMovementModel.StatusCompleted, 100, "done")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "target-new", envRepo.updatedTargetID)
	assert.Equal(t, "cluster-new", envRepo.updatedClusterName)
	assert.Equal(t, "https://new.example", envRepo.updatedClusterServer)
	require.NotNil(t, movementRepo.updated)
	assert.Equal(t, environmentMovementModel.StatusCompleted, movementRepo.updated.Status)
	assert.Equal(t, 100, movementRepo.updated.ProgressPercent)
	assert.Equal(t, "done", movementRepo.updated.Message)
	assert.NotNil(t, movementRepo.updated.CompletedAt)
	assert.WithinDuration(t, time.Now(), *movementRepo.updated.CompletedAt, 2*time.Second)
}
