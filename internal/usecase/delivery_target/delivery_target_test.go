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
	created   *deliveryTargetModel.DeliveryTarget
	updated   *deliveryTargetModel.DeliveryTarget
	deletedID string
}

func (r *targetTestDeliveryTargetRepo) GetByID(ctx context.Context, id string) (*deliveryTargetModel.DeliveryTarget, error) {
	if r.target != nil && r.target.ID == id {
		copy := *r.target
		return &copy, nil
	}
	return nil, nil
}

func (r *targetTestDeliveryTargetRepo) Create(ctx context.Context, target *deliveryTargetModel.DeliveryTarget) error {
	copy := *target
	r.created = &copy
	r.target = &copy
	return nil
}

func (r *targetTestDeliveryTargetRepo) Update(ctx context.Context, target *deliveryTargetModel.DeliveryTarget) error {
	copy := *target
	r.updated = &copy
	r.target = &copy
	return nil
}

func (r *targetTestDeliveryTargetRepo) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return false, nil
}

func (r *targetTestDeliveryTargetRepo) ExistsBySlugExcludingID(ctx context.Context, slug, id string) (bool, error) {
	return false, nil
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

func TestCreatePersistsControlPlaneMetadata(t *testing.T) {
	targetRepo := &targetTestDeliveryTargetRepo{}
	uc := New(Dependencies{
		DeliveryTargetRepo: targetRepo,
	})

	result, err := uc.Create(context.Background(), &deliveryTargetModel.CreateDeliveryTargetRequest{
		Name:                "target-a",
		ClusterName:         "cluster-a",
		ClusterServer:       "https://cluster-a",
		ControlPlaneName:    "cp-a",
		ControlPlaneType:    "argocd",
		KubeconfigContext:   "ctx-a",
		ArgoCDNamespace:     "argocd-a",
		ArgoCDServer:        "https://argocd-a",
		CredentialReference: "secret/a",
	})

	require.NoError(t, err)
	require.NotNil(t, targetRepo.created)
	assert.Equal(t, "cp-a", targetRepo.created.ControlPlaneName)
	assert.Equal(t, "argocd", targetRepo.created.ControlPlaneType)
	assert.Equal(t, "ctx-a", targetRepo.created.KubeconfigContext)
	assert.Equal(t, "argocd-a", targetRepo.created.ArgoCDNamespace)
	assert.Equal(t, "https://argocd-a", targetRepo.created.ArgoCDServer)
	assert.Equal(t, "secret/a", targetRepo.created.CredentialReference)
	require.NotNil(t, result)
	assert.Equal(t, "cp-a", result.ControlPlaneName)
	assert.Equal(t, "argocd", result.ControlPlaneType)
	assert.Equal(t, "ctx-a", result.KubeconfigContext)
	assert.Equal(t, "argocd-a", result.ArgoCDNamespace)
	assert.Equal(t, "https://argocd-a", result.ArgoCDServer)
	assert.Equal(t, "secret/a", result.CredentialReference)
}

func TestCreateRejectsIncompleteControlPlaneMetadata(t *testing.T) {
	targetRepo := &targetTestDeliveryTargetRepo{}
	uc := New(Dependencies{
		DeliveryTargetRepo: targetRepo,
	})

	result, err := uc.Create(context.Background(), &deliveryTargetModel.CreateDeliveryTargetRequest{
		Name:              "target-a",
		ClusterName:       "cluster-a",
		ControlPlaneName:  "cp-a",
		KubeconfigContext: "ctx-a",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIncompleteControlPlaneMetadata)
	assert.Nil(t, result)
	assert.Nil(t, targetRepo.created)
}

func TestUpdateRejectsIncompleteControlPlaneMetadata(t *testing.T) {
	targetRepo := &targetTestDeliveryTargetRepo{target: &deliveryTargetModel.DeliveryTarget{
		ID:          "target-1",
		Name:        "target-a",
		Slug:        "target-a",
		ClusterName: "cluster-a",
	}}
	uc := New(Dependencies{
		DeliveryTargetRepo: targetRepo,
	})

	controlPlaneName := "cp-b"
	result, err := uc.Update(context.Background(), "target-1", &deliveryTargetModel.UpdateDeliveryTargetRequest{
		ControlPlaneName: &controlPlaneName,
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIncompleteControlPlaneMetadata)
	assert.Nil(t, result)
	assert.Nil(t, targetRepo.updated)
}

func TestUpdatePersistsControlPlaneMetadata(t *testing.T) {
	targetRepo := &targetTestDeliveryTargetRepo{target: &deliveryTargetModel.DeliveryTarget{
		ID:            "target-1",
		Name:          "target-a",
		Slug:          "target-a",
		ClusterName:   "cluster-a",
		ClusterServer: "https://cluster-a",
	}}
	uc := New(Dependencies{
		DeliveryTargetRepo: targetRepo,
	})

	controlPlaneName := "cp-b"
	controlPlaneType := "argocd"
	kubeconfigContext := "ctx-b"
	argoNamespace := "argocd-b"
	argoServer := "https://argocd-b"
	credentialReference := "secret/b"

	result, err := uc.Update(context.Background(), "target-1", &deliveryTargetModel.UpdateDeliveryTargetRequest{
		ControlPlaneName:    &controlPlaneName,
		ControlPlaneType:    &controlPlaneType,
		KubeconfigContext:   &kubeconfigContext,
		ArgoCDNamespace:     &argoNamespace,
		ArgoCDServer:        &argoServer,
		CredentialReference: &credentialReference,
	})

	require.NoError(t, err)
	require.NotNil(t, targetRepo.updated)
	assert.Equal(t, "cp-b", targetRepo.updated.ControlPlaneName)
	assert.Equal(t, "argocd", targetRepo.updated.ControlPlaneType)
	assert.Equal(t, "ctx-b", targetRepo.updated.KubeconfigContext)
	assert.Equal(t, "argocd-b", targetRepo.updated.ArgoCDNamespace)
	assert.Equal(t, "https://argocd-b", targetRepo.updated.ArgoCDServer)
	assert.Equal(t, "secret/b", targetRepo.updated.CredentialReference)
	require.NotNil(t, result)
	assert.Equal(t, "cp-b", result.ControlPlaneName)
	assert.Equal(t, "argocd", result.ControlPlaneType)
	assert.Equal(t, "ctx-b", result.KubeconfigContext)
	assert.Equal(t, "argocd-b", result.ArgoCDNamespace)
	assert.Equal(t, "https://argocd-b", result.ArgoCDServer)
	assert.Equal(t, "secret/b", result.CredentialReference)
}
