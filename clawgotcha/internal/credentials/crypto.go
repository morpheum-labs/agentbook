// Package credentials provides AES-256-GCM encryption for stored agent credential payloads.
package credentials

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	nonceSize = 12
	keySize   = 32
)

// Sealed is nonce + ciphertext (no AAD).
type Sealed struct {
	Nonce      []byte
	Ciphertext []byte
}

// ParseMasterKey decodes CLAWGOTCHA_CREDENTIALS_ENCRYPTION_KEY: 32 raw bytes, or base64.StdEncoding, or 64 hex chars.
func ParseMasterKey(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty encryption key")
	}
	if len(s) == keySize {
		return []byte(s), nil
	}
	if len(s) == 64 {
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("hex key: %w", err)
		}
		if len(b) != keySize {
			return nil, fmt.Errorf("hex key must decode to %d bytes", keySize)
		}
		return b, nil
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("base64 key: %w", err)
	}
	if len(b) != keySize {
		return nil, fmt.Errorf("key must be %d bytes (got %d after base64 decode)", keySize, len(b))
	}
	return b, nil
}

// Encrypt seals plaintext with AES-256-GCM.
func Encrypt(plaintext, key []byte) (*Sealed, error) {
	if len(key) != keySize {
		return nil, fmt.Errorf("key must be %d bytes", keySize)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return &Sealed{Nonce: nonce, Ciphertext: ciphertext}, nil
}

// Decrypt opens a sealed payload.
func Decrypt(s *Sealed, key []byte) ([]byte, error) {
	if s == nil || len(s.Nonce) != nonceSize || len(s.Ciphertext) == 0 {
		return nil, errors.New("invalid sealed payload")
	}
	if len(key) != keySize {
		return nil, fmt.Errorf("key must be %d bytes", keySize)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return gcm.Open(nil, s.Nonce, s.Ciphertext, nil)
}
