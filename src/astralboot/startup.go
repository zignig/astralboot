// questions for start up configuration
package main

import (
	"net"
	"os"
)

//Set up the queries/
// functions in questions.go

func (c *Config) Setup() {
	slug(preamble)
	interfaceList := getInterf()
	if len(interfaceList) < 2 {
		slug(singleInterface)
		useSingleInterface := yesNoQuestion{text: "Single Interface", deflt: true}.Ask()
		if !useSingleInterface {
			logger.Fatal("No selected interface exiting. BYE!")
		}
	}
	c.Interf = listQuestion{text: "Select Interface to run services on", list: interfaceList}.Ask()
	slug(enableIPFS)
	c.IPFS = yesNoQuestion{text: "Enable IPFS data source", deflt: true}.Ask()
	if c.IPFS {
		logger.Critical("Help With IPFS setup")
		ipfsHelper()
	} else {
		logger.Critical("Help With file system setup")
		fileHelper()
	}
	slug(enableSpawn)
	c.Spawn = yesNoQuestion{text: "Enable Spawn", deflt: true}.Ask()
	slug(extraNetwork)
	runExtra := yesNoQuestion{text: "Extra Network Configuration", deflt: false}.Ask()
	if runExtra {
		serverIP := getAddr(c.Interf)
		c.DNSServer = ipAddrQuestion{text: "IP address for a dns server", ip: serverIP}.Ask()
		c.Gateway = ipAddrQuestion{text: "IP address of the gateway for dhcp clients", ip: serverIP}.Ask()
	}
	slug(finishUP)
	c.PrintConfig()
	saveConfig := yesNoQuestion{text: "Save Config", deflt: true}.Ask()
	if saveConfig {
		c.Save(configFile)
	} else {
		logger.Fatal("Configuration Failed")
	}
	slug(thanks)
	slug(banner)
}

func ipfsHelper() {
	_, err := os.Stat("refs.toml")
	if err != nil {
		logger.Critical("%s", err)
		if err == os.ErrNotExist {
			logger.Critical("file does not exist %s", err)
		}
	}
}

func fileHelper() {

}

func getInterf() (in map[string]string) {
	in = map[string]string{}
	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Error("Startup Interface %v", err)
	}
	for _, n := range interfaces[1:] { //ignore lo
		addr, _ := n.Addrs()
		if len(addr) > 0 {
			in[n.Name] = addr[0].String()
		}
	}
	return
}

func getAddr(ifaceName string) (i net.IP) {
	iface, _ := net.InterfaceByName(ifaceName)
	addressList, _ := iface.Addrs()
	i, _, _ = net.ParseCIDR(addressList[0].String())
	return
}

const preamble = `Welcome to Astralboot,

This program is a boot server that provides dhcp/tftp/http services to automate the boot sequence.

First some questions to help getting set up:
`

const singleInterface = `Having a single interface can be dangerous if there is another DHCP server on the local network
Are you sure you want to use a single interface?`

const enableIPFS = `IPFS is a distributed data store , find out more at http://ipfs.io/
To access this service you will need to have a local IPFS node running.
If you select NO you will need to have some local files and folders in ./data/ for astralboot to work. `

const enableSpawn = `Spawn is a coreos bootstrapper that talks to fleetd and auto starts unit files listed.
The source code is included in spawn directory.`

const extraNetwork = `Edit the DNS server and gateway ip addresses? these will default to this server if not specified.`
const finishUP = `This is the config setup so far
`
const thanks = `Thanks for trying Astral boot ! :)
If you have an patches / issues / flames please participate at https://github.com/zignig/astralboot

`
