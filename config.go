package main

import (
	"bytes"
	"fmt"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/spf13/afero"
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
	// not exported generated config parts
	fs     *afero.Fs
	OSList []OS
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

	// distributions

	var j []OS
	li := c.OSListGet()
	if err != nil {
		fmt.Println("OS listing error")
		return
	}
	for _, i := range li {
		j = append(j, OS{i, i})
	}
	c.OSList = j
	return
}

func (c *Config) SaveConfig() {
	buf := new(bytes.Buffer)
	err := toml.NewEncoder(buf).Encode(c)
	fmt.Println(buf.String(), err)
}

func (c *Config) OSListGet() (list []string) {
	list, _ = c.cache.Listing(c.Ref + "/boot/")
	fmt.Println("OS listing ", list)
	return
}
