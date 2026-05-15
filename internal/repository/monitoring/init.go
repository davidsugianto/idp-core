package monitoring

import (
	"context"
	"time"

	prometheusPkg "github.com/davidsugianto/idp-core/internal/pkg/prometheus"
)

type Repository interface {
	Query(ctx context.Context, query string) ([]prometheusPkg.QueryResult, error)
	QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]prometheusPkg.RangeQueryResult, error)
}

type repository struct {
	promClient *prometheusPkg.Client
}

type Dependencies struct {
	PromClient *prometheusPkg.Client
}

func New(deps Dependencies) Repository {
	return &repository{
		promClient: deps.PromClient,
	}
}
