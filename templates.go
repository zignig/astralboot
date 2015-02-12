package main

import "fmt"

// dealing with os layouts and templates

// get a listing of the operating systems
func (c *Config) OSListGet() (os []operatingSystem) {
	list, _ := c.fs.List("boot")
	fmt.Println("OS listing ", list)
	for _, i := range list {
		os = append(os, operatingSystem{i, i})
		subList, err := c.fs.List("boot/" + i)
		fmt.Println(subList, err)
	}
	return
}

func (os operatingSystem) templates() {
	// grab and build all the templates for the operating systems.
}
