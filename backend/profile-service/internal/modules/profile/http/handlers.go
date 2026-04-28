package profilehttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/profile-service/internal/modules/profile/app"
	"github.com/ilpaka/landmark_app/backend/profile-service/internal/modules/profile/domain"
)

type Handler struct{ Svc *app.Service }

func userID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetHeader("X-User-Id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}
	return id, true
}

func writeErr(c *gin.Context, err error) {
	switch err {
	case domain.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case domain.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": "nickname taken"})
	case domain.ErrBadRequest:
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}

func (h *Handler) GetMe(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	p, err := h.Svc.GetProfile(c.Request.Context(), uid)
	if err == domain.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		return
	}
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		Nickname    *string `json:"nickname"`
		DisplayName *string `json:"display_name"`
		Bio         *string `json:"bio"`
		City        *string `json:"city"`
		Country     *string `json:"country"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	p, err := h.Svc.UpdateProfile(c.Request.Context(), uid, domain.UpdateProfileInput{
		Nickname: body.Nickname, DisplayName: body.DisplayName,
		Bio: body.Bio, City: body.City, Country: body.Country,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) GetPublic(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad user_id"})
		return
	}
	p, err := h.Svc.GetProfile(c.Request.Context(), targetID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) GetPrivacy(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	priv, err := h.Svc.GetPrivacy(c.Request.Context(), uid)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, priv)
}

func (h *Handler) UpdatePrivacy(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		ProfileVisibility *string `json:"profile_visibility"`
		TripsVisibility   *string `json:"trips_visibility"`
		JournalVisibility *string `json:"journal_visibility"`
		AnalyticsEnabled  *bool   `json:"analytics_enabled"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	priv, err := h.Svc.UpdatePrivacy(c.Request.Context(), uid, domain.UpdatePrivacyInput{
		ProfileVisibility: body.ProfileVisibility,
		TripsVisibility:   body.TripsVisibility,
		JournalVisibility: body.JournalVisibility,
		AnalyticsEnabled:  body.AnalyticsEnabled,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, priv)
}
