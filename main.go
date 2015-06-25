// Combined boot server
package main

import (
	"flag"
	"fmt"
)

func main() {
	lowDebug := flag.Bool("v", false, "Notice and Above")
	medDebug := flag.Bool("vv", false, "Info")
	highDebug := flag.Bool("vvv", false, "Debug")
	flag.Parse()
	var logLevel int
	fmt.Println(*lowDebug, *medDebug, *highDebug)
	if *lowDebug {
		fmt.Println("DEBUG")
		logLevel = 0
	}
	if *medDebug {
		logLevel = 1
	}
	if *highDebug {
		logLevel = 2
	}
	LogSetup(logLevel)
	fmt.Println(banner)
	logger.Notice("Starting Astralboot Server")
	conf := GetConfig("config.toml")
	logger.Notice("Using interface : %s", conf.Interf)
	if logLevel > 0 {
		logger.Notice("-- Implied Config Start --")
		conf.PrintConfig()
		logger.Notice("-- Implied Config Finish --")
	}
	// leases json database
	leases := NewStore(conf)
	logger.Info("starting tftp")
	go tftpServer(conf)
	logger.Info("start dhcp")
	go dhcpServer(conf, leases)
	logger.Info("start web server")
	wh := NewWebServer(conf, leases, logLevel)
	go wh.Run()
	logger.Notice("Serving ...")
	// goroutine spinner
	c := make(chan int, 1)
	<-c
}

const banner = `
┌──────────────────────────────┐
│┏━┓┏━┓╺┳╸┏━┓┏━┓╻  ┏┓ ┏━┓┏━┓╺┳╸│
│┣━┫┗━┓ ┃ ┣┳┛┣━┫┃  ┣┻┓┃ ┃┃ ┃ ┃ │
│╹ ╹┗━┛ ╹ ╹┗╸╹ ╹┗━╸┗━┛┗━┛┗━┛ ╹ │
└──────────────────────────────┘
https://github.com/zignig/astralboot
`
