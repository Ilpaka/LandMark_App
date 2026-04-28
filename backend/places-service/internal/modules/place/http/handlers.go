package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/app"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/domain"
)

// Handlers holds application service and config.
type Handlers struct {
	Svc         *app.Service
	InternalKey string
}

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

func userIDFromCtx(c *gin.Context) (*uuid.UUID, string) {
	raw := c.GetHeader("X-User-Id")
	role := c.GetHeader("X-Role")
	if raw == "" {
		return nil, role
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, role
	}
	return &id, role
}

func (h *Handlers) ListCategories(c *gin.Context) {
	cats, err := h.Svc.ListCategories(c.Request.Context())
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

func (h *Handlers) ListPlaces(c *gin.Context) {
	f := domain.ListFilter{
		Q:            c.Query("q"),
		CategorySlug: c.Query("category"),
	}
	if lim, err := strconv.Atoi(c.Query("limit")); err == nil {
		f.Limit = lim
	}
	if bbox := c.Query("bbox"); bbox != "" {
		parts := strings.Split(bbox, ",")
		if len(parts) == 4 {
			var b domain.BBox
			b.SouthLat, _ = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
			b.WestLng, _ = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			b.NorthLat, _ = strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
			b.EastLng, _ = strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
			f.BBox = &b
		}
	}
	places, err := h.Svc.ListPlaces(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"places": places})
}

func (h *Handlers) GetPlace(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	userID, role := userIDFromCtx(c)
	p, err := h.Svc.GetPlace(c.Request.Context(), id, userID, role)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

type createPlaceReq struct {
	Title     string  `json:"title" binding:"required,min=1,max=120"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (h *Handlers) CreatePlace(c *gin.Context) {
	userID, _ := userIDFromCtx(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req createPlaceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	p, err := h.Svc.CreateDraft(c.Request.Context(), *userID, req.Title, req.Latitude, req.Longitude)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handlers) UpdatePlace(c *gin.Context) {
	userID, _ := userIDFromCtx(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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
	p, err := h.Svc.UpdateDraft(c.Request.Context(), *userID, id, updates)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handlers) SubmitPlace(c *gin.Context) {
	userID, _ := userIDFromCtx(c)
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	if err := h.Svc.Submit(c.Request.Context(), *userID, id); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) ApprovePlace(c *gin.Context) {
	if c.GetHeader("X-Internal-Key") != h.InternalKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	adminID := uuid.New()
	if err := h.Svc.ApprovePlace(c.Request.Context(), id, adminID); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handlers) RejectPlace(c *gin.Context) {
	if c.GetHeader("X-Internal-Key") != h.InternalKey {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation_error"})
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)
	adminID := uuid.New()
	if err := h.Svc.RejectPlace(c.Request.Context(), id, adminID, req.Reason); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
