package crypto

import (
	"encoding/base64"
	"testing"
)

func TestRandomDigits(t *testing.T) {
	t.Parallel()
	if _, err := RandomDigits(0); err == nil {
		t.Fatal("expected error for n=0")
	}
	s, err := RandomDigits(8)
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 8 {
		t.Fatalf("len %d", len(s))
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			t.Fatalf("non-digit: %q", s)
		}
	}
}

func TestHashOTP(t *testing.T) {
	t.Parallel()
	salt := []byte{1, 2, 3}
	a := HashOTP(salt, "123456")
	b := HashOTP(salt, "123456")
	if a != b {
		t.Fatal("hash not stable")
	}
	if HashOTP(salt, "654321") == a {
		t.Fatal("different code should differ")
	}
}

func TestArgon2idHasher_HashVerify(t *testing.T) {
	t.Parallel()
	h := &Argon2idHasher{
		Pepper:      "pepper",
		MemoryKiB:   8,
		Time:        1,
		Parallelism: 1,
		SaltLength:  8,
		KeyLength:   16,
	}
	enc, err := h.Hash("correct-horse-1")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := h.Verify(enc, "correct-horse-1")
	if err != nil || !ok {
		t.Fatalf("verify same: ok=%v err=%v", ok, err)
	}
	ok, err = h.Verify(enc, "wrong")
	if err != nil || ok {
		t.Fatalf("verify wrong: ok=%v err=%v", ok, err)
	}
	ok, err = h.Verify("$argon2id$v=19$m=1,t=1,p=1$bad$bad", "x")
	if err != nil || ok {
		t.Fatalf("malformed encoding: ok=%v err=%v", ok, err)
	}
}

func TestRandomRefreshToken(t *testing.T) {
	t.Parallel()
	a, err := RandomRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := RandomRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("tokens should differ")
	}
	raw, err := base64.RawURLEncoding.DecodeString(a)
	if err != nil || len(raw) != 32 {
		t.Fatalf("decode token: len=%d err=%v", len(raw), err)
	}
}

func TestRefreshTokenHash(t *testing.T) {
	t.Parallel()
	h := RefreshTokenHash("p", "token")
	if h != RefreshTokenHash("p", "token") {
		t.Fatal("not deterministic")
	}
	if RefreshTokenHash("p", "token") == RefreshTokenHash("p", "other") {
		t.Fatal("different token should hash differently")
	}
}
