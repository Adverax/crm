package testutil

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TestJWTSecret is a fixed secret used only in unit tests. Never use in production.
const TestJWTSecret = "test-secret-for-unit-tests-only"

// IssueTestJWT generates a valid JWT access token for testing.
func IssueTestJWT(userID, profileID uuid.UUID, roleID *uuid.UUID) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"pid": profileID.String(),
		"exp": now.Add(15 * time.Minute).Unix(),
		"iat": now.Unix(),
	}
	if roleID != nil {
		claims["rid"] = roleID.String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(TestJWTSecret))
	if err != nil {
		panic("testutil.IssueTestJWT: " + err.Error())
	}
	return signed
}

// IssueExpiredJWT generates an expired JWT for negative tests.
func IssueExpiredJWT(userID, profileID uuid.UUID) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"pid": profileID.String(),
		"exp": now.Add(-1 * time.Hour).Unix(),
		"iat": now.Add(-2 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(TestJWTSecret))
	if err != nil {
		panic("testutil.IssueExpiredJWT: " + err.Error())
	}
	return signed
}
