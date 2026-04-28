package notifhttp

import "github.com/gin-gonic/gin"

func Mount(r *gin.Engine, h *Handler) {
	v1 := r.Group("/v1/notifications")
	v1.GET("", h.List)
	v1.POST("/:id/read", h.MarkRead)
	v1.POST("/read-all", h.MarkAllRead)
	v1.POST("/devices", h.RegisterDevice)
	v1.DELETE("/devices/:id", h.UnregisterDevice)
	v1.GET("/preferences", h.GetPreferences)
	v1.PATCH("/preferences", h.UpdatePreferences)

	internal := r.Group("/v1/internal/notifications", internalKeyMiddleware())
	internal.POST("/send", h.InternalSend)
}
