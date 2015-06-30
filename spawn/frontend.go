package main

// frontend talks to unit file service and lister
// this is only astralboot for now

// TODO investigate a direct ipfs frontent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-systemd/unit"
	"io"
)

// take a unit file as string and convert it to a unit object.
func Load(name string, unitAsFile io.Reader) (u *Unit) {
	opts, err := unit.Deserialize(unitAsFile)
	if err != nil {
		panic(err)
	}
	u = &Unit{}
	u.Options = opts
	u.Name = name + ".service"
	return
}

// run through all of the units
func (sp *spawn) SourceAll() {
	for i, _ := range sp.targetList {
		theUnit := sp.SourceUnit(i)
		sp.units[i] = theUnit
	}
}

func (sp *spawn) SourceUnit(name string) (u *Unit) {
	unitString, err := sp.SourceGet("unit/" + name)
	if err != nil {
		panic(err)
	}
	// sp.unitText[name] = string(unitString)
	//fmt.Println("--- %s ---", name)
	//fmt.Println(string(unitString))
	asFile := bytes.NewBuffer(unitString)
	u = Load(name, asFile)
	return
}

func (sp *spawn) SourceList() (li map[string]string) {
	li = make(map[string]string)
	t, err := sp.SourceGet("list")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(t, &li)
	if err != nil {
		fmt.Println(err)
	}
	sp.targetList = li
	return
}

func (sp *spawn) SourceGet(path string) (data []byte, err error) {
	data, err = sp.Get(path, sp.sourceapi)
	return
}
