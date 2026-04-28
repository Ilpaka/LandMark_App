package mediahttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/media-service/internal/modules/media/app"
	"github.com/ilpaka/landmark_app/backend/media-service/internal/modules/media/domain"
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
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case domain.ErrBadRequest:
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}

func (h *Handler) RequestUpload(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		Kind      string `json:"kind"`
		Mime      string `json:"mime"`
		SizeBytes int64  `json:"size_bytes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	result, err := h.Svc.RequestUpload(c.Request.Context(), uid, body.Kind, body.Mime, body.SizeBytes)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) FinalizeUpload(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	media, err := h.Svc.FinalizeUpload(c.Request.Context(), id, uid)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, media)
}

func (h *Handler) GetMedia(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	m, err := h.Svc.GetMedia(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) GetMediaURL(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	u, err := h.Svc.GetMediaURL(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"url": u})
}

func (h *Handler) DeleteMedia(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Svc.DeleteMedia(c.Request.Context(), id, uid); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
