package security_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

type mockSharingRuleRepo struct {
	rules map[uuid.UUID]*security.SharingRule
}

func newMockSharingRuleRepo() *mockSharingRuleRepo {
	return &mockSharingRuleRepo{rules: make(map[uuid.UUID]*security.SharingRule)}
}

func (r *mockSharingRuleRepo) Create(_ context.Context, _ pgx.Tx, input security.CreateSharingRuleInput) (*security.SharingRule, error) {
	rule := &security.SharingRule{
		ID:            uuid.New(),
		ObjectID:      input.ObjectID,
		RuleType:      input.RuleType,
		SourceGroupID: input.SourceGroupID,
		TargetGroupID: input.TargetGroupID,
		AccessLevel:   input.AccessLevel,
		CriteriaField: input.CriteriaField,
		CriteriaOp:    input.CriteriaOp,
		CriteriaValue: input.CriteriaValue,
	}
	r.rules[rule.ID] = rule
	return rule, nil
}

func (r *mockSharingRuleRepo) GetByID(_ context.Context, id uuid.UUID) (*security.SharingRule, error) {
	return r.rules[id], nil
}

func (r *mockSharingRuleRepo) ListByObjectID(_ context.Context, objectID uuid.UUID) ([]security.SharingRule, error) {
	var result []security.SharingRule
	for _, rule := range r.rules {
		if rule.ObjectID == objectID {
			result = append(result, *rule)
		}
	}
	return result, nil
}

func (r *mockSharingRuleRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input security.UpdateSharingRuleInput) (*security.SharingRule, error) {
	rule := r.rules[id]
	if rule == nil {
		return nil, nil
	}
	rule.TargetGroupID = input.TargetGroupID
	rule.AccessLevel = input.AccessLevel
	return rule, nil
}

func (r *mockSharingRuleRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	delete(r.rules, id)
	return nil
}

func TestSharingRuleService_Create(t *testing.T) {
	sourceGroupID := uuid.New()
	targetGroupID := uuid.New()
	objectID := uuid.New()

	tests := []struct {
		name    string
		input   security.CreateSharingRuleInput
		setup   func(*mockGroupRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "creates owner_based rule successfully",
			input: security.CreateSharingRuleInput{
				ObjectID:      objectID,
				RuleType:      security.RuleTypeOwnerBased,
				SourceGroupID: sourceGroupID,
				TargetGroupID: targetGroupID,
				AccessLevel:   "read",
			},
			setup: func(r *mockGroupRepo) {
				r.groups[sourceGroupID] = &security.Group{ID: sourceGroupID}
				r.groups[targetGroupID] = &security.Group{ID: targetGroupID}
			},
			wantErr: false,
		},
		{
			name: "returns error when rule_type is invalid",
			input: security.CreateSharingRuleInput{
				ObjectID:      objectID,
				RuleType:      "invalid",
				SourceGroupID: sourceGroupID,
				TargetGroupID: targetGroupID,
				AccessLevel:   "read",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns error when access_level is invalid",
			input: security.CreateSharingRuleInput{
				ObjectID:      objectID,
				RuleType:      security.RuleTypeOwnerBased,
				SourceGroupID: sourceGroupID,
				TargetGroupID: targetGroupID,
				AccessLevel:   "write",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns error when source group not found",
			input: security.CreateSharingRuleInput{
				ObjectID:      objectID,
				RuleType:      security.RuleTypeOwnerBased,
				SourceGroupID: uuid.New(),
				TargetGroupID: targetGroupID,
				AccessLevel:   "read",
			},
			setup: func(r *mockGroupRepo) {
				r.groups[targetGroupID] = &security.Group{ID: targetGroupID}
			},
			wantErr: true,
			errCode: "NOT_FOUND",
		},
		{
			name: "returns error when criteria_based without criteria fields",
			input: security.CreateSharingRuleInput{
				ObjectID:      objectID,
				RuleType:      security.RuleTypeCriteriaBased,
				SourceGroupID: sourceGroupID,
				TargetGroupID: targetGroupID,
				AccessLevel:   "read",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleRepo := newMockSharingRuleRepo()
			groupRepo := newMockGroupRepo()
			outboxRepo := &mockOutboxRepo{}

			if tt.setup != nil {
				tt.setup(groupRepo)
			}

			svc := security.NewSharingRuleService(&mockTxBeginner{}, ruleRepo, groupRepo, outboxRepo)
			result, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errCode != "" {
					var appErr *apperror.AppError
					if errors.As(err, &appErr) {
						if string(appErr.Code) != tt.errCode {
							t.Errorf("expected error code %s, got %s", tt.errCode, appErr.Code)
						}
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.ObjectID != tt.input.ObjectID {
				t.Errorf("expected object_id %s, got %s", tt.input.ObjectID, result.ObjectID)
			}
		})
	}
}

func TestSharingRuleService_GetByID(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockSharingRuleRepo)
		wantErr bool
	}{
		{
			name: "returns rule when exists",
			id:   existingID,
			setup: func(r *mockSharingRuleRepo) {
				r.rules[existingID] = &security.SharingRule{ID: existingID, ObjectID: uuid.New()}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent rule",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleRepo := newMockSharingRuleRepo()
			if tt.setup != nil {
				tt.setup(ruleRepo)
			}

			svc := security.NewSharingRuleService(&mockTxBeginner{}, ruleRepo, newMockGroupRepo(), &mockOutboxRepo{})
			result, err := svc.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.ID != tt.id {
				t.Errorf("expected ID %s, got %s", tt.id, result.ID)
			}
		})
	}
}

func TestSharingRuleService_ListByObjectID(t *testing.T) {
	objectID := uuid.New()
	ruleRepo := newMockSharingRuleRepo()
	ruleID := uuid.New()
	ruleRepo.rules[ruleID] = &security.SharingRule{ID: ruleID, ObjectID: objectID}

	svc := security.NewSharingRuleService(&mockTxBeginner{}, ruleRepo, newMockGroupRepo(), &mockOutboxRepo{})
	rules, err := svc.ListByObjectID(context.Background(), objectID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(rules))
	}
}

func TestSharingRuleService_Update(t *testing.T) {
	existingID := uuid.New()
	targetGroupID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		input   security.UpdateSharingRuleInput
		setup   func(*mockSharingRuleRepo)
		wantErr bool
	}{
		{
			name: "updates rule successfully",
			id:   existingID,
			input: security.UpdateSharingRuleInput{
				TargetGroupID: targetGroupID,
				AccessLevel:   "read_write",
			},
			setup: func(r *mockSharingRuleRepo) {
				r.rules[existingID] = &security.SharingRule{ID: existingID, ObjectID: uuid.New(), AccessLevel: "read"}
			},
			wantErr: false,
		},
		{
			name: "returns not found for non-existent rule",
			id:   uuid.New(),
			input: security.UpdateSharingRuleInput{
				TargetGroupID: targetGroupID,
				AccessLevel:   "read",
			},
			wantErr: true,
		},
		{
			name: "returns validation error for invalid access_level",
			id:   existingID,
			input: security.UpdateSharingRuleInput{
				TargetGroupID: targetGroupID,
				AccessLevel:   "write",
			},
			setup: func(r *mockSharingRuleRepo) {
				r.rules[existingID] = &security.SharingRule{ID: existingID, ObjectID: uuid.New()}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleRepo := newMockSharingRuleRepo()
			if tt.setup != nil {
				tt.setup(ruleRepo)
			}

			svc := security.NewSharingRuleService(&mockTxBeginner{}, ruleRepo, newMockGroupRepo(), &mockOutboxRepo{})
			result, err := svc.Update(context.Background(), tt.id, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.AccessLevel != tt.input.AccessLevel {
				t.Errorf("expected access_level %s, got %s", tt.input.AccessLevel, result.AccessLevel)
			}
		})
	}
}

func TestSharingRuleService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockSharingRuleRepo)
		wantErr bool
	}{
		{
			name: "deletes existing rule",
			id:   uuid.New(),
			setup: func(r *mockSharingRuleRepo) {
				// id will be set in the test
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent rule",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleRepo := newMockSharingRuleRepo()
			groupRepo := newMockGroupRepo()
			outboxRepo := &mockOutboxRepo{}

			if !tt.wantErr {
				ruleRepo.rules[tt.id] = &security.SharingRule{ID: tt.id, ObjectID: uuid.New()}
			}

			svc := security.NewSharingRuleService(&mockTxBeginner{}, ruleRepo, groupRepo, outboxRepo)
			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
