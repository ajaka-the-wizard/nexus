package middlewares

import (
	"auth/common"
	"auth/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateRegisterRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		var request models.RegisterRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Warn("Request provided an invalid registration body", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
			})
			c.Abort()
			return
		}

		if err := validate.Struct(request); err != nil {
			logger.Warn("Request provided invalid registration details", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid registration details",
			})
			c.Abort()
			return
		}

		logger.Info("Successfully validated registration request")
		c.Set("registerRequest", request)
		c.Next()
	}
}

func ValidateLoginRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		var request models.LoginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			logger.Warn("Request provided an invalid login body", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request body",
			})
			c.Abort()
			return
		}

		if err := validate.Struct(request); err != nil {
			logger.Warn("Request provided invalid login details", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid login details",
			})
			c.Abort()
			return
		}

		logger.Info("Successfully validated login request")
		c.Set("loginRequest", request)
		c.Next()
	}
}
