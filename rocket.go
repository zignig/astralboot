// Serve rocket files and systemd units
package main

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

var tmpl *template.Template // discovery template

// run the rocket aci out of a local fs for now

// RocketACI : local read only file system for rocket system
var RocketACI ROfs

type rktTmpl struct {
	BaseIP  net.IP
	AciName string
}

//SpawnAPI  spawn api construct
// TODO move into spawn module
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
	}
	// build the templates
	NewTemplates := template.New("")
	sa.units = make(map[string]string)
	for _, i := range unitlist {
		if strings.HasSuffix(i, ".service") {
			shortName := strings.TrimSuffix(i, ".service")
			logger.Debug("unit: %s", i)
			template, err := sa.fs.Get("units/" + i)
			defer template.Close()
			if err != nil {
				logger.Critical("template error , %s", err)
			}
			data, err := ioutil.ReadAll(template)
			_, err = NewTemplates.New(shortName).Parse(string(data))
			if err != nil {
				logger.Critical("template parse error , %s", err)
			}
			sa.units[shortName] = "active" // TODO put unit info in here
		}
	}
	sa.templates = NewTemplates
	logger.Debug("%s", sa.units)
}

// RocketHandler : add the rocket and spawn parts to the wh router
func (wh *WebHandler) RocketHandler() {
	rocketRef := wh.config.Refs.Rocket
	var fs ROfs
	if (rocketRef == "") || (wh.config.Local == true) {
		fs = &Diskfs{"./rocket"}
	} else {
		fs = &IPfsfs{rocketRef}
	}
	RocketACI = fs

	t, err := template.New("rocket").Parse(MetaDiscovery)
	tmpl = t
	if err != nil {
		logger.Error("template error", err)
	}
	//wh.router.GET("/rocket", wh.Discovery)
	wh.router.GET("/rocket/:name", wh.Discovery)
	wh.router.GET("/images/:source/:rocket/:imageName", wh.AciImage)
	// access for spawn
	if wh.config.Spawn {
		// create the spawn api
		TheSpawn = NewSpawnAPI(fs)
		wh.router.GET("/spawn/list", wh.UnitList)
		wh.router.GET("/spawn/unit/:name", wh.GetUnit)
	}
}

// UnitList : send the list of units as a JSON list
func (wh *WebHandler) UnitList(c *gin.Context) {
	c.IndentedJSON(200, TheSpawn.units)
}

// GetUnit : send an individual Unit file as text
func (wh *WebHandler) GetUnit(c *gin.Context) {
	// TODO hand list out of current available unit files
	UnitName := c.Params.ByName("name")
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

// AciImage : sends athe requested rocket aci file
func (wh *WebHandler) AciImage(c *gin.Context) {
	logger.Debug(c.Request.RequestURI)
	AciName := c.Params.ByName("imageName")
	logger.Debug(AciName)
	fh, err := RocketACI.Get(AciName)
	if err != nil {
		logger.Debug("Rocket file error : %s", err)
		c.AbortWithStatus(404)
	}
	io.Copy(c.Writer, fh)
}

// Discovery : templates the rkt discovery data
func (wh *WebHandler) Discovery(c *gin.Context) {
	logger.Debug(c.Request.RequestURI)
	queryMap := c.Request.URL.Query()
	_, ok := queryMap["ac-discovery"]
	if ok {
		t := rktTmpl{}
		t.BaseIP = wh.config.BaseIP
		t.AciName = c.Params.ByName("name")
		// random etags for the win
		c.Header("ETag", time.Now().String())
		err := tmpl.ExecuteTemplate(c.Writer, "rocket", t)
		if err != nil {
			logger.Error("template error ", err)
		}
		return
	}
}

// MetaDiscovery : template for aci dicovery
var MetaDiscovery = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
      <meta name="ac-discovery" content="{{ .BaseIP }}/rocket/{{ .AciName }} http://{{ .BaseIP }}/images/{name}-{version}-{os}-{arch}.{ext}">
  </head>
<html>
`

// AUTH line ( unused )
//<meta name="ac-discovery-pubkeys" content="example.com/{{ .AciName }} https://example.com/pubkeys.gpg">
