package main

// local dns server
// using skydns http://github.com/skynetservices/skydns
// loads local nodes and astralboot dns entries

import (
	//"github.com/miekg/dns"

	"github.com/skynetservices/skydns/msg"
	//	"github.com/skynetservices/skydns/server"
	"sync"
)

type Domain struct {
	prefix  string
	entries map[string]*msg.Service
	addLock sync.Mutex
}

func NewDomain(prefix string) (d *Domain) {
	msg.PathPrefix = "astralboot"
	d = &Domain{}
	d.prefix = msg.Path(prefix)
	d.entries = make(map[string]*msg.Service)
	return d
}

func NewDnsServer(c *Config, leases *Store) (d *Domain) {
	d = NewDomain(c.Domain)
	active := leases.leases.Active()
	d.Add("astralboot", c.BaseIP.String())
	for _, j := range active {
		d.Add(j.Name, j.IP)
	}
	return d
}

func (d *Domain) Run() {
	for i, j := range d.entries {
		logger.Critical("%s %v", i, j)
	}
	logger.Critical("RUN DNS SERVER HERE")
}

func (d *Domain) LongName(name string) (long string) {
	long = d.prefix + "/" + name
	return long
}

func (d *Domain) Add(name string, address string) {
	d.addLock.Lock()
	defer d.addLock.Unlock()
	logger.Debug("Adding DNS entry : %s %s", name, address)
	mess := &msg.Service{
		Host: address,
	}
	d.entries[d.LongName(name)] = mess
}
