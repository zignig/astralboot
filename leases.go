package main

// lease database for dhcp server

import (
	"database/sql"
	"fmt"
	"net"
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
}

func NewStore(dbName string) *Store {
	// create a new store
	store := Store{}
	if dbName == "" {
		dbName = "./leases.db"
	}

	db, err := sql.Open("sqlite3", dbName)
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
	Id  int64
	MAC string
	//IP       int64
	Active  bool
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
