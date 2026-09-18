package common

import (
	"auth/internal/cache"
	"auth/internal/configs"
	"auth/internal/errs"
	"auth/internal/models"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func GetLogger(c *gin.Context) *slog.Logger {
	if l, ok := c.Get("logger"); ok {
		if logger, ok := l.(*slog.Logger); ok && logger != nil {
			return logger
		}
	}
	return slog.Default()
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
	if err == nil {
		return true, nil
	}
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return false, err
}

func GenerateJWT(payload models.MinimalUserStruct, secret string) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(secret))
}

func VerifyJWT(ctx context.Context, c *cache.Cache, tokenString, secret string, claims jwt.Claims) error {
	blacklisted, err := c.CheckBlackList(ctx, TokenDigest(tokenString))
	if err != nil {
		return err
	}
	if blacklisted {
		return errs.ERR_BLACKLISTED_TOKEN
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errs.ERR_INVALID_METHOD
		}
		return []byte(secret), nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return jwt.ErrTokenInvalidClaims
	}
	expiresAt, err := claims.GetExpirationTime()
	if err != nil {
		return err
	}
	if expiresAt == nil || expiresAt.Time.Before(time.Now()) {
		return jwt.ErrTokenExpired
	}
	return nil
}

func SetCookie(c *gin.Context, name, value string, maxAge int, secure bool) {
	c.SetCookie(name, value, maxAge, "/", "", secure, true)
}

func HandleLoginActivity(c *gin.Context, payload models.MinimalUserStruct, env *configs.Env) error {
	sessionDuration := time.Duration(env.JWT_SESSION_DURATION * float64(time.Hour))
	refreshDuration := time.Duration(env.JWT_REFRESH_KEY_DURATION * float64(time.Hour))

	payload.ExpiresAt = jwt.NewNumericDate(time.Now().Add(sessionDuration))
	sessionJTI, err := uuid.NewV7()
	if err != nil {
		return err
	}
	payload.ID = sessionJTI.String()
	sessionToken, err := GenerateJWT(payload, env.JWT_SHARED_SECRET_KEY)
	if err != nil {
		return err
	}

	payload.ExpiresAt = jwt.NewNumericDate(time.Now().Add(refreshDuration))
	refreshJTI, err := uuid.NewV7()
	if err != nil {
		return err
	}
	payload.ID = refreshJTI.String()
	refreshToken, err := GenerateJWT(payload, env.JWT_REFRESH_KEY)
	if err != nil {
		return err
	}

	c.SetSameSite(http.SameSiteLaxMode)
	SetCookie(
		c,
		"JWT_SECRET",
		sessionToken,
		int(sessionDuration/time.Second),
		env.PRODUCTION,
	)
	SetCookie(
		c,
		"JWT_REFRESH_SECRET",
		refreshToken,
		int(refreshDuration/time.Second),
		env.PRODUCTION,
	)
	return nil
}

func TokenDigest(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func BlacklistTokens(ctx context.Context, c *cache.Cache, sessionDigest, refreshDigest string, sessionDuration, refreshDuration time.Duration) error {
	if sessionDigest != "" && !c.AddToBlacklist(ctx, sessionDigest, sessionDuration) {
		return fmt.Errorf("failed to blacklist session token")
	}
	if refreshDigest != "" && !c.AddToBlacklist(ctx, refreshDigest, refreshDuration) {
		return fmt.Errorf("failed to blacklist refresh token")
	}
	return nil
}

func HandleLogoutActivity(c *gin.Context, blacklist *cache.Cache, env *configs.Env) error {
	sessionToken, _ := c.Cookie("JWT_SECRET")
	refreshToken, _ := c.Cookie("JWT_REFRESH_SECRET")
	sessionDuration := time.Duration(env.JWT_SESSION_DURATION * float64(time.Hour))
	refreshDuration := time.Duration(env.JWT_REFRESH_KEY_DURATION * float64(time.Hour))

	if err := BlacklistTokens(
		c.Request.Context(),
		blacklist,
		TokenDigest(sessionToken),
		TokenDigest(refreshToken),
		sessionDuration,
		refreshDuration,
	); err != nil {
		return err
	}

	c.SetSameSite(http.SameSiteLaxMode)
	SetCookie(c, "JWT_SECRET", "", -1, env.PRODUCTION)
	SetCookie(c, "JWT_REFRESH_SECRET", "", -1, env.PRODUCTION)
	return nil
}
