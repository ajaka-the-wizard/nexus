package middlewares

import (
	"gateway/internal/cache"
	"gateway/internal/common"
	"gateway/internal/configs"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func RateLimit(cfg *configs.ServiceConfig, redisCache *cache.Redis) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		path, _ := strings.CutPrefix(c.Request.URL.Path, "/api")
		service, serviceKey, err := common.GetServiceFromServiceConfig(cfg, path)
		if err != nil {
			logger.Warn("Failed to resolve service for rate limiting", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			c.Abort()
			return
		}

		key := "limiter:{" + c.ClientIP() + ":" + serviceKey + "}"

		allowed, err := redisCache.Allow(
			c.Request.Context(),
			logger,
			key,
			service.RateLimitCost,
		)
		if err != nil {
			logger.Error("Failed to check rate limit", "service", serviceKey, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			c.Abort()
			return
		}
		c.Header("X-RateLimit-Remaining", strconv.FormatUint(allowed.Remaining, 10))

		if !allowed.Allowed {
			logger.Warn("Request rejected by rate limit", "service", serviceKey, "remaining", allowed.Remaining, "retryAfter", allowed.RetryAfter)
			c.Header("Retry-After", strconv.FormatUint(allowed.RetryAfter, 10))
			c.JSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "Too many requests"})
			c.Abort()
			return
		}

		logger.Info("Request allowed by rate limit", "service", serviceKey, "remaining", allowed.Remaining)
		c.Next()
	}
}
