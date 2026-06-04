package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/app"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/domain"
)

type Handlers struct{ Svc *app.Service }

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
	case errors.Is(err, domain.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, domain.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
	case errors.Is(err, domain.ErrBadRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}

func userIDFromCtx(c *gin.Context) *uuid.UUID {
	raw := c.GetHeader("X-User-Id")
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func requireUser(c *gin.Context) (*uuid.UUID, bool) {
	id := userIDFromCtx(c)
	if id == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, false
	}
	return id, true
}

// ListEntries handles GET /v1/journal/entries
func (h *Handlers) ListEntries(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	var tripID *uuid.UUID
	if raw := c.Query("trip_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
			return
		}
		tripID = &id
	}
	q := c.Query("q")
	cursor := c.Query("cursor")
	limit := 20
	if l, err := strconv.Atoi(c.Query("limit")); err == nil {
		limit = l
	}
	entries, next, err := h.Svc.List(c.Request.Context(), *userID, tripID, q, cursor, limit)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries, "next_cursor": next})
}

// CreateEntry handles POST /v1/journal/entries
func (h *Handlers) CreateEntry(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	var req struct {
		TripID     *uuid.UUID `json:"trip_id"`
		PlaceID    *uuid.UUID `json:"place_id"`
		Title      *string    `json:"title"`
		Body       string     `json:"body"`
		Mood       *string    `json:"mood"`
		Rating     *int       `json:"rating"`
		OccurredAt *time.Time `json:"occurred_at"`
		Tags       []string   `json:"tags"`
		Latitude   *float64   `json:"latitude"`
		Longitude  *float64   `json:"longitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	input := domain.Entry{
		TripID:    req.TripID,
		PlaceID:   req.PlaceID,
		Title:     req.Title,
		Body:      req.Body,
		Mood:      req.Mood,
		Rating:    req.Rating,
		Tags:      req.Tags,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}
	if req.OccurredAt != nil {
		input.OccurredAt = *req.OccurredAt
	}
	e, err := h.Svc.Create(c.Request.Context(), *userID, input)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, e)
}

// GetEntry handles GET /v1/journal/entries/:id
func (h *Handlers) GetEntry(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	e, err := h.Svc.Get(c.Request.Context(), *userID, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, e)
}

// UpdateEntry handles PATCH /v1/journal/entries/:id
func (h *Handlers) UpdateEntry(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	e, err := h.Svc.Update(c.Request.Context(), *userID, id, updates)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, e)
}

// DeleteEntry handles DELETE /v1/journal/entries/:id
func (h *Handlers) DeleteEntry(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	if err := h.Svc.Delete(c.Request.Context(), *userID, id); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListMedia handles GET /v1/journal/entries/:id/media
func (h *Handlers) ListMedia(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	mediaIDs, err := h.Svc.ListMedia(c.Request.Context(), *userID, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"media_ids": mediaIDs})
}

// AddMedia handles POST /v1/journal/entries/:id/media
func (h *Handlers) AddMedia(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	var req struct {
		MediaID   uuid.UUID `json:"media_id" binding:"required"`
		SortOrder int       `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	if err := h.Svc.AddMedia(c.Request.Context(), *userID, id, req.MediaID, req.SortOrder); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

// RemoveMedia handles DELETE /v1/journal/entries/:id/media/:mid
func (h *Handlers) RemoveMedia(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	mid, err := uuid.Parse(c.Param("mid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	if err := h.Svc.RemoveMedia(c.Request.Context(), *userID, id, mid); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ListReactions handles GET /v1/journal/entries/:id/reactions
func (h *Handlers) ListReactions(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	counts, mine, err := h.Svc.ListReactions(c.Request.Context(), id, *userID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"counts": counts, "mine": mine})
}

// AddReaction handles POST /v1/journal/entries/:id/reactions
func (h *Handlers) AddReaction(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	var req struct {
		Emoji string `json:"emoji" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	if err := h.Svc.AddReaction(c.Request.Context(), *userID, id, req.Emoji); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

// RemoveReaction handles DELETE /v1/journal/entries/:id/reactions/:emoji
func (h *Handlers) RemoveReaction(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	emoji := c.Param("emoji")
	if err := h.Svc.RemoveReaction(c.Request.Context(), *userID, id, emoji); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
