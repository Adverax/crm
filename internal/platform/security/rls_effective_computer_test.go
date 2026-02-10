package security_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/platform/security"
)

// Mock RLS effective cache repo
type mockRLSCacheRepo struct {
	roleHierarchy   []security.EffectiveRoleHierarchy
	visibleOwners   map[uuid.UUID][]security.EffectiveVisibleOwner
	groupMembers    map[uuid.UUID][]security.EffectiveGroupMember
	objectHierarchy []security.EffectiveObjectHierarchy
}

func newMockRLSCacheRepo() *mockRLSCacheRepo {
	return &mockRLSCacheRepo{
		visibleOwners: make(map[uuid.UUID][]security.EffectiveVisibleOwner),
		groupMembers:  make(map[uuid.UUID][]security.EffectiveGroupMember),
	}
}

func (r *mockRLSCacheRepo) ReplaceRoleHierarchy(_ context.Context, _ pgx.Tx, entries []security.EffectiveRoleHierarchy) error {
	r.roleHierarchy = entries
	return nil
}

func (r *mockRLSCacheRepo) ReplaceVisibleOwners(_ context.Context, _ pgx.Tx, userID uuid.UUID, entries []security.EffectiveVisibleOwner) error {
	r.visibleOwners[userID] = entries
	return nil
}

func (r *mockRLSCacheRepo) ReplaceVisibleOwnersAll(_ context.Context, _ pgx.Tx, entries []security.EffectiveVisibleOwner) error {
	r.visibleOwners = make(map[uuid.UUID][]security.EffectiveVisibleOwner)
	for _, e := range entries {
		r.visibleOwners[e.UserID] = append(r.visibleOwners[e.UserID], e)
	}
	return nil
}

func (r *mockRLSCacheRepo) ReplaceGroupMembers(_ context.Context, _ pgx.Tx, groupID uuid.UUID, entries []security.EffectiveGroupMember) error {
	r.groupMembers[groupID] = entries
	return nil
}

func (r *mockRLSCacheRepo) ReplaceGroupMembersAll(_ context.Context, _ pgx.Tx, entries []security.EffectiveGroupMember) error {
	r.groupMembers = make(map[uuid.UUID][]security.EffectiveGroupMember)
	for _, e := range entries {
		r.groupMembers[e.GroupID] = append(r.groupMembers[e.GroupID], e)
	}
	return nil
}

func (r *mockRLSCacheRepo) ReplaceObjectHierarchy(_ context.Context, _ pgx.Tx, entries []security.EffectiveObjectHierarchy) error {
	r.objectHierarchy = entries
	return nil
}

func (r *mockRLSCacheRepo) GetVisibleOwners(_ context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for _, e := range r.visibleOwners[userID] {
		ids = append(ids, e.VisibleOwnerID)
	}
	return ids, nil
}

func (r *mockRLSCacheRepo) GetGroupMemberships(_ context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for gid, members := range r.groupMembers {
		for _, m := range members {
			if m.UserID == userID {
				ids = append(ids, gid)
			}
		}
	}
	return ids, nil
}

func (r *mockRLSCacheRepo) GetRoleDescendants(_ context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for _, e := range r.roleHierarchy {
		if e.AncestorRoleID == roleID && e.Depth > 0 {
			ids = append(ids, e.DescendantRoleID)
		}
	}
	return ids, nil
}

func (r *mockRLSCacheRepo) ListAllRoles(_ context.Context) ([]security.EffectiveRoleHierarchy, error) {
	return r.roleHierarchy, nil
}

// Mock user role repo for RLS tests
type mockUserRoleRepo struct {
	roles []security.UserRole
}

func (r *mockUserRoleRepo) Create(_ context.Context, _ pgx.Tx, input security.CreateUserRoleInput) (*security.UserRole, error) {
	role := &security.UserRole{ID: uuid.New(), APIName: input.APIName, Label: input.Label, ParentID: input.ParentID}
	r.roles = append(r.roles, *role)
	return role, nil
}

func (r *mockUserRoleRepo) GetByID(_ context.Context, id uuid.UUID) (*security.UserRole, error) {
	for _, role := range r.roles {
		if role.ID == id {
			return &role, nil
		}
	}
	return nil, nil
}

func (r *mockUserRoleRepo) GetByAPIName(_ context.Context, apiName string) (*security.UserRole, error) {
	for _, role := range r.roles {
		if role.APIName == apiName {
			return &role, nil
		}
	}
	return nil, nil
}

func (r *mockUserRoleRepo) List(_ context.Context, _, _ int32) ([]security.UserRole, error) {
	return r.roles, nil
}

func (r *mockUserRoleRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input security.UpdateUserRoleInput) (*security.UserRole, error) {
	for i, role := range r.roles {
		if role.ID == id {
			r.roles[i].Label = input.Label
			r.roles[i].Description = input.Description
			r.roles[i].ParentID = input.ParentID
			return &r.roles[i], nil
		}
	}
	return nil, nil
}

func (r *mockUserRoleRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	for i, role := range r.roles {
		if role.ID == id {
			r.roles = append(r.roles[:i], r.roles[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *mockUserRoleRepo) Count(_ context.Context) (int64, error) {
	return int64(len(r.roles)), nil
}

// Mock user repo for RLS tests
type mockUserRepo struct {
	users []security.User
}

func (r *mockUserRepo) Create(_ context.Context, _ pgx.Tx, input security.CreateUserInput) (*security.User, error) {
	u := security.User{
		ID:        uuid.New(),
		Username:  input.Username,
		Email:     input.Email,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		ProfileID: input.ProfileID,
		RoleID:    input.RoleID,
		IsActive:  true,
	}
	r.users = append(r.users, u)
	return &u, nil
}

func (r *mockUserRepo) GetByID(_ context.Context, id uuid.UUID) (*security.User, error) {
	for i, u := range r.users {
		if u.ID == id {
			return &r.users[i], nil
		}
	}
	return nil, nil
}

func (r *mockUserRepo) GetByUsername(_ context.Context, username string) (*security.User, error) {
	for i, u := range r.users {
		if u.Username == username {
			return &r.users[i], nil
		}
	}
	return nil, nil
}

func (r *mockUserRepo) List(_ context.Context, _, _ int32) ([]security.User, error) {
	return r.users, nil
}

func (r *mockUserRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input security.UpdateUserInput) (*security.User, error) {
	for i, u := range r.users {
		if u.ID == id {
			r.users[i].Email = input.Email
			r.users[i].FirstName = input.FirstName
			r.users[i].LastName = input.LastName
			r.users[i].ProfileID = input.ProfileID
			r.users[i].RoleID = input.RoleID
			r.users[i].IsActive = input.IsActive
			return &r.users[i], nil
		}
	}
	return nil, nil
}

func (r *mockUserRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	for i, u := range r.users {
		if u.ID == id {
			r.users = append(r.users[:i], r.users[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *mockUserRepo) Count(_ context.Context) (int64, error) {
	return int64(len(r.users)), nil
}

// Mock metadata adapter
type mockMetadataRLSAdapter struct {
	compositionFields []security.CompositionFieldInfo
}

func (m *mockMetadataRLSAdapter) GetObjectVisibility(_ context.Context, _ uuid.UUID) (string, error) {
	return "private", nil
}

func (m *mockMetadataRLSAdapter) GetObjectTableName(_ context.Context, _ uuid.UUID) (string, error) {
	return "obj_test", nil
}

func (m *mockMetadataRLSAdapter) ListCompositionFields(_ context.Context) ([]security.CompositionFieldInfo, error) {
	return m.compositionFields, nil
}

func TestRLSEffectiveComputer_RecomputeRoleHierarchy(t *testing.T) {
	ceoID := uuid.New()
	vpID := uuid.New()
	mgrID := uuid.New()

	tests := []struct {
		name           string
		roles          []security.UserRole
		wantEntryCount int
	}{
		{
			name:           "empty roles produces no entries",
			roles:          nil,
			wantEntryCount: 0,
		},
		{
			name: "single role produces self-entry",
			roles: []security.UserRole{
				{ID: ceoID, APIName: "ceo", Label: "CEO"},
			},
			wantEntryCount: 1,
		},
		{
			name: "linear hierarchy CEO->VP->Manager produces 6 entries",
			roles: []security.UserRole{
				{ID: ceoID, APIName: "ceo", Label: "CEO"},
				{ID: vpID, APIName: "vp", Label: "VP", ParentID: &ceoID},
				{ID: mgrID, APIName: "mgr", Label: "Manager", ParentID: &vpID},
			},
			// CEO: self(0)
			// VP: self(0), CEO->VP(1)
			// MGR: self(0), VP->MGR(1), CEO->MGR(2)
			wantEntryCount: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := &mockUserRoleRepo{roles: tt.roles}
			userRepo := &mockUserRepo{}
			groupRepo := newMockGroupRepo()
			memberRepo := newMockGroupMemberRepo()
			rlsCache := newMockRLSCacheRepo()
			metadataAdapter := &mockMetadataRLSAdapter{}

			computer := security.NewRLSEffectiveComputer(
				&mockTxBeginner{}, roleRepo, userRepo, groupRepo, memberRepo, rlsCache, metadataAdapter,
			)

			err := computer.RecomputeRoleHierarchy(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(rlsCache.roleHierarchy) != tt.wantEntryCount {
				t.Errorf("expected %d entries, got %d", tt.wantEntryCount, len(rlsCache.roleHierarchy))
			}
		})
	}
}

func TestRLSEffectiveComputer_RecomputeGroupMembers(t *testing.T) {
	groupID := uuid.New()
	user1 := uuid.New()
	user2 := uuid.New()

	tests := []struct {
		name          string
		setupMembers  func(*mockGroupMemberRepo)
		wantUserCount int
	}{
		{
			name:          "empty group has no members",
			setupMembers:  func(_ *mockGroupMemberRepo) {},
			wantUserCount: 0,
		},
		{
			name: "group with direct user members",
			setupMembers: func(r *mockGroupMemberRepo) {
				r.members[groupID] = []security.GroupMember{
					{GroupID: groupID, MemberUserID: &user1},
					{GroupID: groupID, MemberUserID: &user2},
				}
			},
			wantUserCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := &mockUserRoleRepo{}
			userRepo := &mockUserRepo{}
			groupRepo := newMockGroupRepo()
			memberRepo := newMockGroupMemberRepo()
			rlsCache := newMockRLSCacheRepo()
			metadataAdapter := &mockMetadataRLSAdapter{}

			tt.setupMembers(memberRepo)

			computer := security.NewRLSEffectiveComputer(
				&mockTxBeginner{}, roleRepo, userRepo, groupRepo, memberRepo, rlsCache, metadataAdapter,
			)

			err := computer.RecomputeGroupMembersForGroup(context.Background(), groupID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(rlsCache.groupMembers[groupID]) != tt.wantUserCount {
				t.Errorf("expected %d members, got %d", tt.wantUserCount, len(rlsCache.groupMembers[groupID]))
			}
		})
	}
}

func TestRLSEffectiveComputer_RecomputeObjectHierarchy(t *testing.T) {
	parentObj := uuid.New()
	childObj := uuid.New()

	tests := []struct {
		name           string
		fields         []security.CompositionFieldInfo
		wantEntryCount int
	}{
		{
			name:           "no composition fields",
			fields:         nil,
			wantEntryCount: 0,
		},
		{
			name: "single parent-child produces 3 entries",
			fields: []security.CompositionFieldInfo{
				{ChildObjectID: childObj, ParentObjectID: parentObj},
			},
			// parent: self(0)
			// child: self(0), parent->child(1)
			wantEntryCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleRepo := &mockUserRoleRepo{}
			userRepo := &mockUserRepo{}
			groupRepo := newMockGroupRepo()
			memberRepo := newMockGroupMemberRepo()
			rlsCache := newMockRLSCacheRepo()
			metadataAdapter := &mockMetadataRLSAdapter{compositionFields: tt.fields}

			computer := security.NewRLSEffectiveComputer(
				&mockTxBeginner{}, roleRepo, userRepo, groupRepo, memberRepo, rlsCache, metadataAdapter,
			)

			err := computer.RecomputeObjectHierarchy(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(rlsCache.objectHierarchy) != tt.wantEntryCount {
				t.Errorf("expected %d entries, got %d", tt.wantEntryCount, len(rlsCache.objectHierarchy))
			}
		})
	}
}
