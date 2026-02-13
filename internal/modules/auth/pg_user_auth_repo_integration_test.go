//go:build integration

package auth

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/testutil"
)

func setupTestUserWithPassword(t *testing.T, pool *pgxpool.Pool, passwordHash string) (uuid.UUID, string, string) {
	t.Helper()
	ctx := context.Background()

	psID := uuid.New()
	profileID := uuid.New()
	userID := uuid.New()
	username := "testuser_" + userID.String()[:8]
	email := "test_" + userID.String()[:8] + "@example.com"

	_, err := pool.Exec(ctx,
		`INSERT INTO iam.permission_sets (id, api_name, label, ps_type) VALUES ($1, $2, $3, $4)`,
		psID, "test_ps_"+psID.String()[:8], "Test PS", "grant",
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO iam.profiles (id, api_name, label, base_permission_set_id) VALUES ($1, $2, $3, $4)`,
		profileID, "test_profile_"+profileID.String()[:8], "Test Profile", psID,
	)
	require.NoError(t, err)

	_, err = pool.Exec(ctx,
		`INSERT INTO iam.users (id, username, email, first_name, last_name, profile_id, password_hash) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, username, email, "Test", "User", profileID, passwordHash,
	)
	require.NoError(t, err)

	return userID, username, email
}

func TestPgUserAuthRepo_Integration(t *testing.T) {
	pool := testutil.SetupTestPool(t)
	testutil.TruncateTables(t, pool, "iam.users", "iam.profiles", "iam.permission_sets")

	repo := NewPgUserAuthRepository(pool)
	ctx := context.Background()

	t.Run("GetByUsername returns correct user with password_hash", func(t *testing.T) {
		userID, username, email := setupTestUserWithPassword(t, pool, "$2a$12$fakehashvalue")

		fetched, err := repo.GetByUsername(ctx, username)
		require.NoError(t, err)
		require.NotNil(t, fetched)

		assert.Equal(t, userID, fetched.ID)
		assert.Equal(t, username, fetched.Username)
		assert.Equal(t, email, fetched.Email)
		assert.Equal(t, "Test", fetched.FirstName)
		assert.Equal(t, "User", fetched.LastName)
		assert.NotEqual(t, uuid.Nil, fetched.ProfileID)
		assert.Nil(t, fetched.RoleID)
		assert.True(t, fetched.IsActive)
		assert.Equal(t, "$2a$12$fakehashvalue", fetched.PasswordHash)
	})

	t.Run("GetByID returns correct user", func(t *testing.T) {
		userID, username, email := setupTestUserWithPassword(t, pool, "$2a$12$hashbyid")

		fetched, err := repo.GetByID(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, fetched)

		assert.Equal(t, userID, fetched.ID)
		assert.Equal(t, username, fetched.Username)
		assert.Equal(t, email, fetched.Email)
		assert.Equal(t, "$2a$12$hashbyid", fetched.PasswordHash)
	})

	t.Run("GetByEmail returns correct user", func(t *testing.T) {
		userID, _, email := setupTestUserWithPassword(t, pool, "$2a$12$hashbyemail")

		fetched, err := repo.GetByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, fetched)

		assert.Equal(t, userID, fetched.ID)
		assert.Equal(t, email, fetched.Email)
		assert.Equal(t, "$2a$12$hashbyemail", fetched.PasswordHash)
	})

	t.Run("GetByUsername returns nil for non-existent", func(t *testing.T) {
		fetched, err := repo.GetByUsername(ctx, "nonexistent_user_"+uuid.NewString()[:8])
		require.NoError(t, err)
		assert.Nil(t, fetched)
	})

	t.Run("SetPassword updates password_hash", func(t *testing.T) {
		userID, _, _ := setupTestUserWithPassword(t, pool, "$2a$12$oldhash")

		err := repo.SetPassword(ctx, userID, "$2a$12$newhash")
		require.NoError(t, err)

		fetched, err := repo.GetByID(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		assert.Equal(t, "$2a$12$newhash", fetched.PasswordHash)
	})

	t.Run("SetPassword returns error for non-existent user ID", func(t *testing.T) {
		err := repo.SetPassword(ctx, uuid.New(), "$2a$12$somehash")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}
