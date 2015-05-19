package main

// lease database for dhcp server

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// struct for dhcp data store
type Store struct {
	DBname string
	sessMu sync.Mutex
	leases LeaseList
	config *Config
}

// store functions
func NewStore(c *Config) *Store {
	// create a new store
	store := Store{}
	store.config = c
	store.DBname = c.DBname
	// check if the file exists
	var build bool
	_, err := os.Stat(c.DBname)
	if err != nil {
		logger.Critical("error on stat , %s", err)
		build = true
	}
	// if it is a new file build some tables
	if build {
		store.Build(c)
	}
	store.leases = Load(c.DBname)
	return &store
}

// build some initial tables
func (s Store) Build(c *Config) {
	logger.Critical("Building lease tables")
	leaseList := NetList(c.BaseIP, c.Subnet)
	ll := LeaseList{}
	for count, i := range leaseList {
		//fmt.Println("add a lease for ", i)
		l := &Lease{}
		l.Id = int64(count)
		l.Created = time.Now()
		l.IP = i.String()
		ll.Leases = append(ll.Leases, l)
	}
	s.leases = ll
	// TODO
	// need to disable
	// - network address
	s.Reserve(leaseList[0])
	// - self
	s.Reserve(c.BaseIP)
	// - broadcast
	s.Reserve(leaseList[len(leaseList)-1])
	s.leases.Save(s.DBname)
}

// close the store
func (s Store) Close() {
	// TODO  write and close json file
}

// return a net.IP from the lease
func (l Lease) GetIP() (ip net.IP) {
	return net.ParseIP(l.IP)
}

// mark a lease as reserved
func (s Store) Reserve(ip net.IP) {
	l := &Lease{}
	l, err := s.leases.IP(ip)
	if err != nil {
		logger.Error("No such IP , %s", err)
		return
	}
	l.Reserved = true
	if err != nil {
		logger.Error("Lease Reserve Fail , %s", err)
	}
	logger.Info("Reserved IP address %s", ip)
	s.leases.Save(s.DBname)
}

// update active
func (s Store) UpdateActive(mac net.HardwareAddr, name string) bool {
	l := &Lease{}
	logger.Info("Update ", mac, " to active")
	l, err := s.leases.Mac(mac)
	if err != nil {
		fmt.Printf("lease error %s", err)
		return false
	}
	l.Active = true
	l.Distro = name
	s.leases.Save(s.DBname)
	return true
}

// check lease
func (s Store) CheckLease(mac net.HardwareAddr) bool {
	l := &Lease{}
	l, err := s.leases.Mac(mac)
	if err != nil {
		fmt.Printf("lease error %s", err)
		return false
	}
	if &l != nil {
		return true
	}
	return false
}

// get ip
func (s Store) GetIP(mac net.HardwareAddr) (ip net.IP, err error) {
	l := &Lease{}
	l, err = s.leases.Mac(mac)
	if err != nil {
		fmt.Printf("lease error %s", err)
		return nil, err
	}
	ip = net.ParseIP(l.IP)
	logger.Critical("Lease IP : %s", ip)
	return ip, nil
}

// get a list of ips for a distro
// coreos cluster testing
// look into using subclass
func (s Store) DistLease(dist string) (ll LeaseList) {
	var classes []string
	classes, err := s.leases.GetClasses()
	if err != nil {
		logger.Debug("Class list error %s", err)
	}
	logger.Debug("%s", classes)
	ll, err = s.leases.GetDist("etcd")
	if err != nil {
		logger.Debug("Lease search error %s ", err)
		return
	}
	return
}

// get a lease from an IP
func (s Store) GetFromIP(ip net.IP) (l *Lease, err error) {
	newl := &Lease{}
	newl, err = s.leases.IP(ip)
	return newl, err
}

func (s Store) Release(mac net.HardwareAddr) {
	//TODO update lease to be active false
}

//  Find a  free address
// 1. unused
// 2. inactive
// 3. expired
// 4. fail
func (s Store) GetLease(mac net.HardwareAddr) (l *Lease, err error) {
	newl := &Lease{}
	// do I have a lease for this mac address
	logger.Debug("Find Lease for %v", mac)
	newl, err = s.leases.Mac(mac)
	if err == nil {
		return newl, err
	}
	logger.Debug("No existing lease %s ", err)
	// find a lease that is inactive and not reserved
	l, err = s.leases.Free(mac)
	if err != nil {
		logger.Debug("Lease search error %s ", err)
	} else {
		// get one lease and update it's mac address
		logger.Debug("found lease, updating")
		l.MAC = mac.String()
		l.Created = time.Now()
		if l.Name == "" {
			l.Name = fmt.Sprintf("node%d", l.Id)
		}
		logger.Debug("updated lease")
		s.leases.Save(s.DBname)
		return l, nil
	}

	return l, err
}

//helper functions
func NetList(ip net.IP, subnet net.IP) (IPlist []net.IP) {
	//ip, ipnet, err := net.ParseCIDR(cidrNet)
	mask := net.IPv4Mask(subnet[0], subnet[1], subnet[2], subnet[3])
	ipnet := net.IPNet{ip, mask}
	for ip := ip.Mask(mask); ipnet.Contains(ip); incIP(ip) {
		IPlist = append(IPlist, net.IP{ip[0], ip[1], ip[2], ip[3]})
	}
	return
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
