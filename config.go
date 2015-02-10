package main

import (
	"bytes"
	"fmt"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/zignig/cohort/assets"
)

type operatingSystem struct {
	Name        string
	Description string
}

type Config struct {
	Ref    string `toml:"ref"`
	Interf string `toml:"interface"`
	BaseIP net.IP
	DBname string
	// not exported generated config parts
	fs     ROfs
	OSList []operatingSystem
	cache  *assets.Cache
}

func GetConfig(path string, cache *assets.Cache) (c *Config) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		fmt.Println("Config file does not exists,create config")
		fmt.Println(err)
		return
	}
	// bind the cache (not exported)
	c.cache = cache
	// Add items from system not in config file
	if c.Interf == "" {
		c.Interf = "eth0"
	}
	interf, err := net.InterfaceByName(c.Interf)
	if err != nil {
		fmt.Println("Interface error ", err)
	}
	addressList, _ := interf.Addrs()
	serverAddress, _, _ := net.ParseCIDR(addressList[0].String())
	fmt.Println("Server Address  :", serverAddress)
	c.BaseIP = serverAddress

	// database file name
	if c.DBname == "" {
		c.DBname = "./leases.db"
	}
	// mount file system
	//TODO select file system from flag or config

	var filesystem ROfs = &Diskfs{"./data"}
	//var filesystem ROfs = &IPfsfs{c.Ref, c.cache}
	c.fs = filesystem
	// distributions
	c.OSList = c.OSListGet()

	return
}

func (c *Config) PrintConfig() {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	fmt.Println(buf.String(), err)
}

func (c *Config) OSListGet() (os []operatingSystem) {
	list, _ := c.fs.List("boot")
	fmt.Println("OS listing ", list)
	for _, i := range list {
		os = append(os, operatingSystem{i, i})
	}
	return
}
