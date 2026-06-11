package gitops

import (
	"context"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/pkg/argocd"
)

// Repository defines the interface for GitOps/ArgoCD operations
type Repository interface {
	CreateApplication(ctx context.Context, spec argocd.ApplicationSpec) error
	GetApplicationStatus(ctx context.Context, name string) (*environment.ArgoStatus, error)
	SyncApplication(ctx context.Context, name string) error
	DeleteApplication(ctx context.Context, name string) error
}

type repository struct {
	client   *argocd.Client
	argocdNS string
}

type Dependencies struct {
	ArgoCDClient    *argocd.Client
	ArgoCDNamespace string
}

type ClientFactory func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (*argocd.Client, error)

type ProviderDependencies struct {
	DefaultRepository Repository
	Defaults          deliveryTargetModel.TargetControlPlaneDefaults
	ClientFactory     ClientFactory
}

type Provider struct {
	defaultRepository Repository
	defaults          deliveryTargetModel.TargetControlPlaneDefaults
	clientFactory     ClientFactory
}

func New(deps Dependencies) Repository {
	argocdNS := deps.ArgoCDNamespace
	if argocdNS == "" {
		argocdNS = "argocd"
	}

	return &repository{
		client:   deps.ArgoCDClient,
		argocdNS: argocdNS,
	}
}

func NewProvider(deps ProviderDependencies) *Provider {
	return &Provider{
		defaultRepository: deps.DefaultRepository,
		defaults:          deps.Defaults,
		clientFactory:     deps.ClientFactory,
	}
}
