// Package cache provides cache invalidation infrastructure for SOQL.
package cache

import (
	"context"

	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
	soqlModel "github.com/proxima-research/proxima.crm.platform/internal/data/soql/domain"
	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/infrastructure/metadata"
)

// AggregatedInvalidator implements soqlModel.CacheInvalidator by coordinating
// invalidation across both metadata cache and query cache.
type AggregatedInvalidator struct {
	metadataAdapter *metadata.MetadataAdapter
	engine          *engine.Engine
}

// NewAggregatedInvalidator creates a new AggregatedInvalidator.
func NewAggregatedInvalidator(
	metadataAdapter *metadata.MetadataAdapter,
	engine *engine.Engine,
) *AggregatedInvalidator {
	return &AggregatedInvalidator{
		metadataAdapter: metadataAdapter,
		engine:          engine,
	}
}

// InvalidateObject invalidates the metadata cache for a specific object.
func (a *AggregatedInvalidator) InvalidateObject(ctx context.Context, objectApiName string) {
	if a.metadataAdapter != nil {
		a.metadataAdapter.InvalidateObject(ctx, objectApiName)
	}
}

// InvalidateAll invalidates all metadata cache entries.
func (a *AggregatedInvalidator) InvalidateAll(ctx context.Context) {
	if a.metadataAdapter != nil {
		a.metadataAdapter.InvalidateAll(ctx)
	}
}

// ClearQueryCache clears all cached compiled queries.
func (a *AggregatedInvalidator) ClearQueryCache(ctx context.Context) {
	if a.engine != nil {
		a.engine.ClearQueryCache(ctx)
	}
}

// InvalidateQueriesByObject removes cached queries that depend on the given object.
func (a *AggregatedInvalidator) InvalidateQueriesByObject(ctx context.Context, objectApiName string) {
	if a.engine != nil {
		a.engine.InvalidateQueriesByObject(ctx, objectApiName)
	}
}

// Ensure AggregatedInvalidator implements soqlModel.CacheInvalidator.
var _ soqlModel.CacheInvalidator = (*AggregatedInvalidator)(nil)
