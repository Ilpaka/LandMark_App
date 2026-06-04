//go:build e2e

package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/config"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/httpserver"
	jwtadapter "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/jwt"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
	"github.com/ilpaka/landmark_app/backend/auth-service/tests/testutil"
)

var otpRe = regexp.MustCompile(`code=(\d{6})`)

type harn struct {
	*testing.T
	*gin.Engine
	Log  *bytes.Buffer
	Pool *pgxpool.Pool
}

func setup(t *testing.T) *harn {
	t.Helper()
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, nil))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)
	return &harn{T: t, Engine: stack.Engine, Log: &buf, Pool: deps.Pool}
}

func (h *harn) post(path string, body any) *httptest.ResponseRecorder {
	raw, err := json.Marshal(body)
	require.NoError(h, err)
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1"+path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	h.Engine.ServeHTTP(w, req)
	return w
}

func (h *harn) postAuth(path, bearer string, body any) *httptest.ResponseRecorder {
	raw, err := json.Marshal(body)
	require.NoError(h, err)
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1"+path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	h.Engine.ServeHTTP(w, req)
	return w
}

func (h *harn) get(path, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1"+path, nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	h.Engine.ServeHTTP(w, req)
	return w
}

func (h *harn) otp() string {
	all := otpRe.FindAllStringSubmatch(h.Log.String(), -1)
	require.NotEmpty(h, all, "log=%q", h.Log.String())
	return all[len(all)-1][1]
}

func uniquePhone() string {
	return fmt.Sprintf("+7916%07d", time.Now().UnixNano()%10000000)
}

func phoneCodeAndTokens(h *harn, phone string) (access, refresh string) {
	t := h.T
	t.Helper()
	w := h.post("/v1/auth/phone/code", map[string]any{"phone": phone})
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var reg struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &reg))
	wv := h.post("/v1/auth/phone/verify", map[string]any{
		"verification_id": reg.VerificationID,
		"code":            h.otp(),
	})
	require.Equal(t, http.StatusOK, wv.Code, wv.Body.String())
	var tok struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
	}
	require.NoError(t, json.Unmarshal(wv.Body.Bytes(), &tok))
	return tok.Access, tok.Refresh
}

func TestE1_PhoneCodeVerifyMe(t *testing.T) {
	h := setup(t)
	at, _ := phoneCodeAndTokens(h, uniquePhone())
	wm := h.get("/v1/auth/me", at)
	require.Equal(t, http.StatusOK, wm.Code)
}

func TestE2_RepeatPhoneCodeSameAccount(t *testing.T) {
	h := setup(t)
	ph := uniquePhone()
	require.Equal(t, http.StatusOK, h.post("/v1/auth/phone/code", map[string]any{"phone": ph}).Code)
	require.Equal(t, http.StatusOK, h.post("/v1/auth/phone/code", map[string]any{"phone": ph}).Code)
}

func TestE3_Reserved(t *testing.T) { t.Skip("nickname reservation moved to profile service") }

func TestE4_VerifyBeforeCodeNotFound(t *testing.T) {
	h := setup(t)
	w := h.post("/v1/auth/phone/verify", map[string]any{
		"verification_id": "00000000-0000-0000-0000-000000000001",
		"code":            "000000",
	})
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestE5_Reserved2(t *testing.T) { t.Skip("password login removed; use phone OTP") }
func TestE6_Reserved3(t *testing.T) { t.Skip("password login removed; use phone OTP") }

func TestE7_RefreshReuse(t *testing.T) {
	h := setup(t)
	_, rt := phoneCodeAndTokens(h, uniquePhone())
	w1 := h.post("/v1/auth/refresh", map[string]any{"refresh_token": rt})
	require.Equal(t, http.StatusOK, w1.Code, "body=%s", w1.Body.String())
	var tok struct {
		Refresh string `json:"refresh_token"`
	}
	require.NoError(t, json.Unmarshal(w1.Body.Bytes(), &tok))
	// Same old refresh token is accepted again while Redis grace window is active (retry safety).
	time.Sleep(11 * time.Second)
	w2 := h.post("/v1/auth/refresh", map[string]any{"refresh_token": rt})
	require.Equal(t, http.StatusUnauthorized, w2.Code, "reuse body=%s", w2.Body.String())
	_ = tok
}

func TestE8_LogoutThenRefreshInvalid(t *testing.T) {
	h := setup(t)
	at, rt := phoneCodeAndTokens(h, uniquePhone())
	wlo := h.post("/v1/auth/logout", map[string]any{"access_token": at, "refresh_token": rt})
	require.Equal(t, http.StatusOK, wlo.Code, "body=%s", wlo.Body.String())
	wr := h.post("/v1/auth/refresh", map[string]any{"refresh_token": rt})
	require.Equal(t, http.StatusUnauthorized, wr.Code, "body=%s", wr.Body.String())
}

func TestE9_LogoutAllSessionsEmpty(t *testing.T) {
	h := setup(t)
	at, _ := phoneCodeAndTokens(h, uniquePhone())
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/logout-all", nil)
	req.Header.Set("Authorization", "Bearer "+at)
	w := httptest.NewRecorder()
	h.Engine.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
	ws := h.get("/v1/auth/sessions", at)
	require.Equal(t, http.StatusUnauthorized, ws.Code)
}

func TestE12_ForgotPasswordSameShape(t *testing.T) {
	h := setup(t)
	a := h.post("/v1/auth/forgot-password", map[string]any{"email": "nobody-here@example.com"})
	b := h.post("/v1/auth/forgot-password", map[string]any{"email": "other-nope@example.com"})
	require.Equal(t, http.StatusOK, a.Code)
	require.Equal(t, http.StatusOK, b.Code)
	require.JSONEq(t, a.Body.String(), b.Body.String())
}

func TestE13_AccessTokenExpires(t *testing.T) {
	t.Setenv("ACCESS_TOKEN_TTL", "1s")
	h := setup(t)
	at, _ := phoneCodeAndTokens(h, uniquePhone())
	time.Sleep(1500 * time.Millisecond)
	wm := h.get("/v1/auth/me", at)
	require.Equal(t, http.StatusUnauthorized, wm.Code)
}

func TestE14_TamperedAccessToken(t *testing.T) {
	h := setup(t)
	at, _ := phoneCodeAndTokens(h, uniquePhone())
	parts := strings.Split(at, ".")
	require.Len(t, parts, 3)
	tampered := parts[0] + "." + strings.Repeat("A", len(parts[1])) + "." + parts[2]
	wm := h.get("/v1/auth/me", tampered)
	require.Equal(t, http.StatusUnauthorized, wm.Code)
}

func writeTempRSAKeyPair(t *testing.T, dir string) (privPath, pubPath string) {
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

func TestE15_AccessTokenWrongSigningKey(t *testing.T) {
	h := setup(t)
	at, _ := phoneCodeAndTokens(h, uniquePhone())
	var me struct {
		UserID string `json:"user_id"`
	}
	require.NoError(t, json.Unmarshal(h.get("/v1/auth/me", at).Body.Bytes(), &me))
	uid, err := uuid.Parse(me.UserID)
	require.NoError(t, err)

	dir := t.TempDir()
	priv, pub := writeTempRSAKeyPair(t, dir)
	wrongSigner, err := jwtadapter.LoadSigner(priv, pub, "", "https://wrong.issuer", "audience")
	require.NoError(t, err)
	badTok, _, _, err := wrongSigner.SignAccess(context.Background(), uid, uuid.New(), domain.RoleUser, time.Minute)
	require.NoError(t, err)
	wm := h.get("/v1/auth/me", badTok)
	require.Equal(t, http.StatusUnauthorized, wm.Code)
}

func TestE18_ConcurrentRefresh(t *testing.T) {
	h := setup(t)
	_, rt := phoneCodeAndTokens(h, uniquePhone())
	const n = 2
	var wg sync.WaitGroup
	codes := make(chan int, n)
	for range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := h.post("/v1/auth/refresh", map[string]any{"refresh_token": rt})
			codes <- w.Code
		}()
	}
	wg.Wait()
	close(codes)
	var got []int
	for c := range codes {
		got = append(got, c)
	}
	slices.Sort(got)
	require.Equal(t, []int{http.StatusOK, http.StatusUnauthorized}, got, "concurrent refresh codes")
}

func TestE16_AdminBlockBlocksPhoneCode(t *testing.T) {
	h := setup(t)
	victimPhone := uniquePhone()
	atUser, _ := phoneCodeAndTokens(h, victimPhone)
	var me struct {
		UserID string `json:"user_id"`
	}
	wm := h.get("/v1/auth/me", atUser)
	require.NoError(t, json.Unmarshal(wm.Body.Bytes(), &me))

	adminPhone := uniquePhone()
	_, _ = phoneCodeAndTokens(h, adminPhone)
	_, err := h.Pool.Exec(context.Background(),
		`UPDATE auth_accounts SET role='admin' WHERE phone_e164=$1`, adminPhone)
	require.NoError(t, err)
	atAdmin, _ := phoneCodeAndTokens(h, adminPhone)

	blockPath := "/v1/admin/users/" + me.UserID + "/block"
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1"+blockPath, strings.NewReader(`{"reason":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+atAdmin)
	wb := httptest.NewRecorder()
	h.Engine.ServeHTTP(wb, req)
	require.Equal(t, http.StatusOK, wb.Code, "body=%s", wb.Body.String())

	wl := h.post("/v1/auth/phone/code", map[string]any{"phone": victimPhone})
	require.Equal(t, http.StatusForbidden, wl.Code)
}

var resetIDRe = regexp.MustCompile(`reset_id=([0-9a-f-]{36})`)

func TestE11_PasswordResetCompletes(t *testing.T) {
	h := setup(t)
	email := fmt.Sprintf("e11_%d@ex.com", time.Now().UnixNano())
	enorm := strings.ToLower(email)
	ph := uniquePhone()
	at, _ := phoneCodeAndTokens(h, ph)
	var me struct {
		UserID string `json:"user_id"`
	}
	require.NoError(t, json.Unmarshal(h.get("/v1/auth/me", at).Body.Bytes(), &me))
	_, err := h.Pool.Exec(context.Background(),
		`UPDATE auth_accounts SET email=$1, email_normalized=$2, email_verified_at=now() WHERE id=$3::uuid`,
		email, enorm, me.UserID)
	require.NoError(t, err)
	cp := h.postAuth("/v1/auth/change-password", at, map[string]any{"new_password": "correcthorse1"})
	require.Equal(t, http.StatusOK, cp.Code, cp.Body.String())
	h.post("/v1/auth/forgot-password", map[string]any{"email": email})
	m := resetIDRe.FindStringSubmatch(h.Log.String())
	require.Len(t, m, 2, "log=%q", h.Log.String())
	codes := otpRe.FindAllStringSubmatch(h.Log.String(), -1)
	require.NotEmpty(t, codes)
	resetCode := codes[len(codes)-1][1]
	newPass := "newcorrect1"
	wr := h.post("/v1/auth/reset-password", map[string]any{
		"reset_id":     m[1],
		"code":         resetCode,
		"new_password": newPass,
	})
	require.Equal(t, http.StatusOK, wr.Code)
	wl := h.post("/v1/auth/login", map[string]any{"email": email, "password": newPass})
	require.Equal(t, http.StatusOK, wl.Code, wl.Body.String())
	var loginTok struct {
		Access string `json:"access_token"`
	}
	require.NoError(t, json.Unmarshal(wl.Body.Bytes(), &loginTok))
	require.NotEmpty(t, loginTok.Access)
	c2 := h.postAuth("/v1/auth/change-password", loginTok.Access, map[string]any{"current_password": newPass, "new_password": "newcorrect2"})
	require.Equal(t, http.StatusOK, c2.Code, c2.Body.String())
}

func TestE19_VerifyAfterExpiry(t *testing.T) {
	h := setup(t)
	email := fmt.Sprintf("e19_%d@ex.com", time.Now().UnixNano())
	at, _ := phoneCodeAndTokens(h, uniquePhone())
	wb := h.postAuth("/v1/auth/me/email", at, map[string]any{"email": email})
	require.Equal(t, http.StatusOK, wb.Code, wb.Body.String())
	var start struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(wb.Body.Bytes(), &start))
	require.NotEmpty(t, start.VerificationID)
	code := h.otp()
	_, err := h.Pool.Exec(context.Background(),
		`UPDATE auth_email_verifications SET expires_at = now() - interval '1 hour' WHERE id = $1::uuid`,
		start.VerificationID)
	require.NoError(t, err)
	wv := h.post("/v1/auth/verify-email", map[string]any{
		"verification_id": start.VerificationID,
		"code":            code,
	})
	require.Equal(t, http.StatusBadRequest, wv.Code, wv.Body.String())
}

func TestE10_ResetAfterExpiry(t *testing.T) {
	h := setup(t)
	email := fmt.Sprintf("e10_%d@ex.com", time.Now().UnixNano())
	enorm := strings.ToLower(email)
	at, _ := phoneCodeAndTokens(h, uniquePhone())
	var me struct {
		UserID string `json:"user_id"`
	}
	require.NoError(t, json.Unmarshal(h.get("/v1/auth/me", at).Body.Bytes(), &me))
	_, err := h.Pool.Exec(context.Background(),
		`UPDATE auth_accounts SET email=$1, email_normalized=$2, email_verified_at=now() WHERE id=$3::uuid`,
		email, enorm, me.UserID)
	require.NoError(t, err)
	_ = h.postAuth("/v1/auth/change-password", at, map[string]any{"new_password": "correcthorse1"})
	h.post("/v1/auth/forgot-password", map[string]any{"email": email})
	m := resetIDRe.FindStringSubmatch(h.Log.String())
	require.Len(t, m, 2)
	_, err = h.Pool.Exec(context.Background(),
		`UPDATE auth_password_resets SET expires_at = now() - interval '1 hour' WHERE id = $1::uuid`, m[1])
	require.NoError(t, err)
	codes := otpRe.FindAllStringSubmatch(h.Log.String(), -1)
	resetCode := codes[len(codes)-1][1]
	wr := h.post("/v1/auth/reset-password", map[string]any{
		"reset_id": m[1], "code": resetCode, "new_password": "anothergood1",
	})
	require.Equal(t, http.StatusBadRequest, wr.Code)
}

func TestE17_AdminForceLogoutRevokesRefresh(t *testing.T) {
	h := setup(t)
	vPh := uniquePhone()
	atV, rtV := phoneCodeAndTokens(h, vPh)
	aPh := uniquePhone()
	_, _ = phoneCodeAndTokens(h, aPh)
	_, err := h.Pool.Exec(context.Background(),
		`UPDATE auth_accounts SET role='admin' WHERE phone_e164=$1`, aPh)
	require.NoError(t, err)
	atA, _ := phoneCodeAndTokens(h, aPh)

	var me struct {
		UserID string `json:"user_id"`
	}
	require.NoError(t, json.Unmarshal(h.get("/v1/auth/me", atV).Body.Bytes(), &me))

	path := "/v1/admin/users/" + me.UserID + "/force-logout"
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1"+path, nil)
	req.Header.Set("Authorization", "Bearer "+atA)
	wf := httptest.NewRecorder()
	h.Engine.ServeHTTP(wf, req)
	require.Equal(t, http.StatusOK, wf.Code)

	wr := h.post("/v1/auth/refresh", map[string]any{"refresh_token": rtV})
	require.NotEqual(t, http.StatusOK, wr.Code, "body=%s", wr.Body.String())
}

func TestE20_PhonePasswordResetHappyPath(t *testing.T) {
	h := setup(t)
	ph := uniquePhone()
	at, _ := phoneCodeAndTokens(h, ph)
	var me struct {
		UserID string `json:"user_id"`
	}
	require.NoError(t, json.Unmarshal(h.get("/v1/auth/me", at).Body.Bytes(), &me))
	email := fmt.Sprintf("e20_%d@ex.com", time.Now().UnixNano())
	enorm := strings.ToLower(email)
	_, err := h.Pool.Exec(context.Background(),
		`UPDATE auth_accounts SET email=$1, email_normalized=$2, email_verified_at=now() WHERE id=$3::uuid`,
		email, enorm, me.UserID)
	require.NoError(t, err)
	cp := h.postAuth("/v1/auth/change-password", at, map[string]any{"new_password": "correcthorse1"})
	require.Equal(t, http.StatusOK, cp.Code, cp.Body.String())

	wf := h.post("/v1/auth/forgot-password/phone", map[string]any{"phone": ph})
	require.Equal(t, http.StatusOK, wf.Code, wf.Body.String())
	var forgot struct {
		Status         string `json:"status"`
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(wf.Body.Bytes(), &forgot))
	require.Equal(t, "ok", forgot.Status)
	require.NotEmpty(t, forgot.VerificationID)

	newPass := "phonereset1"
	wr := h.post("/v1/auth/reset-password/phone", map[string]any{
		"verification_id": forgot.VerificationID,
		"code":            h.otp(),
		"new_password":    newPass,
	})
	require.Equal(t, http.StatusOK, wr.Code, wr.Body.String())
	var st struct {
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(wr.Body.Bytes(), &st))
	require.Equal(t, "ok", st.Status)
	require.NotContains(t, wr.Body.String(), "access_token")

	wl := h.post("/v1/auth/login", map[string]any{"email": email, "password": newPass})
	require.Equal(t, http.StatusOK, wl.Code, wl.Body.String())
}

func TestE21_PhonePasswordResetWrongCode(t *testing.T) {
	h := setup(t)
	ph := uniquePhone()
	at, _ := phoneCodeAndTokens(h, ph)
	cp := h.postAuth("/v1/auth/change-password", at, map[string]any{"new_password": "correcthorse1"})
	require.Equal(t, http.StatusOK, cp.Code, cp.Body.String())

	wf := h.post("/v1/auth/forgot-password/phone", map[string]any{"phone": ph})
	require.Equal(t, http.StatusOK, wf.Code, wf.Body.String())
	var forgot struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(wf.Body.Bytes(), &forgot))
	require.NotEmpty(t, forgot.VerificationID)

	wr := h.post("/v1/auth/reset-password/phone", map[string]any{
		"verification_id": forgot.VerificationID,
		"code":            "000000",
		"new_password":    "someother1",
	})
	require.Equal(t, http.StatusBadRequest, wr.Code, wr.Body.String())
}

func TestE22_PhoneVerifyRejectsPasswordResetVerification(t *testing.T) {
	h := setup(t)
	ph := uniquePhone()
	at, _ := phoneCodeAndTokens(h, ph)
	cp := h.postAuth("/v1/auth/change-password", at, map[string]any{"new_password": "correcthorse1"})
	require.Equal(t, http.StatusOK, cp.Code, cp.Body.String())

	wf := h.post("/v1/auth/forgot-password/phone", map[string]any{"phone": ph})
	require.Equal(t, http.StatusOK, wf.Code, wf.Body.String())
	var forgot struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(wf.Body.Bytes(), &forgot))
	code := h.otp()

	wbad := h.post("/v1/auth/phone/verify", map[string]any{
		"verification_id": forgot.VerificationID,
		"code":            code,
	})
	require.Equal(t, http.StatusConflict, wbad.Code, wbad.Body.String())
}
