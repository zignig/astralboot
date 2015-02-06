package main

import (
	"fmt"

	"github.com/zignig/cohort/assets"
)

func main() {
	info()
	fmt.Println("loading config")
	conf := GetConfig("config.toml")
	// cache for data files
	cache := assets.NewCache()
	// leases sql database
	leases := NewStore("")

	fmt.Println("starting tftp")
	go tftpServer(conf, cache)
	fmt.Println("start dhcp")
	go dhcpServer(conf, leases)

	// gorotiune spinner
	c := make(chan int, 1)
	<-c
}
