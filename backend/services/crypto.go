package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

// encryptBytes encrypts plaintext using AES-256-GCM with the given key.
// The returned ciphertext includes the nonce prepended.
func encryptBytes(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decryptBytes decrypts AES-256-GCM ciphertext (nonce-prepended) with the given key.
func decryptBytes(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ct := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}

// EncryptString encrypts a plaintext string using the hex-encoded AES-256 key.
// Returns a base64-encoded ciphertext suitable for storage in a TEXT column.
// Returns "" if plaintext is empty.
func EncryptString(keyHex, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		return "", fmt.Errorf("invalid encryption key")
	}
	ct, err := encryptBytes(key, []byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ct), nil
}

// DecryptString decrypts a base64-encoded ciphertext using the hex-encoded AES-256 key.
// Returns "" if ciphertext is empty.
func DecryptString(keyHex, ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		return "", fmt.Errorf("invalid encryption key")
	}
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext encoding: %w", err)
	}
	pt, err := decryptBytes(key, raw)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// EncryptBytes encrypts a plaintext []byte using the hex-encoded AES-256 key.
// Returns nil if plaintext is empty.
func EncryptBytesWithKey(keyHex string, plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return []byte{}, nil
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key")
	}
	return encryptBytes(key, plaintext)
}

// DecryptBytesWithKey decrypts a []byte ciphertext using the hex-encoded AES-256 key.
// Returns nil if ciphertext is empty.
func DecryptBytesWithKey(keyHex string, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return []byte{}, nil
	}
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key")
	}
	return decryptBytes(key, ciphertext)
}
