package internal

import (
	"context"
	"gateway/internal/cache"
	"gateway/internal/configs"
	"gateway/internal/middlewares"
	"gateway/internal/routes"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func Listen() error {
	// Initialize context
	ctx := context.Background()

	// Initialize background loggers
	logger := slog.Default()

	// Load configurations
	env := configs.LoadEnv(logger)
	cfg := configs.LoadServiceConfig(logger)

	// connect to services
	c := cache.InitializeRedis(ctx, env, logger)

	router := gin.New()
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	if env.PRODUCTION {
		gin.SetMode(gin.ReleaseMode)
	}

	// General middlewares
	router.Use(middlewares.GenerateRequestID())
	router.Use(middlewares.AttachScopedLogger(env))
	router.Use(middlewares.RateLimit(cfg, c))
	router.Use(middlewares.AuthenticatePrivateRoutes(env))

	router.Use(routes.Proxy(cfg))

	return router.Run(env.SERVER_ADDR)
}
