package common

import (
	"crypto/sha256"
	"encoding/hex"
	"gateway/internal/configs"
	"gateway/internal/domain"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetLogger(c *gin.Context) *slog.Logger {
	if l, ok := c.Get("logger"); ok {
		if logger, ok := l.(*slog.Logger); ok && logger != nil {
			return logger
		}
	}
	return slog.Default()
}

func TokenDigest(token string) string {
	digest := sha256.Sum256([]byte(token))
	return hex.EncodeToString(digest[:])
}

func GetServiceFromServiceConfig(cfg *configs.ServiceConfig, path string) (*configs.ServiceDefinition, string, error) {
	var key string

	switch {
	case strings.HasPrefix(path, "/auth"):
		key = domain.AUTH_ROUTES
	case strings.HasPrefix(path, "/users"):
		key = domain.USER_ROUTES
	default:
		return nil, "", domain.ErrNoSuchService
	}
	service, ok := cfg.Services[key]
	if !ok {
		return nil, "", domain.ErrNoSuchService
	}
	return &service, key, nil
}

func GetUrlFromPath(path string, cfg *configs.ServiceConfig) (url.URL, error) {
	service, _, err := GetServiceFromServiceConfig(cfg, path)
	if err != nil {
		return url.URL{}, err
	}

	url := url.URL{
		Scheme: service.Protocol,
		Host:   net.JoinHostPort(service.Host, strconv.Itoa(service.Port)),
		Path:   path,
	}

	return url, nil
}
