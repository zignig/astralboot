// lease database for dhcp server
package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"sync"
	"time"
)

//LeaseList : Leases stored on disk as JSON file
type LeaseList struct {
	Leases []*Lease
}

var leaseLock sync.Mutex

//Lease : storage structure for each lease
type Lease struct {
	ID       int64     // id of the machine
	MAC      string    // mac address as a string
	IP       string    // use the SetIP and GetIP funcs
	Active   bool      // lease is active
	Reserved bool      // lease is reserved
	Distro   string    // linux distro
	Name     string    // host name
	Class    string    // sub class of the machine
	Created  time.Time // when the machine is created
	// add more stuff
	pending bool
}

// Lease List functions

// Get all the active leases
func (ll LeaseList) Active() (l []*Lease) {
	l = make([]*Lease, 0)
	for _, i := range ll.Leases {
		if i.Active {
			l = append(l, i)
			logger.Debug("%v", i)
		}
	}
	return l
}

//IP return a lease for the given IP addresss
func (ll LeaseList) IP(ip net.IP) (l *Lease, err error) {
	for _, i := range ll.Leases {
		if i.IP == ip.String() {
			return i, nil
		}
	}
	return nil, errors.New("no lease")
}

//Mac return a lease for the given Hardwareaddr
func (ll LeaseList) Mac(mac net.HardwareAddr) (l *Lease, err error) {
	for _, i := range ll.Leases {
		if i.MAC == mac.String() {
			return i, nil
		}
	}
	return l, errors.New("no lease for mac")
}

//Free : returns an unused address
func (ll LeaseList) Free(mac net.HardwareAddr) (l *Lease, err error) {
	for _, i := range ll.Leases {
		if (i.Active == false) && (i.pending == false) && (i.Reserved == false) {
			logger.Critical("New Lease %s for mac %s", i.IP, i.MAC)
			i.pending = true
			return i, err
		}
	}
	logger.Critical("No leases available")
	return nil, errors.New("no available leases")
}

//Append : lease appender
func (ll *LeaseList) Append(l *Lease) {
	ll.Leases = append(ll.Leases, l)
}

//GetDist :  Returns a map of leaselist for classes of a given disto
func (ll LeaseList) GetDist(dist string) (le map[string]*LeaseList, err error) {
	le = make(map[string]*LeaseList)
	for _, i := range ll.Leases {
		if i.Distro == dist {
			_, ok := le[i.Class]
			if ok == false {
				le[i.Class] = &LeaseList{}
			}
			le[i.Class].Append(i)
			//logger.Critical("%v", le)
		}
	}
	return
}

//Load :  load the leases from the json file on disk
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
	active := 0
	reserved := 0
	for _, i := range ll.Leases {
		if i.Active == true {
			active = active + 1
		}
		if i.Reserved == true {
			reserved = reserved + 1
		}
	}
	logger.Info("%v leases in file", len(ll.Leases))
	logger.Notice("%v active leases in file", active)
	logger.Debug("%v reserved leases in file", reserved)
	return ll
}

//Save : write the leases to disk
func (ll LeaseList) Save(name string) {
	leaseLock.Lock()
	enc, err := json.MarshalIndent(ll, "", " ")
	if err != nil {
		logger.Critical("Lease Marshal fail , %v", err)
	}
	err = ioutil.WriteFile(name, enc, 0644)
	if err != nil {
		logger.Critical("Lease save fail , %v", err)
	}
	logger.Info("Leases Saved")
	leaseLock.Unlock()
}
