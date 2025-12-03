package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    uint32 = 1         // number of iterations
	argonMemory  uint32 = 64 * 1024 // 64 MB
	argonThreads uint8  = 4
	argonKeyLen  uint32 = 32
	saltLen             = 16
)

// HashPassword returns an encoded Argon2id hash that you can store in the DB.
//
// Format: argon2id$v=19$m=65536,t=1,p=4$<salt_b64>$<hash_b64>
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory, argonTime, argonThreads, b64Salt, b64Hash)

	return encoded, nil
}

// VerifyPassword checks a plaintext password against an encoded Argon2id hash.
func VerifyPassword(encoded, password string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) != 5 {
		return errors.New("invalid hash format")
	}

	// parts[0] = "argon2id"
	// parts[1] = "v=19"
	// parts[2] = "m=...,t=...,p=..."
	// parts[3] = salt_b64
	// parts[4] = hash_b64

	var mem uint32
	var time uint32
	var threads uint8

	_, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &mem, &time, &threads)
	if err != nil {
		return fmt.Errorf("parse params: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return fmt.Errorf("decode salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return fmt.Errorf("decode hash: %w", err)
	}

	calculated := argon2.IDKey([]byte(password), salt, time, mem, threads, uint32(len(hash)))

	if subtle.ConstantTimeCompare(hash, calculated) == 1 {
		return nil
	}

	return errors.New("invalid password")
}
