package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ilpaka/landmark_app/backend/journal-service/internal/modules/journal/app"
)

type Stack struct{ Engine *gin.Engine }

func Mount(svc *app.Service) *Stack {
	r := gin.New()
	r.Use(gin.Recovery())

	h := &Handlers{Svc: svc}

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1 := r.Group("/v1")
	{
		entries := v1.Group("/journal/entries")
		{
			entries.GET("", h.ListEntries)
			entries.POST("", h.CreateEntry)
			entries.GET("/:id", h.GetEntry)
			entries.PATCH("/:id", h.UpdateEntry)
			entries.DELETE("/:id", h.DeleteEntry)
			entries.GET("/:id/media", h.ListMedia)
			entries.POST("/:id/media", h.AddMedia)
			entries.DELETE("/:id/media/:mid", h.RemoveMedia)
			entries.GET("/:id/reactions", h.ListReactions)
			entries.POST("/:id/reactions", h.AddReaction)
			entries.DELETE("/:id/reactions/:emoji", h.RemoveReaction)
		}
	}

	return &Stack{Engine: r}
}
