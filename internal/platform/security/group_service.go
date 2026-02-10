package security

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/pkg/apperror"
)

type groupServiceImpl struct {
	txBeginner TxBeginner
	groupRepo  GroupRepository
	memberRepo GroupMemberRepository
	outboxRepo OutboxRepository
}

// NewGroupService creates a new GroupService.
func NewGroupService(
	txBeginner TxBeginner,
	groupRepo GroupRepository,
	memberRepo GroupMemberRepository,
	outboxRepo OutboxRepository,
) GroupService {
	return &groupServiceImpl{
		txBeginner: txBeginner,
		groupRepo:  groupRepo,
		memberRepo: memberRepo,
		outboxRepo: outboxRepo,
	}
}

func (s *groupServiceImpl) Create(ctx context.Context, input CreateGroupInput) (*Group, error) {
	if err := ValidateCreateGroup(input); err != nil {
		return nil, fmt.Errorf("groupService.Create: %w", err)
	}

	existing, _ := s.groupRepo.GetByAPIName(ctx, input.APIName)
	if existing != nil {
		return nil, fmt.Errorf("groupService.Create: %w",
			apperror.Conflict(fmt.Sprintf("group with api_name '%s' already exists", input.APIName)))
	}

	var result *Group
	err := withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		created, err := s.groupRepo.Create(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("groupService.Create: %w", err)
		}
		result = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *groupServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*Group, error) {
	group, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("groupService.GetByID: %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("groupService.GetByID: %w",
			apperror.NotFound("Group", id.String()))
	}
	return group, nil
}

func (s *groupServiceImpl) List(ctx context.Context, page, perPage int32) ([]Group, int64, error) {
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * perPage

	groups, err := s.groupRepo.List(ctx, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("groupService.List: %w", err)
	}

	total, err := s.groupRepo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("groupService.List: count: %w", err)
	}

	return groups, total, nil
}

func (s *groupServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.groupRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("groupService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("groupService.Delete: %w",
			apperror.NotFound("Group", id.String()))
	}

	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.memberRepo.DeleteByGroupID(ctx, tx, id); err != nil {
			return fmt.Errorf("groupService.Delete: delete members: %w", err)
		}
		if err := s.groupRepo.Delete(ctx, tx, id); err != nil {
			return fmt.Errorf("groupService.Delete: %w", err)
		}
		return nil
	})
}

func (s *groupServiceImpl) AddMember(ctx context.Context, input AddGroupMemberInput) (*GroupMember, error) {
	if err := ValidateAddGroupMember(input); err != nil {
		return nil, fmt.Errorf("groupService.AddMember: %w", err)
	}

	group, err := s.groupRepo.GetByID(ctx, input.GroupID)
	if err != nil {
		return nil, fmt.Errorf("groupService.AddMember: %w", err)
	}
	if group == nil {
		return nil, fmt.Errorf("groupService.AddMember: %w",
			apperror.NotFound("Group", input.GroupID.String()))
	}

	var result *GroupMember
	err = withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		member, err := s.memberRepo.Add(ctx, tx, input)
		if err != nil {
			return fmt.Errorf("groupService.AddMember: %w", err)
		}

		if err := s.emitGroupChanged(ctx, tx, input.GroupID); err != nil {
			return err
		}

		result = member
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *groupServiceImpl) RemoveMember(ctx context.Context, groupID uuid.UUID, memberUserID *uuid.UUID, memberGroupID *uuid.UUID) error {
	return withTx(ctx, s.txBeginner, func(tx pgx.Tx) error {
		if err := s.memberRepo.Remove(ctx, tx, groupID, memberUserID, memberGroupID); err != nil {
			return fmt.Errorf("groupService.RemoveMember: %w", err)
		}

		if err := s.emitGroupChanged(ctx, tx, groupID); err != nil {
			return err
		}

		return nil
	})
}

func (s *groupServiceImpl) ListMembers(ctx context.Context, groupID uuid.UUID) ([]GroupMember, error) {
	members, err := s.memberRepo.ListByGroupID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("groupService.ListMembers: %w", err)
	}
	return members, nil
}

func (s *groupServiceImpl) emitGroupChanged(ctx context.Context, tx pgx.Tx, groupID uuid.UUID) error {
	return s.outboxRepo.Insert(ctx, tx, OutboxEvent{
		EventType:  "group_changed",
		EntityType: "group",
		EntityID:   groupID,
		Payload:    []byte("{}"),
	})
}
