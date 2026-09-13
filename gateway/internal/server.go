package internal

import (
	"gateway/internal/configs"
	"gateway/internal/middlewares"
	"gateway/internal/routes"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func Listen() error {
	// Initialize background loggers
	logger := slog.Default()

	// Load configurations
	env := configs.LoadEnv(logger)
	services := configs.LoadServiceConfig(logger)

	router := gin.New()
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	if env.PRODUCTION {
		gin.SetMode(gin.ReleaseMode)
	}

	// General middlewares
	router.Use(middlewares.GenerateRequestID())
	router.Use(middlewares.AttachScopedLogger(env))

	router.Use(routes.Proxy(services))

	return router.Run(env.SERVER_ADDR)
}
