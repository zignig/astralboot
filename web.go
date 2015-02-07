package main

import (
	"github.com/gin-gonic/gin"
)

func webServer(c *Config, l *Store) {
	router := gin.Default()
	router.GET("/ipxe/start", func(c *gin.Context) {
		c.String(200, defaultText)
	})
	router.Static("/boot", "./data/boot/debian")
	router.Run(":80")
}

var defaultText = `#!ipxe

kernel http://192.168.2.1/boot/linux priority=critical auto=true url=http://192.168.2.1/preseed
initrd http://192.168.2.1/boot/initrd.gz
boot

`
