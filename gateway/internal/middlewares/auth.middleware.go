package middlewares

import (
	"gateway/internal/cache"
	"gateway/internal/common"
	"gateway/internal/configs"
	"gateway/internal/domain"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthenticatePrivateRoutes(env *configs.Env, r *cache.Redis) gin.HandlerFunc {

	return func(c *gin.Context) {
		var user domain.MinimalUserStruct
		var err error
		logger := common.GetLogger(c)
		now := time.Now()
		if checkIfPathIsAllowed(c.Request.URL.Path) {
			c.Next()
			return
		}
		secret, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")

		if secret == "" || !found {
			secret, err = c.Cookie(domain.USER_JWT_COOKIE_KEY)
			if err != nil || secret == "" {
				logger.Warn("Request missing authentication token")
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
				c.Abort()
				return
			}
		}

		blacklisted, err := r.CheckBlackList(c.Request.Context(), "session", common.TokenDigest(secret))
		if err != nil {
			logger.Error("Failed to check JWT blacklist", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			c.Abort()
			return
		}
		if blacklisted {
			logger.Warn("Request provided a blacklisted JWT")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(secret, &user, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, domain.ErrInvalidMethod
			}
			return env.JWT_SHARED_SECRET_KEY, nil
		})

		if err != nil || !token.Valid {
			logger.Warn("Request provided invalid JWT", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			c.Abort()
			return
		}

		if user.ExpiresAt == nil || user.ExpiresAt.Time.Before(now) {
			logger.Warn("Request provided expired JWT")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
			c.Abort()
			return
		}

		logger.Info("Successfully authenticated the request")
		c.Set("user", user)
		c.Next()
	}
}

func checkIfPathIsAllowed(path string) bool {
	allowedPaths := []string{
		"/api/auth/login",
		"/api/auth/register",
		"/api/auth/refresh",
		"/api/auth/password/reset",
		"/api/auth/password/forgot",
	}

	for _, allowedPath := range allowedPaths {
		if strings.HasPrefix(path, allowedPath) {
			return true
		}
	}
	return false
}
