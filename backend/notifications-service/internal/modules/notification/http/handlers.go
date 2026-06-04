package notifhttp

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/app"
	"github.com/ilpaka/landmark_app/backend/notifications-service/internal/modules/notification/domain"
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
	case domain.ErrBadRequest:
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}

func internalKeyMiddleware() gin.HandlerFunc {
	key := os.Getenv("INTERNAL_API_KEY")
	return func(c *gin.Context) {
		if c.GetHeader("X-Internal-Key") != key {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func (h *Handler) List(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	f := domain.ListFilter{UserID: uid, Limit: 20}
	f.Cursor = c.Query("cursor")
	if c.Query("unread") == "true" {
		f.UnreadOnly = true
	}
	page, err := h.Svc.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	count, _ := h.Svc.UnreadCount(c.Request.Context(), uid)
	c.JSON(http.StatusOK, gin.H{"items": page.Items, "next_cursor": page.NextCursor, "unread_count": count})
}

func (h *Handler) MarkRead(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Svc.MarkRead(c.Request.Context(), id, uid); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	if err := h.Svc.MarkAllRead(c.Request.Context(), uid); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) RegisterDevice(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		Platform  string `json:"platform"`
		PushToken string `json:"push_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	dev, err := h.Svc.UpsertDevice(c.Request.Context(), uid, body.Platform, body.PushToken)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, dev)
}

func (h *Handler) UnregisterDevice(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Svc.DeleteDevice(c.Request.Context(), id, uid); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetPreferences(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	prefs, err := h.Svc.GetPreferences(c.Request.Context(), uid)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, prefs)
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	var patch map[string]any
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	prefs, err := h.Svc.UpdatePreferences(c.Request.Context(), uid, patch)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, prefs)
}

func (h *Handler) InternalSend(c *gin.Context) {
	var body struct {
		UserID   string         `json:"user_id"`
		Type     string         `json:"type"`
		Title    string         `json:"title"`
		Body     string         `json:"body"`
		DeepLink *string        `json:"deep_link"`
		Meta     map[string]any `json:"meta"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	uid, err := uuid.Parse(body.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad user_id"})
		return
	}
	n, err := h.Svc.Send(c.Request.Context(), domain.SendInput{
		UserID: uid, Type: body.Type, Title: body.Title, Body: body.Body,
		DeepLink: body.DeepLink, Meta: body.Meta,
	})
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, n)
}
