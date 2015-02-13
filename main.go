package main

import (
"fmt"
)
func main() {
 	LogSetup()
	logger.Critical("STARTING DHCP SERVER")
	fmt.Println("loading config")
	conf := GetConfig("config.toml")
	conf.PrintConfig()
	// leases sql database
	leases := NewStore(conf)
	logger.Info("starting tftp")
	go tftpServer(conf)
	fmt.Println("start dhcp")
	go dhcpServer(conf, leases)

	fmt.Println("start web server")
	wh := NewWebServer(conf, leases)
	go wh.Run()
	// gorotiune spinner
	c := make(chan int, 1)
	<-c
}
