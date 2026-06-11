package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ilpaka/landmark_app/backend/trips-service/internal/modules/trip/app"
)

type Stack struct{ Engine *gin.Engine }

func Mount(svc *app.Service) *Stack {
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

	h := &Handlers{Svc: svc}

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1 := r.Group("/v1")
	{
		v1.GET("/trips/public", h.ListPublicTrips)
		v1.GET("/trips", h.ListTrips)
		v1.POST("/trips", h.CreateTrip)
		v1.GET("/trips/:id", h.GetTrip)
		v1.PATCH("/trips/:id", h.UpdateTrip)
		v1.DELETE("/trips/:id", h.DeleteTrip)
		v1.GET("/trips/:id/route", h.GetRoute)
		v1.POST("/trips/:id/optimize", h.OptimizeRoute)
		v1.GET("/trips/:id/stops", h.ListStops)
		v1.POST("/trips/:id/stops", h.AddStop)
		v1.POST("/trips/:id/stops/:sid/visit", h.MarkVisited)
	}

	return &Stack{Engine: r}
}
