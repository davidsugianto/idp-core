package gitops

import (
	"context"
	"fmt"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
)

func (p *Provider) ForTarget(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (Repository, error) {
	resolved := target.Resolve(p.defaults)
	if resolved.UsesDefaultControlPlane {
		if p.defaultRepository == nil {
			return nil, fmt.Errorf("gitops repository not configured")
		}
		return p.defaultRepository, nil
	}

	if p.clientFactory == nil {
		return nil, fmt.Errorf("gitops client factory not configured")
	}

	client, err := p.clientFactory(ctx, resolved)
	if err != nil {
		return nil, err
	}

	return New(Dependencies{
		ArgoCDClient:    client,
		ArgoCDNamespace: resolved.ArgoCDNamespace,
	}), nil
}
