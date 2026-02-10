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

// Mock implementations

type mockGroupRepo struct {
	groups   map[uuid.UUID]*security.Group
	byName   map[string]*security.Group
	byRoleID map[string]*security.Group // key: roleID+groupType
	byUserID map[uuid.UUID]*security.Group
}

func newMockGroupRepo() *mockGroupRepo {
	return &mockGroupRepo{
		groups:   make(map[uuid.UUID]*security.Group),
		byName:   make(map[string]*security.Group),
		byRoleID: make(map[string]*security.Group),
		byUserID: make(map[uuid.UUID]*security.Group),
	}
}

func (r *mockGroupRepo) Create(_ context.Context, _ pgx.Tx, input security.CreateGroupInput) (*security.Group, error) {
	g := &security.Group{
		ID:            uuid.New(),
		APIName:       input.APIName,
		Label:         input.Label,
		GroupType:     input.GroupType,
		RelatedRoleID: input.RelatedRoleID,
		RelatedUserID: input.RelatedUserID,
	}
	r.groups[g.ID] = g
	r.byName[g.APIName] = g
	if input.RelatedRoleID != nil {
		r.byRoleID[input.RelatedRoleID.String()+string(input.GroupType)] = g
	}
	if input.RelatedUserID != nil {
		r.byUserID[*input.RelatedUserID] = g
	}
	return g, nil
}

func (r *mockGroupRepo) GetByID(_ context.Context, id uuid.UUID) (*security.Group, error) {
	return r.groups[id], nil
}

func (r *mockGroupRepo) GetByAPIName(_ context.Context, apiName string) (*security.Group, error) {
	return r.byName[apiName], nil
}

func (r *mockGroupRepo) GetByRelatedRoleID(_ context.Context, roleID uuid.UUID, groupType security.GroupType) (*security.Group, error) {
	return r.byRoleID[roleID.String()+string(groupType)], nil
}

func (r *mockGroupRepo) GetByRelatedUserID(_ context.Context, userID uuid.UUID) (*security.Group, error) {
	return r.byUserID[userID], nil
}

func (r *mockGroupRepo) List(_ context.Context, _, _ int32) ([]security.Group, error) {
	result := make([]security.Group, 0, len(r.groups))
	for _, g := range r.groups {
		result = append(result, *g)
	}
	return result, nil
}

func (r *mockGroupRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	delete(r.groups, id)
	return nil
}

func (r *mockGroupRepo) Count(_ context.Context) (int64, error) {
	return int64(len(r.groups)), nil
}

type mockGroupMemberRepo struct {
	members map[uuid.UUID][]security.GroupMember
}

func newMockGroupMemberRepo() *mockGroupMemberRepo {
	return &mockGroupMemberRepo{
		members: make(map[uuid.UUID][]security.GroupMember),
	}
}

func (r *mockGroupMemberRepo) Add(_ context.Context, _ pgx.Tx, input security.AddGroupMemberInput) (*security.GroupMember, error) {
	m := security.GroupMember{
		ID:            uuid.New(),
		GroupID:       input.GroupID,
		MemberUserID:  input.MemberUserID,
		MemberGroupID: input.MemberGroupID,
	}
	r.members[input.GroupID] = append(r.members[input.GroupID], m)
	return &m, nil
}

func (r *mockGroupMemberRepo) Remove(_ context.Context, _ pgx.Tx, groupID uuid.UUID, memberUserID *uuid.UUID, _ *uuid.UUID) error {
	members := r.members[groupID]
	for i, m := range members {
		if memberUserID != nil && m.MemberUserID != nil && *m.MemberUserID == *memberUserID {
			r.members[groupID] = append(members[:i], members[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *mockGroupMemberRepo) ListByGroupID(_ context.Context, groupID uuid.UUID) ([]security.GroupMember, error) {
	return r.members[groupID], nil
}

func (r *mockGroupMemberRepo) ListByUserID(_ context.Context, userID uuid.UUID) ([]security.GroupMember, error) {
	var result []security.GroupMember
	for _, members := range r.members {
		for _, m := range members {
			if m.MemberUserID != nil && *m.MemberUserID == userID {
				result = append(result, m)
			}
		}
	}
	return result, nil
}

func (r *mockGroupMemberRepo) DeleteByGroupID(_ context.Context, _ pgx.Tx, groupID uuid.UUID) error {
	delete(r.members, groupID)
	return nil
}

type mockOutboxRepo struct {
	events []security.OutboxEvent
}

func (r *mockOutboxRepo) Insert(_ context.Context, _ pgx.Tx, event security.OutboxEvent) error {
	r.events = append(r.events, event)
	return nil
}

func (r *mockOutboxRepo) ListUnprocessed(_ context.Context, _ int) ([]security.OutboxEvent, error) {
	return r.events, nil
}

func (r *mockOutboxRepo) MarkProcessed(_ context.Context, _ int64) error {
	return nil
}

type mockTxBeginner struct{}

func (m *mockTxBeginner) Begin(ctx context.Context) (pgx.Tx, error) {
	return &mockTx{}, nil
}

type mockTx struct{ pgx.Tx }

func (t *mockTx) Commit(_ context.Context) error   { return nil }
func (t *mockTx) Rollback(_ context.Context) error { return nil }

func TestGroupService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   security.CreateGroupInput
		setup   func(*mockGroupRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "creates group successfully",
			input: security.CreateGroupInput{
				APIName:   "test_group",
				Label:     "Test Group",
				GroupType: security.GroupTypePublic,
			},
			wantErr: false,
		},
		{
			name: "returns error when api_name already exists",
			input: security.CreateGroupInput{
				APIName:   "existing_group",
				Label:     "Existing",
				GroupType: security.GroupTypePublic,
			},
			setup: func(r *mockGroupRepo) {
				r.byName["existing_group"] = &security.Group{ID: uuid.New(), APIName: "existing_group"}
			},
			wantErr: true,
			errCode: "CONFLICT",
		},
		{
			name: "returns error when label is empty",
			input: security.CreateGroupInput{
				APIName:   "no_label",
				Label:     "",
				GroupType: security.GroupTypePublic,
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns error when group_type is invalid",
			input: security.CreateGroupInput{
				APIName:   "bad_type",
				Label:     "Bad Type",
				GroupType: "invalid",
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupRepo := newMockGroupRepo()
			memberRepo := newMockGroupMemberRepo()
			outboxRepo := &mockOutboxRepo{}
			if tt.setup != nil {
				tt.setup(groupRepo)
			}

			svc := security.NewGroupService(&mockTxBeginner{}, groupRepo, memberRepo, outboxRepo)
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
			if result.APIName != tt.input.APIName {
				t.Errorf("expected api_name %s, got %s", tt.input.APIName, result.APIName)
			}
		})
	}
}

func TestGroupService_AddMember(t *testing.T) {
	tests := []struct {
		name    string
		input   security.AddGroupMemberInput
		setup   func(*mockGroupRepo)
		wantErr bool
	}{
		{
			name: "adds user member successfully",
			input: security.AddGroupMemberInput{
				GroupID:      uuid.New(), // will be set in setup
				MemberUserID: ptrUUID(uuid.New()),
			},
			wantErr: false,
		},
		{
			name: "returns error when both user and group set",
			input: security.AddGroupMemberInput{
				GroupID:       uuid.New(),
				MemberUserID:  ptrUUID(uuid.New()),
				MemberGroupID: ptrUUID(uuid.New()),
			},
			wantErr: true,
		},
		{
			name: "returns error when neither user nor group set",
			input: security.AddGroupMemberInput{
				GroupID: uuid.New(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupRepo := newMockGroupRepo()
			memberRepo := newMockGroupMemberRepo()
			outboxRepo := &mockOutboxRepo{}

			// Ensure group exists for the valid test case
			if !tt.wantErr {
				groupID := tt.input.GroupID
				groupRepo.groups[groupID] = &security.Group{
					ID:        groupID,
					APIName:   "test",
					Label:     "Test",
					GroupType: security.GroupTypePublic,
				}
			}

			svc := security.NewGroupService(&mockTxBeginner{}, groupRepo, memberRepo, outboxRepo)
			_, err := svc.AddMember(context.Background(), tt.input)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGroupService_GetByID(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockGroupRepo)
		wantErr bool
	}{
		{
			name: "returns group when exists",
			id:   existingID,
			setup: func(r *mockGroupRepo) {
				r.groups[existingID] = &security.Group{ID: existingID, APIName: "test_group"}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent group",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupRepo := newMockGroupRepo()
			if tt.setup != nil {
				tt.setup(groupRepo)
			}

			svc := security.NewGroupService(&mockTxBeginner{}, groupRepo, newMockGroupMemberRepo(), &mockOutboxRepo{})
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

func TestGroupService_List(t *testing.T) {
	groupRepo := newMockGroupRepo()
	id := uuid.New()
	groupRepo.groups[id] = &security.Group{ID: id, APIName: "test"}

	svc := security.NewGroupService(&mockTxBeginner{}, groupRepo, newMockGroupMemberRepo(), &mockOutboxRepo{})
	groups, total, err := svc.List(context.Background(), 1, 20)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}

func TestGroupService_Delete(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockGroupRepo)
		wantErr bool
	}{
		{
			name: "deletes group successfully",
			id:   existingID,
			setup: func(r *mockGroupRepo) {
				r.groups[existingID] = &security.Group{ID: existingID, APIName: "to_delete"}
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent group",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groupRepo := newMockGroupRepo()
			if tt.setup != nil {
				tt.setup(groupRepo)
			}

			svc := security.NewGroupService(&mockTxBeginner{}, groupRepo, newMockGroupMemberRepo(), &mockOutboxRepo{})
			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGroupService_RemoveMember(t *testing.T) {
	groupID := uuid.New()
	userID := uuid.New()

	svc := security.NewGroupService(&mockTxBeginner{}, newMockGroupRepo(), newMockGroupMemberRepo(), &mockOutboxRepo{})
	err := svc.RemoveMember(context.Background(), groupID, &userID, nil)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGroupService_ListMembers(t *testing.T) {
	groupID := uuid.New()
	userID := uuid.New()

	memberRepo := newMockGroupMemberRepo()
	memberRepo.members[groupID] = []security.GroupMember{
		{ID: uuid.New(), GroupID: groupID, MemberUserID: &userID},
	}

	svc := security.NewGroupService(&mockTxBeginner{}, newMockGroupRepo(), memberRepo, &mockOutboxRepo{})
	members, err := svc.ListMembers(context.Background(), groupID)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("expected 1 member, got %d", len(members))
	}
}

func ptrUUID(id uuid.UUID) *uuid.UUID {
	return &id
}
