package profilehttp

import "github.com/gin-gonic/gin"

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
}
