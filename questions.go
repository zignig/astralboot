// Questions for initial config
package main

import (
	"fmt"
	"net"
	"strconv"
)

// Yes or No question
type yesNoQuestion struct {
	text  string
	deflt bool
	val   bool
}

func (q *yesNoQuestion) Ask() (b bool) {
	header(q.text)
	if q.deflt {
		fmt.Print("(Y/n)>")
	} else {
		fmt.Print("(y/N)>")
	}
	var response string
	_, err := fmt.Scanln(&response)
	if len(response) == 0 {
		fmt.Printf("Default answer : %v\n", q.deflt)
		return q.deflt
	}
	if err != nil {
		fmt.Println(len(response))
		logger.Error("readline error %v", err)
	}
	fmt.Println(response)
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return q.Ask()
	}
}

// IP address
type ipAddrQuestion struct {
	text string
	ip   string
}

func (q *ipAddrQuestion) Ask() (addr net.IPAddr) {
	header(q.text)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(len(response))
		logger.Error("readline error %v", err)
	}
	val := net.ParseIP(response)
	if val == nil {
		fmt.Println("Bad ip address")
		return q.Ask()
	}
	fmt.Println(val)
	return
}

// Select from list
type listQuestion struct {
	text string
	list map[string]string
}

func (q *listQuestion) Ask() (response string) {
	header(q.text)
	counter := 0
	value := 0
	asList := []string{}
	for i, item := range q.list {
		fmt.Printf("%d : %s -  %s\n", counter, i, item)
		counter = counter + 1
		asList = append(asList, i)
	}
	fmt.Print("Select Item >")
	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(len(response))
		logger.Error("readline error %v", err)
	}
	value, err = strconv.Atoi(response)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Bad Format")
		q.Ask()
	}
	if (value >= 0) && (value <= len(q.list)) {
		return asList[value]
	} else {
		fmt.Println("Out of range")
		return q.Ask()
	}
	return
}

//helper functions
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

func header(t string) {
	fmt.Println()
	fmt.Println(t)
	fmt.Println()
}
