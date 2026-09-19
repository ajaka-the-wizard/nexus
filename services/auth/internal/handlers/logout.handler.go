package handlers

import (
	"auth/internal/cache"
	"auth/internal/common"
	"auth/internal/configs"
	"auth/internal/errs"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleLogout(env *configs.Env, cc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		if err := common.HandleLogoutActivity(c, cc, env); err != nil {
			if errors.Is(err, errs.ERR_INVALID_METHOD) || errors.Is(err, errs.ERR_NO_TOKENS_PROVIDED) {
				logger.Warn("Logout request provided a token with an invalid signing method")
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid or missing token"})
				return
			}
			if errors.Is(err, errs.ERR_BLACKLISTED_TOKEN) {
				logger.Warn("Logout request provided a blacklisted token")
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
				return
			}
			logger.Error("Failed to blacklist logout tokens", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		logger.Info("Successfully logged out user")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Logout successful"})
	}
}
