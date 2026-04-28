package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// argon2 supports key length up to 2^32-1; cap decoded blobs from untrusted strings.
const maxArgonKeyBytes = 1024

// Argon2idHasher hashes and verifies passwords using Argon2id.
type Argon2idHasher struct {
	Pepper      string
	MemoryKiB   uint32
	Time        uint32
	Parallelism uint8
	SaltLength  int
	KeyLength   uint32
}

func (h *Argon2idHasher) Hash(password string) (string, error) {
	salt := make([]byte, h.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password+h.Pepper), salt, h.Time, h.MemoryKiB, h.Parallelism, h.KeyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", h.MemoryKiB, h.Time, h.Parallelism, b64Salt, b64Key), nil
}

func (h *Argon2idHasher) Verify(encoded, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, nil
	}
	var memory, time uint32
	var threads uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, nil
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, nil
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, nil
	}
	lw := len(want)
	if lw == 0 || lw > maxArgonKeyBytes {
		return false, nil
	}
	got := argon2.IDKey([]byte(password+h.Pepper), salt, time, memory, threads, uint32(lw))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
