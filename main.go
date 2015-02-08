package main

import (
	"fmt"

	"github.com/zignig/cohort/assets"
)

func main() {
	fmt.Println("loading config")

	conf := GetConfig("config.toml")
	conf.SaveConfig()
	// cache for data files
	cache := assets.NewCache()
	// leases sql database
	leases := NewStore(conf)
	fmt.Println("starting tftp")
	go tftpServer(conf, cache)
	fmt.Println("start dhcp")
	go dhcpServer(conf, leases)

	fmt.Println("start web server")
	wh := NewWebServer(conf, leases)
	go wh.Run()
	// gorotiune spinner
	c := make(chan int, 1)
	<-c
}
