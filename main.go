// Combined boot server
package main

import (
	"flag"
	"fmt"
)

func main() {
	logFlag := flag.Int("v", 0, "logging level 0=critcal , 5=debug")
	flag.Parse()
	LogSetup(*logFlag)
	fmt.Println(banner)
	logger.Notice("Starting Astralboot Server")
	conf := GetConfig("config.toml")
	logger.Notice("Using interface : %s", conf.Interf)
	if *logFlag > 0 {
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
	wh := NewWebServer(conf, leases, *logFlag)
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
