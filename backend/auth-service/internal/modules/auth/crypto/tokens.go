package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
)

// RandomRefreshToken returns a URL-safe opaque refresh token.
func RandomRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// RefreshTokenHash returns hex(SHA256(pepper || refreshToken)).
func RefreshTokenHash(pepper, refreshToken string) string {
	h := sha256.Sum256([]byte(pepper + refreshToken))
	return hex.EncodeToString(h[:])
}
