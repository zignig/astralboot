// questions for start up configuration
package main

import (
	"net"
)

//Set up the queries/
// functions in questions.go

func (c *Config) Setup() {
	interfaceList := getInterf()
	interfaceQuestion := &listQuestion{text: "Select Interface", list: interfaceList}
	c.Interf = interfaceQuestion.Ask()

	enableIPFS := &yesNoQuestion{text: "Enable IPFS source", deflt: false}
	c.IPFS = enableIPFS.Ask()
	enableSpawn := &yesNoQuestion{text: "Enable Spawn", deflt: true}
	c.Spawn = enableSpawn.Ask()
	//		&ipAddrQuestion{text: "IP address", ip: "test"},
	c.Domain = ""
	c.Save()
}

func getInterf() (in map[string]string) {
	in = map[string]string{}
	interfaces, err := net.Interfaces()
	if err != nil {
		logger.Error("Startup Interface %v", err)
	}
	for _, n := range interfaces {
		addr, _ := n.Addrs()
		in[n.Name] = addr[0].String()
	}
	return
}
