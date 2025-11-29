package airtable

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

func generateCodeVerifier() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	verifier := base64.RawURLEncoding.EncodeToString(b)
	if len(verifier) < 43 || len(verifier) > 128 {
		return "", errors.New("invalid verifier length")
	}

	return verifier, nil
}

func codeChallengeFromVerifier(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
