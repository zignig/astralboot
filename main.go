package main

import (
	"fmt"

	"github.com/zignig/cohort/assets"
)

func main() {
	fmt.Println("loading config")
	conf := GetConfig("config.toml")
	// cache for data files
	cache := assets.NewCache()
	// leases sql database
	leases := NewStore("")

	fmt.Println("starting tftp")
	go tftpServer()
	fmt.Println("start dhcp")
	go dhcpServer(leases)
	a, err := cache.Ls(conf.Ref)
	fmt.Println(string(a), err)
	c := make(chan int, 1)
	<-c
}
