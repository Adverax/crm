package metadata

import (
	"context"

	"github.com/google/uuid"
)

// ProcedureRepository provides CRUD operations for procedures and their versions.
type ProcedureRepository interface {
	// Procedure CRUD
	Create(ctx context.Context, input CreateProcedureInput) (*Procedure, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Procedure, error)
	GetByCode(ctx context.Context, code string) (*Procedure, error)
	ListAll(ctx context.Context) ([]Procedure, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)
	UpdateMetadata(ctx context.Context, id uuid.UUID, input UpdateProcedureMetadataInput) (*Procedure, error)
	SetDraftVersionID(ctx context.Context, id uuid.UUID, versionID *uuid.UUID) error
	SetPublishedVersionID(ctx context.Context, id uuid.UUID, versionID *uuid.UUID) error

	// Version CRUD
	CreateVersion(ctx context.Context, procID uuid.UUID, version int, def ProcedureDefinition, changeSummary string, createdBy *uuid.UUID) (*ProcedureVersion, error)
	GetVersionByID(ctx context.Context, id uuid.UUID) (*ProcedureVersion, error)
	GetDraftVersion(ctx context.Context, procID uuid.UUID) (*ProcedureVersion, error)
	GetPublishedVersion(ctx context.Context, procID uuid.UUID) (*ProcedureVersion, error)
	UpdateDraft(ctx context.Context, versionID uuid.UUID, def ProcedureDefinition, changeSummary string) (*ProcedureVersion, error)
	DeleteVersion(ctx context.Context, versionID uuid.UUID) error
	ListVersions(ctx context.Context, procID uuid.UUID) ([]ProcedureVersion, error)
	UpdateVersionStatus(ctx context.Context, versionID uuid.UUID, status VersionStatus) error
	SetVersionPublishedAt(ctx context.Context, versionID uuid.UUID) error
	GetMaxVersion(ctx context.Context, procID uuid.UUID) (int, error)
	GetPreviousPublished(ctx context.Context, procID uuid.UUID, beforeVersion int) (*ProcedureVersion, error)
	CountSuperseded(ctx context.Context, procID uuid.UUID) (int, error)
	DeleteOldestSuperseded(ctx context.Context, procID uuid.UUID, keepCount int) error
}
