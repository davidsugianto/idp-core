package gitops

import (
	"context"

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
