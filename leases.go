package main

// lease database for dhcp server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"time"
)

// Leases stored on disk as JSON file
type LeaseList struct {
	Leases []*Lease
}

// Leases storage
type Lease struct {
	Id       int64     // id of the machine
	MAC      string    // mac address as a string
	IP       string    // use the SetIP and GetIP funcs
	Active   bool      // lease is active
	Reserved bool      // lease is reserved
	Distro   string    // linux distro
	Name     string    // host name
	Class    string    // sub class of the machine
	Created  time.Time // when the machine is created
	// add more stuff
}

// Lease List functions

func (ll LeaseList) IP(ip net.IP) (l *Lease, err error) {
	for _, i := range ll.Leases {
		if i.IP == ip.String() {
			return i, nil
		}
	}
	return nil, errors.New("no lease")
}

func (ll LeaseList) Mac(mac net.HardwareAddr) (l *Lease, err error) {
	for _, i := range ll.Leases {
		if i.MAC == mac.String() {
			return i, nil
		}
	}
	return l, errors.New("no lease for mac")
}

func (ll LeaseList) Free(mac net.HardwareAddr) (l *Lease, err error) {
	logger.Critical("%v leases", len(ll.Leases))
	for _, i := range ll.Leases {
		logger.Debug("%v", i)
		if (i.Active == false) && (i.Reserved == false) {
			return i, err
		}
	}
	logger.Critical("No leases available")
	return nil, errors.New("No available leases")
}

func (ll LeaseList) GetDist(dist string) (le LeaseList, err error) {
	// TODO get dist list
	//	distList := make(map[string]int)
	for _, i := range ll.Leases {
		logger.Critical("%v", i.Distro)
	}
	return
}

func (ll LeaseList) GetClasses() (classes []string, err error) {
	for i := range ll.Leases {
		logger.Critical("%v", i)
	}
	logger.Critical("TODO class list")
	return
}

// load the leases from the json file on disk
func Load(name string) (ll LeaseList) {
	content, err := ioutil.ReadFile(name)
	if err != nil {
		logger.Critical("Load Fail : %v", err)
	}
	ll = LeaseList{}
	err = json.Unmarshal(content, &ll)
	if err != nil {
		logger.Critical("Lease Marshall fail , %v", err)
	}
	logger.Info("%v leases in file", len(ll.Leases))
	return ll
}

// Save the leases to disk
// TODO needs locking , perhaps a channel system for linear updates
func (ll LeaseList) Save(name string) {
	enc, err := json.MarshalIndent(ll, "", " ")
	if err != nil {
		logger.Critical("Lease Marshal fail , %v", err)
	}
	err = ioutil.WriteFile(name, enc, 0644)
	if err != nil {
		logger.Critical("Lease save fail , %v", err)
	}
	logger.Info("Leases Saved")
}
