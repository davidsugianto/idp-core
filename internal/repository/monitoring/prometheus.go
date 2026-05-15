package monitoring

import (
	"context"
	"time"

	prometheusPkg "github.com/davidsugianto/idp-core/internal/pkg/prometheus"
)

func (r *repository) Query(ctx context.Context, query string) ([]prometheusPkg.QueryResult, error) {
	return r.promClient.Query(ctx, query)
}

func (r *repository) QueryRange(ctx context.Context, query string, start, end time.Time, step time.Duration) ([]prometheusPkg.RangeQueryResult, error) {
	return r.promClient.QueryRange(ctx, query, start, end, step)
}
