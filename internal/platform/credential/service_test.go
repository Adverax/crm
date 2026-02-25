package credential

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepo implements Repository for testing.
type mockRepo struct {
	credentials map[uuid.UUID]*credEntry
	byCode      map[string]uuid.UUID
	tokens      map[uuid.UUID]*tokenEntry
	usageLog    []UsageLogEntry
}

type credEntry struct {
	cred    Credential
	encAuth []byte
	nonce   []byte
}

type tokenEntry struct {
	encToken  []byte
	nonce     []byte
	tokenType string
	expiresAt time.Time
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		credentials: make(map[uuid.UUID]*credEntry),
		byCode:      make(map[string]uuid.UUID),
		tokens:      make(map[uuid.UUID]*tokenEntry),
	}
}

func (m *mockRepo) Create(_ context.Context, cred *Credential, encAuth, nonce []byte) error {
	m.credentials[cred.ID] = &credEntry{cred: *cred, encAuth: encAuth, nonce: nonce}
	m.byCode[cred.Code] = cred.ID
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*Credential, []byte, []byte, error) {
	e, ok := m.credentials[id]
	if !ok {
		return nil, nil, nil, nil
	}
	cred := e.cred
	return &cred, e.encAuth, e.nonce, nil
}

func (m *mockRepo) GetByCode(_ context.Context, code string) (*Credential, []byte, []byte, error) {
	id, ok := m.byCode[code]
	if !ok {
		return nil, nil, nil, nil
	}
	return m.GetByID(context.Background(), id)
}

func (m *mockRepo) ListAll(_ context.Context) ([]Credential, error) {
	var result []Credential
	for _, e := range m.credentials {
		result = append(result, e.cred)
	}
	return result, nil
}

func (m *mockRepo) Update(_ context.Context, id uuid.UUID, input UpdateCredentialInput, encAuth, nonce []byte) error {
	e, ok := m.credentials[id]
	if !ok {
		return nil
	}
	e.cred.Name = input.Name
	e.cred.Description = input.Description
	e.cred.BaseURL = input.BaseURL
	if encAuth != nil {
		e.encAuth = encAuth
		e.nonce = nonce
	}
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	if e, ok := m.credentials[id]; ok {
		delete(m.byCode, e.cred.Code)
		delete(m.credentials, id)
	}
	return nil
}

func (m *mockRepo) SetActive(_ context.Context, id uuid.UUID, active bool) error {
	if e, ok := m.credentials[id]; ok {
		e.cred.IsActive = active
	}
	return nil
}

func (m *mockRepo) GetToken(_ context.Context, credentialID uuid.UUID) ([]byte, []byte, string, time.Time, error) {
	t, ok := m.tokens[credentialID]
	if !ok {
		return nil, nil, "", time.Time{}, nil
	}
	return t.encToken, t.nonce, t.tokenType, t.expiresAt, nil
}

func (m *mockRepo) UpsertToken(_ context.Context, credentialID uuid.UUID, encToken, nonce []byte, tokenType string, expiresAt time.Time) error {
	m.tokens[credentialID] = &tokenEntry{encToken: encToken, nonce: nonce, tokenType: tokenType, expiresAt: expiresAt}
	return nil
}

func (m *mockRepo) DeleteToken(_ context.Context, credentialID uuid.UUID) error {
	delete(m.tokens, credentialID)
	return nil
}

func (m *mockRepo) LogUsage(_ context.Context, entry *UsageLogEntry) error {
	m.usageLog = append(m.usageLog, *entry)
	return nil
}

func (m *mockRepo) GetUsageLog(_ context.Context, credentialID uuid.UUID, limit int) ([]UsageLogEntry, error) {
	var result []UsageLogEntry
	for _, e := range m.usageLog {
		if e.CredentialID == credentialID {
			result = append(result, e)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func testEncryptionKey() []byte {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return key
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateCredentialInput
		wantErr bool
		errMsg  string
	}{
		{
			name: "creates api_key credential",
			input: CreateCredentialInput{
				Code:        "test_api",
				Name:        "Test API",
				Description: "Test",
				Type:        CredentialTypeAPIKey,
				BaseURL:     "https://api.example.com",
				AuthData:    mustJSON(t, ApiKeyAuth{Header: "X-API-Key", Value: "sk-123"}),
			},
			wantErr: false,
		},
		{
			name: "creates basic credential",
			input: CreateCredentialInput{
				Code:     "test_basic",
				Name:     "Test Basic",
				Type:     CredentialTypeBasic,
				BaseURL:  "https://api.example.com",
				AuthData: mustJSON(t, BasicAuth{Username: "user", Password: "pass"}),
			},
			wantErr: false,
		},
		{
			name: "rejects invalid type",
			input: CreateCredentialInput{
				Code:     "test_bad",
				Name:     "Bad",
				Type:     "invalid",
				BaseURL:  "https://api.example.com",
				AuthData: []byte(`{}`),
			},
			wantErr: true,
			errMsg:  "invalid credential type",
		},
		{
			name: "rejects non-https base URL",
			input: CreateCredentialInput{
				Code:     "test_http",
				Name:     "HTTP",
				Type:     CredentialTypeAPIKey,
				BaseURL:  "http://api.example.com",
				AuthData: mustJSON(t, ApiKeyAuth{Header: "X-Key", Value: "val"}),
			},
			wantErr: true,
			errMsg:  "https://",
		},
		{
			name: "rejects empty auth data",
			input: CreateCredentialInput{
				Code:    "test_empty",
				Name:    "Empty",
				Type:    CredentialTypeAPIKey,
				BaseURL: "https://api.example.com",
			},
			wantErr: true,
			errMsg:  "auth_data is required",
		},
		{
			name: "rejects api_key without header",
			input: CreateCredentialInput{
				Code:     "test_no_header",
				Name:     "No Header",
				Type:     CredentialTypeAPIKey,
				BaseURL:  "https://api.example.com",
				AuthData: mustJSON(t, ApiKeyAuth{Header: "", Value: "val"}),
			},
			wantErr: true,
			errMsg:  "header and value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newMockRepo()
			svc := NewService(repo, testEncryptionKey())

			cred, err := svc.Create(context.Background(), tt.input)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.input.Code, cred.Code)
			assert.Equal(t, tt.input.Name, cred.Name)
			assert.True(t, cred.IsActive)
		})
	}
}

func TestService_ResolveAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		credType   CredentialType
		authData   any
		wantHeader string
		wantPrefix string
	}{
		{
			name:       "api_key returns custom header",
			credType:   CredentialTypeAPIKey,
			authData:   ApiKeyAuth{Header: "X-API-Key", Value: "sk-12345"},
			wantHeader: "X-API-Key",
			wantPrefix: "sk-12345",
		},
		{
			name:       "basic returns Authorization header",
			credType:   CredentialTypeBasic,
			authData:   BasicAuth{Username: "admin", Password: "secret"},
			wantHeader: "Authorization",
			wantPrefix: "Basic ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := newMockRepo()
			svc := NewService(repo, testEncryptionKey())

			_, err := svc.Create(context.Background(), CreateCredentialInput{
				Code:     "test_resolve",
				Name:     "Test",
				Type:     tt.credType,
				BaseURL:  "https://api.example.com",
				AuthData: mustJSON(t, tt.authData),
			})
			require.NoError(t, err)

			headerKey, headerValue, baseURL, err := svc.ResolveAuth(context.Background(), "test_resolve")
			require.NoError(t, err)
			assert.Equal(t, tt.wantHeader, headerKey)
			assert.Contains(t, headerValue, tt.wantPrefix)
			assert.Equal(t, "https://api.example.com", baseURL)
		})
	}
}

func TestService_ActivateDeactivate(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := NewService(repo, testEncryptionKey())

	cred, err := svc.Create(context.Background(), CreateCredentialInput{
		Code:     "toggle_test",
		Name:     "Toggle",
		Type:     CredentialTypeAPIKey,
		BaseURL:  "https://api.example.com",
		AuthData: mustJSON(t, ApiKeyAuth{Header: "X-Key", Value: "val"}),
	})
	require.NoError(t, err)
	assert.True(t, cred.IsActive)

	// Deactivate
	err = svc.Deactivate(context.Background(), cred.ID)
	require.NoError(t, err)

	// ResolveAuth should fail
	_, _, _, err = svc.ResolveAuth(context.Background(), "toggle_test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "inactive")

	// Activate
	err = svc.Activate(context.Background(), cred.ID)
	require.NoError(t, err)

	// ResolveAuth should work
	_, _, _, err = svc.ResolveAuth(context.Background(), "toggle_test")
	require.NoError(t, err)
}

func TestService_Delete(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := NewService(repo, testEncryptionKey())

	cred, err := svc.Create(context.Background(), CreateCredentialInput{
		Code:     "delete_test",
		Name:     "Delete",
		Type:     CredentialTypeAPIKey,
		BaseURL:  "https://api.example.com",
		AuthData: mustJSON(t, ApiKeyAuth{Header: "X-Key", Value: "val"}),
	})
	require.NoError(t, err)

	err = svc.Delete(context.Background(), cred.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(context.Background(), cred.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestService_LogUsage(t *testing.T) {
	t.Parallel()

	repo := newMockRepo()
	svc := NewService(repo, testEncryptionKey())

	cred, err := svc.Create(context.Background(), CreateCredentialInput{
		Code:     "usage_test",
		Name:     "Usage",
		Type:     CredentialTypeAPIKey,
		BaseURL:  "https://api.example.com",
		AuthData: mustJSON(t, ApiKeyAuth{Header: "X-Key", Value: "val"}),
	})
	require.NoError(t, err)

	status := 200
	entry := &UsageLogEntry{
		CredentialID:   cred.ID,
		ProcedureCode:  "test_proc",
		RequestURL:     "https://api.example.com/data",
		ResponseStatus: &status,
		Success:        true,
		DurationMs:     150,
	}

	err = svc.LogUsage(context.Background(), entry)
	require.NoError(t, err)

	log, err := svc.GetUsageLog(context.Background(), cred.ID, 10)
	require.NoError(t, err)
	assert.Len(t, log, 1)
	assert.Equal(t, "test_proc", log[0].ProcedureCode)
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}
