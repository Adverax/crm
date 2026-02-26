package metadata

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProcedureRepo is a minimal in-memory ProcedureRepository for unit tests.
type mockProcedureRepo struct {
	procedures map[uuid.UUID]*Procedure
	versions   map[uuid.UUID]*ProcedureVersion
}

func newMockProcedureRepo() *mockProcedureRepo {
	return &mockProcedureRepo{
		procedures: make(map[uuid.UUID]*Procedure),
		versions:   make(map[uuid.UUID]*ProcedureVersion),
	}
}

func (r *mockProcedureRepo) Create(_ context.Context, input CreateProcedureInput) (*Procedure, error) {
	p := &Procedure{
		ID:          uuid.New(),
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
	}
	r.procedures[p.ID] = p
	return p, nil
}

func (r *mockProcedureRepo) GetByID(_ context.Context, id uuid.UUID) (*Procedure, error) {
	p := r.procedures[id]
	return p, nil
}

func (r *mockProcedureRepo) GetByCode(_ context.Context, code string) (*Procedure, error) {
	for _, p := range r.procedures {
		if p.Code == code {
			return p, nil
		}
	}
	return nil, nil
}

func (r *mockProcedureRepo) ListAll(_ context.Context) ([]Procedure, error) {
	var result []Procedure
	for _, p := range r.procedures {
		result = append(result, *p)
	}
	return result, nil
}

func (r *mockProcedureRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.procedures, id)
	return nil
}

func (r *mockProcedureRepo) Count(_ context.Context) (int, error) {
	return len(r.procedures), nil
}

func (r *mockProcedureRepo) UpdateMetadata(_ context.Context, id uuid.UUID, input UpdateProcedureMetadataInput) (*Procedure, error) {
	p := r.procedures[id]
	if p == nil {
		return nil, nil
	}
	p.Name = input.Name
	p.Description = input.Description
	return p, nil
}

func (r *mockProcedureRepo) SetDraftVersionID(_ context.Context, id uuid.UUID, versionID *uuid.UUID) error {
	if p := r.procedures[id]; p != nil {
		p.DraftVersionID = versionID
	}
	return nil
}

func (r *mockProcedureRepo) SetPublishedVersionID(_ context.Context, id uuid.UUID, versionID *uuid.UUID) error {
	if p := r.procedures[id]; p != nil {
		p.PublishedVersionID = versionID
	}
	return nil
}

func (r *mockProcedureRepo) CreateVersion(_ context.Context, procID uuid.UUID, version int, def ProcedureDefinition, changeSummary string, createdBy *uuid.UUID) (*ProcedureVersion, error) {
	v := &ProcedureVersion{
		ID:            uuid.New(),
		ProcedureID:   procID,
		Version:       version,
		Definition:    def,
		Status:        VersionStatusDraft,
		ChangeSummary: changeSummary,
		CreatedBy:     createdBy,
	}
	r.versions[v.ID] = v
	return v, nil
}

func (r *mockProcedureRepo) GetVersionByID(_ context.Context, id uuid.UUID) (*ProcedureVersion, error) {
	v := r.versions[id]
	return v, nil
}

func (r *mockProcedureRepo) GetDraftVersion(_ context.Context, procID uuid.UUID) (*ProcedureVersion, error) {
	for _, v := range r.versions {
		if v.ProcedureID == procID && v.Status == VersionStatusDraft {
			return v, nil
		}
	}
	return nil, nil
}

func (r *mockProcedureRepo) GetPublishedVersion(_ context.Context, procID uuid.UUID) (*ProcedureVersion, error) {
	for _, v := range r.versions {
		if v.ProcedureID == procID && v.Status == VersionStatusPublished {
			return v, nil
		}
	}
	return nil, nil
}

func (r *mockProcedureRepo) UpdateDraft(_ context.Context, versionID uuid.UUID, def ProcedureDefinition, changeSummary string) (*ProcedureVersion, error) {
	v := r.versions[versionID]
	if v == nil || v.Status != VersionStatusDraft {
		return nil, nil
	}
	v.Definition = def
	v.ChangeSummary = changeSummary
	return v, nil
}

func (r *mockProcedureRepo) DeleteVersion(_ context.Context, versionID uuid.UUID) error {
	delete(r.versions, versionID)
	return nil
}

func (r *mockProcedureRepo) ListVersions(_ context.Context, procID uuid.UUID) ([]ProcedureVersion, error) {
	var result []ProcedureVersion
	for _, v := range r.versions {
		if v.ProcedureID == procID {
			result = append(result, *v)
		}
	}
	return result, nil
}

func (r *mockProcedureRepo) UpdateVersionStatus(_ context.Context, versionID uuid.UUID, status VersionStatus) error {
	if v := r.versions[versionID]; v != nil {
		v.Status = status
	}
	return nil
}

func (r *mockProcedureRepo) SetVersionPublishedAt(_ context.Context, versionID uuid.UUID) error {
	return nil
}

func (r *mockProcedureRepo) GetMaxVersion(_ context.Context, procID uuid.UUID) (int, error) {
	maxVer := 0
	for _, v := range r.versions {
		if v.ProcedureID == procID && v.Version > maxVer {
			maxVer = v.Version
		}
	}
	return maxVer, nil
}

func (r *mockProcedureRepo) GetPreviousPublished(_ context.Context, procID uuid.UUID, beforeVersion int) (*ProcedureVersion, error) {
	var best *ProcedureVersion
	for _, v := range r.versions {
		if v.ProcedureID == procID && v.Status == VersionStatusSuperseded && v.Version < beforeVersion {
			if best == nil || v.Version > best.Version {
				best = v
			}
		}
	}
	return best, nil
}

func (r *mockProcedureRepo) CountSuperseded(_ context.Context, procID uuid.UUID) (int, error) {
	count := 0
	for _, v := range r.versions {
		if v.ProcedureID == procID && v.Status == VersionStatusSuperseded {
			count++
		}
	}
	return count, nil
}

func (r *mockProcedureRepo) DeleteOldestSuperseded(_ context.Context, _ uuid.UUID, _ int) error {
	return nil
}

// mockCacheForProcedures provides a minimal MetadataCache for procedure tests.
func mockCacheForProcedures() *MetadataCache {
	return &MetadataCache{
		objectsByID:               make(map[uuid.UUID]ObjectDefinition),
		objectsByAPIName:          make(map[string]ObjectDefinition),
		fieldsByID:                make(map[uuid.UUID]FieldDefinition),
		fieldsByObjectID:          make(map[uuid.UUID][]FieldDefinition),
		forwardRels:               make(map[uuid.UUID][]RelationshipInfo),
		reverseRels:               make(map[uuid.UUID][]RelationshipInfo),
		validationRulesByObjectID: make(map[uuid.UUID][]ValidationRule),
		functionsByName:           make(map[string]Function),
		objectViewsByAPIName:      make(map[string]ObjectView),
		proceduresByCode:          make(map[string]Procedure),
		automationRulesByObjectID: make(map[uuid.UUID][]AutomationRule),
		loader:                    &noopCacheLoader{},
	}
}

type noopCacheLoader struct{}

func (l *noopCacheLoader) LoadAllObjects(_ context.Context) ([]ObjectDefinition, error) {
	return nil, nil
}
func (l *noopCacheLoader) LoadAllFields(_ context.Context) ([]FieldDefinition, error) {
	return nil, nil
}
func (l *noopCacheLoader) LoadRelationships(_ context.Context) ([]RelationshipInfo, error) {
	return nil, nil
}
func (l *noopCacheLoader) LoadAllValidationRules(_ context.Context) ([]ValidationRule, error) {
	return nil, nil
}
func (l *noopCacheLoader) LoadAllFunctions(_ context.Context) ([]Function, error) { return nil, nil }
func (l *noopCacheLoader) LoadAllObjectViews(_ context.Context) ([]ObjectView, error) {
	return nil, nil
}
func (l *noopCacheLoader) LoadAllProcedures(_ context.Context) ([]Procedure, error) { return nil, nil }
func (l *noopCacheLoader) LoadAllAutomationRules(_ context.Context) ([]AutomationRule, error) {
	return nil, nil
}
func (l *noopCacheLoader) LoadAllLayouts(_ context.Context) ([]Layout, error) { return nil, nil }
func (l *noopCacheLoader) LoadAllSharedLayouts(_ context.Context) ([]SharedLayout, error) {
	return nil, nil
}
func (l *noopCacheLoader) RefreshMaterializedView(_ context.Context) error { return nil }

func TestProcedureService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateProcedureInput
		wantErr bool
	}{
		{
			name: "creates procedure with initial draft",
			input: CreateProcedureInput{
				Code: "send_welcome",
				Name: "Send Welcome",
			},
			wantErr: false,
		},
		{
			name: "rejects invalid code",
			input: CreateProcedureInput{
				Code: "INVALID-CODE",
				Name: "Bad",
			},
			wantErr: true,
		},
		{
			name: "rejects empty name",
			input: CreateProcedureInput{
				Code: "test",
				Name: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := newMockProcedureRepo()
			cache := mockCacheForProcedures()
			svc := NewProcedureService(repo, cache, nil)

			result, err := svc.Create(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.input.Code, result.Procedure.Code)
			assert.NotNil(t, result.DraftVersion)
			assert.Equal(t, 1, result.DraftVersion.Version)
		})
	}
}

func TestProcedureService_DuplicateCode(t *testing.T) {
	t.Parallel()

	repo := newMockProcedureRepo()
	cache := mockCacheForProcedures()
	svc := NewProcedureService(repo, cache, nil)

	_, err := svc.Create(context.Background(), CreateProcedureInput{
		Code: "dup_test",
		Name: "First",
	})
	require.NoError(t, err)

	_, err = svc.Create(context.Background(), CreateProcedureInput{
		Code: "dup_test",
		Name: "Second",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestProcedureService_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*procedureService) uuid.UUID
		wantErr bool
	}{
		{
			name: "returns procedure with versions",
			setup: func(svc *procedureService) uuid.UUID {
				result, _ := svc.Create(context.Background(), CreateProcedureInput{
					Code: "test_get",
					Name: "Test Get",
				})
				return result.Procedure.ID
			},
			wantErr: false,
		},
		{
			name: "returns not found for nonexistent",
			setup: func(_ *procedureService) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := newMockProcedureRepo()
			cache := mockCacheForProcedures()
			svc := NewProcedureService(repo, cache, nil).(*procedureService)

			id := tt.setup(svc)
			result, err := svc.GetByID(context.Background(), id)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, id, result.Procedure.ID)
		})
	}
}

func TestProcedureService_SaveDraft(t *testing.T) {
	t.Parallel()

	repo := newMockProcedureRepo()
	cache := mockCacheForProcedures()
	svc := NewProcedureService(repo, cache, nil)

	result, err := svc.Create(context.Background(), CreateProcedureInput{
		Code: "save_test",
		Name: "Save Test",
	})
	require.NoError(t, err)

	def := ProcedureDefinition{
		Commands: []CommandDef{
			{Type: "compute.transform", As: "step1", Value: map[string]string{"x": "1"}},
		},
	}

	version, err := svc.SaveDraft(context.Background(), result.Procedure.ID, SaveDraftInput{
		Definition:    def,
		ChangeSummary: "Added transform step",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, len(version.Definition.Commands))
}

func TestProcedureService_PublishAndRollback(t *testing.T) {
	t.Parallel()

	repo := newMockProcedureRepo()
	cache := mockCacheForProcedures()
	svc := NewProcedureService(repo, cache, nil)

	result, err := svc.Create(context.Background(), CreateProcedureInput{
		Code: "pub_test",
		Name: "Publish Test",
	})
	require.NoError(t, err)

	// Save a non-empty draft
	_, err = svc.SaveDraft(context.Background(), result.Procedure.ID, SaveDraftInput{
		Definition: ProcedureDefinition{
			Commands: []CommandDef{
				{Type: "compute.fail", Code: "ERR", Message: "test"},
			},
		},
	})
	require.NoError(t, err)

	// Publish
	published, err := svc.Publish(context.Background(), result.Procedure.ID)
	require.NoError(t, err)
	assert.Equal(t, VersionStatusPublished, published.Status)

	// Create and publish a second draft
	_, err = svc.CreateDraftFromPublished(context.Background(), result.Procedure.ID)
	require.NoError(t, err)

	published2, err := svc.Publish(context.Background(), result.Procedure.ID)
	require.NoError(t, err)
	assert.Equal(t, VersionStatusPublished, published2.Status)

	// Rollback
	rolled, err := svc.Rollback(context.Background(), result.Procedure.ID)
	require.NoError(t, err)
	assert.Equal(t, VersionStatusPublished, rolled.Status)
	assert.Equal(t, published.ID, rolled.ID)
}

func TestProcedureService_DiscardDraft(t *testing.T) {
	t.Parallel()

	repo := newMockProcedureRepo()
	cache := mockCacheForProcedures()
	svc := NewProcedureService(repo, cache, nil)

	result, err := svc.Create(context.Background(), CreateProcedureInput{
		Code: "discard_test",
		Name: "Discard Test",
	})
	require.NoError(t, err)

	err = svc.DiscardDraft(context.Background(), result.Procedure.ID)
	require.NoError(t, err)

	// Second discard should fail
	err = svc.DiscardDraft(context.Background(), result.Procedure.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no draft")
}

func TestValidateDefinition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		def     ProcedureDefinition
		wantErr bool
	}{
		{
			name:    "empty definition is valid",
			def:     ProcedureDefinition{Commands: []CommandDef{}},
			wantErr: false,
		},
		{
			name: "valid single command",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{Type: "compute.transform", As: "step1"},
				},
			},
			wantErr: false,
		},
		{
			name: "unknown command type",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{Type: "unknown.cmd"},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate as names",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{Type: "compute.transform", As: "step1"},
					{Type: "compute.transform", As: "step1"},
				},
			},
			wantErr: true,
		},
		{
			name: "missing command type",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{As: "step1"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid retry config",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{
						Type:  "compute.transform",
						As:    "step1",
						Retry: &RetryConfig{MaxAttempts: 3, DelayMs: 1000},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid retry max_attempts too high",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{
						Type:  "compute.transform",
						Retry: &RetryConfig{MaxAttempts: 10, DelayMs: 1000},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid retry max_attempts zero",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{
						Type:  "compute.transform",
						Retry: &RetryConfig{MaxAttempts: 0, DelayMs: 1000},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid retry delay_ms too low",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{
						Type:  "compute.transform",
						Retry: &RetryConfig{MaxAttempts: 2, DelayMs: 10},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid retry delay_ms too high",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{
						Type:  "compute.transform",
						Retry: &RetryConfig{MaxAttempts: 2, DelayMs: 100000},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid flow.try with try and catch",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{
						Type:  "flow.try",
						As:    "result",
						Try:   []CommandDef{{Type: "compute.transform", As: "s1"}},
						Catch: []CommandDef{{Type: "compute.transform", As: "s2"}},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "flow.try nested depth counts correctly",
			def: ProcedureDefinition{
				Commands: []CommandDef{
					{
						Type: "flow.try",
						Try: []CommandDef{
							{
								Type: "flow.try",
								Try: []CommandDef{
									{Type: "compute.transform", As: "deep"},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateDefinition(tt.def)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
