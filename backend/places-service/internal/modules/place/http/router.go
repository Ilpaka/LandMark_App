package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ilpaka/landmark_app/backend/places-service/internal/modules/place/app"
)

// Stack wraps the gin engine.
type Stack struct{ Engine *gin.Engine }

// Mount registers all routes and returns the stack.
func Mount(svc *app.Service, internalKey string) *Stack {
	r := gin.New()
	r.Use(gin.Recovery())

	h := &Handlers{Svc: svc, InternalKey: internalKey}

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1 := r.Group("/v1")
	{
		v1.GET("/places/categories", h.ListCategories)
		v1.GET("/places", h.ListPlaces)
		v1.GET("/places/:id", h.GetPlace)
		v1.POST("/places", h.CreatePlace)
		v1.PATCH("/places/:id", h.UpdatePlace)
		v1.POST("/places/:id/submit", h.SubmitPlace)

		internal := v1.Group("/places/internal")
		{
			internal.POST("/:id/approve", h.ApprovePlace)
			internal.POST("/:id/reject", h.RejectPlace)
		}
	}

	return &Stack{Engine: r}
}
