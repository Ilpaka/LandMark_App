package mediahttp

import "github.com/gin-gonic/gin"

func Mount(r *gin.Engine, h *Handler) {
	v1 := r.Group("/v1/media")
	v1.POST("/uploads", h.RequestUpload)
	v1.POST("/uploads/:id/finalize", h.FinalizeUpload)
	v1.GET("/:id", h.GetMedia)
	v1.GET("/:id/url", h.GetMediaURL)
	v1.DELETE("/:id", h.DeleteMedia)
}
