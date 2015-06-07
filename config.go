package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"text/template"

	"github.com/BurntSushi/toml"
)

// operating system struct
// loaded in templates.go
type operatingSystem struct {
	Name        string
	Description string
	HasClasses  bool
	Classes     []string
	templates   *template.Template
}

// c.lass list
// sub type of the operating system type
type classes struct {
	Classes []string `toml:"classes"`
}

// references for ipfs loading
type Refs struct {
	Boot   string `toml:"boot"`
	Rocket string `toml:"rocket"`
}

// base configuration structure
type Config struct {
	Interf string `toml:"interface"`
	// switchable services
	Spawn     bool
	BaseIP    net.IP
	Gateway   net.IP
	Subnet    net.IP
	DNSServer net.IP
	Domain    string
	DBname    string
	Local     bool
	Refs      *Refs // ipfs references
	// not exported generated config parts
	fs     ROfs
	OSList map[string]*operatingSystem
}

// load config and settings
func GetConfig(path string) (c *Config) {
	if _, err := toml.DecodeFile(path, &c); err != nil {
		logger.Critical("Config file does not exists,create config")
		return
	}
	// loading the refs from seperate file
	re := &Refs{}
	if _, err := toml.DecodeFile("refs.toml", &re); err != nil {
		logger.Critical("Reference file does not exists")
	}
	fmt.Println(re)
	c.Refs = re
	// bind the cache (not exported)
	// Add items from system not in config file
	if c.Interf == "" {
		c.Interf = "eth0"
	}
	interf, err := net.InterfaceByName(c.Interf)
	if err != nil {
		logger.Critical("Interface error ", err)
	}
	//TODO fix interface checks
	addressList, _ := interf.Addrs()
	serverAddress, ipnet, _ := net.ParseCIDR(addressList[0].String())
	logger.Critical("Server Address  : %s", serverAddress)
	c.BaseIP = serverAddress
	b := ipnet.Mask
	c.Subnet = net.IP{b[0], b[1], b[2], b[3]}
	if c.Gateway == nil {
		c.Gateway = serverAddress
	}
	if c.DNSServer == nil {
		c.DNSServer = serverAddress
	}
	// database file name
	if c.DBname == "" {
		c.DBname = "./machines.json"
	}

	if c.Domain == "" {
		c.Domain = "erf"
	}
	//TODO select file system from flag or config

	fileFlag := flag.Bool("l", false, "Use local file sytem")
	var filesystem ROfs
	flag.Parse()
	if *fileFlag {
		c.Local = true
		filesystem = &Diskfs{"./data"}
	} else {
		filesystem = &IPfsfs{c.Refs.Boot}
	}
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
