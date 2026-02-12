package metadata

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
	metaModel "github.com/proxima-research/proxima.crm.platform/internal/metadata/domain"
)

// mockObjectRepository is a configurable mock for testing.
type mockObjectRepository struct {
	objects    map[string]*metaModel.ObjectDescription
	objectList []*metaModel.ObjectDefinition
	callCount  atomic.Int32
	listCount  atomic.Int32
	err        error
	delay      time.Duration
}

func (m *mockObjectRepository) GetObjectDescription(_ context.Context, name metaModel.ObjectApiName, _ metaModel.ObjectDescriptionParams) (*metaModel.ObjectDescription, error) {
	m.callCount.Add(1)
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.objects[string(name)], nil
}

func (m *mockObjectRepository) GetListObjects(_ context.Context) ([]*metaModel.ObjectDefinition, error) {
	m.listCount.Add(1)
	if m.err != nil {
		return nil, m.err
	}
	return m.objectList, nil
}

func TestMetadataAdapter_GetObject_CachesResult(t *testing.T) {
	repo := &mockObjectRepository{
		objects: map[string]*metaModel.ObjectDescription{
			"Account": {
				ObjectDefinition: metaModel.ObjectDefinition{
					ApiName: "Account",
					Control: metaModel.ObjectManagedKernel,
				},
				Fields: []*metaModel.FieldDefinition{},
			},
		},
	}

	adapter := NewMetadataAdapter(repo)

	ctx := context.Background()

	// First call should hit the repository
	obj1, err := adapter.GetObject(ctx, "Account")
	if err != nil {
		t.Fatalf("first GetObject failed: %v", err)
	}
	if obj1 == nil {
		t.Fatal("expected object, got nil")
	}
	if repo.callCount.Load() != 1 {
		t.Errorf("expected 1 repo call, got %d", repo.callCount.Load())
	}

	// Second call should use cache
	obj2, err := adapter.GetObject(ctx, "Account")
	if err != nil {
		t.Fatalf("second GetObject failed: %v", err)
	}
	if obj2 == nil {
		t.Fatal("expected object, got nil")
	}
	if repo.callCount.Load() != 1 {
		t.Errorf("expected still 1 repo call, got %d", repo.callCount.Load())
	}

	// Verify stats
	stats := adapter.Stats()
	if stats.ObjectCacheStats.TotalEntries != 1 {
		t.Errorf("cache size = %d, want 1", stats.ObjectCacheStats.TotalEntries)
	}
}

func TestMetadataAdapter_GetObject_CaseInsensitive(t *testing.T) {
	repo := &mockObjectRepository{
		objects: map[string]*metaModel.ObjectDescription{
			"Account": {
				ObjectDefinition: metaModel.ObjectDefinition{
					ApiName: "Account",
					Control: metaModel.ObjectManagedKernel,
				},
				Fields: []*metaModel.FieldDefinition{},
			},
		},
	}

	adapter := NewMetadataAdapter(repo)
	ctx := context.Background()

	// Call with different cases
	_, _ = adapter.GetObject(ctx, "Account")
	_, _ = adapter.GetObject(ctx, "account")
	_, _ = adapter.GetObject(ctx, "ACCOUNT")

	// Should only call repo once
	if repo.callCount.Load() != 1 {
		t.Errorf("expected 1 repo call, got %d", repo.callCount.Load())
	}
}

func TestMetadataAdapter_GetObject_CachesNegativeResult(t *testing.T) {
	repo := &mockObjectRepository{
		objects: map[string]*metaModel.ObjectDescription{},
	}

	adapter := NewMetadataAdapter(repo)
	ctx := context.Background()

	// First call for non-existent object
	obj1, err := adapter.GetObject(ctx, "NonExistent")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if obj1 != nil {
		t.Fatal("expected nil for non-existent object")
	}

	// Second call should use cached negative result
	obj2, err := adapter.GetObject(ctx, "NonExistent")
	if err != nil {
		t.Fatalf("second GetObject failed: %v", err)
	}
	if obj2 != nil {
		t.Fatal("expected nil for non-existent object")
	}

	if repo.callCount.Load() != 1 {
		t.Errorf("expected 1 repo call, got %d", repo.callCount.Load())
	}
}

func TestMetadataAdapter_GetObject_TTLExpiration(t *testing.T) {
	repo := &mockObjectRepository{
		objects: map[string]*metaModel.ObjectDescription{
			"Account": {
				ObjectDefinition: metaModel.ObjectDefinition{
					ApiName: "Account",
					Control: metaModel.ObjectManagedKernel,
				},
				Fields: []*metaModel.FieldDefinition{},
			},
		},
	}

	// Short TTL for testing
	adapter := NewMetadataAdapterWithTTL(repo, 50*time.Millisecond)
	ctx := context.Background()

	// First call
	_, _ = adapter.GetObject(ctx, "Account")
	if repo.callCount.Load() != 1 {
		t.Errorf("expected 1 repo call, got %d", repo.callCount.Load())
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Should call repo again
	_, _ = adapter.GetObject(ctx, "Account")
	if repo.callCount.Load() != 2 {
		t.Errorf("expected 2 repo calls after expiration, got %d", repo.callCount.Load())
	}
}

func TestMetadataAdapter_GetObject_ReturnsError(t *testing.T) {
	expectedErr := errors.New("database error")
	repo := &mockObjectRepository{
		err: expectedErr,
	}

	adapter := NewMetadataAdapter(repo)
	ctx := context.Background()

	_, err := adapter.GetObject(ctx, "Account")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestMetadataAdapter_GetObject_ConcurrentAccess(t *testing.T) {
	repo := &mockObjectRepository{
		objects: map[string]*metaModel.ObjectDescription{
			"Account": {
				ObjectDefinition: metaModel.ObjectDefinition{
					ApiName: "Account",
					Control: metaModel.ObjectManagedKernel,
				},
				Fields: []*metaModel.FieldDefinition{},
			},
		},
		delay: 10 * time.Millisecond,
	}

	adapter := NewMetadataAdapter(repo)
	ctx := context.Background()

	const goroutines = 10
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()
			_, err := adapter.GetObject(ctx, "Account")
			if err != nil {
				t.Errorf("GetObject failed: %v", err)
			}
		}()
	}

	wg.Wait()

	// With double-check locking, only 1 call should be made
	// (though in practice might be slightly more due to timing)
	if repo.callCount.Load() > 2 {
		t.Errorf("expected at most 2 repo calls with concurrent access, got %d", repo.callCount.Load())
	}
}

func TestMetadataAdapter_ListObjects_CachesResult(t *testing.T) {
	repo := &mockObjectRepository{
		objectList: []*metaModel.ObjectDefinition{
			{ApiName: "Account"},
			{ApiName: "Contact"},
		},
	}

	adapter := NewMetadataAdapter(repo)
	ctx := context.Background()

	// First call
	list1, err := adapter.ListObjects(ctx)
	if err != nil {
		t.Fatalf("first ListObjects failed: %v", err)
	}
	if len(list1) != 2 {
		t.Errorf("expected 2 objects, got %d", len(list1))
	}

	// Second call should use cache
	list2, err := adapter.ListObjects(ctx)
	if err != nil {
		t.Fatalf("second ListObjects failed: %v", err)
	}
	if len(list2) != 2 {
		t.Errorf("expected 2 objects, got %d", len(list2))
	}

	if repo.listCount.Load() != 1 {
		t.Errorf("expected 1 repo call, got %d", repo.listCount.Load())
	}

	// Verify returned slice is a copy
	list1[0] = "Modified"
	list3, _ := adapter.ListObjects(ctx)
	if list3[0] == "Modified" {
		t.Error("cache returned same slice reference instead of copy")
	}
}

func TestMetadataAdapter_InvalidateObject(t *testing.T) {
	repo := &mockObjectRepository{
		objects: map[string]*metaModel.ObjectDescription{
			"Account": {
				ObjectDefinition: metaModel.ObjectDefinition{
					ApiName: "Account",
					Control: metaModel.ObjectManagedKernel,
				},
				Fields: []*metaModel.FieldDefinition{},
			},
		},
	}

	adapter := NewMetadataAdapter(repo)
	ctx := context.Background()

	// Populate cache
	_, _ = adapter.GetObject(ctx, "Account")
	if repo.callCount.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", repo.callCount.Load())
	}

	// Invalidate
	adapter.InvalidateObject(ctx, "Account")

	// Should call repo again
	_, _ = adapter.GetObject(ctx, "Account")
	if repo.callCount.Load() != 2 {
		t.Errorf("expected 2 calls after invalidation, got %d", repo.callCount.Load())
	}
}

func TestMetadataAdapter_InvalidateAll(t *testing.T) {
	repo := &mockObjectRepository{
		objects: map[string]*metaModel.ObjectDescription{
			"Account": {
				ObjectDefinition: metaModel.ObjectDefinition{
					ApiName: "Account",
					Control: metaModel.ObjectManagedKernel,
				},
			},
			"Contact": {
				ObjectDefinition: metaModel.ObjectDefinition{
					ApiName: "Contact",
					Control: metaModel.ObjectManagedKernel,
				},
			},
		},
		objectList: []*metaModel.ObjectDefinition{
			{ApiName: "Account"},
		},
	}

	adapter := NewMetadataAdapter(repo)
	ctx := context.Background()

	// Populate caches
	_, _ = adapter.GetObject(ctx, "Account")
	_, _ = adapter.GetObject(ctx, "Contact")
	_, _ = adapter.ListObjects(ctx)

	stats := adapter.Stats()
	if stats.ObjectCacheStats.TotalEntries != 2 {
		t.Errorf("cache size = %d, want 2", stats.ObjectCacheStats.TotalEntries)
	}
	if stats.ObjectListCacheStats.TotalEntries != 1 {
		t.Errorf("object list size = %d, want 1", stats.ObjectListCacheStats.TotalEntries)
	}

	// Invalidate all
	adapter.InvalidateAll(ctx)

	stats = adapter.Stats()
	if stats.ObjectCacheStats.TotalEntries != 0 {
		t.Errorf("cache size after invalidate = %d, want 0", stats.ObjectCacheStats.TotalEntries)
	}
	if stats.ObjectListCacheStats.TotalEntries != 0 {
		t.Errorf("object list size after invalidate = %d, want 0", stats.ObjectListCacheStats.TotalEntries)
	}
}

func TestMetadataAdapter_ConvertObjectMeta(t *testing.T) {
	repo := &mockObjectRepository{
		objects: map[string]*metaModel.ObjectDescription{
			"Account": {
				ObjectDefinition: metaModel.ObjectDefinition{
					ApiName: "Account",
					Control: metaModel.ObjectManagedKernel,
				},
				Fields: []*metaModel.FieldDefinition{
					{
						ApiName:  "Name",
						Type:     metaModel.FieldTypeText,
						Required: true,
					},
					{
						ApiName: "Revenue",
						Type:    metaModel.FieldTypeNumber,
						Subtype: metaModel.FieldSubtypeInteger,
					},
					{
						ApiName: "IsActive",
						Type:    metaModel.FieldTypeBoolean,
					},
					{
						ApiName: "ContactId",
						Type:    metaModel.FieldTypeReference,
						Config: map[string]any{
							"referenceObject": "Contact",
						},
					},
				},
			},
		},
	}

	adapter := NewMetadataAdapter(repo)
	ctx := context.Background()

	obj, err := adapter.GetObject(ctx, "Account")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}

	// Check object meta
	if obj.Name != "Account" {
		t.Errorf("Name = %q, want Account", obj.Name)
	}
	if obj.SchemeName != "dataview" {
		t.Errorf("SchemeName = %q, want dataview", obj.SchemeName)
	}
	if obj.TableName != "account" {
		t.Errorf("TableName = %q, want account", obj.TableName)
	}

	// Check fields count (4 custom + 6 system fields)
	expectedFieldCount := 4 + 6
	if len(obj.Fields) != expectedFieldCount {
		t.Errorf("Fields count = %d, want %d", len(obj.Fields), expectedFieldCount)
	}

	// Check system fields
	if _, ok := obj.Fields["Id"]; !ok {
		t.Error("missing Id system field")
	}
	if _, ok := obj.Fields["CreatedAt"]; !ok {
		t.Error("missing CreatedAt system field")
	}

	// Check custom fields
	nameField := obj.Fields["Name"]
	if nameField == nil {
		t.Fatal("missing Name field")
	}
	if nameField.Type != engine.FieldTypeString {
		t.Errorf("Name type = %v, want String", nameField.Type)
	}
	if nameField.Nullable {
		t.Error("Name should not be nullable (required)")
	}

	revenueField := obj.Fields["Revenue"]
	if revenueField == nil {
		t.Fatal("missing Revenue field")
	}
	if revenueField.Type != engine.FieldTypeInteger {
		t.Errorf("Revenue type = %v, want Integer", revenueField.Type)
	}

	// Check lookup
	if len(obj.Lookups) != 1 {
		t.Errorf("Lookups count = %d, want 1", len(obj.Lookups))
	}
	contactLookup := obj.Lookups["Contact"]
	if contactLookup == nil {
		t.Fatal("missing Contact lookup")
	}
	if contactLookup.TargetObject != "Contact" {
		t.Errorf("lookup target = %q, want Contact", contactLookup.TargetObject)
	}
}

func TestDefaultCacheTTL(t *testing.T) {
	if DefaultCacheTTL != 5*time.Minute {
		t.Errorf("DefaultCacheTTL = %v, want 5m", DefaultCacheTTL)
	}
}
