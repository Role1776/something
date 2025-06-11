package auth

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateMailCode() (string, error) {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf)[:6], nil
}
