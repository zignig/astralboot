package main

// lease database for dhcp server

import (
	"database/sql"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/coopernurse/gorp"
	dhcp "github.com/krolaw/dhcp4"
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
	return &store
}

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

// Session storage
type Lease struct {
	Id      int64
	MAC     string
	Active  bool
	Distro  string
	Name    string
	Created time.Time
}

// new lease
func (s Store) NewLease(mac net.HardwareAddr) {
	l := &Lease{}
	l.MAC = mac.String()
	l.Created = time.Now()
	fmt.Println(l)
	err := s.dbmap.Insert(l)
	fmt.Println(err)
}

// update active
func (s Store) UpdateActive(mac net.HardwareAddr) bool {
	l := &Lease{}
	fmt.Println("Update ", mac, " to active")
	err := s.dbmap.SelectOne(&l, "select * from Lease where MAC = ?", mac.String())
	if err != nil {
		fmt.Printf("lease error %s", err)
		return false
	}
	l.Active = true
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
	fmt.Println(s.config.BaseIP)
	ip = dhcp.IPAdd(s.config.BaseIP, int(l.Id))
	//ip = net.IP{192, 168, 2, 4}
	return ip, nil
}

func (s Store) Release(mac net.HardwareAddr) {
	//TODO update lease to be active false
}

//  Find a  free address
// 1. search for an unused
// 2. search for inactive
// 3. expired

func (s Store) FindFree(mac net.HardwareAddr) (ip net.IP, err error) {
	var l Lease
	err = s.dbmap.SelectOne(&l, "select * from Lease where MAC = ''")
	return
}

func cidr(cidrNet string) {
	ip, ipnet, err := net.ParseCIDR(cidrNet)
	if err != nil {
		fmt.Println(err)
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		fmt.Println(ip)
	}
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
