package templates

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/metadata"
	"github.com/adverax/crm/internal/platform/security"
)

// --- mocks ---

type mockObjectService struct {
	createFunc func(ctx context.Context, input metadata.CreateObjectInput) (*metadata.ObjectDefinition, error)
}

func (m *mockObjectService) Create(ctx context.Context, input metadata.CreateObjectInput) (*metadata.ObjectDefinition, error) {
	return m.createFunc(ctx, input)
}
func (m *mockObjectService) GetByID(_ context.Context, _ uuid.UUID) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *mockObjectService) List(_ context.Context, _ metadata.ObjectFilter) ([]metadata.ObjectDefinition, int64, error) {
	return nil, 0, nil
}
func (m *mockObjectService) Update(_ context.Context, _ uuid.UUID, _ metadata.UpdateObjectInput) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *mockObjectService) Delete(_ context.Context, _ uuid.UUID) error { return nil }

type mockFieldService struct {
	createFunc func(ctx context.Context, input metadata.CreateFieldInput) (*metadata.FieldDefinition, error)
}

func (m *mockFieldService) Create(ctx context.Context, input metadata.CreateFieldInput) (*metadata.FieldDefinition, error) {
	return m.createFunc(ctx, input)
}
func (m *mockFieldService) GetByID(_ context.Context, _ uuid.UUID) (*metadata.FieldDefinition, error) {
	return nil, nil
}
func (m *mockFieldService) ListByObjectID(_ context.Context, _ uuid.UUID) ([]metadata.FieldDefinition, error) {
	return nil, nil
}
func (m *mockFieldService) Update(_ context.Context, _ uuid.UUID, _ metadata.UpdateFieldInput) (*metadata.FieldDefinition, error) {
	return nil, nil
}
func (m *mockFieldService) Delete(_ context.Context, _ uuid.UUID) error { return nil }

type mockObjectRepo struct {
	countResult int64
	countErr    error
}

func (m *mockObjectRepo) Count(_ context.Context) (int64, error) {
	return m.countResult, m.countErr
}
func (m *mockObjectRepo) Create(_ context.Context, _ pgx.Tx, _ metadata.CreateObjectInput) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *mockObjectRepo) GetByID(_ context.Context, _ uuid.UUID) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *mockObjectRepo) GetByAPIName(_ context.Context, _ string) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *mockObjectRepo) List(_ context.Context, _, _ int32) ([]metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *mockObjectRepo) ListAll(_ context.Context) ([]metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *mockObjectRepo) Update(_ context.Context, _ pgx.Tx, _ uuid.UUID, _ metadata.UpdateObjectInput) (*metadata.ObjectDefinition, error) {
	return nil, nil
}
func (m *mockObjectRepo) Delete(_ context.Context, _ pgx.Tx, _ uuid.UUID) error { return nil }

type mockPermissionService struct {
	setObjectCalls []security.SetObjectPermissionInput
	setFieldCalls  []security.SetFieldPermissionInput
}

func (m *mockPermissionService) SetObjectPermission(_ context.Context, _ uuid.UUID, input security.SetObjectPermissionInput) (*security.ObjectPermission, error) {
	m.setObjectCalls = append(m.setObjectCalls, input)
	return &security.ObjectPermission{ID: uuid.New()}, nil
}
func (m *mockPermissionService) ListObjectPermissions(_ context.Context, _ uuid.UUID) ([]security.ObjectPermission, error) {
	return nil, nil
}
func (m *mockPermissionService) RemoveObjectPermission(_ context.Context, _, _ uuid.UUID) error {
	return nil
}
func (m *mockPermissionService) SetFieldPermission(_ context.Context, _ uuid.UUID, input security.SetFieldPermissionInput) (*security.FieldPermission, error) {
	m.setFieldCalls = append(m.setFieldCalls, input)
	return &security.FieldPermission{ID: uuid.New()}, nil
}
func (m *mockPermissionService) ListFieldPermissions(_ context.Context, _ uuid.UUID) ([]security.FieldPermission, error) {
	return nil, nil
}
func (m *mockPermissionService) RemoveFieldPermission(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

type mockCacheInvalidator struct {
	called bool
}

func (m *mockCacheInvalidator) Invalidate(_ context.Context) error {
	m.called = true
	return nil
}

// --- tests ---

func TestApplier_Apply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		countResult int64
		countErr    error
		objectErr   error
		fieldErr    error
		wantErr     bool
		errCode     apperror.Code
	}{
		{
			name:        "succeeds with empty database",
			countResult: 0,
		},
		{
			name:        "returns conflict when objects exist",
			countResult: 5,
			wantErr:     true,
			errCode:     apperror.CodeConflict,
		},
		{
			name:     "returns error when count fails",
			countErr: errors.New("db error"),
			wantErr:  true,
		},
		{
			name:      "returns error when object creation fails",
			objectErr: errors.New("create failed"),
			wantErr:   true,
		},
		{
			name:     "returns error when field creation fails",
			fieldErr: errors.New("create failed"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			objectIDs := make(map[string]uuid.UUID)

			objectSvc := &mockObjectService{
				createFunc: func(_ context.Context, input metadata.CreateObjectInput) (*metadata.ObjectDefinition, error) {
					if tt.objectErr != nil {
						return nil, tt.objectErr
					}
					id := uuid.New()
					objectIDs[input.APIName] = id
					return &metadata.ObjectDefinition{ID: id, APIName: input.APIName}, nil
				},
			}

			fieldSvc := &mockFieldService{
				createFunc: func(_ context.Context, input metadata.CreateFieldInput) (*metadata.FieldDefinition, error) {
					if tt.fieldErr != nil {
						return nil, tt.fieldErr
					}
					return &metadata.FieldDefinition{ID: uuid.New(), APIName: input.APIName}, nil
				},
			}

			objectRepo := &mockObjectRepo{
				countResult: tt.countResult,
				countErr:    tt.countErr,
			}

			permSvc := &mockPermissionService{}
			cache := &mockCacheInvalidator{}

			applier := NewApplier(objectSvc, fieldSvc, objectRepo, permSvc, cache)

			tmpl := Template{
				ID:    "test",
				Label: "Test",
				Objects: []ObjectTemplate{
					{APIName: "obj_a", Label: "A", PluralLabel: "As", Visibility: metadata.VisibilityPrivate, IsCreateable: true},
					{APIName: "obj_b", Label: "B", PluralLabel: "Bs", Visibility: metadata.VisibilityPrivate, IsCreateable: true},
				},
				Fields: []FieldTemplate{
					{ObjectAPIName: "obj_a", APIName: "name", Label: "Name", FieldType: metadata.FieldTypeText, SortOrder: 1},
					{ObjectAPIName: "obj_b", APIName: "ref", Label: "Ref", FieldType: metadata.FieldTypeReference, ReferencedObjectAPIName: "obj_a", SortOrder: 1},
				},
			}

			err := applier.Apply(context.Background(), tmpl)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errCode != "" {
					var appErr *apperror.AppError
					if errors.As(err, &appErr) {
						if appErr.Code != tt.errCode {
							t.Errorf("expected error code %s, got %s", tt.errCode, appErr.Code)
						}
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify OLS was set for both objects.
			if len(permSvc.setObjectCalls) != 2 {
				t.Errorf("expected 2 OLS calls, got %d", len(permSvc.setObjectCalls))
			}
			for _, call := range permSvc.setObjectCalls {
				if call.Permissions != security.OLSAll {
					t.Errorf("expected OLS permissions %d, got %d", security.OLSAll, call.Permissions)
				}
			}

			// Verify cache was invalidated.
			if !cache.called {
				t.Error("expected cache invalidation")
			}
		})
	}
}

func TestApplier_Apply_SalesCRM(t *testing.T) {
	t.Parallel()

	objectSvc := &mockObjectService{
		createFunc: func(_ context.Context, input metadata.CreateObjectInput) (*metadata.ObjectDefinition, error) {
			return &metadata.ObjectDefinition{ID: uuid.New(), APIName: input.APIName}, nil
		},
	}

	fieldSvc := &mockFieldService{
		createFunc: func(_ context.Context, input metadata.CreateFieldInput) (*metadata.FieldDefinition, error) {
			return &metadata.FieldDefinition{ID: uuid.New(), APIName: input.APIName}, nil
		},
	}

	objectRepo := &mockObjectRepo{countResult: 0}
	permSvc := &mockPermissionService{}
	cache := &mockCacheInvalidator{}

	applier := NewApplier(objectSvc, fieldSvc, objectRepo, permSvc, cache)

	err := applier.Apply(context.Background(), SalesCRM())
	if err != nil {
		t.Fatalf("failed to apply SalesCRM template: %v", err)
	}

	if len(permSvc.setObjectCalls) != 4 {
		t.Errorf("expected 4 OLS calls for SalesCRM, got %d", len(permSvc.setObjectCalls))
	}
}

func TestApplier_Apply_Recruiting(t *testing.T) {
	t.Parallel()

	objectSvc := &mockObjectService{
		createFunc: func(_ context.Context, input metadata.CreateObjectInput) (*metadata.ObjectDefinition, error) {
			return &metadata.ObjectDefinition{ID: uuid.New(), APIName: input.APIName}, nil
		},
	}

	fieldSvc := &mockFieldService{
		createFunc: func(_ context.Context, input metadata.CreateFieldInput) (*metadata.FieldDefinition, error) {
			return &metadata.FieldDefinition{ID: uuid.New(), APIName: input.APIName}, nil
		},
	}

	objectRepo := &mockObjectRepo{countResult: 0}
	permSvc := &mockPermissionService{}
	cache := &mockCacheInvalidator{}

	applier := NewApplier(objectSvc, fieldSvc, objectRepo, permSvc, cache)

	err := applier.Apply(context.Background(), Recruiting())
	if err != nil {
		t.Fatalf("failed to apply Recruiting template: %v", err)
	}

	if len(permSvc.setObjectCalls) != 4 {
		t.Errorf("expected 4 OLS calls for Recruiting, got %d", len(permSvc.setObjectCalls))
	}
}
