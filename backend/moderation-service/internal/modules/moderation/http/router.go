package modhttp

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

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

func Mount(r *gin.Engine, h *Handler) {
	v1 := r.Group("/v1")
	v1.POST("/moderation/submit", h.Submit)

	admin := v1.Group("/moderation")
	admin.GET("/stats", h.GetStats)
	admin.GET("/queue", h.ListQueue)
	admin.POST("/queue/:id/approve", h.Approve)
	admin.POST("/queue/:id/reject", h.Reject)

	internal := r.Group("/v1/internal/moderation", internalKeyMiddleware())
	internal.POST("/submit", h.InternalSubmit)
}
