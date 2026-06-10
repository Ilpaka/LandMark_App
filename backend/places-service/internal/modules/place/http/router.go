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

	// Unified JSON error envelope for unmatched routes and methods (BUG-03)
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})

	h := &Handlers{Svc: svc, InternalKey: internalKey}

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1 := r.Group("/v1")
	{
		v1.GET("/places/categories", h.ListCategories)
		v1.POST("/places/categories", h.CreateCategory)
		v1.PATCH("/places/categories/:id", h.UpdateCategory)
		v1.DELETE("/places/categories/:id", h.DeleteCategory)
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
