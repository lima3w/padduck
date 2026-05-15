package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// generateAgentToken creates a cryptographically random raw token and its SHA-256 hash.
func generateAgentToken() (rawToken, tokenHash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("rand.Read: %w", err)
	}
	rawToken = base64.RawURLEncoding.EncodeToString(b)
	tokenHash = hashAgentToken(rawToken)
	return rawToken, tokenHash, nil
}

// hashAgentToken returns the hex-encoded SHA-256 of a raw token.
func hashAgentToken(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}
