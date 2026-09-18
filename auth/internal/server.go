package internal

import (
	"auth/internal/cache"
	"auth/internal/configs"
	"auth/internal/middlewares"
	"auth/internal/repositories"
	"auth/internal/routes"
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func Listen() error {
	ctx := context.Background()
	logger := slog.Default()

	env := configs.LoadEnv(logger)
	pool := configs.ConnectDB(ctx, logger, env.DATABASE_URL)

	repo := repositories.InitRepository(pool)
	cache := cache.Initcache(ctx, env, logger)

	router := gin.New()
	router.Use(gin.Recovery())
	router.SetTrustedProxies(nil)

	router.Use(middlewares.AttachScopedLogger(env))

	g := router.Group("/auth")

	routes.MountRoutes(g, repo, cache, env)

	return router.Run(env.SERVER_ADDR)
}
