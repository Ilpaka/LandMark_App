package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/app"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/domain"
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

// ListTrips handles GET /v1/trips
func (h *Handlers) ListTrips(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	status := c.Query("status")
	q := c.Query("q")
	cursor := c.Query("cursor")
	limit := 20
	if l, err := strconv.Atoi(c.Query("limit")); err == nil {
		limit = l
	}
	trips, next, err := h.Svc.List(c.Request.Context(), *userID, status, q, cursor, limit)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"trips": trips, "next_cursor": next})
}

// CreateTrip handles POST /v1/trips
func (h *Handlers) CreateTrip(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	t, err := h.Svc.CreateDraft(c.Request.Context(), *userID, req.Title)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

// GetTrip handles GET /v1/trips/:id
func (h *Handlers) GetTrip(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	t, err := h.Svc.Get(c.Request.Context(), *userID, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

// UpdateTrip handles PATCH /v1/trips/:id
func (h *Handlers) UpdateTrip(c *gin.Context) {
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
	t, err := h.Svc.Update(c.Request.Context(), *userID, id, updates)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

// DeleteTrip handles DELETE /v1/trips/:id
func (h *Handlers) DeleteTrip(c *gin.Context) {
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

// ListStops handles GET /v1/trips/:id/stops
func (h *Handlers) ListStops(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	stops, err := h.Svc.ListStops(c.Request.Context(), *userID, tripID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"stops": stops})
}

// AddStop handles POST /v1/trips/:id/stops
func (h *Handlers) AddStop(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	var req struct {
		Title     string     `json:"title" binding:"required"`
		Latitude  float64    `json:"latitude"`
		Longitude float64    `json:"longitude"`
		PlaceID   *uuid.UUID `json:"place_id"`
		PlannedAt *time.Time `json:"planned_at"`
		Notes     *string    `json:"notes"`
		SortOrder int        `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	stop := domain.TripStop{
		Title:     req.Title,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		PlaceID:   req.PlaceID,
		PlannedAt: req.PlannedAt,
		Notes:     req.Notes,
		SortOrder: req.SortOrder,
	}
	st, err := h.Svc.AddStop(c.Request.Context(), *userID, tripID, stop)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, st)
}

// ListPublicTrips handles GET /v1/trips/public
func (h *Handlers) ListPublicTrips(c *gin.Context) {
	var excludeOwner uuid.UUID
	if raw := c.GetHeader("X-User-Id"); raw != "" {
		excludeOwner, _ = uuid.Parse(raw)
	}
	cursor := c.Query("cursor")
	limit := 20
	if l, err := strconv.Atoi(c.Query("limit")); err == nil {
		limit = l
	}
	trips, next, err := h.Svc.ListPublic(c.Request.Context(), excludeOwner, cursor, limit)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"trips": trips, "next_cursor": next})
}

// GetRoute handles GET /v1/trips/:id/route
func (h *Handlers) GetRoute(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	tripID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	coords, err := h.Svc.GetRoute(c.Request.Context(), *userID, tripID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"coordinates": coords})
}

// MarkVisited handles POST /v1/trips/:id/stops/:sid/visit
func (h *Handlers) MarkVisited(c *gin.Context) {
	userID, ok := requireUser(c)
	if !ok {
		return
	}
	stopID, err := uuid.Parse(c.Param("sid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	var req struct {
		VisitedAt *time.Time `json:"visited_at"`
	}
	_ = c.ShouldBindJSON(&req)
	at := time.Now()
	if req.VisitedAt != nil {
		at = *req.VisitedAt
	}
	if err := h.Svc.MarkVisited(c.Request.Context(), *userID, stopID, at); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
