package procedure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/adverax/crm/internal/pkg/apperror"
	"github.com/adverax/crm/internal/platform/credential"
	"github.com/adverax/crm/internal/platform/metadata"
)

// IntegrationCommandExecutor handles integration.http commands.
type IntegrationCommandExecutor struct {
	credSvc    credential.Service
	resolver   *ExpressionResolver
	httpClient *http.Client
}

// NewIntegrationCommandExecutor creates a new IntegrationCommandExecutor.
func NewIntegrationCommandExecutor(credSvc credential.Service, resolver *ExpressionResolver) *IntegrationCommandExecutor {
	return &IntegrationCommandExecutor{
		credSvc:  credSvc,
		resolver: resolver,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Category returns "integration".
func (e *IntegrationCommandExecutor) Category() string {
	return "integration"
}

// Execute runs an integration command.
func (e *IntegrationCommandExecutor) Execute(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	parts := strings.SplitN(cmd.Type, ".", 2)
	if len(parts) != 2 {
		return nil, apperror.BadRequest("invalid integration command type")
	}

	switch parts[1] {
	case "http":
		return e.executeHTTP(ctx, cmd, execCtx)
	default:
		return nil, apperror.BadRequest(fmt.Sprintf("unknown integration command: %s", cmd.Type))
	}
}

func (e *IntegrationCommandExecutor) executeHTTP(ctx context.Context, cmd metadata.CommandDef, execCtx *ExecutionContext) (any, error) {
	if execCtx.HTTPCount >= MaxHTTPCalls {
		return nil, apperror.BadRequest(fmt.Sprintf("max HTTP calls exceeded (%d)", MaxHTTPCalls))
	}

	if cmd.Credential == "" {
		return nil, apperror.BadRequest("integration.http requires a credential")
	}

	// Dry-run: return placeholder
	if execCtx.DryRun {
		execCtx.HTTPCount++
		return map[string]any{
			"status": 200,
			"body":   map[string]any{"dry_run": true},
		}, nil
	}

	// Resolve credential
	headerKey, headerValue, baseURL, err := e.credSvc.ResolveAuth(ctx, cmd.Credential)
	if err != nil {
		return nil, fmt.Errorf("resolve credential %q: %w", cmd.Credential, err)
	}

	// Resolve method
	method := strings.ToUpper(cmd.Method)
	if method == "" {
		method = http.MethodGet
	}

	// Resolve path
	path, err := e.resolver.ResolveString(cmd.Path, execCtx)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	// Build full URL
	fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")

	// Validate URL against SSRF
	if err := credential.ValidateRequestURL(fullURL, baseURL); err != nil {
		return nil, fmt.Errorf("SSRF validation: %w", err)
	}

	// Resolve body
	var bodyReader io.Reader
	if cmd.Body != "" {
		resolvedBody, err := e.resolver.ResolveString(cmd.Body, execCtx)
		if err != nil {
			return nil, fmt.Errorf("resolve body: %w", err)
		}
		bodyReader = strings.NewReader(resolvedBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set auth header
	req.Header.Set(headerKey, headerValue)

	// Resolve and set custom headers
	if cmd.Headers != nil {
		for k, v := range cmd.Headers {
			resolved, err := e.resolver.ResolveString(v, execCtx)
			if err != nil {
				return nil, fmt.Errorf("resolve header %q: %w", k, err)
			}
			req.Header.Set(k, resolved)
		}
	}

	// Set Content-Type for body requests
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	start := time.Now()
	resp, err := e.httpClient.Do(req)
	duration := time.Since(start)
	durationMs := int(duration.Milliseconds())

	// Log usage
	credCode := cmd.Credential
	usageEntry := &credential.UsageLogEntry{
		ProcedureCode: getProcedureCode(execCtx),
		RequestURL:    fullURL,
		Success:       err == nil,
		DurationMs:    durationMs,
	}

	if err != nil {
		usageEntry.ErrorMessage = err.Error()
		_ = e.logUsage(ctx, credCode, usageEntry)
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	status := resp.StatusCode
	usageEntry.ResponseStatus = &status
	usageEntry.Success = status < 400

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		usageEntry.ErrorMessage = "failed to read response body"
		usageEntry.Success = false
		_ = e.logUsage(ctx, credCode, usageEntry)
		return nil, fmt.Errorf("read response body: %w", err)
	}

	_ = e.logUsage(ctx, credCode, usageEntry)

	execCtx.HTTPCount++

	// Parse response
	var respData any
	if err := json.Unmarshal(respBody, &respData); err != nil {
		// Return as string if not JSON
		respData = string(respBody)
	}

	return map[string]any{
		"status": status,
		"body":   respData,
	}, nil
}

func (e *IntegrationCommandExecutor) logUsage(ctx context.Context, credCode string, entry *credential.UsageLogEntry) error {
	// Resolve credential ID from code
	cred, err := e.credSvc.GetByCode(ctx, credCode)
	if err != nil {
		return err
	}
	entry.CredentialID = cred.ID
	return e.credSvc.LogUsage(ctx, entry)
}

func getProcedureCode(execCtx *ExecutionContext) string {
	if len(execCtx.CallStack) > 0 {
		return execCtx.CallStack[0]
	}
	return ""
}
