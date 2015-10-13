package main

// local dns server
// using skydns http://github.com/skynetservices/skydns
// loads local nodes and astralboot dns entries

import (
	"errors"
	"github.com/skynetservices/skydns/msg"
	"github.com/skynetservices/skydns/server"
	"sync"
)

type Domain struct {
	prefix    string
	entries   map[string]*msg.Service
	addLock   sync.Mutex
	DNSConfig *server.Config
}

func NewDomain(prefix string) (d *Domain) {
	d = &Domain{}
	msg.PathPrefix = "astralboot"
	d.prefix = msg.Path(prefix)
	d.entries = make(map[string]*msg.Service)
	return d
}

func NewDnsServer(c *Config, leases *Store) (d *Domain) {
	d = NewDomain(c.Domain)
	active := leases.leases.Active()
	// add the astralboot
	d.Add("astralboot", c.BaseIP.String())
	// add all the active nodes
	for _, j := range active {
		d.Add(j.Name, j.IP)
	}
	// Create the config
	d.DNSConfig = &server.Config{
		Domain:  c.Domain,
		DnsAddr: c.BaseIP.String() + ":53",
		Verbose: false,
	}
	return d
}

func (d *Domain) Run() {
	for i, j := range d.entries {
		logger.Info("%s %v", i, j)
	}
	server.Metrics()
	server.SetDefaults(d.DNSConfig)
	s := server.New(d, d.DNSConfig)
	s.Run()
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

// skydns interface for serving

func (d *Domain) Records(name string, exact bool) ([]msg.Service, error) {
	path, _ := msg.PathWithWildcard(name)
	logger.Debug("fetch record %s", path)
	val, ok := d.entries[path]
	if ok {
		l := make([]msg.Service, 1)
		l[0] = *val
		return l, nil
	}
	return nil, errors.New("FAIL")
}

func (d *Domain) ReverseRecord(name string) (*msg.Service, error) {
	return nil, errors.New("FAIL REVERSE")
}
