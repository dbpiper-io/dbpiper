package airtable

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"time"
)

type oauthStatePayload struct {
	UserID       string `json:"uid"`
	CodeVerifier string `json:"cv"`
	Exp          int64  `json:"exp"`
	Nonce        string `json:"n"`
}

func getStateSecret() ([]byte, error) {
	secret := os.Getenv("AIRTABLE_STATE_SECRET")
	if secret == "" {
		return nil, errors.New("AIRTABLE_STATE_SECRET not set")
	}
	return []byte(secret), nil
}

func generateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func signState(userID, codeVerifier string, ttl time.Duration) (string, error) {
	secret, err := getStateSecret()
	if err != nil {
		return "", err
	}

	nonce, err := generateNonce()
	if err != nil {
		return "", err
	}

	p := oauthStatePayload{
		UserID:       userID,
		CodeVerifier: codeVerifier,
		Exp:          time.Now().Add(ttl).Unix(),
		Nonce:        nonce,
	}

	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	payload := base64.RawURLEncoding.EncodeToString(b)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return payload + "." + sig, nil
}

func verifyState(s string) (*oauthStatePayload, error) {
	secret, err := getStateSecret()
	if err != nil {
		return nil, err
	}

	parts := splitOnce(s, '.')
	if parts == nil {
		return nil, errors.New("invalid state format")
	}

	payloadB, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	sigB, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0]))
	expected := mac.Sum(nil)

	if !hmac.Equal(expected, sigB) {
		return nil, errors.New("invalid state signature")
	}

	var p oauthStatePayload
	if err := json.Unmarshal(payloadB, &p); err != nil {
		return nil, err
	}

	if time.Now().Unix() > p.Exp {
		return nil, errors.New("state expired")
	}

	return &p, nil
}

func splitOnce(s string, sep rune) []string {
	for i, r := range s {
		if r == sep {
			return []string{s[:i], s[i+1:]}
		}
	}
	return nil
}
