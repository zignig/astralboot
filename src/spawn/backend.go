package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// backend talks to the fleet api

// post a unit object  to fleet with the desired state

func (sp *spawn) SendUnit(u *Unit, state string) (data []byte, err error) {
	u.DesiredState = state
	unitJson, err := json.Marshal(u)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(u)
	//fmt.Println("-----")
	//fmt.Println(string(unitJson))

	target := sp.api.String() + "units/" + u.Name
	fmt.Println(target)

	client := http.Client{}
	request, err := http.NewRequest("PUT", target, strings.NewReader(string(unitJson)))
	request.Header.Set("Content-Type", "application/json")
	request.ContentLength = int64(len(unitJson))
	response, err := client.Do(request)
	fmt.Println(response, err)
	return
}

// get wrapper function for http request to fleet
func (sp *spawn) TargetGet(path string) (data []byte, err error) {
	data, err = sp.Get(path, sp.api)
	return
}

// get a list of units from fleet
func (sp *spawn) GetUnits() (u *UnitList) {
	u = &UnitList{}
	t, err := sp.TargetGet("units")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(t, u)
	if err != nil {
		fmt.Println(err)
	}
	return
}
