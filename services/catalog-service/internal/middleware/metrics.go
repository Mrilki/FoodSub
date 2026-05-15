package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalog_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "catalog_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	cacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "catalog_cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	cacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "catalog_cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	kafkaEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "catalog_kafka_events_total",
			Help: "Total number of Kafka events processed",
		},
		[]string{"event_type", "status"},
	)
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, string(rune(status))).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

func IncCacheHit() {
	cacheHitsTotal.Inc()
}

func IncCacheMiss() {
	cacheMissesTotal.Inc()
}

func IncKafkaEvent(eventType, status string) {
	kafkaEventsTotal.WithLabelValues(eventType, status).Inc()
}
