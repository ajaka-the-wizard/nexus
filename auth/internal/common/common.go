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

func HoursToDuration(hours float64) time.Duration {
	return time.Duration(hours * float64(time.Hour))
}

func GenerateJWT(secret string, duration float64, payload jwt.Claims) (string, time.Duration, error) {
	if duration <= 0 {
		return "", 0, fmt.Errorf("JWT duration must be greater than zero")
	}

	tokenDuration := HoursToDuration(duration)
	expiresAt := jwt.NewNumericDate(time.Now().Add(tokenDuration))
	switch claims := payload.(type) {
	case *models.MinimalUserStruct:
		claims.ExpiresAt = expiresAt
	case *models.LoneEmailPayload:
		claims.ExpiresAt = expiresAt
	default:
		return "", 0, fmt.Errorf("unsupported JWT claims type %T", payload)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	return signedToken, tokenDuration, nil
}

func VerifyJWT(ctx context.Context, c *cache.Cache, tokenString, prefix, secret string, claims jwt.Claims) error {
	blacklisted, err := c.CheckBlackList(ctx, prefix, TokenDigest(tokenString))
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
	sessionToken, sessionDuration, err := GenerateJWT(env.JWT_SHARED_SECRET_KEY, env.JWT_SESSION_DURATION, &payload)
	if err != nil {
		return err
	}

	refreshToken, refreshDuration, err := GenerateJWT(env.JWT_REFRESH_KEY, env.JWT_REFRESH_KEY_DURATION, &payload)
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

func BlacklistToken(ctx context.Context, c *cache.Cache, token, prefix string, duration time.Duration) error {
	if !c.AddToBlacklist(ctx, prefix, TokenDigest(token), duration) {
		return errs.ERR_BLACKLIST_TOKEN_FAILURE
	}
	return nil
}

func HandleLogoutActivity(c *gin.Context, cc *cache.Cache, env *configs.Env) error {
	sessionToken, _ := c.Cookie("JWT_SECRET")
	refreshToken, _ := c.Cookie("JWT_REFRESH_SECRET")

	if sessionToken == "" && refreshToken == "" {
		return errs.ERR_NO_TOKENS_PROVIDED
	}

	var sessionClaims models.MinimalUserStruct
	if err := VerifyJWT(c.Request.Context(), cc, sessionToken, "session", env.JWT_SHARED_SECRET_KEY, &sessionClaims); err != nil {
		return err
	}
	if err := BlacklistToken(c.Request.Context(), cc, sessionToken, "session", time.Until(sessionClaims.ExpiresAt.Time)); err != nil {
		return err
	}
	var refreshClaims models.MinimalUserStruct
	if err := VerifyJWT(c.Request.Context(), cc, refreshToken, "refresh", env.JWT_REFRESH_KEY, &refreshClaims); err != nil {
		return err
	}
	if err := BlacklistToken(c.Request.Context(), cc, refreshToken, "refresh", time.Until(refreshClaims.ExpiresAt.Time)); err != nil {
		return err
	}

	c.SetSameSite(http.SameSiteLaxMode)
	SetCookie(c, "JWT_SECRET", "", -1, env.PRODUCTION)
	SetCookie(c, "JWT_REFRESH_SECRET", "", -1, env.PRODUCTION)
	return nil
}
