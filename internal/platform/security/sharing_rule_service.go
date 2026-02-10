package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
)

type sharingRuleServiceImpl struct {
	txBeginner TxBeginner
	ruleRepo   SharingRuleRepository
	groupRepo  GroupRepository
	outboxRepo OutboxRepository
}

// NewSharingRuleService creates a new SharingRuleService.
func NewSharingRuleService(
	txBeginner TxBeginner,
	ruleRepo SharingRuleRepository,
	groupRepo GroupRepository,
	outboxRepo OutboxRepository,
) SharingRuleService {
	return &sharingRuleServiceImpl{
		txBeginner: txBeginner,
		ruleRepo:   ruleRepo,
		groupRepo:  groupRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *sharingRuleServiceImpl) Create(ctx context.Context, input CreateSharingRuleInput) (*SharingRule, error) {
	if err := ValidateCreateSharingRule(input); err != nil {
		return nil, fmt.Errorf("sharingRuleService.Create: %w", err)
	}

	// Validate source and target groups exist
	sourceGroup, err := s.groupRepo.GetByID(ctx, input.SourceGroupID)
	if err != nil {
		return nil, fmt.Errorf("sharingRuleService.Create: lookup source group: %w", err)
	}
	if sourceGroup == nil {
		return nil, fmt.Errorf("sharingRuleService.Create: %w",
			apperror.NotFound("Group", input.SourceGroupID.String()))
	}

	targetGroup, err := s.groupRepo.GetByID(ctx, input.TargetGroupID)
	if err != nil {
		return nil, fmt.Errorf("sharingRuleService.Create: lookup target group: %w", err)
	}
	if targetGroup == nil {
		return nil, fmt.Errorf("sharingRuleService.Create: %w",
			apperror.NotFound("Group", input.TargetGroupID.String()))
	}

	var result *SharingRule
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.ruleRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("sharingRuleService.Create: %w", err)
		}

		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "sharing_rule_changed",
			EntityType: "sharing_rule",
			EntityID:   created.ID,
			Payload:    []byte(`{"action":"create"}`),
		}); err != nil {
			return fmt.Errorf("sharingRuleService.Create: outbox: %w", err)
		}

		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *sharingRuleServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*SharingRule, error) {
	rule, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sharingRuleService.GetByID: %w", err)
	}
	if rule == nil {
		return nil, fmt.Errorf("sharingRuleService.GetByID: %w",
			apperror.NotFound("SharingRule", id.String()))
	}
	return rule, nil
}

func (s *sharingRuleServiceImpl) ListByObjectID(ctx context.Context, objectID uuid.UUID) ([]SharingRule, error) {
	rules, err := s.ruleRepo.ListByObjectID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("sharingRuleService.ListByObjectID: %w", err)
	}
	return rules, nil
}

func (s *sharingRuleServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateSharingRuleInput) (*SharingRule, error) {
	existing, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sharingRuleService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("sharingRuleService.Update: %w",
			apperror.NotFound("SharingRule", id.String()))
	}

	if err := ValidateUpdateSharingRule(input); err != nil {
		return nil, fmt.Errorf("sharingRuleService.Update: %w", err)
	}

	var result *SharingRule
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		updated, err := s.ruleRepo.Update(ctx, tx, id, input)
		if err != nil {
			return fmt.Errorf("sharingRuleService.Update: %w", err)
		}

		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "sharing_rule_changed",
			EntityType: "sharing_rule",
			EntityID:   id,
			Payload:    []byte(`{"action":"update"}`),
		}); err != nil {
			return fmt.Errorf("sharingRuleService.Update: outbox: %w", err)
		}

		result = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *sharingRuleServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.ruleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("sharingRuleService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("sharingRuleService.Delete: %w",
			apperror.NotFound("SharingRule", id.String()))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.ruleRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("sharingRuleService.Delete: %w", err)
		}

		if err := s.outboxRepo.Insert(ctx, tx, OutboxEvent{
			EventType:  "sharing_rule_changed",
			EntityType: "sharing_rule",
			EntityID:   id,
			Payload:    []byte(`{"action":"delete"}`),
		}); err != nil {
			return fmt.Errorf("sharingRuleService.Delete: outbox: %w", err)
		}

		return nil
	})
}
