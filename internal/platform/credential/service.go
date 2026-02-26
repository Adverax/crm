package credential

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

// Service defines the business logic interface for credentials.
type Service interface {
	Create(ctx context.Context, input CreateCredentialInput) (*Credential, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Credential, error)
	GetByCode(ctx context.Context, code string) (*Credential, error)
	ListAll(ctx context.Context) ([]Credential, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateCredentialInput) (*Credential, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	TestConnection(ctx context.Context, id uuid.UUID) error
	GetUsageLog(ctx context.Context, id uuid.UUID, limit int) ([]UsageLogEntry, error)

	// ResolveAuth resolves the credential for HTTP usage: returns auth header key/value + base URL.
	ResolveAuth(ctx context.Context, code string) (headerKey, headerValue, baseURL string, err error)

	// LogUsage records a usage entry for audit.
	LogUsage(ctx context.Context, entry *UsageLogEntry) error
}

// ServiceImpl implements the Service interface.
type ServiceImpl struct {
	repo          Repository
	encryptionKey []byte
	httpClient    *http.Client
}

// NewService creates a new credential Service.
func NewService(repo Repository, encryptionKey []byte) *ServiceImpl {
	return &ServiceImpl{
		repo:          repo,
		encryptionKey: encryptionKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *ServiceImpl) Create(ctx context.Context, input CreateCredentialInput) (*Credential, error) {
	if err := validateCredentialType(input.Type); err != nil {
		return nil, err
	}

	if !strings.HasPrefix(input.BaseURL, "https://") {
		return nil, apperror.BadRequest("base_url must start with https://")
	}

	if err := validateAuthData(input.Type, input.AuthData); err != nil {
		return nil, err
	}

	encryptedAuth, nonce, err := Encrypt(s.encryptionKey, input.AuthData)
	if err != nil {
		return nil, fmt.Errorf("credentialService.Create: %w", err)
	}

	now := time.Now().UTC()
	cred := &Credential{
		ID:          uuid.New(),
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
		Type:        input.Type,
		BaseURL:     input.BaseURL,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, cred, encryptedAuth, nonce); err != nil {
		return nil, fmt.Errorf("credentialService.Create: %w", err)
	}

	return cred, nil
}

func (s *ServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (*Credential, error) {
	cred, _, _, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("credentialService.GetByID: %w", err)
	}
	if cred == nil {
		return nil, apperror.NotFound("credential", id.String())
	}
	return cred, nil
}

func (s *ServiceImpl) GetByCode(ctx context.Context, code string) (*Credential, error) {
	cred, _, _, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("credentialService.GetByCode: %w", err)
	}
	if cred == nil {
		return nil, apperror.NotFound("credential", code)
	}
	return cred, nil
}

func (s *ServiceImpl) ListAll(ctx context.Context) ([]Credential, error) {
	creds, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("credentialService.ListAll: %w", err)
	}
	return creds, nil
}

func (s *ServiceImpl) Update(ctx context.Context, id uuid.UUID, input UpdateCredentialInput) (*Credential, error) {
	cred, _, _, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("credentialService.Update: %w", err)
	}
	if cred == nil {
		return nil, apperror.NotFound("credential", id.String())
	}

	var encryptedAuth, nonce []byte
	if input.AuthData != nil {
		if err := validateAuthData(cred.Type, input.AuthData); err != nil {
			return nil, err
		}

		encryptedAuth, nonce, err = Encrypt(s.encryptionKey, input.AuthData)
		if err != nil {
			return nil, fmt.Errorf("credentialService.Update: %w", err)
		}
	}

	if err := s.repo.Update(ctx, id, input, encryptedAuth, nonce); err != nil {
		return nil, fmt.Errorf("credentialService.Update: %w", err)
	}

	cred.Name = input.Name
	cred.Description = input.Description
	cred.BaseURL = input.BaseURL
	cred.UpdatedAt = time.Now().UTC()

	return cred, nil
}

func (s *ServiceImpl) Delete(ctx context.Context, id uuid.UUID) error {
	cred, _, _, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("credentialService.Delete: %w", err)
	}
	if cred == nil {
		return apperror.NotFound("credential", id.String())
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("credentialService.Delete: %w", err)
	}
	return nil
}

func (s *ServiceImpl) Activate(ctx context.Context, id uuid.UUID) error {
	cred, _, _, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("credentialService.Activate: %w", err)
	}
	if cred == nil {
		return apperror.NotFound("credential", id.String())
	}

	if err := s.repo.SetActive(ctx, id, true); err != nil {
		return fmt.Errorf("credentialService.Activate: %w", err)
	}
	return nil
}

func (s *ServiceImpl) Deactivate(ctx context.Context, id uuid.UUID) error {
	cred, _, _, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("credentialService.Deactivate: %w", err)
	}
	if cred == nil {
		return apperror.NotFound("credential", id.String())
	}

	if err := s.repo.SetActive(ctx, id, false); err != nil {
		return fmt.Errorf("credentialService.Deactivate: %w", err)
	}
	return nil
}

func (s *ServiceImpl) TestConnection(ctx context.Context, id uuid.UUID) error {
	cred, encAuth, nonce, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("credentialService.TestConnection: %w", err)
	}
	if cred == nil {
		return apperror.NotFound("credential", id.String())
	}

	if !cred.IsActive {
		return apperror.BadRequest("credential is inactive")
	}

	headerKey, headerValue, err := s.resolveAuthHeader(ctx, cred, encAuth, nonce)
	if err != nil {
		return fmt.Errorf("credentialService.TestConnection: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cred.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("credentialService.TestConnection: %w", err)
	}
	req.Header.Set(headerKey, headerValue)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return apperror.BadRequest(fmt.Sprintf("connection test failed: %s", err.Error()))
	}
	defer func() { _ = resp.Body.Close() }()
	// Drain body to allow connection reuse
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return apperror.BadRequest(fmt.Sprintf("connection test returned HTTP %d", resp.StatusCode))
	}

	return nil
}

func (s *ServiceImpl) GetUsageLog(ctx context.Context, id uuid.UUID, limit int) ([]UsageLogEntry, error) {
	cred, _, _, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("credentialService.GetUsageLog: %w", err)
	}
	if cred == nil {
		return nil, apperror.NotFound("credential", id.String())
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	entries, err := s.repo.GetUsageLog(ctx, id, limit)
	if err != nil {
		return nil, fmt.Errorf("credentialService.GetUsageLog: %w", err)
	}
	return entries, nil
}

func (s *ServiceImpl) ResolveAuth(ctx context.Context, code string) (string, string, string, error) {
	cred, encAuth, nonce, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return "", "", "", fmt.Errorf("credentialService.ResolveAuth: %w", err)
	}
	if cred == nil {
		return "", "", "", apperror.NotFound("credential", code)
	}

	if !cred.IsActive {
		return "", "", "", apperror.BadRequest(fmt.Sprintf("credential %q is inactive", code))
	}

	headerKey, headerValue, err := s.resolveAuthHeader(ctx, cred, encAuth, nonce)
	if err != nil {
		return "", "", "", fmt.Errorf("credentialService.ResolveAuth: %w", err)
	}

	return headerKey, headerValue, cred.BaseURL, nil
}

func (s *ServiceImpl) LogUsage(ctx context.Context, entry *UsageLogEntry) error {
	entry.ID = uuid.New()
	entry.CreatedAt = time.Now().UTC()

	if err := s.repo.LogUsage(ctx, entry); err != nil {
		return fmt.Errorf("credentialService.LogUsage: %w", err)
	}
	return nil
}

// resolveAuthHeader decrypts auth data and returns the HTTP header key/value.
func (s *ServiceImpl) resolveAuthHeader(ctx context.Context, cred *Credential, encAuth, nonce []byte) (string, string, error) {
	plaintext, err := Decrypt(s.encryptionKey, encAuth, nonce)
	if err != nil {
		return "", "", fmt.Errorf("decrypt auth: %w", err)
	}

	switch cred.Type {
	case CredentialTypeAPIKey:
		var auth ApiKeyAuth
		if err := json.Unmarshal(plaintext, &auth); err != nil {
			return "", "", fmt.Errorf("unmarshal api_key auth: %w", err)
		}
		return auth.Header, auth.Value, nil

	case CredentialTypeBasic:
		var auth BasicAuth
		if err := json.Unmarshal(plaintext, &auth); err != nil {
			return "", "", fmt.Errorf("unmarshal basic auth: %w", err)
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(auth.Username + ":" + auth.Password))
		return "Authorization", "Basic " + encoded, nil

	case CredentialTypeOAuth2Client:
		return s.resolveOAuth2Auth(ctx, cred, plaintext)

	default:
		return "", "", apperror.BadRequest(fmt.Sprintf("unsupported credential type: %s", cred.Type))
	}
}

// resolveOAuth2Auth handles OAuth2 client credentials flow with token caching.
func (s *ServiceImpl) resolveOAuth2Auth(ctx context.Context, cred *Credential, plaintext []byte) (string, string, error) {
	var auth OAuth2ClientAuth
	if err := json.Unmarshal(plaintext, &auth); err != nil {
		return "", "", fmt.Errorf("unmarshal oauth2 auth: %w", err)
	}

	// Check cached token
	encToken, tokenNonce, tokenType, expiresAt, err := s.repo.GetToken(ctx, cred.ID)
	if err != nil {
		return "", "", fmt.Errorf("get cached token: %w", err)
	}

	if encToken != nil && time.Now().Before(expiresAt.Add(-30*time.Second)) {
		// Token is still valid (with 30s buffer)
		token, err := Decrypt(s.encryptionKey, encToken, tokenNonce)
		if err != nil {
			return "", "", fmt.Errorf("decrypt cached token: %w", err)
		}
		return "Authorization", tokenType + " " + string(token), nil
	}

	// Fetch new token
	token, tType, expiresIn, err := s.fetchOAuth2Token(ctx, auth)
	if err != nil {
		return "", "", err
	}

	// Cache the token
	encNewToken, newNonce, err := Encrypt(s.encryptionKey, []byte(token))
	if err != nil {
		return "", "", fmt.Errorf("encrypt token: %w", err)
	}

	newExpiry := time.Now().Add(time.Duration(expiresIn) * time.Second)
	if err := s.repo.UpsertToken(ctx, cred.ID, encNewToken, newNonce, tType, newExpiry); err != nil {
		// Log but don't fail — we still have the token
		_ = err
	}

	return "Authorization", tType + " " + token, nil
}

// fetchOAuth2Token performs the OAuth2 client credentials grant.
func (s *ServiceImpl) fetchOAuth2Token(ctx context.Context, auth OAuth2ClientAuth) (token, tokenType string, expiresIn int, err error) {
	data := fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s",
		auth.ClientID, auth.ClientSecret)
	if auth.Scope != "" {
		data += "&scope=" + auth.Scope
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, auth.TokenURL, strings.NewReader(data))
	if err != nil {
		return "", "", 0, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("token request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", 0, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", "", 0, fmt.Errorf("parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", "", 0, fmt.Errorf("empty access_token in response")
	}

	tType := tokenResp.TokenType
	if tType == "" {
		tType = "Bearer"
	}

	return tokenResp.AccessToken, tType, tokenResp.ExpiresIn, nil
}

func validateCredentialType(t CredentialType) error {
	switch t {
	case CredentialTypeAPIKey, CredentialTypeBasic, CredentialTypeOAuth2Client:
		return nil
	default:
		return apperror.BadRequest(fmt.Sprintf("invalid credential type: %s", t))
	}
}

func validateAuthData(credType CredentialType, authData []byte) error {
	if len(authData) == 0 {
		return apperror.BadRequest("auth_data is required")
	}

	switch credType {
	case CredentialTypeAPIKey:
		var auth ApiKeyAuth
		if err := json.Unmarshal(authData, &auth); err != nil {
			return apperror.BadRequest("invalid api_key auth data: " + err.Error())
		}
		if auth.Header == "" || auth.Value == "" {
			return apperror.BadRequest("api_key auth requires header and value")
		}

	case CredentialTypeBasic:
		var auth BasicAuth
		if err := json.Unmarshal(authData, &auth); err != nil {
			return apperror.BadRequest("invalid basic auth data: " + err.Error())
		}
		if auth.Username == "" || auth.Password == "" {
			return apperror.BadRequest("basic auth requires username and password")
		}

	case CredentialTypeOAuth2Client:
		var auth OAuth2ClientAuth
		if err := json.Unmarshal(authData, &auth); err != nil {
			return apperror.BadRequest("invalid oauth2_client auth data: " + err.Error())
		}
		if auth.ClientID == "" || auth.ClientSecret == "" || auth.TokenURL == "" {
			return apperror.BadRequest("oauth2_client auth requires client_id, client_secret, and token_url")
		}
	}

	return nil
}
