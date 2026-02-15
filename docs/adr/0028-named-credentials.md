# ADR-0028: Named Credentials — Secure Secret Storage for Integrations

**Status:** Accepted
**Date:** 2026-02-15
**Participants:** @roman_myakotin

## Context

### Problem: HTTP Integrations Require Authentication

Procedure Engine (ADR-0024) provides the `integration.http` command type for calling external APIs. Each HTTP call requires authentication — an API key, username/password, or OAuth2 token.

The current abstraction in ADR-0024 is the `$.secrets` namespace:

```json
{
  "type": "integration.http",
  "method": "POST",
  "url": "https://api.payment.com/charge",
  "headers": {
    "Authorization": "Bearer $.secrets.stripe_key"
  },
  "body": { "amount": "$.input.amount" }
}
```

This creates **four problems**:

| Problem | Description |
|---------|-------------|
| **Where are `$.secrets` stored?** | ADR-0024 does not define storage. Env vars? DB? Config file? |
| **How to rotate?** | Updating a secret requires knowing where it is stored and who uses it |
| **How to audit?** | There is no record of which Procedure used which secret and when |
| **How to protect against SSRF?** | The URL is constructed from strings — a Procedure could call `http://localhost:5432` |

### Salesforce Named Credentials

Salesforce solves this problem through **Named Credentials** — named accounts bound to an endpoint:

```
Named Credential = endpoint URL + auth method + secrets
```

A Procedure references a credential by name, without knowing the authentication details. When a secret is rotated, only the credential is updated — all Procedures continue to work.

### Why Not Env Vars

| Aspect | Environment Variables | Named Credentials |
|--------|----------------------|-------------------|
| Storage | File / CI secrets | DB (encrypted) |
| Management | DevOps + deployment | Administrator through UI |
| Rotation | Redeployment | Update through API without deployment |
| Audit | Impossible | Full usage log |
| SSRF | No protection | base_url restricts target |
| OAuth2 | Manual token updates | Automatic refresh |
| Visibility | Available to entire application | Only through Credential Service |

## Considered Options

### Option A — Named Credentials in DB (chosen)

Named credentials are stored in `metadata.credentials`, secrets are encrypted with AES-256-GCM. A Procedure references a credential by code. The Credential Service decrypts secrets at runtime and builds the auth header.

**Pros:**
- Centralized management: single source of truth for all secrets
- Encryption at rest: AES-256-GCM with unique nonce
- SSRF protection: base_url restricts allowed hosts
- Rotation without deployment: administrator updates through UI
- OAuth2 auto-refresh: the platform refreshes expired tokens automatically
- Audit: every usage is logged
- Integration with Procedure: `credential` field instead of inline secrets

**Cons:**
- Master key management: a separate secret is needed for encryption (ENV)
- Additional complexity: encryption service, token cache, audit log
- Single point of failure: master key compromise = all secrets

### Option B — Environment Variables

Secrets are stored in env vars, accessible via `$.env.STRIPE_KEY` in Procedure.

**Pros:**
- Simplicity: standard approach, no new code
- Compatibility: works with any CI/CD

**Cons:**
- No UI: administrator cannot manage secrets
- No rotation without deployment: changing an env var = restart
- No audit: impossible to track usage
- No SSRF protection: URL is fully controlled by Procedure
- No OAuth2: manual token updates
- Global visibility: env vars are available to entire application

### Option C — External Vault (HashiCorp Vault / AWS Secrets Manager)

Integration with an external secret manager.

**Pros:**
- Enterprise-grade: battle-tested solutions
- Key rotation: built-in rotation
- Fine-grained ACL: per-secret permissions

**Cons:**
- External dependency: Vault needs to be deployed, configured, maintained
- Latency: network call on every secret access
- Complexity for self-hosted: our target model is single-tenant self-hosted (ADR-0016)
- Overkill for MVP: tens to hundreds of secrets, not thousands

### Option D — Inline Secrets in Procedure JSON

Secrets are stored directly in the Procedure JSON definition.

**Pros:**
- No abstraction: everything in one place

**Cons:**
- Secrets are visible to everyone who can read Procedure definitions
- Logged as part of Procedure (in logs, audit, dumps)
- Duplication: one API key in N Procedures
- No rotation: update = editing all Procedures

## Decision

**Option A chosen: Named Credentials in DB with AES-256-GCM encryption.**

### Named Credential Definition

A Named Credential is a named account that encapsulates:
- **Endpoint** (base URL) — where requests can be sent
- **Auth method** (authentication type) — how to authenticate
- **Secrets** (encrypted data) — what to authenticate with

```json
{
  "code": "stripe_api",
  "name": "Stripe API",
  "description": "Production Stripe account",
  "type": "api_key",
  "base_url": "https://api.stripe.com",
  "auth": {
    "placement": "header",
    "header_name": "Authorization",
    "header_value": "Bearer sk_live_xxx"
  }
}
```

### Credential Types

| Type | Description | Auth flow |
|------|-------------|-----------|
| `api_key` | Static token | Header / Query param |
| `basic` | Username + Password | `Authorization: Basic base64(user:pass)` |
| `oauth2_client` | Client Credentials Grant | Automatic token fetch + refresh |

#### API Key

```json
{
  "code": "sendgrid_api",
  "type": "api_key",
  "base_url": "https://api.sendgrid.com",
  "auth": {
    "placement": "header",
    "header_name": "Authorization",
    "header_value": "Bearer SG.xxx"
  }
}
```

The Credential Service builds the header: `Authorization: Bearer SG.xxx`.

#### Basic Auth

```json
{
  "code": "legacy_erp",
  "type": "basic",
  "base_url": "https://erp.company.com",
  "auth": {
    "username": "api_user",
    "password": "secret123"
  }
}
```

The Credential Service builds: `Authorization: Basic YXBpX3VzZXI6c2VjcmV0MTIz`.

#### OAuth2 Client Credentials

```json
{
  "code": "salesforce_api",
  "type": "oauth2_client",
  "base_url": "https://company.my.salesforce.com",
  "auth": {
    "token_url": "https://login.salesforce.com/services/oauth2/token",
    "client_id": "3MVG9...",
    "client_secret": "xxx",
    "scope": "api refresh_token"
  }
}
```

The Credential Service automatically:
1. Requests an access_token via Client Credentials Grant
2. Caches it until `expires_at` in `metadata.credential_tokens`
3. Refreshes on expiry (transparent to the Procedure)

### Usage in Procedure

The `integration.http` command gets a `credential` field instead of inline auth:

```json
{
  "type": "integration.http",
  "credential": "stripe_api",
  "method": "POST",
  "path": "/v1/charges",
  "body": {
    "amount": "$.input.amount",
    "currency": "usd"
  },
  "as": "charge"
}
```

**What happens during execution:**

```
integration.http command
    |
    +-- 1. Resolve credential by code ("stripe_api")
    |       -> metadata.credentials WHERE code = 'stripe_api'
    |
    +-- 2. Decrypt auth data
    |       -> AES-256-GCM decrypt (master_key, nonce, ciphertext)
    |
    +-- 3. Build auth header
    |       -> For api_key: header from auth config
    |       -> For basic: Authorization: Basic base64(user:pass)
    |       -> For oauth2: resolve/refresh token, Authorization: Bearer <token>
    |
    +-- 4. Validate URL (SSRF protection)
    |       -> base_url + path -> HTTPS only, host match, no internal IPs
    |
    +-- 5. Execute HTTP request
    |       -> request.Header.Set(auth_header)
    |       -> client.Do(request) [timeout: 10s]
    |
    +-- 6. Log usage (without secrets)
            -> credential_id, procedure_code, url, status, duration
```

**`$.secrets` is replaced by `credential`:** ADR-0024 defined `$.secrets` as a runtime namespace for secrets. Named Credentials make this namespace unnecessary — a Procedure does not access secrets directly, but references a credential by code. The Credential Service injects auth transparently.

### Encryption

```
                    +-------------+
                    | Master Key  |
                    | (from ENV)  |
                    +------+------+
                           |
    +----------------------+----------------------+
    |                      |                      |
    |  +----------+   +----+----+   +----------+  |
    |  |  Nonce   |   | AES-256 |   |Auth Data |  |
    |  | (random) |-->|   GCM   |<--|(plaintext)|  |
    |  +----------+   +----+----+   +----------+  |
    |                      |                      |
    |                      v                      |
    |         +------------------------+          |
    |         |  auth_data_encrypted   |          |
    |         |  + auth_data_nonce     |          |
    |         |  (stored in DB)        |          |
    |         +------------------------+          |
    +---------------------------------------------+
```

**Why AES-256-GCM:**
- Authenticated encryption: integrity + confidentiality in one algorithm
- Industry standard (NIST recommendation)
- Hardware acceleration (AES-NI) on all modern CPUs
- Unique nonce per record prevents replay attacks

**Master Key:**

| Environment | Storage |
|-------------|---------|
| Development | `.env` (gitignored) |
| Production | Environment variable `CREDENTIAL_ENCRYPTION_KEY` |
| Enterprise (future) | HashiCorp Vault / KMS via interface |

```bash
# Generation (32 bytes = 256 bits)
openssl rand -base64 32
```

### SSRF Protection

Each credential has a required `base_url`. When executing `integration.http`:

1. Full URL = `base_url` + `path`
2. HTTPS only (HTTP is forbidden)
3. The host of the full URL must match the host from `base_url`
4. Internal IPs are blocked (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, ::1)

```
credential.base_url = "https://api.stripe.com"

path: "/v1/charges"        -> OK https://api.stripe.com/v1/charges
path: "/../internal"       -> DENIED host mismatch (path traversal)
url: "http://localhost:5432" -> DENIED HTTP + internal IP
```

### Storage

Table `metadata.credentials`:

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | Unique ID |
| code | VARCHAR(100) UNIQUE | Code for referencing from Procedure |
| name | VARCHAR(255) | Human-readable name |
| description | TEXT | Purpose description |
| type | VARCHAR(20) | `api_key` / `basic` / `oauth2_client` |
| base_url | VARCHAR(500) NOT NULL | Base URL (SSRF protection) |
| auth_data_encrypted | BYTEA NOT NULL | Encrypted auth data (AES-256-GCM) |
| auth_data_nonce | BYTEA NOT NULL | Unique nonce |
| is_active | BOOLEAN DEFAULT true | Whether the credential is active (deactivation without deletion) |
| created_at | TIMESTAMPTZ | Creation time |
| updated_at | TIMESTAMPTZ | Update time |

Table `metadata.credential_tokens` (OAuth2 token cache):

| Column | Type | Description |
|--------|------|-------------|
| credential_id | UUID PK FK->credentials | 1:1 with credential |
| access_token_encrypted | BYTEA | Encrypted access token |
| access_token_nonce | BYTEA | Nonce for access token |
| token_type | VARCHAR(50) DEFAULT 'Bearer' | Token type |
| expires_at | TIMESTAMPTZ | Expiry time |
| created_at | TIMESTAMPTZ | Time obtained |
| updated_at | TIMESTAMPTZ | Update time |

Table `metadata.credential_usage_log` (usage audit):

| Column | Type | Description |
|--------|------|-------------|
| id | UUID PK | Unique ID |
| credential_id | UUID FK->credentials | Which credential was used |
| procedure_code | VARCHAR(100) | Which Procedure initiated |
| request_url | VARCHAR(500) | Request URL (without query params) |
| response_status | INT | HTTP response status |
| success | BOOLEAN | Success indicator |
| error_message | TEXT | Error message (if any) |
| duration_ms | INT | Request duration |
| created_at | TIMESTAMPTZ | Usage time |
| user_id | UUID FK->users | Who initiated the Procedure |

**Note on `is_active`:** Unlike soft delete for business records (which we do not have, ADR-0003), `is_active` for credentials is a **security mechanism**: temporary deactivation on suspicion of compromise, without losing configuration. A deactivated credential blocks all Procedures that use it.

### API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/admin/credentials` | List credentials (without secrets) |
| POST | `/api/v1/admin/credentials` | Create credential |
| GET | `/api/v1/admin/credentials/:id` | Get credential (without secrets) |
| PUT | `/api/v1/admin/credentials/:id` | Update credential |
| DELETE | `/api/v1/admin/credentials/:id` | Delete (409 if used in Procedures) |
| POST | `/api/v1/admin/credentials/:id/test` | Test connection (GET on base_url) |
| GET | `/api/v1/admin/credentials/:id/usage` | Usage log |
| POST | `/api/v1/admin/credentials/:id/deactivate` | Deactivate |
| POST | `/api/v1/admin/credentials/:id/activate` | Activate |

**GET/PUT response — auth_data is masked:**

```json
{
  "id": "uuid",
  "code": "stripe_api",
  "name": "Stripe API",
  "type": "api_key",
  "base_url": "https://api.stripe.com",
  "is_active": true,
  "auth_masked": {
    "placement": "header",
    "header_name": "Authorization",
    "header_value": "Bearer sk_l***xxx"
  },
  "created_at": "2026-02-15T10:00:00Z"
}
```

Secrets are **never** returned in full in API responses. When updating auth data, the new plaintext is submitted and encrypted on save.

### Security

| Threat | Protection |
|--------|------------|
| DB leak (SQL dump) | AES-256-GCM encryption; without master key the data is useless |
| Master key leak | Stored in ENV (not in code, not in DB); single point — a deliberate trade-off |
| SSRF (Server-Side Request Forgery) | base_url constraint; host match; internal IP blocklist; HTTPS only |
| Logging secrets | Auth data is **never** logged; usage log stores only URL + status |
| Unauthorized API access | Admin-only endpoints (middleware); future: OLS on credentials |
| Credential in inactive Procedure | Validation on Procedure save: credential must exist |
| Deletion of a used credential | 409 Conflict: dependency check (where it is used) |
| Compromise of one credential | Deactivation (`is_active = false`) blocks all usages instantly |

### Limits

| Limit | Rationale |
|-------|-----------|
| HTTPS only | Security (by design) |
| No mTLS | Complexity; to be added in the future |
| No automatic key rotation | Requires integration with each provider; manual update |
| Single master key | Simplicity; key rotation via re-encrypting all records |
| No per-user credentials | All Procedures use one credential; per-user OAuth2 — future |
| Max 100 credentials | Sufficient for production; prevents bloat |

### Validation on Procedure Save

When saving a Procedure, if the command type = `integration.http`:
1. The `credential` field is required (inline URL/auth is forbidden)
2. A credential with the specified code must exist
3. The credential must be `is_active = true`

```json
// OK
{
  "type": "integration.http",
  "credential": "stripe_api",
  "method": "POST",
  "path": "/v1/charges",
  "body": { "amount": "$.input.amount" }
}

// FORBIDDEN: inline URL without credential
{
  "type": "integration.http",
  "method": "POST",
  "url": "https://api.stripe.com/v1/charges",
  "headers": { "Authorization": "Bearer sk_xxx" },
  "body": { "amount": "$.input.amount" }
}
```

### Constructor UI

Admin page for managing credentials:

1. **Credential list**: code, name, type, base_url, is_active, last_used_at
2. **Create/edit**: form with fields by credential type (api_key/basic/oauth2)
3. **Test connection**: button to verify the connection (GET base_url with auth)
4. **Usage log**: usage table with filters (date range, procedure, status)
5. **Deactivate/Activate**: toggle with confirmation (shows affected Procedures)

## Consequences

### Positive

- **Centralization** — all secrets for integrations in one place
- **Encryption at rest** — AES-256-GCM, secrets are unreadable from DB without master key
- **SSRF protection** — base_url + host match + internal IP blocklist
- **Rotation without deployment** — update through UI/API, Procedures remain unchanged
- **OAuth2 auto-refresh** — the platform refreshes tokens automatically
- **Audit** — full log: who, when, which credential, with what result
- **DRY** — one credential is used in N Procedures
- **Integration with Procedure** — `credential` field instead of `$.secrets`; validation on save
- **Instant deactivation** — `is_active = false` blocks all usages

### Negative

- **Master key** — single point; compromise = all secrets. Mitigation: ENV + access control
- **Additional complexity** — encryption service, token cache, usage log
- **No per-user OAuth2** — all users use one credential; Connected Apps — future
- **Manual rotation** — no automatic rotation of API keys (only OAuth2 tokens auto-refresh)

## Related ADRs

- **ADR-0024** — Procedure Engine: `integration.http` command uses the `credential` field. Named Credentials replace the `$.secrets` namespace
- **ADR-0025** — Scenario Engine: steps can call Procedures with `integration.http`, which use Named Credentials
- **ADR-0016** — Single-tenant: master key per instance; no multi-tenant key management
- **ADR-0009** — Security: Admin-only access to credentials API
