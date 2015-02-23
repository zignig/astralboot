package main

// lease database for dhcp server

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
)

// struct for thing store
type Store struct {
	db     *sql.DB
	dbmap  *gorp.DbMap
	sessMu sync.Mutex
	leases map[string]*Lease
	config *Config
}

func NewStore(c *Config) *Store {
	// create a new store
	store := Store{}
	store.config = c
	// check if the file exists
	var build bool
	_, err := os.Stat(c.DBname)
	if err != nil {
		logger.Critical("error on stat , %s", err)
		build = true
	}
	db, err := sql.Open("sqlite3", c.DBname)
	store.db = db
	if err != nil {
		fmt.Println(err)
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	store.dbmap = dbmap

	// map the objects
	dbmap.AddTable(Lease{}).SetKeys(true, "Id")

	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		fmt.Print(err)
	}
	// if it is a new file build some tables
	if build {
		store.Build(c)
	}
	return &store
}

// build some initial tables
func (s Store) Build(c *Config) {
	logger.Critical("Building lease tables")
	leaseList := NetList(c.BaseIP, c.Subnet)
	for _, i := range leaseList {
		fmt.Println("add a lease for ", i)
		l := &Lease{}
		l.Created = time.Now()
		l.IP = i.String()
		err := s.dbmap.Insert(l)
		if err != nil {
			logger.Error("Lease insert error %s", err)
		}
	}
	// TODO
	// need to disable
	// - network address
	s.Reserve(leaseList[0])
	// - self
	s.Reserve(&c.BaseIP)
	// - broadcast
	s.Reserve(leaseList[len(leaseList)-1])
	// possibly ping check and reserve those addresses
}

// close the store
func (s Store) Close() {
	s.db.Close()
}

// access methods
func (s Store) Query(q string) error {
	rows, err := s.db.Query(q)
	if err != nil {
		return err
	}
	fmt.Println(rows)
	return nil
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

// return a net.IP from the lease ( stored as string in sql )
func (l Lease) GetIP() (ip net.IP) {
	return net.ParseIP(l.IP)
}

// mark a lease as reserved
func (s Store) Reserve(ip *net.IP) {
	l := &Lease{}
	err := s.dbmap.SelectOne(&l, "select * from Lease where IP = ?", ip.String())
	if err != nil {
		logger.Error("No such IP , %s", err)
		return
	}
	l.Reserved = true
	_, err = s.dbmap.Update(l)
	if err != nil {
		logger.Error("Lease Reserve Fail , %s", err)
	}
	logger.Info("Reserved IP address %s", ip)
}

// update active
func (s Store) UpdateActive(mac net.HardwareAddr, name string) bool {
	l := &Lease{}
	fmt.Println("Update ", mac, " to active")
	err := s.dbmap.SelectOne(&l, "select * from Lease where MAC = ?", mac.String())
	if err != nil {
		fmt.Printf("lease error %s", err)
		return false
	}
	l.Active = true
	l.Distro = name
	count, err := s.dbmap.Update(l)
	fmt.Println(count, err)
	return true
}

// check lease
func (s Store) CheckLease(mac net.HardwareAddr) bool {
	var l Lease
	err := s.dbmap.SelectOne(&l, "select * from Lease where MAC = ?", mac.String())
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
	var l Lease
	err = s.dbmap.SelectOne(&l, "select * from Lease where MAC = ?", mac.String())
	if err != nil {
		fmt.Printf("lease error %s", err)
		return nil, err
	}
	ip = net.ParseIP(l.IP)
	logger.Critical("Lease IP : %s", ip)
	return ip, nil
}

func (s Store) GetDist(mac net.HardwareAddr) (name string, err error) {
	var l Lease
	err = s.dbmap.SelectOne(&l, "select dist from Lease where MAC = '' and active == True")
	return l.Name, err
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
	err = s.dbmap.SelectOne(&newl, "select * from Lease where MAC = ?", mac.String())
	if err == nil {
		return newl, err
	}
	logger.Debug("No existing lease %s ", err)
	// find a lease that is inactive and not reserved
	var leaseList []Lease
	logger.Debug("PREFAIL")
	_, err = s.dbmap.Select(&leaseList, "select * from Lease where Active = 0 and Reserved = 0 limit 1")
	fmt.Println(leaseList)
	if err != nil {
		logger.Debug("Lease search error %s ", err)
	}
	if len(leaseList) == 1 {
		// get one lease and update it's mac address
		theLease := leaseList[0]
		theLease.MAC = mac.String()
		theLease.Created = time.Now()
		_, err := s.dbmap.Update(&theLease)
		if err != nil {
			logger.Critical("Lease Update Fail %s", err)
			return nil, err
		}
		return &theLease, nil
	}

	return newl, err
}

//helper functions
func NetList(ip net.IP, subnet net.IP) (IPlist []*net.IP) {
	//ip, ipnet, err := net.ParseCIDR(cidrNet)
	mask := net.IPv4Mask(subnet[0], subnet[1], subnet[2], subnet[3])
	ipnet := net.IPNet{ip, mask}
	for ip := ip.Mask(mask); ipnet.Contains(ip); incIP(ip) {
		IPlist = append(IPlist, &net.IP{ip[0], ip[1], ip[2], ip[3]})
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
