package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Mrilki/catalog-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func RateLimiterMiddleware(
	client *redis.Client,
	limit int,
	window time.Duration,
	log *logger.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		clientIP := c.ClientIP()
		key := "rate_limit:" + clientIP

		current, err := client.Incr(ctx, key).Result()
		if err != nil {
			log.Error("Failed to increment rate limit", zap.Error(err))
			c.Next()
			return
		}

		if current == 1 {
			if err := client.Expire(ctx, key, window).Err(); err != nil {
				log.Error("Failed to set rate limit expiry", zap.Error(err))
			}
		}

		if current > int64(limit) {
			log.Warn("Rate limit exceeded",
				zap.String("ip", clientIP),
				zap.Int64("current", current),
				zap.Int("limit", limit))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", string(rune(limit)))
		c.Header("X-RateLimit-Remaining", string(rune(limit-int(current))))

		c.Next()
	}
}

func AdminRateLimiterMiddleware(
	client *redis.Client,
	limit int,
	window time.Duration,
	log *logger.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()

		clientIP := c.ClientIP()
		key := "rate_limit:admin:" + clientIP

		current, err := client.Incr(ctx, key).Result()
		if err != nil {
			log.Error("Failed to increment admin rate limit", zap.Error(err))
			c.Next()
			return
		}

		if current == 1 {
			if err := client.Expire(ctx, key, window).Err(); err != nil {
				log.Error("Failed to set admin rate limit expiry", zap.Error(err))
			}
		}

		if current > int64(limit) {
			log.Warn("Admin rate limit exceeded",
				zap.String("ip", clientIP),
				zap.Int64("current", current),
				zap.Int("limit", limit))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many admin requests",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
