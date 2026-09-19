package handlers

import (
	"auth/internal/cache"
	"auth/internal/common"
	"auth/internal/configs"
	"auth/internal/errs"
	"auth/internal/models"
	"auth/internal/repositories"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

func HandleForgotPassword(repo *repositories.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		value, exists := c.Get("forgotPasswordRequest")
		if !exists {
			logger.Error("Forgot password request was not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		req, ok := value.(models.ForgotPasswordRequest)
		if !ok {
			logger.Error("Forgot password request has an invalid context type", "type", fmt.Sprintf("%T", value))
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		_, err := repo.GetUserByEmail(c.Request.Context(), req.Email)
		if err != nil && !errors.Is(err, errs.ERR_EMAIL_NO_EXISTS) {
			logger.Error("Failed to check email for password reset", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		if err == nil {
			// TODO: Send the password reset URL to the email service through Kafka.
			logger.Info("Password reset requested for existing user")
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "An email has been sent with instructions to reset your password.",
		})
	}
}

func HandleVerifyPasswordReset(env *configs.Env, cc *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		resetToken := c.Query("val")
		if resetToken == "" {
			logger.Warn("Password reset verification request missing token")
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid reset token"})
			return
		}

		var payload models.LoneEmailPayload
		if err := common.VerifyJWT(c.Request.Context(), cc, resetToken, "password", env.JWT_EMAIL_SECRET, &payload); err != nil {
			logger.Warn("Password reset verification failed", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid or expired reset token"})
			return
		}

		remaining := time.Until(payload.ExpiresAt.Time)
		if err := common.BlacklistToken(c.Request.Context(), cc, resetToken, "password", remaining); err != nil {
			logger.Error("Failed to blacklist password reset token", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		codec := securecookie.New([]byte(env.COOKIE_SECRET), nil)
		resetState, err := codec.Encode("password-reset-email", payload.Email)
		if err != nil {
			logger.Error("Failed to sign password reset cookie", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		c.SetSameSite(http.SameSiteLaxMode)
		common.SetCookie(c, "PASSWORD_RESET_USER", resetState, int(remaining/time.Second), env.PRODUCTION)

		logger.Info("Successfully verified password reset token")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password reset token verified"})
	}
}

func HandleResetPassword(repo *repositories.Repository, env *configs.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		value, exists := c.Get("resetPasswordRequest")
		if !exists {
			logger.Error("Reset password request was not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		request, ok := value.(models.ResetPasswordRequest)
		if !ok {
			logger.Error("Reset password request has an invalid context type", "type", fmt.Sprintf("%T", value))
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		encodedEmail, err := c.Cookie("PASSWORD_RESET_USER")
		if err != nil || encodedEmail == "" {
			logger.Warn("Password reset request missing reset cookie")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			return
		}
		codec := securecookie.New([]byte(env.COOKIE_SECRET), nil)
		var email string
		if err := codec.Decode("password-reset-email", encodedEmail, &email); err != nil {
			logger.Warn("Password reset request provided an invalid reset cookie", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			return
		}

		if err := repo.ResetPassword(c.Request.Context(), email, &request); err != nil {
			if errors.Is(err, errs.ERR_EMAIL_NO_EXISTS) {
				logger.Warn("Password reset requested for an email that does not exist", "email", email)
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
				return
			}
			if errors.Is(err, errs.ERR_PASSWORD_REUSED) {
				logger.Warn("Password reset rejected because the password was used previously")
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Password was used previously"})
				return
			}
			if errors.Is(err, errs.ERR_PASSWORD_RESET_COOLDOWN) {
				logger.Warn("Password reset rejected because the cooldown has not elapsed")
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Password reset is not available yet"})
				return
			}
			logger.Error("Failed to reset user password", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		common.SetCookie(c, "PASSWORD_RESET_USER", "", -1, env.PRODUCTION)
		logger.Info("Successfully reset user password", "email", email)
		c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Password reset successful"})
	}
}
