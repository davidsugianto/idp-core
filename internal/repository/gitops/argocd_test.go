package gitops

import (
	"context"
	"testing"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/pkg/argocd"
	"github.com/stretchr/testify/assert"
)

type fakeGitopsRepo struct {
	name string
}

func (f *fakeGitopsRepo) CreateApplication(ctx context.Context, spec argocd.ApplicationSpec) error {
	return nil
}

func (f *fakeGitopsRepo) GetApplicationStatus(ctx context.Context, name string) (*environment.ArgoStatus, error) {
	return &environment.ArgoStatus{SyncStatus: f.name}, nil
}

func (f *fakeGitopsRepo) SyncApplication(ctx context.Context, name string) error {
	return nil
}

func (f *fakeGitopsRepo) DeleteApplication(ctx context.Context, name string) error {
	return nil
}

func TestProviderForTargetUsesDefaultRepositoryForDefaultControlPlane(t *testing.T) {
	defaultRepo := &fakeGitopsRepo{name: "default"}
	provider := NewProvider(ProviderDependencies{
		DefaultRepository: defaultRepo,
		Defaults: deliveryTargetModel.TargetControlPlaneDefaults{
			ControlPlaneName:  "default-cp",
			ArgoCDNamespace:   "argocd",
			KubeconfigContext: "default-ctx",
		},
	})

	repo, err := provider.ForTarget(context.Background(), &deliveryTargetModel.TargetControlPlane{DeliveryTargetID: "target-a"})
	assert.NoError(t, err)
	assert.Same(t, defaultRepo, repo)
}

func TestProviderForTargetBuildsRepositoryForExplicitTargetControlPlane(t *testing.T) {
	defaultRepo := &fakeGitopsRepo{name: "default"}
	factoryCalls := 0
	provider := NewProvider(ProviderDependencies{
		DefaultRepository: defaultRepo,
		Defaults: deliveryTargetModel.TargetControlPlaneDefaults{
			ControlPlaneName:  "default-cp",
			ArgoCDNamespace:   "argocd",
			KubeconfigContext: "default-ctx",
		},
		ClientFactory: func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (*argocd.Client, error) {
			factoryCalls++
			assert.Equal(t, "target-cp", target.ControlPlaneName)
			assert.Equal(t, "argocd-target", target.ArgoCDNamespace)
			assert.Equal(t, "ctx-target", target.KubeconfigContext)
			return &argocd.Client{}, nil
		},
	})

	repo, err := provider.ForTarget(context.Background(), &deliveryTargetModel.TargetControlPlane{
		DeliveryTargetID:  "target-a",
		ControlPlaneName:  "target-cp",
		ArgoCDNamespace:   "argocd-target",
		KubeconfigContext: "ctx-target",
	})
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.NotSame(t, defaultRepo, repo)
	assert.Equal(t, 1, factoryCalls)
}

func TestProviderForTargetReturnsErrorWhenExplicitTargetHasNoFactory(t *testing.T) {
	provider := NewProvider(ProviderDependencies{
		Defaults: deliveryTargetModel.TargetControlPlaneDefaults{
			ControlPlaneName: "default-cp",
		},
	})

	repo, err := provider.ForTarget(context.Background(), &deliveryTargetModel.TargetControlPlane{
		DeliveryTargetID: "target-a",
		ControlPlaneName: "target-cp",
	})
	assert.Nil(t, repo)
	assert.EqualError(t, err, "gitops client factory not configured")
}

func TestProviderForTargetBuildsStatusRepositoryForExplicitTargetControlPlane(t *testing.T) {
	provider := NewProvider(ProviderDependencies{
		Defaults: deliveryTargetModel.TargetControlPlaneDefaults{
			ControlPlaneName:  "default-cp",
			ArgoCDNamespace:   "argocd",
			KubeconfigContext: "default-ctx",
		},
		ClientFactory: func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (*argocd.Client, error) {
			assert.Equal(t, "target-cp", target.ControlPlaneName)
			assert.Equal(t, "argocd-target", target.ArgoCDNamespace)
			return &argocd.Client{}, nil
		},
	})

	repo, err := provider.ForTarget(context.Background(), &deliveryTargetModel.TargetControlPlane{
		DeliveryTargetID:  "target-a",
		ControlPlaneName:  "target-cp",
		ArgoCDNamespace:   "argocd-target",
		KubeconfigContext: "ctx-target",
	})
	assert.NoError(t, err)

	statusRepo, ok := repo.(*repository)
	assert.True(t, ok)
	assert.Equal(t, "argocd-target", statusRepo.argocdNS)
}
