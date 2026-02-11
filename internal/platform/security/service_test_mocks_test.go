package security_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/adverax/crm/internal/platform/security"
)

// mockPSRepo implements security.PermissionSetRepository.
type mockPSRepo struct {
	sets   map[uuid.UUID]*security.PermissionSet
	byName map[string]*security.PermissionSet
}

func newMockPSRepo() *mockPSRepo {
	return &mockPSRepo{
		sets:   make(map[uuid.UUID]*security.PermissionSet),
		byName: make(map[string]*security.PermissionSet),
	}
}

func (r *mockPSRepo) Create(_ context.Context, _ pgx.Tx, input security.CreatePermissionSetInput) (*security.PermissionSet, error) {
	ps := &security.PermissionSet{
		ID:          uuid.New(),
		APIName:     input.APIName,
		Label:       input.Label,
		Description: input.Description,
		PSType:      input.PSType,
	}
	r.sets[ps.ID] = ps
	r.byName[ps.APIName] = ps
	return ps, nil
}

func (r *mockPSRepo) GetByID(_ context.Context, id uuid.UUID) (*security.PermissionSet, error) {
	return r.sets[id], nil
}

func (r *mockPSRepo) GetByAPIName(_ context.Context, apiName string) (*security.PermissionSet, error) {
	return r.byName[apiName], nil
}

func (r *mockPSRepo) List(_ context.Context, _, _ int32) ([]security.PermissionSet, error) {
	result := make([]security.PermissionSet, 0, len(r.sets))
	for _, ps := range r.sets {
		result = append(result, *ps)
	}
	return result, nil
}

func (r *mockPSRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input security.UpdatePermissionSetInput) (*security.PermissionSet, error) {
	ps := r.sets[id]
	if ps == nil {
		return nil, nil
	}
	ps.Label = input.Label
	ps.Description = input.Description
	return ps, nil
}

func (r *mockPSRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	delete(r.sets, id)
	return nil
}

func (r *mockPSRepo) Count(_ context.Context) (int64, error) {
	return int64(len(r.sets)), nil
}

// mockProfileRepo implements security.ProfileRepository.
type mockProfileRepo struct {
	profiles map[uuid.UUID]*security.Profile
	byName   map[string]*security.Profile
}

func newMockProfileRepo() *mockProfileRepo {
	return &mockProfileRepo{
		profiles: make(map[uuid.UUID]*security.Profile),
		byName:   make(map[string]*security.Profile),
	}
}

func (r *mockProfileRepo) Create(_ context.Context, _ pgx.Tx, profile *security.Profile) (*security.Profile, error) {
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}
	r.profiles[profile.ID] = profile
	r.byName[profile.APIName] = profile
	return profile, nil
}

func (r *mockProfileRepo) GetByID(_ context.Context, id uuid.UUID) (*security.Profile, error) {
	return r.profiles[id], nil
}

func (r *mockProfileRepo) GetByAPIName(_ context.Context, apiName string) (*security.Profile, error) {
	return r.byName[apiName], nil
}

func (r *mockProfileRepo) List(_ context.Context, _, _ int32) ([]security.Profile, error) {
	result := make([]security.Profile, 0, len(r.profiles))
	for _, p := range r.profiles {
		result = append(result, *p)
	}
	return result, nil
}

func (r *mockProfileRepo) Update(_ context.Context, _ pgx.Tx, id uuid.UUID, input security.UpdateProfileInput) (*security.Profile, error) {
	p := r.profiles[id]
	if p == nil {
		return nil, nil
	}
	p.Label = input.Label
	p.Description = input.Description
	return p, nil
}

func (r *mockProfileRepo) Delete(_ context.Context, _ pgx.Tx, id uuid.UUID) error {
	delete(r.profiles, id)
	return nil
}

func (r *mockProfileRepo) Count(_ context.Context) (int64, error) {
	return int64(len(r.profiles)), nil
}

// mockPSToUserRepo implements security.PermissionSetToUserRepository.
type mockPSToUserRepo struct {
	assignments []security.PermissionSetToUser
}

func (r *mockPSToUserRepo) Assign(_ context.Context, _ pgx.Tx, psID, userID uuid.UUID) (*security.PermissionSetToUser, error) {
	a := security.PermissionSetToUser{
		ID:              uuid.New(),
		PermissionSetID: psID,
		UserID:          userID,
	}
	r.assignments = append(r.assignments, a)
	return &a, nil
}

func (r *mockPSToUserRepo) Revoke(_ context.Context, _ pgx.Tx, psID, userID uuid.UUID) error {
	for i, a := range r.assignments {
		if a.PermissionSetID == psID && a.UserID == userID {
			r.assignments = append(r.assignments[:i], r.assignments[i+1:]...)
			return nil
		}
	}
	return nil
}

func (r *mockPSToUserRepo) ListByUserID(_ context.Context, userID uuid.UUID) ([]security.PermissionSetToUser, error) {
	var result []security.PermissionSetToUser
	for _, a := range r.assignments {
		if a.UserID == userID {
			result = append(result, a)
		}
	}
	return result, nil
}

func (r *mockPSToUserRepo) ListByPermissionSetID(_ context.Context, psID uuid.UUID) ([]security.PermissionSetToUser, error) {
	var result []security.PermissionSetToUser
	for _, a := range r.assignments {
		if a.PermissionSetID == psID {
			result = append(result, a)
		}
	}
	return result, nil
}

// mockObjPermRepo implements security.ObjectPermissionRepository.
type mockObjPermRepo struct {
	perms map[string]*security.ObjectPermission
}

func newMockObjPermRepo() *mockObjPermRepo {
	return &mockObjPermRepo{perms: make(map[string]*security.ObjectPermission)}
}

func permKey(a, b uuid.UUID) string {
	return a.String() + ":" + b.String()
}

func (r *mockObjPermRepo) Upsert(_ context.Context, _ pgx.Tx, psID, objectID uuid.UUID, permissions int) (*security.ObjectPermission, error) {
	key := permKey(psID, objectID)
	op := &security.ObjectPermission{
		ID:              uuid.New(),
		PermissionSetID: psID,
		ObjectID:        objectID,
		Permissions:     permissions,
	}
	r.perms[key] = op
	return op, nil
}

func (r *mockObjPermRepo) GetByPSAndObject(_ context.Context, psID, objectID uuid.UUID) (*security.ObjectPermission, error) {
	return r.perms[permKey(psID, objectID)], nil
}

func (r *mockObjPermRepo) ListByPermissionSetID(_ context.Context, psID uuid.UUID) ([]security.ObjectPermission, error) {
	var result []security.ObjectPermission
	for _, op := range r.perms {
		if op.PermissionSetID == psID {
			result = append(result, *op)
		}
	}
	return result, nil
}

func (r *mockObjPermRepo) Delete(_ context.Context, _ pgx.Tx, psID, objectID uuid.UUID) error {
	delete(r.perms, permKey(psID, objectID))
	return nil
}

// mockFieldPermRepo implements security.FieldPermissionRepository.
type mockFieldPermRepo struct {
	perms map[string]*security.FieldPermission
}

func newMockFieldPermRepo() *mockFieldPermRepo {
	return &mockFieldPermRepo{perms: make(map[string]*security.FieldPermission)}
}

func (r *mockFieldPermRepo) Upsert(_ context.Context, _ pgx.Tx, psID, fieldID uuid.UUID, permissions int) (*security.FieldPermission, error) {
	key := permKey(psID, fieldID)
	fp := &security.FieldPermission{
		ID:              uuid.New(),
		PermissionSetID: psID,
		FieldID:         fieldID,
		Permissions:     permissions,
	}
	r.perms[key] = fp
	return fp, nil
}

func (r *mockFieldPermRepo) GetByPSAndField(_ context.Context, psID, fieldID uuid.UUID) (*security.FieldPermission, error) {
	return r.perms[permKey(psID, fieldID)], nil
}

func (r *mockFieldPermRepo) ListByPermissionSetID(_ context.Context, psID uuid.UUID) ([]security.FieldPermission, error) {
	var result []security.FieldPermission
	for _, fp := range r.perms {
		if fp.PermissionSetID == psID {
			result = append(result, *fp)
		}
	}
	return result, nil
}

func (r *mockFieldPermRepo) Delete(_ context.Context, _ pgx.Tx, psID, fieldID uuid.UUID) error {
	delete(r.perms, permKey(psID, fieldID))
	return nil
}
