package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "hashes valid password", password: "securepassword123", wantErr: false},
		{name: "hashes empty password", password: "", wantErr: false},
		{name: "hashes long password", password: "a" + string(make([]byte, 71)), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	t.Parallel()

	hash, err := HashPassword("testpassword")
	if err != nil {
		t.Fatalf("setup: HashPassword failed: %v", err)
	}

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  bool
	}{
		{name: "correct password matches", hash: hash, password: "testpassword", wantErr: false},
		{name: "wrong password fails", hash: hash, password: "wrongpassword", wantErr: true},
		{name: "empty password fails", hash: hash, password: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := CheckPassword(tt.hash, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken() error: %v", err)
	}
	if len(token) != 64 {
		t.Errorf("GenerateToken() length = %d, want 64", len(token))
	}

	token2, _ := GenerateToken()
	if token == token2 {
		t.Error("GenerateToken() produced duplicate tokens")
	}
}

func TestHashToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{name: "hashes non-empty token", input: "abc123"},
		{name: "hashes empty string", input: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hash := HashToken(tt.input)
			if len(hash) != 64 {
				t.Errorf("HashToken() length = %d, want 64 (SHA-256 hex)", len(hash))
			}
			hash2 := HashToken(tt.input)
			if hash != hash2 {
				t.Error("HashToken() is not deterministic")
			}
		})
	}
}
