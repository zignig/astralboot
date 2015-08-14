// Serve rocket files and systemd units
package main

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"io/ioutil"
	"strings"
)

//SpawnAPI  spawn api construct
type SpawnAPI struct {
	units     map[string]string
	fs        ROfs
	templates *template.Template
}

// TheSpawn : local spawn API instance
// this may need to be pushed up into config
var TheSpawn *SpawnAPI

// NewSpawnAPI : create a new spawn instance
func NewSpawnAPI(fs ROfs) (sa *SpawnAPI) {
	sa = &SpawnAPI{}
	sa.fs = fs
	sa.ScanUnits()
	return
}

// ScanUnits : get a list of the local available unit files
func (sa *SpawnAPI) ScanUnits() {
	logger.Debug("Scan for units")
	unitlist, err := sa.fs.List("units")
	if err != nil {
		logger.Error("Unit scan error , %v", err)
		return
	}
	// build the templates
	NewTemplates := template.New("")
	sa.units = make(map[string]string)
	for _, i := range unitlist {
		if strings.HasSuffix(i, ".service") {
			shortName := strings.TrimSuffix(i, ".service")
			logger.Debug("unit: %s", i)
			template, _, err := sa.fs.Get("units/" + i)
			defer template.Close()
			if err != nil {
				logger.Critical("template error , %s", err)
			}
			data, err := ioutil.ReadAll(template)
			_, err = NewTemplates.New(shortName).Parse(string(data))
			if err != nil {
				logger.Error("template parse error , %s", err)
			}
			sa.units[shortName] = "active" // put unit info in here
		}
	}
	sa.templates = NewTemplates
	logger.Debug("%s", sa.units)
}

func (wh *WebHandler) SpawnHandler() {
	spawnRef := wh.config.Refs.Spawn
	var fs ROfs
	if wh.config.IPFS == true {
		// if empty ref assume under base ref
		if spawnRef == "" {
			logger.Debug("using units under boot ref")
			fs = NewIPfsfs(wh.config.Refs.Boot)
		} else {
			fs = NewIPfsfs(spawnRef)
		}
	} else {
		fs = &Diskfs{"./data"}
	}
	TheSpawn = NewSpawnAPI(fs)
	wh.router.GET("/spawn/list", wh.UnitList)
	wh.router.GET("/spawn/unit/:name", wh.GetUnit)
}

// UnitList : send the list of units as a JSON list
func (wh *WebHandler) UnitList(c *gin.Context) {
	logger.Notice("List Available Units")
	c.IndentedJSON(200, TheSpawn.units)
}

// GetUnit : send an individual Unit file as text
func (wh *WebHandler) GetUnit(c *gin.Context) {
	UnitName := c.Params.ByName("name")
	logger.Notice("Get Unit File : %s", UnitName)
	if TheSpawn.templates.Lookup(UnitName) == nil {
		c.AbortWithStatus(404)
		return
	}
	logger.Debug("Unit Requested is %s", UnitName)
	client, err := GetIP(c)
	if err != nil {
		logger.Debug("client fail , %s ", err)
	}
	td := wh.GenTemplateData(client, "coreos") // spawn only works with coreos at the moment
	err = TheSpawn.templates.ExecuteTemplate(c.Writer, UnitName, td)
	if err != nil {
		logger.Critical("unit template faili %s", err)
	}

}
