package credential

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		plaintext string
	}{
		{name: "short text", plaintext: "hello"},
		{name: "json data", plaintext: `{"api_key":"sk-12345","header":"X-API-Key"}`},
		{name: "empty string", plaintext: ""},
		{name: "unicode", plaintext: "Привет мир 🌍"},
	}

	key := make([]byte, 32) // AES-256
	for i := range key {
		key[i] = byte(i)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ciphertext, nonce, err := Encrypt(key, []byte(tt.plaintext))
			require.NoError(t, err)
			assert.NotEmpty(t, ciphertext)
			assert.NotEmpty(t, nonce)

			// Ciphertext should differ from plaintext
			if tt.plaintext != "" {
				assert.NotEqual(t, []byte(tt.plaintext), ciphertext)
			}

			// Decrypt should recover original
			decrypted, err := Decrypt(key, ciphertext, nonce)
			require.NoError(t, err)
			assert.Equal(t, tt.plaintext, string(decrypted))
		})
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	t.Parallel()

	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	for i := range key1 {
		key1[i] = byte(i)
		key2[i] = byte(i + 1)
	}

	ciphertext, nonce, err := Encrypt(key1, []byte("secret"))
	require.NoError(t, err)

	_, err = Decrypt(key2, ciphertext, nonce)
	assert.Error(t, err)
}

func TestEncryptUniqueness(t *testing.T) {
	t.Parallel()

	key := make([]byte, 32)
	plaintext := []byte("same data")

	ct1, n1, err := Encrypt(key, plaintext)
	require.NoError(t, err)

	ct2, n2, err := Encrypt(key, plaintext)
	require.NoError(t, err)

	// Different nonces should produce different ciphertexts
	assert.NotEqual(t, n1, n2)
	assert.NotEqual(t, ct1, ct2)
}
