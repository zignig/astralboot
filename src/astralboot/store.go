// Store interface for lease wrangling.
package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

//Store :  struct for dhcp data store
type Store struct {
	DBname string
	leases LeaseList
	config *Config
}

// store functions

// NewStore : create a new store object
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

// Build : generate the initial lease file
func (s Store) Build(c *Config) {
	logger.Critical("Building lease tables")
	leaseList := NetList(c.BaseIP, c.Subnet)
	ll := LeaseList{}
	for count, i := range leaseList {
		l := &Lease{}
		l.ID = int64(count)
		l.Created = time.Now()
		l.IP = i.String()
		ll.Leases = append(ll.Leases, l)
	}
	s.leases = ll
	// Reserve the following
	// - network address
	s.Reserve(leaseList[0])
	// - self
	s.Reserve(c.BaseIP)
	// - broadcast
	s.Reserve(leaseList[len(leaseList)-1])
	s.leases.Save(s.DBname)
}

// GetIP :  return a net.IP from the lease
func (l Lease) GetIP() (ip net.IP) {
	return net.ParseIP(l.IP)
}

// Reserve : mark a lease as reserved
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

// UpdateActive : update a lease to active
func (s Store) UpdateActive(mac net.HardwareAddr, name string) bool {
	l := &Lease{}
	logger.Info("Update ", mac, " to active")
	l, err := s.leases.Mac(mac)
	if err != nil {
		logger.Error("lease error %s", err)
		return false
	}
	l.Active = true
	l.Distro = name
	s.leases.Save(s.DBname)
	return true
}

// UpdateClass : update class and activate
func (s Store) UpdateClass(mac net.HardwareAddr, name string, class string) bool {
	l := &Lease{}
	l, err := s.leases.Mac(mac)
	if err != nil {
		logger.Error("lease error %s", err)
		return false
	}
	l.Active = true
	l.Distro = name
	l.Class = class
	logger.Critical("Update %v to class %s of %s", l.MAC, l.Class, l.Distro)
	s.leases.Save(s.DBname)
	return true
}

// CheckLease : check if a lease exists
func (s Store) CheckLease(mac net.HardwareAddr) bool {
	l := &Lease{}
	l, err := s.leases.Mac(mac)
	if err != nil {
		logger.Error("lease error %s", err)
		return false
	}
	if &l != nil {
		return true
	}
	return false
}

// GetIP : fore the given mac address get the IP address
func (s Store) GetIP(mac net.HardwareAddr) (ip net.IP, err error) {
	l := &Lease{}
	l, err = s.leases.Mac(mac)
	if err != nil {
		logger.Error("lease error %s", err)
		return nil, err
	}
	ip = net.ParseIP(l.IP)
	logger.Critical("Lease IP : %s", ip)
	return ip, nil
}

// DistLease : for a given distro get a map of the classes as lists
func (s Store) DistLease(dist string) (ll map[string]*LeaseList) {
	ll, err := s.leases.GetDist(dist)
	if err != nil {
		logger.Error("Lease search error %s ", err)
		return
	}
	return
}

// GetFromIP : get a lease from an IP
func (s Store) GetFromIP(ip net.IP) (l *Lease, err error) {
	newl := &Lease{}
	newl, err = s.leases.IP(ip)
	return newl, err
}

// Release : not working
func (s Store) Release(mac net.HardwareAddr) (err error) {
	rell, err := s.leases.Mac(mac)
	if err == nil {
		return err
	}
	rell.Active = false
	s.leases.Save(s.DBname)
	return nil
}

// GetLease : get an existing lease or mark a new one.
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
			l.Name = fmt.Sprintf("node%d", l.ID)
		}
		logger.Debug("updated lease")
		s.leases.Save(s.DBname)
		return l, nil
	}

	return l, err
}

// NetList : subnet helper function
func NetList(ip net.IP, subnet net.IP) (IPlist []net.IP) {
	//ip, ipnet, err := net.ParseCIDR(cidrNet)
	mask := net.IPv4Mask(subnet[0], subnet[1], subnet[2], subnet[3])
	ipnet := net.IPNet{ip, mask}
	for ip := ip.Mask(mask); ipnet.Contains(ip); incIP(ip) {
		IPlist = append(IPlist, net.IP{ip[0], ip[1], ip[2], ip[3]})
	}
	return
}

// incIP : looper function for subnets
func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
