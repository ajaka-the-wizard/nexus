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

func HandleRegister(repo *repositories.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		value, exists := c.Get("registerRequest")
		if !exists {
			logger.Error("Registration request was not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		request, ok := value.(models.RegisterRequest)
		if !ok {
			logger.Error("Registration request has an invalid context type", "type", fmt.Sprintf("%T", value))
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		hashedPassword, err := common.HashPassword(request.Password)
		if err != nil {
			logger.Error("Failed to hash registration password", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		request.Password = hashedPassword

		logger.Info("Creating user from registration request", "email", request.Email)
		if err = repo.CreateUser(c.Request.Context(), &request); err != nil {
			var message string
			var code int
			if errors.Is(err, errs.ERR_DUPLICATE_EMAIL) {
				logger.Warn("Registration rejected because the email already exists", "email", request.Email)
				message = "Invalid credentials"
				code = http.StatusBadRequest
			} else {
				logger.Error("Failed to persist registration request", "error", err)
				message = "Somethig went wrong"
				code = http.StatusInternalServerError
			}
			c.JSON(code, gin.H{"success": false, "message": message})
			return

		}
		logger.Info("Successfully registered user", "email", request.Email)
		c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Registration request accepted"})
	}
}

func HandleLogin(repo *repositories.Repository, env *configs.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		value, exists := c.Get("loginRequest")
		if !exists {
			logger.Error("Login request was not found in context")
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		request, ok := value.(models.LoginRequest)
		if !ok {
			logger.Error("Login request has an invalid context type", "type", fmt.Sprintf("%T", value))
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		user, err := repo.GetUserByEmail(c.Request.Context(), request.Email)
		if err != nil {
			if errors.Is(err, errs.ERR_EMAIL_NO_EXISTS) {
				logger.Warn("Login attempted with an email that does not exist", "email", request.Email)
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid credentials"})
				return
			}
			logger.Error("Failed to retrieve user for login", "email", request.Email, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		matches, err := common.VerifyPassword(request.Password, user.Password)
		if err != nil {
			logger.Error("Failed to verify login password", "email", request.Email, "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		if !matches {
			logger.Warn("Login request provided an incorrect password", "email", request.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid credentials"})
			return
		}

		payload := models.MinimalUserStruct{
			UserId: user.Id,
			Email:  user.Email,
		}
		if err := common.HandleLoginActivity(
			c,
			payload,
			env,
		); err != nil {
			logger.Error("Failed to create login session", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		logger.Info("Successfully authenticated user", "email", request.Email)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Login successful"})
	}
}

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

		if err := common.HandleLoginActivity(c, payload, env); err != nil {
			logger.Error("Failed to refresh login session", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}

		logger.Info("Successfully refreshed login session")
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Session refreshed"})
	}
}

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

		var payload models.ResetPasswordPayload
		if err := common.VerifyJWT(c.Request.Context(), cc, resetToken, "email", env.JWT_EMAIL_SECRET, &payload); err != nil {
			logger.Warn("Password reset verification failed", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid or expired reset token"})
			return
		}

		remaining := time.Until(payload.ExpiresAt.Time)
		if err := common.BlacklistToken(c.Request.Context(), cc, resetToken, "email", remaining); err != nil {
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
