package main

import (
	"github.com/gin-gonic/gin"
)

func webServer(c *Config, l *Store) {
	router := gin.Default()
	router.GET("/boot/stuff", func(c *gin.Context) {
		c.String(200, "#!ipxe\n\nlogin")
	})
	router.Run(":80")
}
