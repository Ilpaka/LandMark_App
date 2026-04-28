package authhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
)

func TestWriteErr_StatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		err  error
		want int
	}{
		{domain.ErrBadRequest, http.StatusBadRequest},
		{domain.ErrConflict, http.StatusConflict},
		{domain.ErrNicknameUnavailable, http.StatusConflict},
		{domain.ErrInvalidCredentials, http.StatusUnauthorized},
		{domain.ErrInvalidOTP, http.StatusBadRequest},
		{domain.ErrTooManyAttempts, http.StatusBadRequest},
		{domain.ErrAccountBlocked, http.StatusForbidden},
		{domain.ErrEmailNotVerified, http.StatusForbidden},
		{domain.ErrInvalidRefresh, http.StatusUnauthorized},
		{domain.ErrRefreshReuse, http.StatusUnauthorized},
		{domain.ErrRateLimited, http.StatusTooManyRequests},
		{domain.ErrServiceUnavailable, http.StatusServiceUnavailable},
		{domain.ErrForbidden, http.StatusForbidden},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		writeErr(c, tc.err)
		require.Equalf(t, tc.want, w.Code, "err=%v", tc.err)
	}
}
