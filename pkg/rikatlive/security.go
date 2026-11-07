package rikatlive

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"os"
)

var defaultSecret = "larapak-rikatlive-secret-key-default-v1"

func getSecretKey() string {
	key := os.Getenv("APP_KEY")
	if key != "" {
		return key
	}
	return defaultSecret
}

// SignSnapshot generates an HMAC-SHA256 signature for the given serialized JSON snapshot.
func SignSnapshot(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(getSecretKey()))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySnapshot checks whether a snapshot matches its cryptographic signature.
func VerifySnapshot(payload []byte, checksum string) bool {
	expectedSignature := SignSnapshot(payload)
	return subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(checksum)) == 1
}
