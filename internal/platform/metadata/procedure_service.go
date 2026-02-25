package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

const (
	maxProcedureCount        = 200
	maxProcedureCommands     = 50
	maxProcedureNestingDepth = 5
	maxProcedureDefSize      = 65536 // 64KB
	maxSupersededVersions    = 10
)

var validProcedureCode = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Known command types for validation.
var knownCommandTypes = map[string]bool{
	"record.create":       true,
	"record.update":       true,
	"record.delete":       true,
	"record.get":          true,
	"record.query":        true,
	"compute.transform":   true,
	"compute.validate":    true,
	"compute.fail":        true,
	"flow.if":             true,
	"flow.match":          true,
	"flow.call":           true,
	"flow.try":            true,
	"integration.http":    true,
	"notification.email":  true,
	"notification.in_app": true,
	"wait.delay":          true,
	"wait.approval":       true,
}

// ProcedureService provides business logic for procedures (ADR-0024, ADR-0029).
type ProcedureService interface {
	Create(ctx context.Context, input CreateProcedureInput) (*ProcedureWithVersions, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ProcedureWithVersions, error)
	GetByCode(ctx context.Context, code string) (*Procedure, error)
	ListAll(ctx context.Context) ([]Procedure, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateMetadata(ctx context.Context, id uuid.UUID, input UpdateProcedureMetadataInput) (*Procedure, error)

	SaveDraft(ctx context.Context, id uuid.UUID, input SaveDraftInput) (*ProcedureVersion, error)
	DiscardDraft(ctx context.Context, id uuid.UUID) error
	CreateDraftFromPublished(ctx context.Context, id uuid.UUID) (*ProcedureVersion, error)

	Publish(ctx context.Context, id uuid.UUID) (*ProcedureVersion, error)
	Rollback(ctx context.Context, id uuid.UUID) (*ProcedureVersion, error)
	ListVersions(ctx context.Context, id uuid.UUID) ([]ProcedureVersion, error)

	GetPublishedDefinition(ctx context.Context, code string) (*ProcedureDefinition, error)
}

// OnProceduresChanged is a callback invoked after procedures are modified.
type OnProceduresChanged func(ctx context.Context) error

type procedureService struct {
	repo     ProcedureRepository
	cache    *MetadataCache
	onChange OnProceduresChanged
}

// NewProcedureService creates a new ProcedureService.
func NewProcedureService(
	repo ProcedureRepository,
	cache *MetadataCache,
	onChange OnProceduresChanged,
) ProcedureService {
	return &procedureService{
		repo:     repo,
		cache:    cache,
		onChange: onChange,
	}
}

func (s *procedureService) Create(ctx context.Context, input CreateProcedureInput) (*ProcedureWithVersions, error) {
	if err := validateProcedureCode(input.Code); err != nil {
		return nil, fmt.Errorf("procedureService.Create: %w", err)
	}
	if input.Name == "" {
		return nil, fmt.Errorf("procedureService.Create: %w",
			apperror.BadRequest("name is required"))
	}

	count, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Create: %w", err)
	}
	if count >= maxProcedureCount {
		return nil, fmt.Errorf("procedureService.Create: %w",
			apperror.BadRequest(fmt.Sprintf("max procedure limit reached (%d)", maxProcedureCount)))
	}

	existing, err := s.repo.GetByCode(ctx, input.Code)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Create: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("procedureService.Create: %w",
			apperror.Conflict(fmt.Sprintf("procedure %q already exists", input.Code)))
	}

	proc, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Create: %w", err)
	}

	emptyDef := ProcedureDefinition{Commands: []CommandDef{}}
	version, err := s.repo.CreateVersion(ctx, proc.ID, 1, emptyDef, "Initial draft", nil)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Create: create initial version: %w", err)
	}

	if err := s.repo.SetDraftVersionID(ctx, proc.ID, &version.ID); err != nil {
		return nil, fmt.Errorf("procedureService.Create: set draft version: %w", err)
	}
	proc.DraftVersionID = &version.ID

	if err := s.reloadAndNotify(ctx); err != nil {
		return nil, fmt.Errorf("procedureService.Create: %w", err)
	}

	return &ProcedureWithVersions{
		Procedure:    *proc,
		DraftVersion: version,
	}, nil
}

func (s *procedureService) GetByID(ctx context.Context, id uuid.UUID) (*ProcedureWithVersions, error) {
	proc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.GetByID: %w", err)
	}
	if proc == nil {
		return nil, fmt.Errorf("procedureService.GetByID: %w",
			apperror.NotFound("procedure", id.String()))
	}

	result := &ProcedureWithVersions{Procedure: *proc}

	if proc.DraftVersionID != nil {
		draft, err := s.repo.GetVersionByID(ctx, *proc.DraftVersionID)
		if err != nil {
			return nil, fmt.Errorf("procedureService.GetByID: load draft: %w", err)
		}
		result.DraftVersion = draft
	}
	if proc.PublishedVersionID != nil {
		published, err := s.repo.GetVersionByID(ctx, *proc.PublishedVersionID)
		if err != nil {
			return nil, fmt.Errorf("procedureService.GetByID: load published: %w", err)
		}
		result.PublishedVersion = published
	}

	return result, nil
}

func (s *procedureService) GetByCode(ctx context.Context, code string) (*Procedure, error) {
	proc, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("procedureService.GetByCode: %w", err)
	}
	if proc == nil {
		return nil, fmt.Errorf("procedureService.GetByCode: %w",
			apperror.NotFound("procedure", code))
	}
	return proc, nil
}

func (s *procedureService) ListAll(ctx context.Context) ([]Procedure, error) {
	procedures, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("procedureService.ListAll: %w", err)
	}
	return procedures, nil
}

func (s *procedureService) Delete(ctx context.Context, id uuid.UUID) error {
	proc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("procedureService.Delete: %w", err)
	}
	if proc == nil {
		return fmt.Errorf("procedureService.Delete: %w",
			apperror.NotFound("procedure", id.String()))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("procedureService.Delete: %w", err)
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return fmt.Errorf("procedureService.Delete: %w", err)
	}

	return nil
}

func (s *procedureService) UpdateMetadata(ctx context.Context, id uuid.UUID, input UpdateProcedureMetadataInput) (*Procedure, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("procedureService.UpdateMetadata: %w",
			apperror.BadRequest("name is required"))
	}

	proc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.UpdateMetadata: %w", err)
	}
	if proc == nil {
		return nil, fmt.Errorf("procedureService.UpdateMetadata: %w",
			apperror.NotFound("procedure", id.String()))
	}

	updated, err := s.repo.UpdateMetadata(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("procedureService.UpdateMetadata: %w", err)
	}
	if updated == nil {
		return nil, fmt.Errorf("procedureService.UpdateMetadata: %w",
			apperror.NotFound("procedure", id.String()))
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return nil, fmt.Errorf("procedureService.UpdateMetadata: %w", err)
	}

	return updated, nil
}

func (s *procedureService) SaveDraft(ctx context.Context, id uuid.UUID, input SaveDraftInput) (*ProcedureVersion, error) {
	if err := validateDefinition(input.Definition); err != nil {
		return nil, fmt.Errorf("procedureService.SaveDraft: %w", err)
	}

	proc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.SaveDraft: %w", err)
	}
	if proc == nil {
		return nil, fmt.Errorf("procedureService.SaveDraft: %w",
			apperror.NotFound("procedure", id.String()))
	}

	if proc.DraftVersionID != nil {
		version, err := s.repo.UpdateDraft(ctx, *proc.DraftVersionID, input.Definition, input.ChangeSummary)
		if err != nil {
			return nil, fmt.Errorf("procedureService.SaveDraft: %w", err)
		}
		if version != nil {
			return version, nil
		}
	}

	maxVer, err := s.repo.GetMaxVersion(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.SaveDraft: %w", err)
	}

	version, err := s.repo.CreateVersion(ctx, id, maxVer+1, input.Definition, input.ChangeSummary, nil)
	if err != nil {
		return nil, fmt.Errorf("procedureService.SaveDraft: create version: %w", err)
	}

	if err := s.repo.SetDraftVersionID(ctx, id, &version.ID); err != nil {
		return nil, fmt.Errorf("procedureService.SaveDraft: set draft: %w", err)
	}

	return version, nil
}

func (s *procedureService) DiscardDraft(ctx context.Context, id uuid.UUID) error {
	proc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("procedureService.DiscardDraft: %w", err)
	}
	if proc == nil {
		return fmt.Errorf("procedureService.DiscardDraft: %w",
			apperror.NotFound("procedure", id.String()))
	}
	if proc.DraftVersionID == nil {
		return fmt.Errorf("procedureService.DiscardDraft: %w",
			apperror.BadRequest("no draft to discard"))
	}

	draftVersionID := *proc.DraftVersionID

	if err := s.repo.SetDraftVersionID(ctx, id, nil); err != nil {
		return fmt.Errorf("procedureService.DiscardDraft: clear draft pointer: %w", err)
	}

	if err := s.repo.DeleteVersion(ctx, draftVersionID); err != nil {
		return fmt.Errorf("procedureService.DiscardDraft: delete version: %w", err)
	}

	return nil
}

func (s *procedureService) CreateDraftFromPublished(ctx context.Context, id uuid.UUID) (*ProcedureVersion, error) {
	proc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.CreateDraftFromPublished: %w", err)
	}
	if proc == nil {
		return nil, fmt.Errorf("procedureService.CreateDraftFromPublished: %w",
			apperror.NotFound("procedure", id.String()))
	}
	if proc.DraftVersionID != nil {
		return nil, fmt.Errorf("procedureService.CreateDraftFromPublished: %w",
			apperror.Conflict("draft already exists"))
	}
	if proc.PublishedVersionID == nil {
		return nil, fmt.Errorf("procedureService.CreateDraftFromPublished: %w",
			apperror.BadRequest("no published version to copy from"))
	}

	published, err := s.repo.GetVersionByID(ctx, *proc.PublishedVersionID)
	if err != nil {
		return nil, fmt.Errorf("procedureService.CreateDraftFromPublished: %w", err)
	}

	maxVer, err := s.repo.GetMaxVersion(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.CreateDraftFromPublished: %w", err)
	}

	draft, err := s.repo.CreateVersion(ctx, id, maxVer+1, published.Definition, "Copied from published", nil)
	if err != nil {
		return nil, fmt.Errorf("procedureService.CreateDraftFromPublished: %w", err)
	}

	if err := s.repo.SetDraftVersionID(ctx, id, &draft.ID); err != nil {
		return nil, fmt.Errorf("procedureService.CreateDraftFromPublished: %w", err)
	}

	return draft, nil
}

func (s *procedureService) Publish(ctx context.Context, id uuid.UUID) (*ProcedureVersion, error) {
	proc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Publish: %w", err)
	}
	if proc == nil {
		return nil, fmt.Errorf("procedureService.Publish: %w",
			apperror.NotFound("procedure", id.String()))
	}
	if proc.DraftVersionID == nil {
		return nil, fmt.Errorf("procedureService.Publish: %w",
			apperror.BadRequest("no draft to publish"))
	}

	draft, err := s.repo.GetVersionByID(ctx, *proc.DraftVersionID)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Publish: load draft: %w", err)
	}

	if err := validateDefinition(draft.Definition); err != nil {
		return nil, fmt.Errorf("procedureService.Publish: %w", err)
	}

	// Supersede current published version
	if proc.PublishedVersionID != nil {
		if err := s.repo.UpdateVersionStatus(ctx, *proc.PublishedVersionID, VersionStatusSuperseded); err != nil {
			return nil, fmt.Errorf("procedureService.Publish: supersede old: %w", err)
		}
	}

	// Promote draft to published
	if err := s.repo.UpdateVersionStatus(ctx, draft.ID, VersionStatusPublished); err != nil {
		return nil, fmt.Errorf("procedureService.Publish: publish draft: %w", err)
	}
	if err := s.repo.SetVersionPublishedAt(ctx, draft.ID); err != nil {
		return nil, fmt.Errorf("procedureService.Publish: set published_at: %w", err)
	}

	if err := s.repo.SetPublishedVersionID(ctx, id, &draft.ID); err != nil {
		return nil, fmt.Errorf("procedureService.Publish: update procedure: %w", err)
	}
	if err := s.repo.SetDraftVersionID(ctx, id, nil); err != nil {
		return nil, fmt.Errorf("procedureService.Publish: clear draft: %w", err)
	}

	// Cleanup old superseded versions (keep max 10)
	if err := s.repo.DeleteOldestSuperseded(ctx, id, maxSupersededVersions); err != nil {
		return nil, fmt.Errorf("procedureService.Publish: cleanup: %w", err)
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return nil, fmt.Errorf("procedureService.Publish: %w", err)
	}

	published, err := s.repo.GetVersionByID(ctx, draft.ID)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Publish: reload version: %w", err)
	}
	return published, nil
}

func (s *procedureService) Rollback(ctx context.Context, id uuid.UUID) (*ProcedureVersion, error) {
	proc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Rollback: %w", err)
	}
	if proc == nil {
		return nil, fmt.Errorf("procedureService.Rollback: %w",
			apperror.NotFound("procedure", id.String()))
	}
	if proc.PublishedVersionID == nil {
		return nil, fmt.Errorf("procedureService.Rollback: %w",
			apperror.BadRequest("no published version to rollback"))
	}

	currentPublished, err := s.repo.GetVersionByID(ctx, *proc.PublishedVersionID)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Rollback: %w", err)
	}

	previous, err := s.repo.GetPreviousPublished(ctx, id, currentPublished.Version)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Rollback: %w", err)
	}
	if previous == nil {
		return nil, fmt.Errorf("procedureService.Rollback: %w",
			apperror.BadRequest("no previous version to rollback to"))
	}

	// Current published → superseded
	if err := s.repo.UpdateVersionStatus(ctx, currentPublished.ID, VersionStatusSuperseded); err != nil {
		return nil, fmt.Errorf("procedureService.Rollback: supersede current: %w", err)
	}

	// Previous superseded → published
	if err := s.repo.UpdateVersionStatus(ctx, previous.ID, VersionStatusPublished); err != nil {
		return nil, fmt.Errorf("procedureService.Rollback: restore previous: %w", err)
	}
	if err := s.repo.SetVersionPublishedAt(ctx, previous.ID); err != nil {
		return nil, fmt.Errorf("procedureService.Rollback: set published_at: %w", err)
	}

	if err := s.repo.SetPublishedVersionID(ctx, id, &previous.ID); err != nil {
		return nil, fmt.Errorf("procedureService.Rollback: update procedure: %w", err)
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return nil, fmt.Errorf("procedureService.Rollback: %w", err)
	}

	restored, err := s.repo.GetVersionByID(ctx, previous.ID)
	if err != nil {
		return nil, fmt.Errorf("procedureService.Rollback: reload version: %w", err)
	}
	return restored, nil
}

func (s *procedureService) ListVersions(ctx context.Context, id uuid.UUID) ([]ProcedureVersion, error) {
	proc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.ListVersions: %w", err)
	}
	if proc == nil {
		return nil, fmt.Errorf("procedureService.ListVersions: %w",
			apperror.NotFound("procedure", id.String()))
	}

	versions, err := s.repo.ListVersions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("procedureService.ListVersions: %w", err)
	}
	return versions, nil
}

func (s *procedureService) GetPublishedDefinition(ctx context.Context, code string) (*ProcedureDefinition, error) {
	proc, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("procedureService.GetPublishedDefinition: %w", err)
	}
	if proc == nil {
		return nil, fmt.Errorf("procedureService.GetPublishedDefinition: %w",
			apperror.NotFound("procedure", code))
	}
	if proc.PublishedVersionID == nil {
		return nil, fmt.Errorf("procedureService.GetPublishedDefinition: %w",
			apperror.BadRequest(fmt.Sprintf("procedure %q has no published version", code)))
	}

	version, err := s.repo.GetVersionByID(ctx, *proc.PublishedVersionID)
	if err != nil {
		return nil, fmt.Errorf("procedureService.GetPublishedDefinition: %w", err)
	}
	return &version.Definition, nil
}

func (s *procedureService) reloadAndNotify(ctx context.Context) error {
	if err := s.cache.LoadProcedures(ctx); err != nil {
		return fmt.Errorf("cache reload: %w", err)
	}
	if s.onChange != nil {
		if err := s.onChange(ctx); err != nil {
			return fmt.Errorf("onChange callback: %w", err)
		}
	}
	return nil
}

// --- validation helpers ---

func validateProcedureCode(code string) error {
	if !validProcedureCode.MatchString(code) {
		return apperror.BadRequest("code must match ^[a-z][a-z0-9_]*$")
	}
	if len(code) > 100 {
		return apperror.BadRequest("code must be at most 100 characters")
	}
	return nil
}

func validateDefinition(def ProcedureDefinition) error {
	defJSON, err := json.Marshal(def)
	if err != nil {
		return apperror.BadRequest("invalid definition")
	}
	if len(defJSON) > maxProcedureDefSize {
		return apperror.BadRequest(fmt.Sprintf("definition must be at most %d bytes", maxProcedureDefSize))
	}

	commandCount := countCommands(def.Commands)
	if commandCount > maxProcedureCommands {
		return apperror.BadRequest(fmt.Sprintf("max %d commands allowed (got %d)", maxProcedureCommands, commandCount))
	}

	depth := maxCommandDepth(def.Commands, 0)
	if depth > maxProcedureNestingDepth {
		return apperror.BadRequest(fmt.Sprintf("max nesting depth is %d (got %d)", maxProcedureNestingDepth, depth))
	}

	asNames := make(map[string]bool)
	if err := validateCommands(def.Commands, asNames); err != nil {
		return err
	}

	return nil
}

func validateCommands(cmds []CommandDef, asNames map[string]bool) error {
	for _, cmd := range cmds {
		if cmd.Type == "" {
			return apperror.BadRequest("command type is required")
		}
		if !knownCommandTypes[cmd.Type] {
			return apperror.BadRequest(fmt.Sprintf("unknown command type: %s", cmd.Type))
		}
		if cmd.As != "" {
			if asNames[cmd.As] {
				return apperror.BadRequest(fmt.Sprintf("duplicate 'as' name: %s", cmd.As))
			}
			asNames[cmd.As] = true
		}

		if err := validateCommands(cmd.Then, asNames); err != nil {
			return err
		}
		if err := validateCommands(cmd.Else, asNames); err != nil {
			return err
		}
		if err := validateCommands(cmd.Rollback, asNames); err != nil {
			return err
		}
		for _, branch := range cmd.Cases {
			if err := validateCommands(branch, asNames); err != nil {
				return err
			}
		}
		if err := validateCommands(cmd.Default, asNames); err != nil {
			return err
		}
		if err := validateCommands(cmd.Try, asNames); err != nil {
			return err
		}
		if err := validateCommands(cmd.Catch, asNames); err != nil {
			return err
		}

		// Validate retry config
		if cmd.Retry != nil {
			if cmd.Retry.MaxAttempts < 1 || cmd.Retry.MaxAttempts > 5 {
				return apperror.BadRequest("retry max_attempts must be 1-5")
			}
			if cmd.Retry.DelayMs < 100 || cmd.Retry.DelayMs > 60000 {
				return apperror.BadRequest("retry delay_ms must be 100-60000")
			}
		}
	}
	return nil
}

func countCommands(cmds []CommandDef) int {
	count := len(cmds)
	for _, cmd := range cmds {
		count += countCommands(cmd.Then)
		count += countCommands(cmd.Else)
		count += countCommands(cmd.Rollback)
		for _, branch := range cmd.Cases {
			count += countCommands(branch)
		}
		count += countCommands(cmd.Default)
		count += countCommands(cmd.Try)
		count += countCommands(cmd.Catch)
	}
	return count
}

func maxCommandDepth(cmds []CommandDef, current int) int {
	if len(cmds) == 0 {
		return current
	}
	maxDepth := current + 1
	for _, cmd := range cmds {
		for _, nested := range [][]CommandDef{cmd.Then, cmd.Else, cmd.Default, cmd.Try, cmd.Catch} {
			d := maxCommandDepth(nested, current+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
		for _, branch := range cmd.Cases {
			d := maxCommandDepth(branch, current+1)
			if d > maxDepth {
				maxDepth = d
			}
		}
	}
	return maxDepth
}
