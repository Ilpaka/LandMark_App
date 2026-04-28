package domain

// Phone OTP purpose distinguishes login SMS from password reset SMS.
const (
	PhoneOTPPurposeLogin          = "login"
	PhoneOTPPurposePasswordReset = "password_reset"
)
