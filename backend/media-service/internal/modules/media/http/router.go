package mediahttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Mount(r *gin.Engine, h *Handler) {
	v1 := r.Group("/v1/media")
	v1.POST("/uploads", h.RequestUpload)
	v1.POST("/uploads/:id/finalize", h.FinalizeUpload)
	v1.GET("/:id", h.GetMedia)
	v1.GET("/:id/url", h.GetMediaURL)
	v1.DELETE("/:id", h.DeleteMedia)

	// Unified JSON error envelope for unmatched routes and methods (BUG-03)
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})
}
