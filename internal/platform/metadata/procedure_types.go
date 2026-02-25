package metadata

import (
	"time"

	"github.com/google/uuid"
)

// VersionStatus represents the status of a procedure version.
type VersionStatus string

const (
	VersionStatusDraft      VersionStatus = "draft"
	VersionStatusPublished  VersionStatus = "published"
	VersionStatusSuperseded VersionStatus = "superseded"
)

// Procedure represents a named business procedure (ADR-0024).
type Procedure struct {
	ID                 uuid.UUID  `json:"id"`
	Code               string     `json:"code"`
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	DraftVersionID     *uuid.UUID `json:"draft_version_id"`
	PublishedVersionID *uuid.UUID `json:"published_version_id"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// ProcedureVersion represents a versioned snapshot of a procedure definition (ADR-0029).
type ProcedureVersion struct {
	ID            uuid.UUID           `json:"id"`
	ProcedureID   uuid.UUID           `json:"procedure_id"`
	Version       int                 `json:"version"`
	Definition    ProcedureDefinition `json:"definition"`
	Status        VersionStatus       `json:"status"`
	ChangeSummary string              `json:"change_summary"`
	CreatedBy     *uuid.UUID          `json:"created_by"`
	CreatedAt     time.Time           `json:"created_at"`
	PublishedAt   *time.Time          `json:"published_at"`
}

// ProcedureDefinition describes the executable body of a procedure.
type ProcedureDefinition struct {
	Commands []CommandDef      `json:"commands"`
	Result   map[string]string `json:"result,omitempty"`
}

// CommandDef describes a single command within a procedure definition.
type CommandDef struct {
	Type      string                  `json:"type"`
	As        string                  `json:"as,omitempty"`
	Optional  bool                    `json:"optional,omitempty"`
	When      string                  `json:"when,omitempty"`
	Rollback  []CommandDef            `json:"rollback,omitempty"`
	Object    string                  `json:"object,omitempty"`
	ID        string                  `json:"id,omitempty"`
	Data      map[string]string       `json:"data,omitempty"`
	Query     string                  `json:"query,omitempty"`
	Value     map[string]string       `json:"value,omitempty"`
	Condition string                  `json:"condition,omitempty"`
	Code      string                  `json:"code,omitempty"`
	Message   string                  `json:"message,omitempty"`
	Then      []CommandDef            `json:"then,omitempty"`
	Else      []CommandDef            `json:"else,omitempty"`
	Cases     map[string][]CommandDef `json:"cases,omitempty"`
	Default   []CommandDef            `json:"default,omitempty"`
	Procedure string                  `json:"procedure,omitempty"`
	Input     map[string]string       `json:"input,omitempty"`
	// integration.http fields
	Credential string            `json:"credential,omitempty"`
	Method     string            `json:"method,omitempty"`
	Path       string            `json:"path,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
	// Expression field for flow.match
	Expression string `json:"expression,omitempty"`
	// flow.try fields
	Try   []CommandDef `json:"try,omitempty"`
	Catch []CommandDef `json:"catch,omitempty"`
	// Retry config
	Retry *RetryConfig `json:"retry,omitempty"`
}

// RetryConfig describes retry behavior for a command.
type RetryConfig struct {
	MaxAttempts int `json:"max_attempts"`
	DelayMs     int `json:"delay_ms"`
	BackoffMult int `json:"backoff_mult,omitempty"` // multiplier (default 1 = fixed delay)
}

// ProcedureWithVersions combines a procedure with its draft and published versions.
type ProcedureWithVersions struct {
	Procedure        Procedure         `json:"procedure"`
	DraftVersion     *ProcedureVersion `json:"draft_version,omitempty"`
	PublishedVersion *ProcedureVersion `json:"published_version,omitempty"`
}

// CreateProcedureInput is the input for creating a new procedure.
type CreateProcedureInput struct {
	Code        string
	Name        string
	Description string
}

// UpdateProcedureMetadataInput is the input for updating procedure metadata.
type UpdateProcedureMetadataInput struct {
	Name        string
	Description string
}

// SaveDraftInput is the input for saving a draft version.
type SaveDraftInput struct {
	Definition    ProcedureDefinition
	ChangeSummary string
}
