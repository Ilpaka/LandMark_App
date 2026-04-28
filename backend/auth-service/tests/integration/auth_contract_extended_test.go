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
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/config"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/httpserver"
	"github.com/ilpaka/landmark_app/backend/auth-service/tests/testutil"
)

var extOTPRe = regexp.MustCompile(`code=(\d{6})`)

func TestAuthContract_OpenAPI_PhoneVerifyAndMe(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)

	var logBuf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logBuf, nil))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)

	doc := loadAndValidateSpec(t, filepath.Join(testutil.BackendRoot(t), "api", "openapi", "auth.yaml"))

	phone := fmt.Sprintf("+7916%07d", time.Now().UnixNano()%10000000)
	regRaw, _ := json.Marshal(map[string]any{"phone": phone})
	regReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/code", bytes.NewReader(regRaw))
	regReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w, regReq)
	regResp := w.Result()
	rb, _ := io.ReadAll(regResp.Body)
	_ = regResp.Body.Close()
	require.Equal(t, http.StatusOK, regResp.StatusCode)

	var regOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(rb, &regOut))
	all := extOTPRe.FindAllStringSubmatch(logBuf.String(), -1)
	require.NotEmpty(t, all)
	m := all[len(all)-1]
	require.Len(t, m, 2)

	verifyBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, regOut.VerificationID, m[1])
	vReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/verify", bytes.NewReader([]byte(verifyBody)))
	vReq.Header.Set("Content-Type", "application/json")
	wv := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wv, vReq)
	vResp := wv.Result()
	vb, _ := io.ReadAll(vResp.Body)
	_ = vResp.Body.Close()
	require.Equal(t, http.StatusOK, vResp.StatusCode)
	valV, err := http.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/verify", bytes.NewReader([]byte(verifyBody)))
	require.NoError(t, err)
	valV.Header.Set("Content-Type", "application/json")
	validateResponse(t, doc, valV, vResp, vb)

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.Unmarshal(vb, &tok))

	meReq := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/v1/auth/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	wm := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wm, meReq)
	meResp := wm.Result()
	mb, _ := io.ReadAll(meResp.Body)
	_ = meResp.Body.Close()
	require.Equal(t, http.StatusOK, meResp.StatusCode)
	valMe, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/v1/auth/me", nil)
	require.NoError(t, err)
	valMe.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	validateResponse(t, doc, valMe, meResp, mb)
}

func TestAuthContract_OpenAPI_PhoneCodeBadBody(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)
	doc := loadAndValidateSpec(t, filepath.Join(testutil.BackendRoot(t), "api", "openapi", "auth.yaml"))

	bad := `{}`
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/code", bytes.NewReader([]byte(bad)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	valReq, err := http.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/code", bytes.NewReader([]byte(bad)))
	require.NoError(t, err)
	valReq.Header.Set("Content-Type", "application/json")
	validateResponseOnly(t, doc, valReq, resp, body)
}

func TestAuthContract_OpenAPI_MeMissingToken(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)
	doc := loadAndValidateSpec(t, filepath.Join(testutil.BackendRoot(t), "api", "openapi", "auth.yaml"))

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/v1/auth/me", nil)
	w := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	valReq, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/v1/auth/me", nil)
	require.NoError(t, err)
	validateResponse(t, doc, valReq, resp, body)
}

func TestAuthContract_OpenAPI_PhoneBindEmailPasswordLogin(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)

	var logBuf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logBuf, nil))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)

	doc := loadAndValidateSpec(t, filepath.Join(testutil.BackendRoot(t), "api", "openapi", "auth.yaml"))

	phone := fmt.Sprintf("+7916%07d", time.Now().UnixNano()%10000000)
	regRaw, _ := json.Marshal(map[string]any{"phone": phone})
	regReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/code", bytes.NewReader(regRaw))
	regReq.Header.Set("Content-Type", "application/json")
	regReq.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w, regReq)
	regResp := w.Result()
	rb, _ := io.ReadAll(regResp.Body)
	_ = regResp.Body.Close()
	require.Equal(t, http.StatusOK, regResp.StatusCode)

	var regOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(rb, &regOut))
	all := extOTPRe.FindAllStringSubmatch(logBuf.String(), -1)
	require.NotEmpty(t, all)
	smsCode := all[len(all)-1][1]

	verifyBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, regOut.VerificationID, smsCode)
	vReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/verify", bytes.NewReader([]byte(verifyBody)))
	vReq.Header.Set("Content-Type", "application/json")
	vReq.RemoteAddr = "127.0.0.1:12345"
	wv := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wv, vReq)
	vResp := wv.Result()
	vb, _ := io.ReadAll(vResp.Body)
	_ = vResp.Body.Close()
	require.Equal(t, http.StatusOK, vResp.StatusCode)

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.Unmarshal(vb, &tok))

	email := fmt.Sprintf("bind%d@example.com", time.Now().UnixNano())
	emailRaw, _ := json.Marshal(map[string]any{"email": email})
	meEmailReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/me/email", bytes.NewReader(emailRaw))
	meEmailReq.Header.Set("Content-Type", "application/json")
	meEmailReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	meEmailReq.RemoteAddr = "127.0.0.1:12345"
	wEmail := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wEmail, meEmailReq)
	emailResp := wEmail.Result()
	eb, _ := io.ReadAll(emailResp.Body)
	_ = emailResp.Body.Close()
	require.Equal(t, http.StatusOK, emailResp.StatusCode, string(eb))
	valEmailReq, err := http.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/me/email", bytes.NewReader(emailRaw))
	require.NoError(t, err)
	valEmailReq.Header.Set("Content-Type", "application/json")
	valEmailReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	validateResponse(t, doc, valEmailReq, emailResp, eb)

	var emailOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(eb, &emailOut))

	all = extOTPRe.FindAllStringSubmatch(logBuf.String(), -1)
	require.NotEmpty(t, all)
	emailCode := all[len(all)-1][1]

	verifyEmailBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, emailOut.VerificationID, emailCode)
	verReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/verify-email", bytes.NewReader([]byte(verifyEmailBody)))
	verReq.Header.Set("Content-Type", "application/json")
	wver := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wver, verReq)
	verResp := wver.Result()
	verb, _ := io.ReadAll(verResp.Body)
	_ = verResp.Body.Close()
	require.Equal(t, http.StatusOK, verResp.StatusCode, string(verb))

	cpRaw, _ := json.Marshal(map[string]any{"new_password": "correcthorse1"})
	cpReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/change-password", bytes.NewReader(cpRaw))
	cpReq.Header.Set("Content-Type", "application/json")
	cpReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	cpReq.RemoteAddr = "127.0.0.1:12345"
	wcp := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wcp, cpReq)
	cpResp := wcp.Result()
	cpb, _ := io.ReadAll(cpResp.Body)
	_ = cpResp.Body.Close()
	require.Equal(t, http.StatusOK, cpResp.StatusCode, string(cpb))

	loginRaw, _ := json.Marshal(map[string]any{"email": email, "password": "correcthorse1"})
	loginReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/login", bytes.NewReader(loginRaw))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.RemoteAddr = "127.0.0.1:12345"
	wl := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wl, loginReq)
	loginResp := wl.Result()
	lb, _ := io.ReadAll(loginResp.Body)
	_ = loginResp.Body.Close()
	require.Equal(t, http.StatusOK, loginResp.StatusCode, string(lb))
	valLogin, err := http.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/login", bytes.NewReader(loginRaw))
	require.NoError(t, err)
	valLogin.Header.Set("Content-Type", "application/json")
	validateResponse(t, doc, valLogin, loginResp, lb)

	var loginTok struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.Unmarshal(lb, &loginTok))
	require.NotEmpty(t, loginTok.AccessToken)
}

func TestAuthContract_OpenAPI_PhonePasswordResetFlow(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)

	var logBuf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logBuf, nil))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)
	doc := loadAndValidateSpec(t, filepath.Join(testutil.BackendRoot(t), "api", "openapi", "auth.yaml"))

	phone := fmt.Sprintf("+7916%07d", time.Now().UnixNano()%10000000)
	regRaw, _ := json.Marshal(map[string]any{"phone": phone})
	regReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/code", bytes.NewReader(regRaw))
	regReq.Header.Set("Content-Type", "application/json")
	regReq.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w, regReq)
	regResp := w.Result()
	rb, _ := io.ReadAll(regResp.Body)
	_ = regResp.Body.Close()
	require.Equal(t, http.StatusOK, regResp.StatusCode)

	var regOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(rb, &regOut))
	all := extOTPRe.FindAllStringSubmatch(logBuf.String(), -1)
	require.NotEmpty(t, all)
	loginCode := all[len(all)-1][1]

	verifyBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, regOut.VerificationID, loginCode)
	vReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/verify", bytes.NewReader([]byte(verifyBody)))
	vReq.Header.Set("Content-Type", "application/json")
	vReq.RemoteAddr = "127.0.0.1:12345"
	wv := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wv, vReq)
	vResp := wv.Result()
	vb, _ := io.ReadAll(vResp.Body)
	_ = vResp.Body.Close()
	require.Equal(t, http.StatusOK, vResp.StatusCode)

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.Unmarshal(vb, &tok))

	cpRaw, _ := json.Marshal(map[string]any{"new_password": "correcthorse1"})
	cpReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/change-password", bytes.NewReader(cpRaw))
	cpReq.Header.Set("Content-Type", "application/json")
	cpReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	cpReq.RemoteAddr = "127.0.0.1:12345"
	wcp := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wcp, cpReq)
	cpResp := wcp.Result()
	cpb, _ := io.ReadAll(cpResp.Body)
	_ = cpResp.Body.Close()
	require.Equal(t, http.StatusOK, cpResp.StatusCode, string(cpb))

	email := fmt.Sprintf("phr%d@example.com", time.Now().UnixNano())
	emailRaw, _ := json.Marshal(map[string]any{"email": email})
	meEmailReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/me/email", bytes.NewReader(emailRaw))
	meEmailReq.Header.Set("Content-Type", "application/json")
	meEmailReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	meEmailReq.RemoteAddr = "127.0.0.1:12345"
	wEmail := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wEmail, meEmailReq)
	emailResp := wEmail.Result()
	eb, _ := io.ReadAll(emailResp.Body)
	_ = emailResp.Body.Close()
	require.Equal(t, http.StatusOK, emailResp.StatusCode, string(eb))

	var emailOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(eb, &emailOut))
	all = extOTPRe.FindAllStringSubmatch(logBuf.String(), -1)
	emailCode := all[len(all)-1][1]
	verifyEmailBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, emailOut.VerificationID, emailCode)
	verReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/verify-email", bytes.NewReader([]byte(verifyEmailBody)))
	verReq.Header.Set("Content-Type", "application/json")
	wver := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wver, verReq)
	verResp := wver.Result()
	verb, _ := io.ReadAll(verResp.Body)
	_ = verResp.Body.Close()
	require.Equal(t, http.StatusOK, verResp.StatusCode, string(verb))

	forgotRaw, _ := json.Marshal(map[string]any{"phone": phone})
	forgotReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/forgot-password/phone", bytes.NewReader(forgotRaw))
	forgotReq.Header.Set("Content-Type", "application/json")
	forgotReq.RemoteAddr = "127.0.0.1:12345"
	wf := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wf, forgotReq)
	forgotResp := wf.Result()
	fb, _ := io.ReadAll(forgotResp.Body)
	_ = forgotResp.Body.Close()
	require.Equal(t, http.StatusOK, forgotResp.StatusCode, string(fb))
	valForgot, err := http.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/forgot-password/phone", bytes.NewReader(forgotRaw))
	require.NoError(t, err)
	valForgot.Header.Set("Content-Type", "application/json")
	validateResponse(t, doc, valForgot, forgotResp, fb)

	var forgotOut struct {
		Status           string `json:"status"`
		VerificationID   string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(fb, &forgotOut))
	require.Equal(t, "ok", forgotOut.Status)
	require.NotEmpty(t, forgotOut.VerificationID)

	all = extOTPRe.FindAllStringSubmatch(logBuf.String(), -1)
	resetSMS := all[len(all)-1][1]

	newPass := "resetphone1"
	resetRaw, _ := json.Marshal(map[string]any{
		"verification_id": forgotOut.VerificationID,
		"code":            resetSMS,
		"new_password":    newPass,
	})
	resetReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/reset-password/phone", bytes.NewReader(resetRaw))
	resetReq.Header.Set("Content-Type", "application/json")
	resetReq.RemoteAddr = "127.0.0.1:12345"
	wr := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wr, resetReq)
	resetResp := wr.Result()
	rsb, _ := io.ReadAll(resetResp.Body)
	_ = resetResp.Body.Close()
	require.Equal(t, http.StatusOK, resetResp.StatusCode, string(rsb))
	valReset, err := http.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/reset-password/phone", bytes.NewReader(resetRaw))
	require.NoError(t, err)
	valReset.Header.Set("Content-Type", "application/json")
	validateResponse(t, doc, valReset, resetResp, rsb)

	loginRaw, _ := json.Marshal(map[string]any{"email": email, "password": newPass})
	loginReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/login", bytes.NewReader(loginRaw))
	loginReq.Header.Set("Content-Type", "application/json")
	loginReq.RemoteAddr = "127.0.0.1:12345"
	wl := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wl, loginReq)
	loginResp := wl.Result()
	lb, _ := io.ReadAll(loginResp.Body)
	_ = loginResp.Body.Close()
	require.Equal(t, http.StatusOK, loginResp.StatusCode, string(lb))
	valLogin, err := http.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/login", bytes.NewReader(loginRaw))
	require.NoError(t, err)
	valLogin.Header.Set("Content-Type", "application/json")
	validateResponse(t, doc, valLogin, loginResp, lb)
}

func TestAuthContract_OpenAPI_PhoneVerifyRejectsPasswordResetOTP(t *testing.T) {
	t.Setenv("GIN_MODE", "release")
	deps := testutil.RequireAuthDocker(t)
	cfg, err := config.Load()
	require.NoError(t, err)

	var logBuf bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logBuf, nil))
	stack, err := httpserver.MountAuth(cfg, log, deps.Pool, deps.RDB)
	require.NoError(t, err)
	doc := loadAndValidateSpec(t, filepath.Join(testutil.BackendRoot(t), "api", "openapi", "auth.yaml"))

	phone := fmt.Sprintf("+7916%07d", time.Now().UnixNano()%10000000)
	regRaw, _ := json.Marshal(map[string]any{"phone": phone})
	regReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/code", bytes.NewReader(regRaw))
	regReq.Header.Set("Content-Type", "application/json")
	regReq.RemoteAddr = "127.0.0.1:12346"
	w := httptest.NewRecorder()
	stack.Engine.ServeHTTP(w, regReq)
	regResp := w.Result()
	rb, _ := io.ReadAll(regResp.Body)
	_ = regResp.Body.Close()
	require.Equal(t, http.StatusOK, regResp.StatusCode)

	var regOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(rb, &regOut))
	all := extOTPRe.FindAllStringSubmatch(logBuf.String(), -1)
	loginCode := all[len(all)-1][1]

	verifyBody := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, regOut.VerificationID, loginCode)
	vReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/verify", bytes.NewReader([]byte(verifyBody)))
	vReq.Header.Set("Content-Type", "application/json")
	vReq.RemoteAddr = "127.0.0.1:12346"
	wv := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wv, vReq)
	vResp := wv.Result()
	vb, _ := io.ReadAll(vResp.Body)
	_ = vResp.Body.Close()
	require.Equal(t, http.StatusOK, vResp.StatusCode)

	var tok struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.Unmarshal(vb, &tok))

	cpRaw, _ := json.Marshal(map[string]any{"new_password": "correcthorse1"})
	cpReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/change-password", bytes.NewReader(cpRaw))
	cpReq.Header.Set("Content-Type", "application/json")
	cpReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	cpReq.RemoteAddr = "127.0.0.1:12346"
	wcp := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wcp, cpReq)
	cpResp := wcp.Result()
	cpb, _ := io.ReadAll(cpResp.Body)
	_ = cpResp.Body.Close()
	require.Equal(t, http.StatusOK, cpResp.StatusCode, string(cpb))

	forgotRaw, _ := json.Marshal(map[string]any{"phone": phone})
	forgotReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/forgot-password/phone", bytes.NewReader(forgotRaw))
	forgotReq.Header.Set("Content-Type", "application/json")
	forgotReq.RemoteAddr = "127.0.0.1:12346"
	wf := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wf, forgotReq)
	forgotResp := wf.Result()
	fb, _ := io.ReadAll(forgotResp.Body)
	_ = forgotResp.Body.Close()
	require.Equal(t, http.StatusOK, forgotResp.StatusCode, string(fb))

	var forgotOut struct {
		VerificationID string `json:"verification_id"`
	}
	require.NoError(t, json.Unmarshal(fb, &forgotOut))
	require.NotEmpty(t, forgotOut.VerificationID)

	all = extOTPRe.FindAllStringSubmatch(logBuf.String(), -1)
	resetSMS := all[len(all)-1][1]

	malVerify := fmt.Sprintf(`{"verification_id":"%s","code":"%s"}`, forgotOut.VerificationID, resetSMS)
	badReq := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/verify", bytes.NewReader([]byte(malVerify)))
	badReq.Header.Set("Content-Type", "application/json")
	badReq.RemoteAddr = "127.0.0.1:12346"
	wbad := httptest.NewRecorder()
	stack.Engine.ServeHTTP(wbad, badReq)
	badResp := wbad.Result()
	badb, _ := io.ReadAll(badResp.Body)
	_ = badResp.Body.Close()
	require.Equal(t, http.StatusConflict, badResp.StatusCode, string(badb))
	valBad, err := http.NewRequest(http.MethodPost, "http://127.0.0.1/v1/auth/phone/verify", bytes.NewReader([]byte(malVerify)))
	require.NoError(t, err)
	valBad.Header.Set("Content-Type", "application/json")
	validateResponse(t, doc, valBad, badResp, badb)
}
