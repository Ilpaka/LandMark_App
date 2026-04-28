package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	AuthRegisterTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_register_total",
		Help: "Successful user registrations",
	})
	AuthLoginTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_login_total",
		Help: "Successful logins",
	})
	AuthRefreshTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_refresh_total",
		Help: "Successful refresh rotations",
	})
	AuthRefreshReuseTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_refresh_reuse_total",
		Help: "Detected refresh token reuse events",
	})
	AuthPhoneOTPStoreErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "auth_phone_otp_store_errors_total",
		Help: "Phone OTP challenge store failures (create/verify)",
	})
	AuthPhoneOTPVerifyOutcomes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "auth_phone_otp_verify_outcomes_total",
		Help: "Phone OTP verify outcomes (result label)",
	}, []string{"result"})
)
