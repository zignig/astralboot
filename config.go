package main

import (
	"bytes"
	"fmt"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/zignig/cohort/assets"
)

type OS struct {
	Name        string
	Description string
}

type Config struct {
	Ref    string `toml:"ref"`
	Interf string `toml:"interface"`
	BaseIP net.IP
	DBname string
	OSList []OS
}

func GetConfig(path string, cache *assets.Cache) (c *Config) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		fmt.Println("Config file does not exists,create config")
		fmt.Println(err)
		return
	}
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

	// distributions

	var j []OS
	j = append(j, OS{"debian", "Debian"})
	j = append(j, OS{"coreos", "CoreOS"})
	c.OSList = j
	return
}

func (c *Config) SaveConfig() {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	fmt.Println(buf.String(), err)
}
