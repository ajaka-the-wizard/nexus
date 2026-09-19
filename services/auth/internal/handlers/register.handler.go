package handlers

import (
	"auth/internal/common"
	"auth/internal/errs"
	"auth/internal/models"
	"auth/internal/repositories"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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
		// TODO: Generate the email verification JWT, build the verification URL, and publish it to Kafka.
		logger.Info("Successfully registered user", "email", request.Email)
		c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Registration request accepted"})
	}
}
