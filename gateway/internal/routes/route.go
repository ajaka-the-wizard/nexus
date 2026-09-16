package routes

import (
	"gateway/internal/common"
	"gateway/internal/configs"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Proxy(s *configs.ServiceConfig) gin.HandlerFunc {
	client := http.Client{}
	return func(c *gin.Context) {
		logger := common.GetLogger(c)
		path, _ := strings.CutPrefix(c.Request.URL.Path, "/api")
		serviceURL, err := common.GetUrlFromPath(path, s)
		if err != nil {
			logger.Error("Failed to resolve upstream service for proxy request",
				"error", err,
			)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
			return
		}

		serviceURL.RawQuery = c.Request.URL.RawQuery

		logger.Info("Proxying request to upstream service")

		req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, serviceURL.String(), c.Request.Body)
		if err != nil {
			logger.Error("Failed to create proxied request",
				"error", err,
			)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		req.Header = c.Request.Header
		logger.Info("header ingoing", "header", req.Header)
		req.ContentLength = c.Request.ContentLength

		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Upstream proxy request failed",
				"error", err,
			)
			c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Something went wrong"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= http.StatusBadRequest {
			logger.Warn("Upstream service returned an error response")
		}

		logger.Info("Header Outgoing", "header", resp.Header)
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}
		c.Status(resp.StatusCode)

		if _, err := io.Copy(c.Writer, resp.Body); err != nil {
			logger.Error("Failed to copy upstream response to client",
				"error", err,
			)
			c.Error(err)
			return
		}
	}
}
