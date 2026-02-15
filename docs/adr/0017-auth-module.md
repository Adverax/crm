# ADR-0017: Auth Module — JWT Authentication

**Status:** Accepted
**Date:** 2026-02-13
**Participants:** @roman_myakotin

## Context

Phases 0–4 are complete: metadata engine, security (OLS/FLS/RLS), SOQL, DML. All requests are authenticated via DevAuth middleware (`X-Dev-User-Id` header), which is only suitable for development. Phase 5 replaces DevAuth with full JWT authentication.

Requirements:
1. Authentication via username + password
2. Stateless access tokens (JWT) for API requests
3. Refresh tokens for session renewal without re-entering the password
4. Password reset flow (forgot password)
5. Rate limiting on the login endpoint for brute-force protection
6. Compatibility with the existing security engine (UserContext -> SOQL/DML)

## Options Considered

### Option A — Session-based auth (server-side sessions)

Classic HTTP sessions: session ID in a cookie, session data in Redis/DB.

**Pros:**
- Simple invalidation (delete session on the server)
- No token size issues

**Cons:**
- Stateful — requires shared storage for sessions
- Every request -> lookup in Redis/DB
- Not suitable for an API-first platform (mobile, integrations)
- CSRF protection required for cookie-based auth

### Option B — JWT access + refresh tokens (chosen)

Short-lived JWT access token (15 min) + long-lived refresh token (7 days) in the DB.

**Pros:**
- Stateless access tokens — no lookup on every request
- All required data (UserID, ProfileID, RoleID) in claims — compatible with UserContext
- API-first: convenient for SPA, mobile, integrations
- Standard approach for enterprise APIs

**Cons:**
- Cannot instantly invalidate an access token (wait for 15 min expiry)
- Refresh token requires DB storage

### Option C — OAuth 2.0 / OIDC

Full OAuth 2.0 authorization server.

**Pros:**
- Industry standard
- SSO support, federated identity

**Cons:**
- Enormous complexity for MVP
- Requires an authorization server (Keycloak, Hydra)
- Overkill for a single-tenant self-hosted CRM

## Decision

**Option B chosen — JWT access + refresh tokens.**

### Implementation Details

#### Tokens
- **Access token**: JWT, signed with HMAC-SHA256, TTL = 15 minutes
- **Refresh token**: crypto/rand 32 bytes -> hex string, TTL = 7 days
- **Refresh token storage**: SHA-256 hash in the `iam.refresh_tokens` table. Only the client knows the raw token.
- **Token rotation**: on refresh the old token is deleted and a new one is issued (prevents replay)

#### JWT Claims
```json
{
  "sub": "<user_id>",
  "pid": "<profile_id>",
  "rid": "<role_id>",
  "exp": 1234567890,
  "iat": 1234567890
}
```

Claims contain all fields for `security.UserContext` — middleware creates UserContext from JWT without a DB call.

#### Passwords
- **Hashing**: bcrypt, cost = 12
- **Storage**: `password_hash VARCHAR(255)` column in `iam.users`
- **Empty hash** (`''`) means "password not set" — login is rejected

#### Registration
- **Admin-only**: the administrator creates a user via the existing CRUD (`POST /admin/security/users`), then sets the password via `PUT /admin/security/users/:id/password`
- **No self-registration** — atypical for enterprise CRM
- **Initial admin password**: env variable `ADMIN_INITIAL_PASSWORD`, set on first launch

#### Password Reset
- Table `iam.password_reset_tokens`: one-time token, TTL = 1 hour
- `POST /auth/forgot-password` — always returns 200 (does not reveal whether the email exists)
- `POST /auth/reset-password` — validates the token, sets a new password, invalidates all refresh tokens (force re-login)
- Email sender — interface. For dev: console implementation (logs URL). For production: SMTP implementation (connected later)

#### Rate Limiting
- In-memory sliding window per IP
- 5 attempts per 15 minutes
- Sufficient for single-tenant (ADR-0016)
- Can be replaced with Redis-based if needed

#### Access Token Blacklisting
- **Not implemented.** Access tokens are short-lived (15 min), they expire naturally
- On logout only the refresh token is deleted
- On password reset all of the user's refresh tokens are deleted

### Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/login` | No | Login: username + password -> token pair |
| POST | `/api/v1/auth/refresh` | No | Refresh: refresh_token -> new token pair |
| POST | `/api/v1/auth/forgot-password` | No | Request password reset by email |
| POST | `/api/v1/auth/reset-password` | No | Reset password by token |
| POST | `/api/v1/auth/logout` | JWT | Logout: deletes refresh token |
| GET | `/api/v1/auth/me` | JWT | Current user |
| PUT | `/api/v1/admin/security/users/:id/password` | JWT | Set password (admin) |

### Compatibility with Security Engine

JWT middleware creates `security.UserContext{UserID, ProfileID, RoleID}` from claims and sets it in Gin + standard context — exactly like DevAuth. SOQL/DML engines require no changes.

## Consequences

- DevAuth middleware is replaced by JWTAuth. DevAuth can be preserved for tests (`MODE=dev`)
- All existing endpoints get JWT protection
- Frontend gets a login page, token management, route guards
- Email infrastructure (SMTP) — stub for MVP, real implementation when needed
- OAuth/OIDC (SSO) — future work, in `ee/` (ADR-0014)

## Related Decisions

- [ADR-0009: Security architecture](0009-security-architecture-overview.md) — 3-layer security, UserContext
- [ADR-0014: Open Core](0014-licensing-and-business-model.md) — SSO in ee/
- [ADR-0016: Single-tenant](0016-single-tenant-architecture.md) — in-memory rate limiting is sufficient
