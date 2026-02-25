package procedure

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/adverax/crm/internal/platform/credential"
	"github.com/adverax/crm/internal/platform/metadata"
)

// mockCredentialService implements credential.Service for testing.
type mockCredentialService struct {
	credentials map[string]*credential.Credential
	authHeaders map[string][3]string // code -> [headerKey, headerValue, baseURL]
	usageLog    []credential.UsageLogEntry
}

func newMockCredentialService() *mockCredentialService {
	return &mockCredentialService{
		credentials: make(map[string]*credential.Credential),
		authHeaders: make(map[string][3]string),
	}
}

func (m *mockCredentialService) Create(_ context.Context, _ credential.CreateCredentialInput) (*credential.Credential, error) {
	return nil, nil
}

func (m *mockCredentialService) GetByID(_ context.Context, id uuid.UUID) (*credential.Credential, error) {
	for _, c := range m.credentials {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockCredentialService) GetByCode(_ context.Context, code string) (*credential.Credential, error) {
	c, ok := m.credentials[code]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (m *mockCredentialService) ListAll(_ context.Context) ([]credential.Credential, error) {
	return nil, nil
}

func (m *mockCredentialService) Update(_ context.Context, _ uuid.UUID, _ credential.UpdateCredentialInput) (*credential.Credential, error) {
	return nil, nil
}

func (m *mockCredentialService) Delete(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCredentialService) Activate(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCredentialService) Deactivate(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCredentialService) TestConnection(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockCredentialService) GetUsageLog(_ context.Context, _ uuid.UUID, _ int) ([]credential.UsageLogEntry, error) {
	return nil, nil
}

func (m *mockCredentialService) ResolveAuth(_ context.Context, code string) (string, string, string, error) {
	auth, ok := m.authHeaders[code]
	if !ok {
		return "", "", "", nil
	}
	return auth[0], auth[1], auth[2], nil
}

func (m *mockCredentialService) LogUsage(_ context.Context, entry *credential.UsageLogEntry) error {
	m.usageLog = append(m.usageLog, *entry)
	return nil
}

func TestIntegrationCommandExecutor_Category(t *testing.T) {
	t.Parallel()

	exec := NewIntegrationCommandExecutor(nil, nil)
	assert.Equal(t, "integration", exec.Category())
}

func TestIntegrationCommandExecutor_DryRun(t *testing.T) {
	t.Parallel()

	credSvc := newMockCredentialService()
	exec := NewIntegrationCommandExecutor(credSvc, nil)

	execCtx := NewExecutionContext(map[string]any{}, true, time.Now().Add(30*time.Second))

	cmd := metadata.CommandDef{
		Type:       "integration.http",
		Credential: "test_api",
		Method:     "GET",
		Path:       "/data",
	}

	result, err := exec.Execute(context.Background(), cmd, execCtx)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, 200, resultMap["status"])
	assert.Equal(t, 1, execCtx.HTTPCount)
}

func TestIntegrationCommandExecutor_MissingCredential(t *testing.T) {
	t.Parallel()

	exec := NewIntegrationCommandExecutor(nil, nil)
	execCtx := NewExecutionContext(map[string]any{}, false, time.Now().Add(30*time.Second))

	cmd := metadata.CommandDef{
		Type:   "integration.http",
		Method: "GET",
		Path:   "/data",
	}

	_, err := exec.Execute(context.Background(), cmd, execCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a credential")
}

func TestIntegrationCommandExecutor_HTTPLimitExceeded(t *testing.T) {
	t.Parallel()

	exec := NewIntegrationCommandExecutor(nil, nil)
	execCtx := NewExecutionContext(map[string]any{}, false, time.Now().Add(30*time.Second))
	execCtx.HTTPCount = MaxHTTPCalls

	cmd := metadata.CommandDef{
		Type:       "integration.http",
		Credential: "test_api",
		Method:     "GET",
		Path:       "/data",
	}

	_, err := exec.Execute(context.Background(), cmd, execCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max HTTP calls exceeded")
}

func TestIntegrationCommandExecutor_UnknownSubtype(t *testing.T) {
	t.Parallel()

	exec := NewIntegrationCommandExecutor(nil, nil)
	execCtx := NewExecutionContext(map[string]any{}, false, time.Now().Add(30*time.Second))

	cmd := metadata.CommandDef{
		Type: "integration.grpc",
	}

	_, err := exec.Execute(context.Background(), cmd, execCtx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown integration command")
}
