package authhttp

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/jwt"
)

func writeRSAKeyPairHTTP(t *testing.T, dir string) (privPath, pubPath string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	pubBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})
	privPath = filepath.Join(dir, "priv.pem")
	pubPath = filepath.Join(dir, "pub.pem")
	require.NoError(t, os.WriteFile(privPath, privBytes, 0o600))
	require.NoError(t, os.WriteFile(pubPath, pubBytes, 0o644))
	return privPath, pubPath
}

func testSigner(t *testing.T) *jwt.Signer {
	t.Helper()
	dir := t.TempDir()
	priv, pub := writeRSAKeyPairHTTP(t, dir)
	s, err := jwt.LoadSigner(priv, pub, "kid", "iss", "aud")
	require.NoError(t, err)
	return s
}

func TestMe_MissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	s := testSigner(t)
	r := gin.New()
	h := &Handlers{}
	g := r.Group("/v1/auth")
	g.Use(AuthMiddleware(s, nil))
	g.GET("/me", h.Me)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestMe_OK(t *testing.T) {
	// /me now loads account from DB; success path is covered in integration (OpenAPI) tests.
	t.Skip("use integration TestAuthContract_OpenAPI_PhoneVerifyAndMe for full /me with DB")
}
