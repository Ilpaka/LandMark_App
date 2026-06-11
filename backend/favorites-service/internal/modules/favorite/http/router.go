package favhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Mount(r *gin.Engine, h *Handler) {
	v1 := r.Group("/v1")
	v1.GET("/favorites", h.ListFavorites)
	v1.POST("/favorites", h.AddFavorite)
	v1.DELETE("/favorites/:type/:id", h.RemoveFavorite)
	v1.GET("/favorites/:type/:id", h.CheckFavorite)

	cols := v1.Group("/favorites/collections")
	cols.GET("", h.ListCollections)
	cols.POST("", h.CreateCollection)
	cols.PATCH("/:id", h.UpdateCollection)
	cols.DELETE("/:id", h.DeleteCollection)

	// Unified JSON error envelope for unmatched routes and methods (BUG-03)
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})
}
