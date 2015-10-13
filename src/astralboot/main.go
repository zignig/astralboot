// Combined boot server
package main

import (
	"flag"
	//"fmt"
)

var configFile string

func main() {
	// Flags
	configFileName := flag.String("c", "config.toml", "name for config file")
	lowDebug := flag.Bool("v", false, "Notice and Above")
	medDebug := flag.Bool("vv", false, "Info")
	highDebug := flag.Bool("vvv", false, "Debug")
	flag.Parse()
	// Config file name
	configFile = *configFileName
	// log level switcher
	var logLevel int
	if *lowDebug {
		logLevel = 1
	}
	if *medDebug {
		logLevel = 2
	}
	if *highDebug {
		logLevel = 3
	}
	LogSetup(logLevel)
	//fmt.Println(banner)
	logger.Notice("Starting Astralboot Server")
	conf := GetConfig(configFile)
	logger.Notice("Using interface : %s", conf.Interf)
	if logLevel > 0 {
		logger.Notice("-- Implied Config Start --")
		conf.PrintConfig()
		logger.Notice("-- Implied Config Finish --")
	}
	// leases json database
	leases := NewStore(conf)

	logger.Info("starting dns")
	d := NewDnsServer(conf, leases)
	go d.Run()
	logger.Info("starting tftp")
	go tftpServer(conf)
	logger.Info("start dhcp")
	go dhcpServer(conf, leases)
	logger.Info("start web server")
	wh := NewWebServer(conf, leases, logLevel)
	go wh.Run()
	logger.Notice("Serving ...")
	// go spinner
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
