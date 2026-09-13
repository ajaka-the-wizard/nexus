package middlewares

import (
	"gateway/internal/configs"
	"log/slog"
	"uuid"

	"github.com/gin-gonic/gin"
)

func GenerateRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := uuid.NewV7()
		c.Request.Header.Add("X-REQUEST-ID", reqId.String())
		c.Writer.Header().Set("X-REQUEST-ID", reqId.String())
		c.Set("requestId", reqId.String())
		c.Next()
	}
}

func AttachScopedLogger(env *configs.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := c.GetString("requestId")

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
