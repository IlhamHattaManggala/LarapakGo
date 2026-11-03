package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

const (
	defaultIterations = 100000
	saltLength        = 16
	keyLength         = 32
)

// HashPassword hashes a plain text password using PBKDF2-HMAC-SHA256 with a cryptographically secure random salt.
// Format: pbkdf2_sha256$iterations$salt$hash
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate random salt: %w", err)
	}

	derivedKey := pbkdf2SHA256([]byte(password), salt, defaultIterations, keyLength)
	saltB64 := base64.RawStdEncoding.EncodeToString(salt)
	hashB64 := base64.RawStdEncoding.EncodeToString(derivedKey)

	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", defaultIterations, saltB64, hashB64), nil
}

// VerifyPassword checks if a plain password matches the hashed password string.
func VerifyPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return false
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil || iterations <= 0 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	derivedKey := pbkdf2SHA256([]byte(password), salt, iterations, len(expectedHash))
	return subtle.ConstantTimeCompare(derivedKey, expectedHash) == 1
}

// NeedsRehash checks if the stored hash was generated with fewer iterations than the current default.
func NeedsRehash(encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return true
	}
	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return true
	}
	return iterations < defaultIterations
}

// pbkdf2SHA256 implements RFC 2898 PBKDF2 using HMAC-SHA256 without external dependencies.
func pbkdf2SHA256(password, salt []byte, iter, keyLen int) []byte {
	prf := hmac.New(sha256.New, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var result []byte
	buf := make([]byte, 4)

	for block := 1; block <= numBlocks; block++ {
		// U_1 = PRF(Password, Salt || INT_32_BE(block))
		prf.Reset()
		prf.Write(salt)
		buf[0] = byte(block >> 24)
		buf[1] = byte(block >> 16)
		buf[2] = byte(block >> 8)
		buf[3] = byte(block)
		prf.Write(buf)
		u := prf.Sum(nil)

		t := make([]byte, hashLen)
		copy(t, u)

		// U_2 ... U_iter
		for i := 2; i <= iter; i++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(nil)
			for j := 0; j < hashLen; j++ {
				t[j] ^= u[j]
			}
		}

		result = append(result, t...)
	}

	return result[:keyLen]
}

