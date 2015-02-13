package main

import "fmt"

// dealing with os layouts and templates

// get a listing of the operating systems
func (c *Config) OSListGet() (os []operatingSystem) {
	list, _ := c.fs.List("boot")
	fmt.Println("OS listing ", list)
	for _, i := range list {
		tmpOS := operatingSystem{Name: i, Description: i}
		if tmpOS.CheckAndLoad() == true {
			os = append(os, tmpOS)
		}
		fmt.Println(" ----- " + i + "-------")
		subList, err := c.fs.List("boot/" + i)
		fmt.Println(subList, err)
	}
	return
}

func (os operatingSystem) CheckAndLoad() (pass bool) {
	return true
	// grab and build all the templates for the operating systems.
	// returns true if the
}
