package metadata

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockObjectRepo is a test mock for ObjectRepository.
type mockObjectRepo struct {
	objects   map[uuid.UUID]*ObjectDefinition
	byName    map[string]*ObjectDefinition
	createErr error
	updateErr error
	deleteErr error
}

func newMockObjectRepo() *mockObjectRepo {
	return &mockObjectRepo{
		objects: make(map[uuid.UUID]*ObjectDefinition),
		byName:  make(map[string]*ObjectDefinition),
	}
}

func (m *mockObjectRepo) Create(_ context.Context, _ pgx.Tx, input CreateObjectInput) (*ObjectDefinition, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	obj := &ObjectDefinition{
		ID:                    uuid.New(),
		APIName:               input.APIName,
		Label:                 input.Label,
		PluralLabel:           input.PluralLabel,
		Description:           input.Description,
		TableName:             GenerateTableName(input.APIName),
		ObjectType:            input.ObjectType,
		IsVisibleInSetup:      input.IsVisibleInSetup,
		IsCustomFieldsAllowed: input.IsCustomFieldsAllowed,
		IsDeleteableObject:    input.IsDeleteableObject,
		IsCreateable:          input.IsCreateable,
		IsUpdateable:          input.IsUpdateable,
		IsDeleteable:          input.IsDeleteable,
		IsQueryable:           input.IsQueryable,
		IsSearchable:          input.IsSearchable,
	}
	m.objects[obj.ID] = obj
	m.byName[obj.APIName] = obj
	return obj, nil
}

func (m *mockObjectRepo) GetByID(_ context.Context, id uuid.UUID) (*ObjectDefinition, error) {
	return m.objects[id], nil
}

func (m *mockObjectRepo) GetByAPIName(_ context.Context, apiName string) (*ObjectDefinition, error) {
	return m.byName[apiName], nil
}

func (m *mockObjectRepo) List(_ context.Context, limit, offset int32) ([]ObjectDefinition, error) {
	result := make([]ObjectDefinition, 0)
	i := int32(0)
	for _, obj := range m.objects {
		if i >= offset && int32(len(result)) < limit {
			result = append(result, *obj)
		}
		i++
	}
	return result, nil
}

func (m *mockObjectRepo) ListAll(_ context.Context) ([]ObjectDefinition, error) {
	result := make([]ObjectDefinition, 0, len(m.objects))
	for _, obj := range m.objects {
		result = append(result, *obj)
	}
	return result, nil
}

func (m *mockObjectRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input UpdateObjectInput) (*ObjectDefinition, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	obj, ok := m.objects[id]
	if !ok {
		return nil, nil
	}
	obj.Label = input.Label
	obj.PluralLabel = input.PluralLabel
	obj.Description = input.Description
	return obj, nil
}

func (m *mockObjectRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	obj := m.objects[id]
	if obj != nil {
		delete(m.byName, obj.APIName)
	}
	delete(m.objects, id)
	return nil
}

func (m *mockObjectRepo) Count(_ context.Context) (int64, error) {
	return int64(len(m.objects)), nil
}

func (m *mockObjectRepo) addObject(obj *ObjectDefinition) {
	m.objects[obj.ID] = obj
	m.byName[obj.APIName] = obj
}

// mockFieldRepo is a test mock for FieldRepository.
type mockFieldRepo struct {
	fields    map[uuid.UUID]*FieldDefinition
	byObjName map[string]*FieldDefinition
	createErr error
}

func newMockFieldRepo() *mockFieldRepo {
	return &mockFieldRepo{
		fields:    make(map[uuid.UUID]*FieldDefinition),
		byObjName: make(map[string]*FieldDefinition),
	}
}

func (m *mockFieldRepo) Create(_ context.Context, _ pgx.Tx, input CreateFieldInput) (*FieldDefinition, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	configBytes, _ := json.Marshal(input.Config)
	var config FieldConfig
	_ = json.Unmarshal(configBytes, &config)

	f := &FieldDefinition{
		ID:                 uuid.New(),
		ObjectID:           input.ObjectID,
		APIName:            input.APIName,
		Label:              input.Label,
		Description:        input.Description,
		HelpText:           input.HelpText,
		FieldType:          input.FieldType,
		FieldSubtype:       input.FieldSubtype,
		ReferencedObjectID: input.ReferencedObjectID,
		IsRequired:         input.IsRequired,
		IsUnique:           input.IsUnique,
		Config:             config,
		IsCustom:           input.IsCustom,
		SortOrder:          input.SortOrder,
	}
	m.fields[f.ID] = f
	m.byObjName[fmt.Sprintf("%s/%s", input.ObjectID, input.APIName)] = f
	return f, nil
}

func (m *mockFieldRepo) GetByID(_ context.Context, id uuid.UUID) (*FieldDefinition, error) {
	return m.fields[id], nil
}

func (m *mockFieldRepo) GetByObjectAndName(_ context.Context, objectID uuid.UUID, apiName string) (*FieldDefinition, error) {
	key := fmt.Sprintf("%s/%s", objectID, apiName)
	return m.byObjName[key], nil
}

func (m *mockFieldRepo) ListByObjectID(_ context.Context, objectID uuid.UUID) ([]FieldDefinition, error) {
	result := make([]FieldDefinition, 0)
	for _, f := range m.fields {
		if f.ObjectID == objectID {
			result = append(result, *f)
		}
	}
	return result, nil
}

func (m *mockFieldRepo) ListAll(_ context.Context) ([]FieldDefinition, error) {
	result := make([]FieldDefinition, 0, len(m.fields))
	for _, f := range m.fields {
		result = append(result, *f)
	}
	return result, nil
}

func (m *mockFieldRepo) ListReferenceFields(_ context.Context) ([]FieldDefinition, error) {
	result := make([]FieldDefinition, 0)
	for _, f := range m.fields {
		if f.FieldType == FieldTypeReference {
			result = append(result, *f)
		}
	}
	return result, nil
}

func (m *mockFieldRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input UpdateFieldInput) (*FieldDefinition, error) {
	f := m.fields[id]
	if f == nil {
		return nil, nil
	}
	f.Label = input.Label
	f.Description = input.Description
	f.HelpText = input.HelpText
	f.IsRequired = input.IsRequired
	f.IsUnique = input.IsUnique
	f.Config = input.Config
	f.SortOrder = input.SortOrder
	return f, nil
}

func (m *mockFieldRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	f := m.fields[id]
	if f != nil {
		key := fmt.Sprintf("%s/%s", f.ObjectID, f.APIName)
		delete(m.byObjName, key)
	}
	delete(m.fields, id)
	return nil
}

// mockPolymorphicRepo is a test mock for PolymorphicTargetRepository.
type mockPolymorphicRepo struct {
	targets map[uuid.UUID][]PolymorphicTarget
}

func newMockPolymorphicRepo() *mockPolymorphicRepo {
	return &mockPolymorphicRepo{
		targets: make(map[uuid.UUID][]PolymorphicTarget),
	}
}

func (m *mockPolymorphicRepo) Create(_ context.Context, _ pgx.Tx, fieldID, objectID uuid.UUID) (*PolymorphicTarget, error) {
	pt := &PolymorphicTarget{
		ID:       uuid.New(),
		FieldID:  fieldID,
		ObjectID: objectID,
	}
	m.targets[fieldID] = append(m.targets[fieldID], *pt)
	return pt, nil
}

func (m *mockPolymorphicRepo) ListByFieldID(_ context.Context, fieldID uuid.UUID) ([]PolymorphicTarget, error) {
	return m.targets[fieldID], nil
}

func (m *mockPolymorphicRepo) ListAll(_ context.Context) ([]PolymorphicTarget, error) {
	result := make([]PolymorphicTarget, 0)
	for _, pts := range m.targets {
		result = append(result, pts...)
	}
	return result, nil
}

func (m *mockPolymorphicRepo) DeleteByFieldID(_ context.Context, _ pgx.Tx, fieldID uuid.UUID) error {
	delete(m.targets, fieldID)
	return nil
}

// mockDDLExec is a test mock for DDLExecutor.
type mockDDLExec struct {
	executed [][]string
	execErr  error
}

func newMockDDLExec() *mockDDLExec {
	return &mockDDLExec{}
}

func (m *mockDDLExec) ExecInTx(_ context.Context, _ pgx.Tx, statements []string) error {
	if m.execErr != nil {
		return m.execErr
	}
	m.executed = append(m.executed, statements)
	return nil
}

// mockCacheInvalidator is a test mock for CacheInvalidator.
type mockCacheInvalidator struct {
	invalidated int
	err         error
}

func newMockCache() *mockCacheInvalidator {
	return &mockCacheInvalidator{}
}

func (m *mockCacheInvalidator) Invalidate(_ context.Context) error {
	m.invalidated++
	return m.err
}

// mockTxBeginner is a test mock for TxBeginner that uses a no-op tx.
type mockTxBeginner struct{}

func (m *mockTxBeginner) Begin(_ context.Context) (pgx.Tx, error) {
	return &mockTx{}, nil
}

// mockTx is a minimal pgx.Tx mock that does nothing.
type mockTx struct{}

func (m *mockTx) Begin(_ context.Context) (pgx.Tx, error) { return m, nil }
func (m *mockTx) Commit(_ context.Context) error          { return nil }
func (m *mockTx) Rollback(_ context.Context) error        { return nil }
func (m *mockTx) CopyFrom(_ context.Context, _ pgx.Identifier, _ []string, _ pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *mockTx) SendBatch(_ context.Context, _ *pgx.Batch) pgx.BatchResults { return nil }
func (m *mockTx) LargeObjects() pgx.LargeObjects                             { return pgx.LargeObjects{} }
func (m *mockTx) Prepare(_ context.Context, _ string, _ string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (m *mockTx) Exec(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (m *mockTx) Query(_ context.Context, _ string, _ ...any) (pgx.Rows, error) {
	return nil, nil
}
func (m *mockTx) QueryRow(_ context.Context, _ string, _ ...any) pgx.Row {
	return nil
}
func (m *mockTx) Conn() *pgx.Conn { return nil }
