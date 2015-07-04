// Loader for operating system templates
package main

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
	//"os"
	"strings"
	"text/template"
)

// dealing with os layouts and templates

// OSListGet :  get a listing of the operating systems
func (c *Config) OSListGet() (osm map[string]*OperatingSystem) {
	osm = make(map[string]*OperatingSystem)
	list, err := c.fs.List("boot")
	if err != nil {
		logger.Error("OS failure %s", err)
		//		os.Exit(1)
	}
	logger.Info("OS listing ", list)
	for _, i := range list {
		logger.Notice("Adding OS %s", i)
		tmpOS := &OperatingSystem{Name: i, Description: i}
		if tmpOS.CheckAndLoad(c) == true {
			osm[i] = tmpOS
		}
	}
	return
}

// CheckAndLoad : check and load the operating system
func (os *OperatingSystem) CheckAndLoad(c *Config) (pass bool) {
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
}

// LoadTemplates : load the templates for the operating system
func (os *OperatingSystem) LoadTemplates(c *Config) (pass bool) {
	path := "boot/" + os.Name + "/template"
	templateList, err := c.fs.List(path)
	if err != nil {
		logger.Critical("Template Load fail %s", err)
		return false
	}
	//load all the templates for the operating sytem
	newTemplates := template.New("")
	for i, j := range templateList {
		template, _, _ := c.fs.Get(path + "/" + j)
		defer template.Close()
		if strings.HasSuffix(j, ".tmpl") {
			data, err := ioutil.ReadAll(template)
			name := strings.TrimSuffix(j, ".tmpl")
			_, err = newTemplates.New(name).Parse(string(data))
			if err != nil {
				logger.Error("template ", name, " -> ", string(data), err)
			}
			logger.Info("#%d : %s", i, j)
		}
	}
	os.templates = newTemplates
	// check for operating system classes
	classPath := "boot/" + os.Name + "/" + "classes.toml"
	classFile, _, err := c.fs.Get(classPath)
	defer classFile.Close()
	if err != nil {
		logger.Info("Class List fail, %s", err)
		// still returns true , so the OS is added
		return true
	}
	// load the clases file
	classString, _ := ioutil.ReadAll(classFile)
	cl := &classes{}
	_, err = toml.Decode(string(classString), &cl)
	// attach to the os list
	os.Classes = cl.Classes
	for _, i := range os.Classes {
		logger.Notice("Class %s", i)
	}
	logger.Debug("Class File : %s", cl)
	os.HasClasses = true
	return true
}
