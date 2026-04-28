package authhttp

import (
	"github.com/gin-gonic/gin"

	redisx "github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/adapters/redis"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/app"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/ports"
)

// Mount registers auth HTTP routes on r (typically /v1).
func Mount(r *gin.RouterGroup, svc *app.Service, jwt ports.JWTSigner, rdb *redisx.Client) {
	h := &Handlers{Svc: svc}
	authz := AuthMiddleware(jwt, rdb)

	r.POST("/auth/register", h.Register)
	r.POST("/auth/phone/code", h.PhoneRequestCode)
	r.POST("/auth/phone/verify", h.PhoneVerify)
	r.POST("/auth/phone/resend", h.PhoneResend)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/verify-email", h.VerifyEmail)
	r.POST("/auth/refresh", h.Refresh)
	r.POST("/auth/logout", h.Logout)

	g := r.Group("/auth")
	g.Use(authz)
	g.GET("/me", h.Me)
	g.POST("/me/email", h.StartEmailVerification)
	g.POST("/me/email/resend", h.ResendEmailVerification)
	g.POST("/logout-all", h.LogoutAll)
	g.GET("/sessions", h.Sessions)
	g.DELETE("/sessions/:id", h.DeleteSession)
	g.POST("/change-password", h.ChangePassword)
	g.DELETE("/account", h.DeleteAccount)

	r.POST("/auth/forgot-password", h.ForgotPassword)
	r.POST("/auth/forgot-password/phone", h.ForgotPasswordPhone)
	r.POST("/auth/reset-password", h.ResetPassword)
	r.POST("/auth/reset-password/phone", h.ResetPasswordPhone)

	adm := r.Group("/admin")
	adm.Use(authz, RequireAdmin())
	adm.POST("/users/:id/block", h.AdminBlock)
	adm.POST("/users/:id/unblock", h.AdminUnblock)
	adm.POST("/users/:id/force-logout", h.AdminForceLogout)
	adm.GET("/users/:id/audit", h.AdminAudit)
}
