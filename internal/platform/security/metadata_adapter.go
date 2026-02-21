package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/platform/metadata"
)

// CacheBackedMetadataLister implements MetadataFieldLister and MetadataRLSAdapter
// by delegating to metadata.MetadataReader instead of raw SQL (ADR-0030).
type CacheBackedMetadataLister struct {
	reader metadata.MetadataReader
}

// NewCacheBackedMetadataLister creates a new CacheBackedMetadataLister.
func NewCacheBackedMetadataLister(reader metadata.MetadataReader) *CacheBackedMetadataLister {
	return &CacheBackedMetadataLister{reader: reader}
}

func (l *CacheBackedMetadataLister) ListFieldsByObjectID(_ context.Context, objectID uuid.UUID) ([]FieldInfo, error) {
	fields := l.reader.GetFieldsByObjectID(objectID)
	result := make([]FieldInfo, len(fields))
	for i, f := range fields {
		result[i] = FieldInfo{ID: f.ID, APIName: f.APIName}
	}
	return result, nil
}

func (l *CacheBackedMetadataLister) ListAllObjectIDs(_ context.Context) ([]uuid.UUID, error) {
	names := l.reader.ListObjectAPINames()
	ids := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		obj, ok := l.reader.GetObjectByAPIName(name)
		if ok {
			ids = append(ids, obj.ID)
		}
	}
	return ids, nil
}

func (l *CacheBackedMetadataLister) GetObjectVisibility(_ context.Context, objectID uuid.UUID) (string, error) {
	obj, ok := l.reader.GetObjectByID(objectID)
	if !ok {
		return "", fmt.Errorf("cacheBackedMetadataLister.GetObjectVisibility: object %s not found", objectID)
	}
	return string(obj.Visibility), nil
}

func (l *CacheBackedMetadataLister) GetObjectTableName(_ context.Context, objectID uuid.UUID) (string, error) {
	obj, ok := l.reader.GetObjectByID(objectID)
	if !ok {
		return "", fmt.Errorf("cacheBackedMetadataLister.GetObjectTableName: object %s not found", objectID)
	}
	return obj.TableName, nil
}

func (l *CacheBackedMetadataLister) ListCompositionFields(_ context.Context) ([]CompositionFieldInfo, error) {
	names := l.reader.ListObjectAPINames()
	var result []CompositionFieldInfo
	for _, name := range names {
		obj, ok := l.reader.GetObjectByAPIName(name)
		if !ok {
			continue
		}
		rels := l.reader.GetForwardRelationships(obj.ID)
		for _, rel := range rels {
			if rel.ReferenceSubtype == metadata.SubtypeComposition {
				result = append(result, CompositionFieldInfo{
					ChildObjectID:  rel.ChildObjectID,
					ParentObjectID: rel.ParentObjectID,
				})
			}
		}
	}
	return result, nil
}
