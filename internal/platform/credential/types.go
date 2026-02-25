package credential

import (
	"time"

	"github.com/google/uuid"
)

// CredentialType represents the authentication type.
type CredentialType string

const (
	CredentialTypeAPIKey       CredentialType = "api_key"
	CredentialTypeBasic        CredentialType = "basic"
	CredentialTypeOAuth2Client CredentialType = "oauth2_client"
)

// Credential represents a named credential for HTTP integrations (ADR-0028).
type Credential struct {
	ID          uuid.UUID      `json:"id"`
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        CredentialType `json:"type"`
	BaseURL     string         `json:"base_url"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ApiKeyAuth holds API key authentication data.
type ApiKeyAuth struct {
	Header string `json:"header"`
	Value  string `json:"value"`
}

// BasicAuth holds basic authentication data.
type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// OAuth2ClientAuth holds OAuth2 client credentials.
type OAuth2ClientAuth struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TokenURL     string `json:"token_url"`
	Scope        string `json:"scope,omitempty"`
}

// CreateCredentialInput is the input for creating a credential.
type CreateCredentialInput struct {
	Code        string
	Name        string
	Description string
	Type        CredentialType
	BaseURL     string
	AuthData    []byte // JSON-encoded auth data (plaintext, will be encrypted)
}

// UpdateCredentialInput is the input for updating a credential.
type UpdateCredentialInput struct {
	Name        string
	Description string
	BaseURL     string
	AuthData    []byte // if non-nil, re-encrypt
}

// UsageLogEntry represents a single usage log entry.
type UsageLogEntry struct {
	ID             uuid.UUID  `json:"id"`
	CredentialID   uuid.UUID  `json:"credential_id"`
	ProcedureCode  string     `json:"procedure_code,omitempty"`
	RequestURL     string     `json:"request_url"`
	ResponseStatus *int       `json:"response_status,omitempty"`
	Success        bool       `json:"success"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	DurationMs     int        `json:"duration_ms"`
	UserID         *uuid.UUID `json:"user_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
