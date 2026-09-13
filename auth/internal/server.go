package internal

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Listen() error {
	router := gin.New()

	router.Use(gin.Recovery())

	router.SetTrustedProxies(nil)

	router.GET("/auth/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	return router.Run(":5000")
}
