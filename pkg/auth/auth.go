package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

type Authenticator struct {
	apiKey     string
	privateKey ed25519.PrivateKey
}

func NewAuthenticator(apiKey, base64PrivateKey string) (*Authenticator, error) {
	privateKeyBytes, err := base64.StdEncoding.DecodeString(base64PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Ed25519 private keys are 64 bytes (32 bytes seed + 32 bytes public key)
	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(privateKeyBytes))
	}

	privateKey := ed25519.PrivateKey(privateKeyBytes)

	return &Authenticator{
		apiKey:     apiKey,
		privateKey: privateKey,
	}, nil
}

// GetAuthHeaders generates the required authentication headers for a request
func (a *Authenticator) GetAuthHeaders(method, path, body string) (map[string]string, error) {
	timestamp := time.Now().Unix()
	
	// Construct the message to sign: api_key + timestamp + path + method + body
	message := a.apiKey + strconv.FormatInt(timestamp, 10) + path + method + body
	
	// Sign the message
	signature := ed25519.Sign(a.privateKey, []byte(message))
	
	// Encode signature to base64
	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	
	headers := map[string]string{
		"x-api-key":   a.apiKey,
		"x-signature": encodedSignature,
		"x-timestamp": strconv.FormatInt(timestamp, 10),
	}
	
	return headers, nil
}

// GenerateKeyPair generates a new Ed25519 key pair for API authentication
func GenerateKeyPair() (privateKeyBase64, publicKeyBase64 string, err error) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key pair: %w", err)
	}
	
	// Encode to base64
	privateKeyBase64 = base64.StdEncoding.EncodeToString(privateKey)
	publicKeyBase64 = base64.StdEncoding.EncodeToString(publicKey)
	
	return privateKeyBase64, publicKeyBase64, nil
}