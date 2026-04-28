package favhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilpaka/landmark_app/backend/favorites-service/internal/modules/favorite/app"
	"github.com/ilpaka/landmark_app/backend/favorites-service/internal/modules/favorite/domain"
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
		c.JSON(http.StatusConflict, gin.H{"error": "conflict"})
	case domain.ErrForbidden:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case domain.ErrBadRequest:
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
	}
}

func (h *Handler) ListCollections(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	cols, err := h.Svc.ListCollections(c.Request.Context(), uid)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"collections": cols})
}

func (h *Handler) CreateCollection(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		Title string `json:"title"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	col, err := h.Svc.CreateCollection(c.Request.Context(), uid, body.Title, body.Color)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, col)
}

func (h *Handler) UpdateCollection(c *gin.Context) {
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
		Title string `json:"title"`
		Color string `json:"color"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	col, err := h.Svc.UpdateCollection(c.Request.Context(), uid, id, body.Title, body.Color)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, col)
}

func (h *Handler) DeleteCollection(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Svc.DeleteCollection(c.Request.Context(), uid, id); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListFavorites(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	f := domain.ListFavoritesFilter{UserID: uid, Limit: 20}
	if tt := c.Query("target_type"); tt != "" {
		t := domain.TargetType(tt)
		f.TargetType = &t
	}
	if colStr := c.Query("collection_id"); colStr != "" {
		colID, err := uuid.Parse(colStr)
		if err == nil {
			f.CollectionID = &colID
		}
	}
	f.Cursor = c.Query("cursor")
	page, err := h.Svc.ListFavorites(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

func (h *Handler) AddFavorite(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	var body struct {
		TargetType   string  `json:"target_type"`
		TargetID     string  `json:"target_id"`
		CollectionID *string `json:"collection_id"`
		Note         *string `json:"note"`
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
	var colID *uuid.UUID
	if body.CollectionID != nil {
		id, err := uuid.Parse(*body.CollectionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad collection_id"})
			return
		}
		colID = &id
	}
	if err := h.Svc.AddFavorite(c.Request.Context(), uid, domain.TargetType(body.TargetType), targetID, colID, body.Note); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusCreated)
}

func (h *Handler) RemoveFavorite(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	tt := domain.TargetType(c.Param("type"))
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	if err := h.Svc.RemoveFavorite(c.Request.Context(), uid, tt, targetID); err != nil {
		writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) CheckFavorite(c *gin.Context) {
	uid, ok := userID(c)
	if !ok {
		return
	}
	tt := domain.TargetType(c.Param("type"))
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}
	fav, err := h.Svc.Store.GetFavorite(c.Request.Context(), uid, tt, targetID)
	if err == domain.ErrNotFound {
		c.JSON(http.StatusOK, gin.H{"is_favorite": false})
		return
	}
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"is_favorite": true, "favorite": fav})
}
