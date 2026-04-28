package httpserver

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const ginCtxRequestID = "request_id"

var (
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_http_requests_total",
			Help: "HTTP requests handled by the auth API",
		},
		[]string{"method", "route", "status"},
	)
	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "route"},
	)
)

// RequestIDMiddleware sets X-Request-ID (from client or generated) and exposes it on the Gin context.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(ginCtxRequestID, rid)
		c.Writer.Header().Set("X-Request-ID", rid)
		c.Next()
	}
}

// PrometheusHTTPMiddleware records request counts and latency after routing.
func PrometheusHTTPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		httpRequests.WithLabelValues(method, route, status).Inc()
		httpDuration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
	}
}
