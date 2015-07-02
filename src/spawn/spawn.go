package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/coreos/go-systemd/unit"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

// api constants for fleet and astralboot
const (
	api        = "/fleet/v1/"
	port       = "9876"
	sourceapi  = "/spawn/"
	sourceport = "80"
)

type spawn struct {
	host       string
	sourcehost string
	api        url.URL
	sourceapi  url.URL
	// list of units that should be running
	targetList map[string]string
	// unit store
	unitText map[string]string
	// mapped array of units ( to jsonify and send to fleet api)
	units map[string]*Unit
}

// struct for sending units to fleetd api

type Unit struct {
	DesiredState string             `json:"desiredState"`
	CurrentState string             `json:"currentState,omitempty"`
	Name         string             `json:"name,omitempty"`
	Options      []*unit.UnitOption `json:"options"`
}

type UnitList struct {
	Units []Unit `json:"units"`
}

// create a spawn instance
func NewSpawn(sourcehost string, host string) (sp *spawn) {
	sp = &spawn{}
	sp.host = host
	sp.sourcehost = sourcehost

	// target host ( has fleetapi running on it )
	u := url.URL{}
	u.Scheme = "http"
	u.Host = sp.host + ":" + port
	u.Path = api

	sp.api = u

	// source host ( has astralboot + spawn running on it )
	u2 := url.URL{}
	u2.Scheme = "http"
	u2.Host = sp.sourcehost + ":" + sourceport
	u2.Path = sourceapi

	sp.sourceapi = u2
	// create the data maps
	sp.unitText = make(map[string]string)
	sp.units = make(map[string]*Unit)
	return
}

// shared get function
func (sp *spawn) Get(path string, api url.URL) (data []byte, err error) {
	resp, err := http.Get(api.String() + path)
	fmt.Println(api.String())
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}
	if err != nil {
		return nil, err
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return data, err
}

func main() {
	fmt.Println("Spawning Spawn")
	// get the source and target
	// flags override enviromental variables
	source := os.Getenv("SPAWN_SOURCE")
	sourceFlag := flag.String("source", "", "spawn source address X.X.X.X")
	// target , running fleetd on port
	target := os.Getenv("SPAWN_TARGET")
	targetFlag := flag.String("target", "", "spawn target address X.X.X.X")
	// parse the flags
	flag.Parse()

	if *sourceFlag != "" {
		source = *sourceFlag
	}
	if source == "" {
		fmt.Println("can not run without spawn source")
		os.Exit(1)
	}

	if *targetFlag != "" {
		target = *targetFlag
	}
	// default to localhost
	if target == "" {
		target = "127.0.0.1"
	}

	fmt.Println("source env :", source)
	fmt.Println("source flag :", *sourceFlag)
	fmt.Println("target env :", target)
	fmt.Println("target flag :", *targetFlag)
	sp := NewSpawn(source, target)
	u := sp.GetUnits()
	for _, i := range u.Units {
		fmt.Println(i)
	}

	//t := Load("test.service")
	//sp.LoadUnit(t, "launched")

	sp.SourceList()
	sp.SourceAll()
	fmt.Println(sp)
	for i, j := range sp.targetList {
		fmt.Println(i, "\n", j)
		sp.SendUnit(sp.units[i], "launched")
	}
}
