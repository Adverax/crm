package security_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/security"
)

func TestUserService_Create(t *testing.T) {
	profileID := uuid.New()
	roleID := uuid.New()

	tests := []struct {
		name    string
		input   security.CreateUserInput
		setup   func(*mockUserRepo, *mockProfileRepo, *mockUserRoleRepo, *mockGroupRepo)
		wantErr bool
		errCode string
	}{
		{
			name: "creates user successfully without role",
			input: security.CreateUserInput{
				Username:  "johndoe",
				Email:     "john@example.com",
				FirstName: "John",
				LastName:  "Doe",
				ProfileID: profileID,
			},
			setup: func(_ *mockUserRepo, pr *mockProfileRepo, _ *mockUserRoleRepo, _ *mockGroupRepo) {
				pr.profiles[profileID] = &security.Profile{ID: profileID, APIName: "standard", Label: "Standard"}
			},
			wantErr: false,
		},
		{
			name: "creates user with role and adds to role groups",
			input: security.CreateUserInput{
				Username:  "janedoe",
				Email:     "jane@example.com",
				ProfileID: profileID,
				RoleID:    &roleID,
			},
			setup: func(_ *mockUserRepo, pr *mockProfileRepo, rr *mockUserRoleRepo, gr *mockGroupRepo) {
				pr.profiles[profileID] = &security.Profile{ID: profileID, APIName: "standard", Label: "Standard"}
				rr.roles = append(rr.roles, security.UserRole{ID: roleID, APIName: "sales", Label: "Sales"})
				gr.byRoleID[roleID.String()+string(security.GroupTypeRole)] = &security.Group{
					ID: uuid.New(), GroupType: security.GroupTypeRole, RelatedRoleID: &roleID,
				}
				gr.byRoleID[roleID.String()+string(security.GroupTypeRoleAndSubordinates)] = &security.Group{
					ID: uuid.New(), GroupType: security.GroupTypeRoleAndSubordinates, RelatedRoleID: &roleID,
				}
			},
			wantErr: false,
		},
		{
			name: "returns validation error for empty username",
			input: security.CreateUserInput{
				Username:  "",
				Email:     "john@example.com",
				ProfileID: profileID,
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns validation error for invalid email",
			input: security.CreateUserInput{
				Username:  "johndoe",
				Email:     "not-an-email",
				ProfileID: profileID,
			},
			wantErr: true,
			errCode: "VALIDATION",
		},
		{
			name: "returns conflict for duplicate username",
			input: security.CreateUserInput{
				Username:  "existing_user",
				Email:     "new@example.com",
				ProfileID: profileID,
			},
			setup: func(ur *mockUserRepo, pr *mockProfileRepo, _ *mockUserRoleRepo, _ *mockGroupRepo) {
				ur.users = append(ur.users, security.User{ID: uuid.New(), Username: "existing_user", Email: "old@example.com", ProfileID: profileID})
				pr.profiles[profileID] = &security.Profile{ID: profileID}
			},
			wantErr: true,
			errCode: "CONFLICT",
		},
		{
			name: "returns not found for non-existent profile",
			input: security.CreateUserInput{
				Username:  "johndoe",
				Email:     "john@example.com",
				ProfileID: uuid.New(),
			},
			wantErr: true,
			errCode: "NOT_FOUND",
		},
		{
			name: "returns not found for non-existent role",
			input: security.CreateUserInput{
				Username:  "johndoe",
				Email:     "john@example.com",
				ProfileID: profileID,
				RoleID:    ptrUUID(uuid.New()),
			},
			setup: func(_ *mockUserRepo, pr *mockProfileRepo, _ *mockUserRoleRepo, _ *mockGroupRepo) {
				pr.profiles[profileID] = &security.Profile{ID: profileID, APIName: "standard", Label: "Standard"}
			},
			wantErr: true,
			errCode: "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{}
			profileRepo := newMockProfileRepo()
			roleRepo := &mockUserRoleRepo{}
			psToUserRepo := &mockPSToUserRepo{}
			outboxRepo := &mockOutboxRepo{}
			groupRepo := newMockGroupRepo()
			memberRepo := newMockGroupMemberRepo()

			if tt.setup != nil {
				tt.setup(userRepo, profileRepo, roleRepo, groupRepo)
			}

			svc := security.NewUserService(
				&mockTxBeginner{}, userRepo, profileRepo, roleRepo,
				psToUserRepo, outboxRepo, groupRepo, memberRepo,
			)
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
			if result.Username != tt.input.Username {
				t.Errorf("expected username %s, got %s", tt.input.Username, result.Username)
			}
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockUserRepo)
		wantErr bool
	}{
		{
			name: "returns user when exists",
			id:   existingID,
			setup: func(ur *mockUserRepo) {
				ur.users = append(ur.users, security.User{ID: existingID, Username: "john"})
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent user",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{}
			if tt.setup != nil {
				tt.setup(userRepo)
			}

			svc := security.NewUserService(
				&mockTxBeginner{}, userRepo, newMockProfileRepo(), &mockUserRoleRepo{},
				&mockPSToUserRepo{}, &mockOutboxRepo{}, newMockGroupRepo(), newMockGroupMemberRepo(),
			)
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

func TestUserService_List(t *testing.T) {
	tests := []struct {
		name      string
		page      int32
		perPage   int32
		setup     func(*mockUserRepo)
		wantCount int
	}{
		{
			name:      "returns empty list",
			page:      1,
			perPage:   20,
			wantCount: 0,
		},
		{
			name:    "returns users",
			page:    1,
			perPage: 10,
			setup: func(ur *mockUserRepo) {
				ur.users = append(ur.users, security.User{ID: uuid.New(), Username: "john"})
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{}
			if tt.setup != nil {
				tt.setup(userRepo)
			}

			svc := security.NewUserService(
				&mockTxBeginner{}, userRepo, newMockProfileRepo(), &mockUserRoleRepo{},
				&mockPSToUserRepo{}, &mockOutboxRepo{}, newMockGroupRepo(), newMockGroupMemberRepo(),
			)
			users, _, err := svc.List(context.Background(), tt.page, tt.perPage)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(users) != tt.wantCount {
				t.Errorf("expected %d users, got %d", tt.wantCount, len(users))
			}
		})
	}
}

func TestUserService_Update(t *testing.T) {
	existingID := uuid.New()
	profileID := uuid.New()
	newProfileID := uuid.New()

	tests := []struct {
		name           string
		id             uuid.UUID
		input          security.UpdateUserInput
		setup          func(*mockUserRepo)
		wantErr        bool
		wantOutboxSize int
	}{
		{
			name: "updates user successfully (profile changed)",
			id:   existingID,
			input: security.UpdateUserInput{
				Email:     "new@example.com",
				ProfileID: newProfileID,
			},
			setup: func(ur *mockUserRepo) {
				ur.users = append(ur.users, security.User{ID: existingID, Username: "john", Email: "old@example.com", ProfileID: profileID})
			},
			wantErr:        false,
			wantOutboxSize: 1,
		},
		{
			name: "updates user without outbox when profile/role unchanged",
			id:   existingID,
			input: security.UpdateUserInput{
				Email:     "new@example.com",
				ProfileID: profileID,
			},
			setup: func(ur *mockUserRepo) {
				ur.users = append(ur.users, security.User{ID: existingID, Username: "john", Email: "old@example.com", ProfileID: profileID})
			},
			wantErr:        false,
			wantOutboxSize: 0,
		},
		{
			name: "returns not found for non-existent user",
			id:   uuid.New(),
			input: security.UpdateUserInput{
				Email:     "new@example.com",
				ProfileID: profileID,
			},
			wantErr: true,
		},
		{
			name: "returns validation error for empty email",
			id:   existingID,
			input: security.UpdateUserInput{
				Email:     "",
				ProfileID: profileID,
			},
			setup: func(ur *mockUserRepo) {
				ur.users = append(ur.users, security.User{ID: existingID, Username: "john", Email: "old@example.com", ProfileID: profileID})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{}
			outboxRepo := &mockOutboxRepo{}
			if tt.setup != nil {
				tt.setup(userRepo)
			}

			svc := security.NewUserService(
				&mockTxBeginner{}, userRepo, newMockProfileRepo(), &mockUserRoleRepo{},
				&mockPSToUserRepo{}, outboxRepo, newMockGroupRepo(), newMockGroupMemberRepo(),
			)
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
			if result.Email != tt.input.Email {
				t.Errorf("expected email %s, got %s", tt.input.Email, result.Email)
			}
			if len(outboxRepo.events) != tt.wantOutboxSize {
				t.Errorf("expected %d outbox events, got %d", tt.wantOutboxSize, len(outboxRepo.events))
			}
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		setup   func(*mockUserRepo)
		wantErr bool
	}{
		{
			name: "deletes user successfully",
			id:   existingID,
			setup: func(ur *mockUserRepo) {
				ur.users = append(ur.users, security.User{ID: existingID, Username: "john"})
			},
			wantErr: false,
		},
		{
			name:    "returns not found for non-existent user",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{}
			if tt.setup != nil {
				tt.setup(userRepo)
			}

			svc := security.NewUserService(
				&mockTxBeginner{}, userRepo, newMockProfileRepo(), &mockUserRoleRepo{},
				&mockPSToUserRepo{}, &mockOutboxRepo{}, newMockGroupRepo(), newMockGroupMemberRepo(),
			)
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

func TestUserService_UpdateWithRoleChange(t *testing.T) {
	existingID := uuid.New()
	profileID := uuid.New()
	oldRoleID := uuid.New()
	newRoleID := uuid.New()

	tests := []struct {
		name       string
		input      security.UpdateUserInput
		setup      func(*mockUserRepo, *mockGroupRepo, *mockUserRoleRepo)
		wantOutbox bool
	}{
		{
			name: "role change triggers group recomputation",
			input: security.UpdateUserInput{
				Email:     "john@example.com",
				ProfileID: profileID,
				RoleID:    &newRoleID,
			},
			setup: func(ur *mockUserRepo, gr *mockGroupRepo, rr *mockUserRoleRepo) {
				ur.users = append(ur.users, security.User{
					ID: existingID, Username: "john", Email: "old@example.com",
					ProfileID: profileID, RoleID: &oldRoleID,
				})
				// Old role groups
				oldRoleGroup := &security.Group{ID: uuid.New(), GroupType: security.GroupTypeRole, RelatedRoleID: &oldRoleID}
				oldRoleSubGroup := &security.Group{ID: uuid.New(), GroupType: security.GroupTypeRoleAndSubordinates, RelatedRoleID: &oldRoleID}
				gr.byRoleID[oldRoleID.String()+string(security.GroupTypeRole)] = oldRoleGroup
				gr.byRoleID[oldRoleID.String()+string(security.GroupTypeRoleAndSubordinates)] = oldRoleSubGroup
				gr.groups[oldRoleGroup.ID] = oldRoleGroup
				gr.groups[oldRoleSubGroup.ID] = oldRoleSubGroup
				// New role groups
				newRoleGroup := &security.Group{ID: uuid.New(), GroupType: security.GroupTypeRole, RelatedRoleID: &newRoleID}
				newRoleSubGroup := &security.Group{ID: uuid.New(), GroupType: security.GroupTypeRoleAndSubordinates, RelatedRoleID: &newRoleID}
				gr.byRoleID[newRoleID.String()+string(security.GroupTypeRole)] = newRoleGroup
				gr.byRoleID[newRoleID.String()+string(security.GroupTypeRoleAndSubordinates)] = newRoleSubGroup
				gr.groups[newRoleGroup.ID] = newRoleGroup
				gr.groups[newRoleSubGroup.ID] = newRoleSubGroup
			},
			wantOutbox: true,
		},
		{
			name: "role removed (set to nil) triggers group removal",
			input: security.UpdateUserInput{
				Email:     "john@example.com",
				ProfileID: profileID,
				RoleID:    nil,
			},
			setup: func(ur *mockUserRepo, gr *mockGroupRepo, _ *mockUserRoleRepo) {
				ur.users = append(ur.users, security.User{
					ID: existingID, Username: "john", Email: "old@example.com",
					ProfileID: profileID, RoleID: &oldRoleID,
				})
				oldRoleGroup := &security.Group{ID: uuid.New(), GroupType: security.GroupTypeRole, RelatedRoleID: &oldRoleID}
				oldRoleSubGroup := &security.Group{ID: uuid.New(), GroupType: security.GroupTypeRoleAndSubordinates, RelatedRoleID: &oldRoleID}
				gr.byRoleID[oldRoleID.String()+string(security.GroupTypeRole)] = oldRoleGroup
				gr.byRoleID[oldRoleID.String()+string(security.GroupTypeRoleAndSubordinates)] = oldRoleSubGroup
				gr.groups[oldRoleGroup.ID] = oldRoleGroup
				gr.groups[oldRoleSubGroup.ID] = oldRoleSubGroup
			},
			wantOutbox: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{}
			groupRepo := newMockGroupRepo()
			memberRepo := newMockGroupMemberRepo()
			roleRepo := &mockUserRoleRepo{}
			outboxRepo := &mockOutboxRepo{}

			if tt.setup != nil {
				tt.setup(userRepo, groupRepo, roleRepo)
			}

			svc := security.NewUserService(
				&mockTxBeginner{}, userRepo, newMockProfileRepo(), roleRepo,
				&mockPSToUserRepo{}, outboxRepo, groupRepo, memberRepo,
			)
			result, err := svc.Update(context.Background(), existingID, tt.input)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if tt.wantOutbox && len(outboxRepo.events) == 0 {
				t.Error("expected outbox event for role change")
			}
		})
	}
}

func TestUserService_AssignPermissionSet(t *testing.T) {
	userID := uuid.New()
	psID := uuid.New()

	svc := security.NewUserService(
		&mockTxBeginner{}, &mockUserRepo{}, newMockProfileRepo(), &mockUserRoleRepo{},
		&mockPSToUserRepo{}, &mockOutboxRepo{}, newMockGroupRepo(), newMockGroupMemberRepo(),
	)

	err := svc.AssignPermissionSet(context.Background(), userID, psID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_RevokePermissionSet(t *testing.T) {
	userID := uuid.New()
	psID := uuid.New()

	svc := security.NewUserService(
		&mockTxBeginner{}, &mockUserRepo{}, newMockProfileRepo(), &mockUserRoleRepo{},
		&mockPSToUserRepo{}, &mockOutboxRepo{}, newMockGroupRepo(), newMockGroupMemberRepo(),
	)

	err := svc.RevokePermissionSet(context.Background(), userID, psID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserService_ListPermissionSets(t *testing.T) {
	userID := uuid.New()
	psToUserRepo := &mockPSToUserRepo{
		assignments: []security.PermissionSetToUser{
			{ID: uuid.New(), PermissionSetID: uuid.New(), UserID: userID},
		},
	}

	svc := security.NewUserService(
		&mockTxBeginner{}, &mockUserRepo{}, newMockProfileRepo(), &mockUserRoleRepo{},
		psToUserRepo, &mockOutboxRepo{}, newMockGroupRepo(), newMockGroupMemberRepo(),
	)

	result, err := svc.ListPermissionSets(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 assignment, got %d", len(result))
	}
}
