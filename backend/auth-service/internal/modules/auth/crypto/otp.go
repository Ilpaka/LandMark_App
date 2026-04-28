package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
)

// RandomDigits returns an n-digit numeric OTP as string.
func RandomDigits(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("invalid length")
	}
	const digits = "0123456789"
	buf := make([]byte, n)
	for i := range buf {
		v, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		d := v.Uint64()
		buf[i] = digits[d]
	}
	return string(buf), nil
}

func RandomSalt(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

// HashOTP returns hex(SHA256(salt || code)).
func HashOTP(salt []byte, code string) string {
	h := sha256.Sum256(append(salt, []byte(code)...))
	return hex.EncodeToString(h[:])
}
