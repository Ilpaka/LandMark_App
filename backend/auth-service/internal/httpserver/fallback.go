package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterFallbackHandlers makes unmatched routes and methods answer with the
// unified JSON error envelope instead of gin's plain-text defaults (BUG-03).
func RegisterFallbackHandlers(engine *gin.Engine) {
	engine.HandleMethodNotAllowed = true
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})
	engine.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	})
}
