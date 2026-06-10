package profilehttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Mount(r *gin.Engine, h *Handler) {
	v1 := r.Group("/v1/profile")
	v1.GET("/me", h.GetMe)
	v1.PATCH("/me", h.UpdateMe)
	v1.GET("/me/privacy", h.GetPrivacy)
	v1.PATCH("/me/privacy", h.UpdatePrivacy)
	v1.GET("/:user_id", h.GetPublic)
	v1.GET("/:user_id/stats", h.GetFollowStats)
	v1.POST("/:user_id/follow", h.FollowUser)
	v1.DELETE("/:user_id/follow", h.UnfollowUser)
	v1.GET("/:user_id/followers", h.ListFollowers)
	v1.GET("/:user_id/following", h.ListFollowing)

	// Unified JSON error envelope for unmatched routes and methods (BUG-03)
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})
}
