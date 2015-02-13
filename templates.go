package main

import "fmt"

// dealing with os layouts and templates

// get a listing of the operating systems
func (c *Config) OSListGet() (os []operatingSystem) {
	list, _ := c.fs.List("boot")
	fmt.Println("OS listing ", list)
	for _, i := range list {
		logger.Info(" ----- " + i + "-------")
		tmpOS := operatingSystem{Name: i, Description: i}
		if tmpOS.CheckAndLoad(c) == true {
			os = append(os, tmpOS)
		}
	}
	return
}

func (os operatingSystem) CheckAndLoad(c *Config) (pass bool) {
	subList, err := c.fs.List("boot/" + os.Name)
	if err != nil {
		logger.Error("SubList Error %s", err)
		return false
	}
	for _, i := range subList {
		logger.Debug("found folder %s", i)
		if i == "template" {
			os.LoadTemplates()
		}
	}
	return true
	// grab and build all the templates for the operating systems.
	// returns true if the
}

func (os operatingSystem) LoadTemplates() {
	logger.Critical("LOAD TEMPLATES HERE ")
}
