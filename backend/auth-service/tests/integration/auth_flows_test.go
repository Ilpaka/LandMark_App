//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/config"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/httpserver"
	"github.com/ilpaka/landmark_app/backend/auth-service/tests/testutil"
)

var otpCodeRe = regexp.MustCompile(`code=(\d{6})`)

func parseVerificationOTP(t *testing.T, logBuf *bytes.Buffer) string {
	t.Helper()
	all := otpCodeRe.FindAllStringSubmatch(logBuf.String(), -1)
	if len(all) == 0 {
		t.Fatalf("no OTP in log: %q", logBuf.String())
	}
	return all[len(all)-1][1]
}

// phoneVerifyAccessToken runs phone/code + phone/verify and returns the access token.
func phoneVerifyAccessToken(t *testing.T, stack *gin.Engine, logBuf *bytes.Buffer) string {
	t.Helper()
	phone := fmt.Sprintf("+7916%07d", time.Now().UnixNano()%10000000)
	regRaw, err := json.Marshal(map[string]any{"phone": phone})
	require.NoError(t, err)
	regReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/code", bytes.NewReader(regRaw))
	regReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	stack.ServeHTTP(w, regReq)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var regOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &regOut))
	code := parseVerificationOTP(t, logBuf)
	verifyBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, regOut.VerificationID, code)
	vReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/verify", bytes.NewReader([]byte(verifyBody)))
	vReq.Header.Set("Content-Type", "application/json")
	wv := httptest.NewRecorder()
	stack.ServeHTTP(wv, vReq)
	require.Equal(t, http.StatusOK, wv.Code, wv.Body.String())
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.Unmarshal(wv.Body.Bytes(), &tok))
	require.NotEmpty(t, tok.AccessToken)
	return tok.AccessToken
}

func TestAuthFlow_PhoneVerifyBindEmailLoginMe(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)

	cfg, err := config.Load()
	require.NoError(t, err)
	cfg.HTTPAddr = "127.0.0.1:0"

	var logBuf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{}))

	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)

	at := phoneVerifyAccessToken(t, stack.Engine, &logBuf)
	email := fmt.Sprintf("flow%d@example.com", time.Now().UnixNano())

	emailRaw, _ := json.Marshal(map[string]any{"email": email})
	emReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/me/email", bytes.NewReader(emailRaw))
	emReq.Header.Set("Content-Type", "application/json")
	emReq.Header.Set("Authorization", "Bearer "+at)
	we := httptest.NewRecorder()
	stack.Engine.ServeHTTP(we, emReq)
	require.Equal(t, http.StatusOK, we.Code, we.Body.String())
	var emOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(we.Body.Bytes(), &emOut))

	evCode := parseVerificationOTP(t, &logBuf)
	verifyBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, emOut.VerificationID, evCode)
	vReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/verify-email", bytes.NewReader([]byte(verifyBody)))
	vReq.Header.Set("Content-Type", "application/json")
	wv := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wv, vReq)
	require.Equal(t, http.StatusOK, wv.Code, wv.Body.String())

	cpRaw, _ := json.Marshal(map[string]any{"new_password": "correcthorse1"})
	cpReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/change-password", bytes.NewReader(cpRaw))
	cpReq.Header.Set("Content-Type", "application/json")
	cpReq.Header.Set("Authorization", "Bearer "+at)
	wcp := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wcp, cpReq)
	require.Equal(t, http.StatusOK, wcp.Code, wcp.Body.String())

	loginBody := fmt.Sprintf(`{"email":%q,"password":"correcthorse1"}`, email)
	lReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/login", bytes.NewReader([]byte(loginBody)))
	lReq.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w3, lReq)
	lResp := w3.Result()
	lb, _ := io.ReadAll(lResp.Body)
	_ = lResp.Body.Close()
	require.Equal(t, http.StatusOK, lResp.StatusCode, "login body=%s", lb)

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.Unmarshal(lb, &tok))
	require.NotEmpty(t, tok.AccessToken)

	meReq := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	w4 := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w4, meReq)
	meResp := w4.Result()
	mb, _ := io.ReadAll(meResp.Body)
	_ = meResp.Body.Close()
	require.Equal(t, http.StatusOK, meResp.StatusCode, "me body=%s", mb)
	var me struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}
	require.NoError(t, json.Unmarshal(mb, &me))
	require.NotEmpty(t, me.UserID)
	require.Equal(t, email, me.Email)
}

func TestAuthFlow_LoginBeforeEmailVerified(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)
	var logBuf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{}))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)

	at := phoneVerifyAccessToken(t, stack.Engine, &logBuf)
	email := fmt.Sprintf("unver%d@example.com", time.Now().UnixNano())
	emailRaw, _ := json.Marshal(map[string]any{"email": email})
	emReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/me/email", bytes.NewReader(emailRaw))
	emReq.Header.Set("Content-Type", "application/json")
	emReq.Header.Set("Authorization", "Bearer "+at)
	we := httptest.NewRecorder()
	stack.Engine.ServeHTTP(we, emReq)
	require.Equal(t, http.StatusOK, we.Code, we.Body.String())

	cpRaw, _ := json.Marshal(map[string]any{"new_password": "correcthorse1"})
	cpReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/change-password", bytes.NewReader(cpRaw))
	cpReq.Header.Set("Content-Type", "application/json")
	cpReq.Header.Set("Authorization", "Bearer "+at)
	wcp := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wcp, cpReq)
	require.Equal(t, http.StatusOK, wcp.Code, wcp.Body.String())

	loginBody := fmt.Sprintf(`{"email":%q,"password":"correcthorse1"}`, email)
	lReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/login", bytes.NewReader([]byte(loginBody)))
	lReq.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w2, lReq)
	require.Equal(t, http.StatusForbidden, w2.Code)
}

func TestAuthFlow_DuplicateEmailBind(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)
	var logBuf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{}))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)

	email := fmt.Sprintf("dup%d@example.com", time.Now().UnixNano())

	at1 := phoneVerifyAccessToken(t, stack.Engine, &logBuf)
	emailRaw, _ := json.Marshal(map[string]any{"email": email})
	em1 := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/me/email", bytes.NewReader(emailRaw))
	em1.Header.Set("Content-Type", "application/json")
	em1.Header.Set("Authorization", "Bearer "+at1)
	w1 := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w1, em1)
	require.Equal(t, http.StatusOK, w1.Code, w1.Body.String())
	var emOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(w1.Body.Bytes(), &emOut))
	evCode := parseVerificationOTP(t, &logBuf)
	verifyBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, emOut.VerificationID, evCode)
	vReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/verify-email", bytes.NewReader([]byte(verifyBody)))
	vReq.Header.Set("Content-Type", "application/json")
	wv := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wv, vReq)
	require.Equal(t, http.StatusOK, wv.Code, wv.Body.String())

	logBuf2 := bytes.Buffer{}
	log2 := slog.New(slog.NewTextHandler(&logBuf2, &slog.HandlerOptions{}))
	stack2, err := httpserver.MountAuth(cfg, log2, deps.Pool, deps.RDB)
	require.NoError(t, err)
	at2 := phoneVerifyAccessToken(t, stack2.Engine, &logBuf2)
	em2 := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/me/email", bytes.NewReader(emailRaw))
	em2.Header.Set("Content-Type", "application/json")
	em2.Header.Set("Authorization", "Bearer "+at2)
	w2 := httptest.NewRecorder()
	stack2.Engine.ServeHTTP(w2, em2)
	require.Equal(t, http.StatusConflict, w2.Code, w2.Body.String())
}

func TestAuthFlow_LoginWrongPassword(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{}))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)

	at := phoneVerifyAccessToken(t, stack.Engine, &buf)
	email := fmt.Sprintf("badpw%d@example.com", time.Now().UnixNano())
	emailRaw, _ := json.Marshal(map[string]any{"email": email})
	emReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/me/email", bytes.NewReader(emailRaw))
	emReq.Header.Set("Content-Type", "application/json")
	emReq.Header.Set("Authorization", "Bearer "+at)
	we := httptest.NewRecorder()
	stack.Engine.ServeHTTP(we, emReq)
	require.Equal(t, http.StatusOK, we.Code, we.Body.String())
	var emOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(we.Body.Bytes(), &emOut))
	code := parseVerificationOTP(t, &buf)
	verifyBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, emOut.VerificationID, code)
	vReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/verify-email", bytes.NewReader([]byte(verifyBody)))
	vReq.Header.Set("Content-Type", "application/json")
	wv := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wv, vReq)
	require.Equal(t, http.StatusOK, wv.Code, wv.Body.String())

	cpRaw, _ := json.Marshal(map[string]any{"new_password": "correcthorse1"})
	cpReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/change-password", bytes.NewReader(cpRaw))
	cpReq.Header.Set("Content-Type", "application/json")
	cpReq.Header.Set("Authorization", "Bearer "+at)
	wcp := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wcp, cpReq)
	require.Equal(t, http.StatusOK, wcp.Code, wcp.Body.String())

	loginBody := fmt.Sprintf(`{"email":%q,"password":"wrongpassword1"}`, email)
	lReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/login", bytes.NewReader([]byte(loginBody)))
	lReq.Header.Set("Content-Type", "application/json")
	wl := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wl, lReq)
	require.Equal(t, http.StatusUnauthorized, wl.Code)
}

func TestAuthFlow_MeWithoutToken(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)

	meReq := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/v1/auth/me", nil)
	w := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w, meReq)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
