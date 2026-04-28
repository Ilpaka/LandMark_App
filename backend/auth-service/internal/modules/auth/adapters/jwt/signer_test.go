package jwt

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
)

func writeRSAKeyPair(t *testing.T, dir string) (privPath, pubPath string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pubBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	privPath = filepath.Join(dir, "priv.pem")
	pubPath = filepath.Join(dir, "pub.pem")
	if err := os.WriteFile(privPath, privBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pubPath, pubBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	return privPath, pubPath
}

func TestLoadSigner_SignParseRoundTrip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	priv, pub := writeRSAKeyPair(t, dir)
	s, err := LoadSigner(priv, pub, "kid-1", "https://issuer.test", "api")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	uid := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sid := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	tok, jti, exp, err := s.SignAccess(ctx, uid, sid, domain.RoleUser, time.Minute)
	if err != nil || tok == "" || jti == "" || exp.IsZero() {
		t.Fatalf("sign: err=%v jti=%q", err, jti)
	}
	cl, err := s.ParseAccess(ctx, tok)
	if err != nil {
		t.Fatal(err)
	}
	if cl.Sub != uid || cl.Session != sid || cl.JTI != jti || cl.Role != domain.RoleUser {
		t.Fatalf("claims mismatch: %+v", cl)
	}
}

func TestSigner_JWKS(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	priv, pub := writeRSAKeyPair(t, dir)
	s, err := LoadSigner(priv, pub, "", "iss", "aud")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := s.JWKS(context.Background())
	if err != nil || len(raw) < 50 {
		t.Fatalf("jwks: len=%d err=%v", len(raw), err)
	}
}
