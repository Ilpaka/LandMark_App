package modhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/app"
	"github.com/ilpaka/landmark_app/backend/moderation-service/internal/modules/moderation/domain"
)

type Handler struct{ Svc *app.Service }

func requireAdmin(c *gin.Context) bool {
	if c.GetHeader("X-Role") != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return false
	}
	return true
}

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
	case domain.ErrConflict:
		c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}

func (h *Handler) Submit(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	targetID, err := uuid.Parse(body.TargetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad target_id"})
		return
	}
	item, err := h.Svc.SubmitItem(c.Request.Context(), domain.TargetType(body.TargetType), targetID, uid)
	if err == domain.ErrConflict {
		c.JSON(http.StatusOK, item)
		return
	}
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) ListQueue(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	f := domain.ListFilter{Limit: 20, Cursor: c.Query("cursor")}
	page, err := h.Svc.ListPending(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *Handler) Approve(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	uid, ok := userID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Svc.Approve(c.Request.Context(), uid, id); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Reject(c *gin.Context) {
	if !requireAdmin(c) {
		return
	}
	uid, ok := userID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	c.ShouldBindJSON(&body)
	if err := h.Svc.Reject(c.Request.Context(), uid, id, body.Note); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// InternalSubmit is called by places-service to enqueue a place for moderation
func (h *Handler) InternalSubmit(c *gin.Context) {
	var body struct {
		TargetType  string `json:"target_type"`
		TargetID    string `json:"target_id"`
		SubmittedBy string `json:"submitted_by"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	targetID, err := uuid.Parse(body.TargetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad target_id"})
		return
	}
	submittedBy, err := uuid.Parse(body.SubmittedBy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad submitted_by"})
		return
	}
	item, err := h.Svc.SubmitItem(c.Request.Context(), domain.TargetType(body.TargetType), targetID, submittedBy)
	if err == domain.ErrConflict {
		c.JSON(http.StatusOK, item)
		return
	}
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}
