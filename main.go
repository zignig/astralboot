package main

func main() {
	LogSetup()
	logger.Critical("STARTING DHCP SERVER")
	conf := GetConfig("config.toml")
	logger.Critical("-- Implied Config Start --")
	conf.PrintConfig()
	logger.Critical("-- Implied Config Finish --")
	// leases sql database
	leases := NewStore(conf)
	logger.Info("starting tftp")
	go tftpServer(conf)
	logger.Info("start dhcp")
	go dhcpServer(conf, leases)
	logger.Info("start web server")
	wh := NewWebServer(conf, leases)
	go wh.Run()
	logger.Info("Serving ...")
	// gorotiune spinner
	c := make(chan int, 1)
	<-c
}
