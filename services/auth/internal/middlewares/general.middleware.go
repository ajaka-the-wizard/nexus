package middlewares

import (
	"auth/internal/configs"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AttachScopedLogger(env *configs.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := c.GetHeader("X_REQUEST_ID")
		if reqId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Request Id not found"})
			c.Abort()
			return
		}

		reqLogger := slog.Default().With(
			slog.String("requestId", reqId),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("Environment", env.ENVIRONMENT),
		)

		c.Set("logger", reqLogger)
		c.Next()
	}
}
