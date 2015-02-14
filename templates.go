package main

import "fmt"

// dealing with os layouts and templates

// get a listing of the operating systems
func (c *Config) OSListGet() (os map[string]*operatingSystem) {
	os = make(map[string]*operatingSystem)
	list, _ := c.fs.List("boot")
	fmt.Println("OS listing ", list)
	for _, i := range list {
		logger.Info(" ----- " + i + "-------")
		tmpOS := &operatingSystem{Name: i, Description: i}
		if tmpOS.CheckAndLoad(c) == true {
			os[i] = tmpOS
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
			templatePass := os.LoadTemplates(c)
			if templatePass == false {
				return false
			}

		}
	}
	return true
	// grab and build all the templates for the operating systems.
	// returns true if the
}

func (os operatingSystem) LoadTemplates(c *Config) (pass bool) {
	templateList, err := c.fs.List("boot/" + os.Name + "/template/")
	if err != nil {
		logger.Critical("Template Load fail %s", err)
		return false
	}
	for i, j := range templateList {

		logger.Critical("#%s : %s", i, j)
	}
	return true

}
