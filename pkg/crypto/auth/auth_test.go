package auth

import (
	"encoding/base64"
	"testing"
)

func TestNewAuthenticator(t *testing.T) {
	// Test data from the documentation
	apiKey := "rh-api-6148effc-c0b1-486c-8940-a1d099456be6"
	// Note: This is the private key from the docs, but it's only the seed portion
	// We need to append the public key to make it a full 64-byte Ed25519 private key
	privateKeySeed := "xQnTJVeQLmw1/Mg2YimEViSpw/SdJcgNXZ5kQkAXNPU="
	publicKeyBase64 := "jPItx4TLjcnSUnmnXQQyAKL4eJj3+oWNNMmmm2vATqk="
	
	// Combine seed and public key to form the full private key
	seed, _ := base64.StdEncoding.DecodeString(privateKeySeed)
	pub, _ := base64.StdEncoding.DecodeString(publicKeyBase64)
	fullPrivateKey := append(seed, pub...)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(fullPrivateKey)

	tests := []struct {
		name      string
		apiKey    string
		privKey   string
		wantErr   bool
		errContains string
	}{
		{
			name:    "Valid authenticator",
			apiKey:  apiKey,
			privKey: privateKeyBase64,
			wantErr: false,
		},
		{
			name:      "Invalid base64",
			apiKey:    apiKey,
			privKey:   "not-valid-base64!@#$",
			wantErr:   true,
			errContains: "failed to decode private key",
		},
		{
			name:      "Wrong key size",
			apiKey:    apiKey,
			privKey:   base64.StdEncoding.EncodeToString([]byte("too-short")),
			wantErr:   true,
			errContains: "invalid private key size",
		},
		{
			name:    "Empty API key",
			apiKey:  "",
			privKey: privateKeyBase64,
			wantErr: false, // API key can be empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewAuthenticator(tt.apiKey, tt.privKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuthenticator() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %v", err, tt.errContains)
				}
			}
			if !tt.wantErr && auth == nil {
				t.Error("expected non-nil authenticator")
			}
		})
	}
}

func TestGetAuthHeaders(t *testing.T) {
	// Test data from the documentation
	apiKey := "rh-api-6148effc-c0b1-486c-8940-a1d099456be6"
	privateKeySeed := "xQnTJVeQLmw1/Mg2YimEViSpw/SdJcgNXZ5kQkAXNPU="
	publicKeyBase64 := "jPItx4TLjcnSUnmnXQQyAKL4eJj3+oWNNMmmm2vATqk="
	
	// Combine seed and public key to form the full private key
	seed, _ := base64.StdEncoding.DecodeString(privateKeySeed)
	pub, _ := base64.StdEncoding.DecodeString(publicKeyBase64)
	fullPrivateKey := append(seed, pub...)
	privateKeyBase64 := base64.StdEncoding.EncodeToString(fullPrivateKey)

	auth, err := NewAuthenticator(apiKey, privateKeyBase64)
	if err != nil {
		t.Fatalf("Failed to create authenticator: %v", err)
	}

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "GET request without body",
			method: "GET",
			path:   "/api/v1/crypto/trading/accounts/",
			body:   "",
		},
		{
			name:   "POST request with body",
			method: "POST",
			path:   "/api/v1/crypto/trading/orders/",
			body:   `{"client_order_id":"131de903-5a9c-4260-abc1-28d562a5dcf0","side":"buy","type":"market","symbol":"BTC-USD","market_order_config":{"asset_quantity":"0.1"}}`,
		},
		{
			name:   "DELETE request",
			method: "DELETE",
			path:   "/api/v1/crypto/trading/orders/123/cancel/",
			body:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers, err := auth.GetAuthHeaders(tt.method, tt.path, tt.body)
			if err != nil {
				t.Errorf("GetAuthHeaders() error = %v", err)
				return
			}

			// Check required headers
			requiredHeaders := []string{"x-api-key", "x-signature", "x-timestamp"}
			for _, h := range requiredHeaders {
				if v, ok := headers[h]; !ok || v == "" {
					t.Errorf("missing or empty required header: %s", h)
				}
			}

			// Verify API key
			if headers["x-api-key"] != apiKey {
				t.Errorf("x-api-key = %v, want %v", headers["x-api-key"], apiKey)
			}

			// Verify timestamp is numeric
			if _, err := parseInt64(headers["x-timestamp"]); err != nil {
				t.Errorf("x-timestamp is not a valid integer: %v", headers["x-timestamp"])
			}

			// Verify signature is base64
			if _, err := base64.StdEncoding.DecodeString(headers["x-signature"]); err != nil {
				t.Errorf("x-signature is not valid base64: %v", err)
			}
		})
	}
}

func TestGenerateKeyPair(t *testing.T) {
	privKey, pubKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair() error = %v", err)
	}

	// Verify keys are valid base64
	privKeyBytes, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		t.Errorf("private key is not valid base64: %v", err)
	}
	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		t.Errorf("public key is not valid base64: %v", err)
	}

	// Verify key sizes
	if len(privKeyBytes) != 64 {
		t.Errorf("private key size = %d, want 64", len(privKeyBytes))
	}
	if len(pubKeyBytes) != 32 {
		t.Errorf("public key size = %d, want 32", len(pubKeyBytes))
	}

	// Verify we can create an authenticator with the generated keys
	auth, err := NewAuthenticator("test-api-key", privKey)
	if err != nil {
		t.Errorf("failed to create authenticator with generated keys: %v", err)
	}
	if auth == nil {
		t.Error("expected non-nil authenticator")
	}
}

// Helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		len(s) >= len(substr) && contains(s[1:], substr)
}

func parseInt64(s string) (int64, error) {
	var result int64
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, &parseError{s}
		}
		result = result*10 + int64(r-'0')
	}
	return result, nil
}

type parseError struct {
	s string
}

func (e *parseError) Error() string {
	return "invalid integer: " + e.s
}