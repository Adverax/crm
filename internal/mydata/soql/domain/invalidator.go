// Package soqlModel defines domain interfaces and models for SOQL cache invalidation.
package soqlModel

import "context"

// CacheInvalidator defines the interface for invalidating SOQL caches.
// It aggregates operations on both metadata cache and query cache.
type CacheInvalidator interface {
	// InvalidateObject invalidates the metadata cache for a specific object.
	// This should be called when an object or its fields are modified.
	InvalidateObject(ctx context.Context, objectApiName string)

	// InvalidateAll invalidates all metadata cache entries.
	// This should be called when a bulk metadata change occurs.
	InvalidateAll(ctx context.Context)

	// ClearQueryCache clears all cached compiled queries.
	// This should be called when metadata changes, as cached queries
	// may reference the changed metadata.
	ClearQueryCache(ctx context.Context)

	// InvalidateQueriesByObject removes cached queries that depend on the given object.
	// This is more efficient than ClearQueryCache as it only removes affected queries.
	InvalidateQueriesByObject(ctx context.Context, objectApiName string)
}
