package handlers

import (
	"auth/internal/cache"
	"auth/internal/common"
	"auth/internal/configs"
	"auth/internal/errs"
	"auth/internal/models"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleRefresh(env *configs.Env, cc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		refreshToken, err := c.Cookie("JWT_REFRESH_SECRET")
		if err != nil || refreshToken == "" {
			logger.Warn("Refresh request missing refresh token")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			return
		}

		var payload models.MinimalUserStruct
		err = common.VerifyJWT(c.Request.Context(), cc, refreshToken, "refresh", env.JWT_REFRESH_KEY, &payload)
		if errors.Is(err, errs.ERR_INVALID_METHOD) {
			logger.Warn("Refresh request provided a token with an invalid signing method")
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid refresh token"})
			return
		}
		if err != nil {
			logger.Warn("Refresh request provided an invalid or expired refresh token", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			return
		}

		if err := common.BlacklistToken(c.Request.Context(), cc, refreshToken, "refresh", time.Until(payload.ExpiresAt.Time)); err != nil {
			logger.Error("Failed to blacklist refresh token", "error", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		if err := common.HandleLoginActivity(c, payload, env); err != nil {
			logger.Error("Failed to refresh login session", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		logger.Info("Successfully refreshed login session")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Session refreshed"})
	}
}
