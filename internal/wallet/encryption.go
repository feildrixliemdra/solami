package wallet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	kdfMemory      = 64 * 1024
	kdfIterations  = 3
	kdfParallelism = 2
	kdfKeyLength   = 32
)

// KDFParams stores password key derivation metadata.
type KDFParams struct {
	Salt        string `json:"salt"`
	Time        uint32 `json:"time"`
	Memory      uint32 `json:"memory"`
	Parallelism uint8  `json:"parallelism"`
	KeyLength   uint32 `json:"key_length"`
}

// EncryptedPayload stores AES-GCM encrypted wallet data.
type EncryptedPayload struct {
	Version    int       `json:"version"`
	Cipher     string    `json:"cipher"`
	KDF        string    `json:"kdf"`
	KDFParams  KDFParams `json:"kdf_params"`
	Nonce      string    `json:"nonce"`
	Ciphertext string    `json:"ciphertext"`
}

// Encrypt encrypts plaintext wallet data with a password.
func Encrypt(password string, plaintext []byte) (EncryptedPayload, error) {
	if password == "" {
		return EncryptedPayload{}, fmt.Errorf("password cannot be empty")
	}
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return EncryptedPayload{}, fmt.Errorf("generate salt: %w", err)
	}
	key := deriveKey(password, salt, kdfIterations, kdfMemory, kdfParallelism, kdfKeyLength)
	block, err := aes.NewCipher(key)
	if err != nil {
		return EncryptedPayload{}, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return EncryptedPayload{}, fmt.Errorf("create gcm: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return EncryptedPayload{}, fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	zero(key)
	return EncryptedPayload{
		Version: 1,
		Cipher:  "AES-256-GCM",
		KDF:     "argon2id",
		KDFParams: KDFParams{
			Salt:        base64.StdEncoding.EncodeToString(salt),
			Time:        kdfIterations,
			Memory:      kdfMemory,
			Parallelism: kdfParallelism,
			KeyLength:   kdfKeyLength,
		},
		Nonce:      base64.StdEncoding.EncodeToString(nonce),
		Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
	}, nil
}

// Decrypt decrypts an encrypted wallet payload.
func Decrypt(password string, payload EncryptedPayload) ([]byte, error) {
	if payload.Version != 1 || payload.Cipher != "AES-256-GCM" || payload.KDF != "argon2id" {
		return nil, fmt.Errorf("unsupported wallet encryption format")
	}
	salt, err := base64.StdEncoding.DecodeString(payload.KDFParams.Salt)
	if err != nil {
		return nil, fmt.Errorf("decode salt: %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(payload.Nonce)
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}
	ciphertext, err := base64.StdEncoding.DecodeString(payload.Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}
	key := deriveKey(password, salt, payload.KDFParams.Time, payload.KDFParams.Memory, payload.KDFParams.Parallelism, payload.KDFParams.KeyLength)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	zero(key)
	if err != nil {
		return nil, ErrInvalidPassword
	}
	return plaintext, nil
}

func deriveKey(password string, salt []byte, time uint32, memory uint32, parallelism uint8, keyLength uint32) []byte {
	return argon2.IDKey([]byte(password), salt, time, memory, parallelism, keyLength)
}

func zero(buf []byte) {
	for i := range buf {
		buf[i] = 0
	}
}

func constantTimeEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}

func bytesContainSecret(haystack []byte, secret string) bool {
	return bytes.Contains(haystack, []byte(secret))
}
