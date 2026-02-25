package credential

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRequestURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		requestURL string
		baseURL    string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "valid HTTPS same host",
			requestURL: "https://api.example.com/v1/users",
			baseURL:    "https://api.example.com",
			wantErr:    false,
		},
		{
			name:       "valid HTTPS with path in base",
			requestURL: "https://api.example.com/v2/data",
			baseURL:    "https://api.example.com/v1",
			wantErr:    false,
		},
		{
			name:       "rejects HTTP scheme",
			requestURL: "http://api.example.com/v1/users",
			baseURL:    "https://api.example.com",
			wantErr:    true,
			errMsg:     "only HTTPS URLs are allowed",
		},
		{
			name:       "rejects host mismatch",
			requestURL: "https://evil.com/v1/users",
			baseURL:    "https://api.example.com",
			wantErr:    true,
			errMsg:     "does not match credential base URL host",
		},
		{
			name:       "rejects empty scheme",
			requestURL: "api.example.com/v1/users",
			baseURL:    "https://api.example.com",
			wantErr:    true,
			errMsg:     "only HTTPS URLs are allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateRequestURL(tt.requestURL, tt.baseURL)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				// May fail on DNS resolution in CI, skip if that's the issue
				if err != nil {
					assert.Contains(t, err.Error(), "failed to resolve host")
				}
			}
		})
	}
}

func TestIsInternalIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ip       string
		internal bool
	}{
		{name: "localhost IPv4", ip: "127.0.0.1", internal: true},
		{name: "localhost range", ip: "127.0.0.2", internal: true},
		{name: "10.x.x.x", ip: "10.0.0.1", internal: true},
		{name: "172.16.x.x", ip: "172.16.0.1", internal: true},
		{name: "172.31.x.x", ip: "172.31.255.255", internal: true},
		{name: "192.168.x.x", ip: "192.168.1.1", internal: true},
		{name: "link-local", ip: "169.254.1.1", internal: true},
		{name: "IPv6 loopback", ip: "::1", internal: true},
		{name: "IPv6 private", ip: "fc00::1", internal: true},
		{name: "IPv6 link-local", ip: "fe80::1", internal: true},
		{name: "public IPv4", ip: "8.8.8.8", internal: false},
		{name: "public IPv4 2", ip: "93.184.216.34", internal: false},
		{name: "172.32.x.x is public", ip: "172.32.0.1", internal: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ip := net.ParseIP(tt.ip)
			assert.Equal(t, tt.internal, isInternalIP(ip), "IP %s", tt.ip)
		})
	}
}
