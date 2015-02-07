package main

import (
	"github.com/gin-gonic/gin"
)

func webServer(c *Config, l *Store) {
	router := gin.Default()
	router.GET("/ipxe/start", func(c *gin.Context) {
		c.String(200, coreText)
	})
	router.Static("/boot", "./data/boot/coreos")
	//router.Static("/boot", "./data/boot/debian")
	router.Run(":80")
}

var defaultText = `#!ipxe

kernel http://192.168.2.1/boot/linux priority=critical auto=true url=http://192.168.2.1/boot/preseed
initrd http://192.168.2.1/boot/initrd.gz
boot

`
var coreText = `#!ipxe

kernel http://192.168.2.1/boot/coreos_production_pxe.vmlinuz root=/dev/sda1 cloud-config-url=http://192.168.2.1/pxe-cloud-config.yml
initrd http://192.168.2.1/boot/coreos_production_pxe_image.cpio.gz
boot

`
