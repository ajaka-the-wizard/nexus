package handlers

import (
	"auth/internal/cache"
	"auth/internal/common"
	"auth/internal/configs"
	"auth/internal/errs"
	"auth/internal/models"
	"auth/internal/repositories"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func HandleVerifyUser(repo *repositories.Repository, env *configs.Env, cc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		verificationToken := c.Query("val")
		if verificationToken == "" {
			logger.Warn("Email verification request missing token")
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid verification token"})
			return
		}

		var payload models.LoneEmailPayload
		if err := common.VerifyJWT(c.Request.Context(), cc, verificationToken, "email", env.JWT_EMAIL_SECRET, &payload); err != nil {
			logger.Warn("Email verification failed", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid or expired verification token"})
			return
		}

		if err := repo.VerifyUser(c.Request.Context(), payload.Email); err != nil {
			if errors.Is(err, errs.ERR_EMAIL_NO_EXISTS) {
				logger.Warn("Email verification requested for a missing or already verified user")
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
				return
			}
			logger.Error("Failed to verify user email", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		remaining := time.Until(payload.ExpiresAt.Time)
		if err := common.BlacklistToken(c.Request.Context(), cc, verificationToken, "email", remaining); err != nil {
			logger.Error("Failed to blacklist email verification token", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		logger.Info("Successfully verified user email")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Email verified successfully"})
	}
}
