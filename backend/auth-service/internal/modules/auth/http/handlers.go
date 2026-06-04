package authhttp

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/app"
	"github.com/ilpaka/landmark_app/backend/auth-service/internal/modules/auth/domain"
)

type Handlers struct {
	Svc *app.Service
}

type registerReq struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=10,max=128"`
}

// Register POST /v1/auth/register — create account with email+password, sends verification OTP.
func (h *Handlers) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if !validPassword(req.Password) {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	verID, plainCode, err := h.Svc.RegisterEmail(c.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		writeErr(c, err)
		return
	}
	// dev_code возвращается для демо: код OTP сразу виден клиенту,
	// чтобы можно было показать его в push-уведомлении без почтовой инфраструктуры.
	c.JSON(http.StatusOK, gin.H{"status": "ok", "verification_id": verID.String(), "dev_code": plainCode})
}

func validPassword(s string) bool {
	if len(s) < 10 {
		return false
	}
	var letter, digit bool
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z':
			letter = true
		case r >= '0' && r <= '9':
			digit = true
		}
	}
	return letter && digit
}

type phoneCodeReq struct {
	Phone  string  `json:"phone" binding:"required"`
	Region *string `json:"region"`
}

// PhoneRequestCode POST /v1/auth/phone/code — send SMS OTP.
func (h *Handlers) PhoneRequestCode(c *gin.Context) {
	var req phoneCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	var def string
	if req.Region != nil {
		def = *req.Region
	}
	ip := c.ClientIP()
	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	out, err := h.Svc.RequestPhoneCode(c.Request.Context(), app.PhoneRequestInput{
		Phone:   req.Phone,
		IP:      ipPtr,
		Default: def,
	})
	if out != nil {
		WriteRateLimitHeaders(c, out.Rate)
	}
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"verification_id": out.VerificationID.String(),
	})
}

type verifyReq struct {
	VerificationID string `json:"verification_id" binding:"required,uuid"`
	Code           string `json:"code" binding:"required,len=6,numeric"`
}

func (h *Handlers) VerifyEmail(c *gin.Context) {
	var req verifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	id, err := uuid.Parse(req.VerificationID)
	if err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if err := h.Svc.VerifyEmail(c.Request.Context(), id, req.Code); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type phoneVerifyReq struct {
	VerificationID string  `json:"verification_id" binding:"required,uuid"`
	Code           string  `json:"code" binding:"required,len=6,numeric"`
	DeviceID       *string `json:"device_id"`
	DeviceName     *string `json:"device_name"`
	Platform       *string `json:"platform"`
}

type loginReq struct {
	Email      string  `json:"email" binding:"required,email"`
	Password   string  `json:"password" binding:"required,min=1"`
	DeviceID   *string `json:"device_id"`
	DeviceName *string `json:"device_name"`
	Platform   *string `json:"platform"`
}

// Login POST /v1/auth/login — email + password; requires verified email on the account.
func (h *Handlers) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	var ipPtr *string
	if ip := c.ClientIP(); ip != "" {
		ipPtr = &ip
	}
	ua := c.GetHeader("User-Agent")
	pair, rl, err := h.Svc.LoginWithPassword(c.Request.Context(), app.LoginPasswordInput{
		Email:      req.Email,
		Password:   req.Password,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
		Platform:   req.Platform,
		IP:         ipPtr,
		UserAgent:  &ua,
	})
	WriteRateLimitHeaders(c, rl)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.AccessExp.UTC().Format(time.RFC3339Nano),
	})
}

// PhoneVerify POST /v1/auth/phone/verify — issue tokens after valid SMS code.
func (h *Handlers) PhoneVerify(c *gin.Context) {
	var req phoneVerifyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	id, err := uuid.Parse(req.VerificationID)
	if err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	var ipPtr *string
	if ip := c.ClientIP(); ip != "" {
		ipPtr = &ip
	}
	ua := c.GetHeader("User-Agent")
	pair, rl, err := h.Svc.VerifyPhoneCode(c.Request.Context(), app.PhoneVerifyInput{
		VerificationID: id,
		Code:           req.Code,
		DeviceID:       req.DeviceID,
		DeviceName:     req.DeviceName,
		Platform:       req.Platform,
		IP:             ipPtr,
		UserAgent:      &ua,
	})
	WriteRateLimitHeaders(c, rl)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.AccessExp.UTC().Format(time.RFC3339Nano),
	})
}

type emailReq struct {
	Email string `json:"email" binding:"required,email"`
}

// PhoneResend POST /v1/auth/phone/resend — resend SMS (same body as phone/code).
func (h *Handlers) PhoneResend(c *gin.Context) {
	var req phoneCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	var def string
	if req.Region != nil {
		def = *req.Region
	}
	ip := c.ClientIP()
	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	rl, err := h.Svc.ResendPhoneCode(c.Request.Context(), app.PhoneRequestInput{Phone: req.Phone, IP: ipPtr, Default: def})
	WriteRateLimitHeaders(c, rl)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type refreshReq struct {
	RefreshToken string  `json:"refresh_token" binding:"required"`
	DeviceID     *string `json:"device_id"`
}

func (h *Handlers) Refresh(c *gin.Context) {
	var req refreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	var ipPtr *string
	if ip := c.ClientIP(); ip != "" {
		ipPtr = &ip
	}
	ua := c.GetHeader("User-Agent")
	pair, err := h.Svc.Refresh(c.Request.Context(), req.RefreshToken, req.DeviceID, ipPtr, &ua)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  pair.AccessToken,
		"refresh_token": pair.RefreshToken,
		"expires_at":    pair.AccessExp.UTC().Format(time.RFC3339Nano),
	})
}

type tokensReq struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handlers) Logout(c *gin.Context) {
	var req tokensReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if err := h.Svc.Logout(c.Request.Context(), req.AccessToken, req.RefreshToken); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) LogoutAll(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.Svc.LogoutAll(c.Request.Context(), cl.Sub, cl.JTI, cl.Expires); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Me returns user id, session, role, and account flags.
func (h *Handlers) Me(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	info, err := h.Svc.Me(c.Request.Context(), cl.Sub)
	if err != nil {
		writeErr(c, err)
		return
	}
	out := gin.H{
		"user_id":        cl.Sub.String(),
		"session_id":     cl.Session.String(),
		"role":           string(cl.Role),
		"has_password":   info.HasPassword,
		"email_verified": info.EmailVerified,
		"phone_verified": info.PhoneVerified,
		"email":          info.Email,
		"phone_e164":     info.PhoneE164,
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handlers) Sessions(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	list, err := h.Svc.ListSessions(c.Request.Context(), cl.Sub)
	if err != nil {
		writeErr(c, err)
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, s := range list {
		out = append(out, gin.H{
			"id":                     s.ID.String(),
			"session_family_id":      s.SessionFamilyID.String(),
			"device_id":              s.DeviceID,
			"device_name":            s.DeviceName,
			"platform":               s.Platform,
			"issued_at":              s.IssuedAt,
			"expires_at":             s.ExpiresAt,
			"last_seen_at":           s.LastSeenAt,
			"revoked_at":             s.RevokedAt,
			"revoke_reason":          s.RevokeReason,
			"replaced_by_session_id": s.ReplacedBySessionID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"sessions": out})
}

func (h *Handlers) DeleteSession(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if err := h.Svc.DeleteSession(c.Request.Context(), cl.Sub, sid); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// StartEmailVerification POST /v1/auth/me/email — send OTP to bind email (authenticated).
func (h *Handlers) StartEmailVerification(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req emailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	vid, err := h.Svc.StartEmailVerification(c.Request.Context(), cl.Sub, req.Email)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "verification_id": vid.String()})
}

// ResendEmailVerification POST /v1/auth/me/email/resend
func (h *Handlers) ResendEmailVerification(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rl, vid, err := h.Svc.ResendEmailVerification(c.Request.Context(), cl.Sub)
	WriteRateLimitHeaders(c, rl)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "verification_id": vid.String()})
}

func (h *Handlers) ForgotPassword(c *gin.Context) {
	var req emailReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	_ = h.Svc.ForgotPassword(c.Request.Context(), req.Email)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ForgotPasswordPhone POST /v1/auth/forgot-password/phone — SMS OTP for password reset (eligible accounts only).
func (h *Handlers) ForgotPasswordPhone(c *gin.Context) {
	var req phoneCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	var def string
	if req.Region != nil {
		def = *req.Region
	}
	ip := c.ClientIP()
	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	out, err := h.Svc.ForgotPasswordPhone(c.Request.Context(), app.PhoneRequestInput{
		Phone:   req.Phone,
		IP:      ipPtr,
		Default: def,
	})
	if out != nil {
		WriteRateLimitHeaders(c, out.Rate)
	}
	if err != nil {
		writeErr(c, err)
		return
	}
	if out != nil && out.VerificationID != nil {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "verification_id": out.VerificationID.String()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type resetReq struct {
	ResetID     string `json:"reset_id" binding:"required,uuid"`
	Code        string `json:"code" binding:"required,len=6,numeric"`
	NewPassword string `json:"new_password" binding:"required,min=10,max=128"`
}

func (h *Handlers) ResetPassword(c *gin.Context) {
	var req resetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	id, err := uuid.Parse(req.ResetID)
	if err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if !validPassword(req.NewPassword) {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if err := h.Svc.ResetPassword(c.Request.Context(), id, req.Code, req.NewPassword); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type resetPhoneReq struct {
	VerificationID string `json:"verification_id" binding:"required,uuid"`
	Code           string `json:"code" binding:"required,len=6,numeric"`
	NewPassword    string `json:"new_password" binding:"required,min=10,max=128"`
}

// ResetPasswordPhone POST /v1/auth/reset-password/phone — set new password after SMS reset OTP.
func (h *Handlers) ResetPasswordPhone(c *gin.Context) {
	var req resetPhoneReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	id, err := uuid.Parse(req.VerificationID)
	if err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if !validPassword(req.NewPassword) {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	var ipPtr *string
	if ip := c.ClientIP(); ip != "" {
		ipPtr = &ip
	}
	ua := c.GetHeader("User-Agent")
	rl, err := h.Svc.ResetPasswordPhone(c.Request.Context(), app.ResetPasswordPhoneInput{
		VerificationID: id,
		Code:           req.Code,
		NewPassword:    req.NewPassword,
		IP:             ipPtr,
		UserAgent:      &ua,
	})
	WriteRateLimitHeaders(c, rl)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type changePwdReq struct {
	CurrentPassword *string `json:"current_password"`
	NewPassword     string  `json:"new_password" binding:"required,min=10,max=128"`
}

// ChangePassword sets the first password when current_password is omitted, otherwise changes the password.
func (h *Handlers) ChangePassword(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req changePwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if !validPassword(req.NewPassword) {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	cur := ""
	if req.CurrentPassword != nil {
		cur = strings.TrimSpace(*req.CurrentPassword)
	}
	if cur == "" {
		if err := h.Svc.SetInitialPassword(c.Request.Context(), cl.Sub, req.NewPassword); err != nil {
			writeErr(c, err)
			return
		}
	} else {
		if err := h.Svc.ChangePassword(c.Request.Context(), cl.Sub, cur, req.NewPassword, cl.Session); err != nil {
			writeErr(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) DeleteAccount(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.Svc.DeleteAccount(c.Request.Context(), cl.Sub); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) AdminBlock(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	target, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	var body struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if err := h.Svc.AdminBlock(c.Request.Context(), cl.Sub, target, body.Reason); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) AdminUnblock(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	target, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if err := h.Svc.AdminUnblock(c.Request.Context(), cl.Sub, target); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) AdminForceLogout(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	target, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	if err := h.Svc.AdminForceLogout(c.Request.Context(), cl.Sub, target); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// AdminListUsers GET /v1/admin/users — paginated list of users for the admin
// dashboard. Query: ?q=fragment&limit=50&offset=0.
func (h *Handlers) AdminListUsers(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	q := c.Query("q")
	limit := atoiOr(c.Query("limit"), 50)
	offset := atoiOr(c.Query("offset"), 0)

	users, err := h.Svc.AdminListUsers(c.Request.Context(), cl.Sub, q, limit, offset)
	if err != nil {
		writeErr(c, err)
		return
	}

	out := make([]gin.H, 0, len(users))
	for _, u := range users {
		out = append(out, gin.H{
			"id":                u.ID.String(),
			"email":             nilStr(u.Email),
			"phone_e164":        nilStr(u.PhoneE164),
			"status":            string(u.Status),
			"role":              string(u.Role),
			"email_verified_at": nilTime(u.EmailVerifiedAt),
			"phone_verified_at": nilTime(u.PhoneVerifiedAt),
			"blocked_at":        nilTime(u.BlockedAt),
			"blocked_reason":    nilStr(u.BlockedReason),
			"last_login_at":     nilTime(u.LastLoginAt),
			"created_at":        u.CreatedAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"users":  out,
		"limit":  limit,
		"offset": offset,
	})
}

func atoiOr(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func nilStr(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func nilTime(p *time.Time) any {
	if p == nil {
		return nil
	}
	return p.Format(time.RFC3339)
}

func (h *Handlers) AdminAudit(c *gin.Context) {
	cl, ok := claimsFromCtx(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	target, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeErr(c, domain.ErrBadRequest)
		return
	}
	list, err := h.Svc.AdminListAudit(c.Request.Context(), cl.Sub, target, 100, 0)
	if err != nil {
		writeErr(c, err)
		return
	}

	out := make([]gin.H, 0, len(list))
	for _, e := range list {
		var userID any
		if e.UserID != nil {
			userID = e.UserID.String()
		}
		out = append(out, gin.H{
			"id":         e.ID.String(),
			"user_id":    userID,
			"event_type": e.EventType,
			"ip":         nilStr(e.IP),
			"user_agent": nilStr(e.UserAgent),
			"meta":       e.Meta,
			"created_at": e.CreatedAt.Format(time.RFC3339),
		})
	}
	c.JSON(http.StatusOK, gin.H{"events": out})
}
