package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"net"
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
	OSList map[string]OS
}

func GetConfig(path string) (c *Config) {
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

	return
}

func (c *Config) SaveConfig() {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	fmt.Println(buf.String(), err)
}
