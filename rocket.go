package main

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"time"
)

var tmpl *template.Template // discovery template

// run the rocket aci out of a local fs for now

var RocketACI ROfs

type rktTmpl struct {
	BaseIP  net.IP
	AciName string
}

// spawn api construct
type SpawnAPI struct {
	units     map[string]string
	fs        ROfs
	templates *template.Template
}

// local spawn API instance
// this may need to be pushed up into config
var TheSpawn *SpawnAPI

func NewSpawnAPI(fs ROfs) (sa *SpawnAPI) {
	sa = &SpawnAPI{}
	sa.fs = fs
	sa.ScanUnits()
	return
}

func (sa *SpawnAPI) ScanUnits() {
	logger.Debug("Scan for units")
	unitlist, err := sa.fs.List("units")
	if err != nil {
		logger.Error("Unit scan error , %s", err)
	}
	// build the templates
	NewTemplates := template.New("")
	sa.units = make(map[string]string)
	for _, i := range unitlist {
		logger.Debug("unit: %s", i)
		template, err := sa.fs.Get("units/" + i)
		defer template.Close()
		if err != nil {
			logger.Critical("template error , %s", err)
		}
		data, err := ioutil.ReadAll(template)
		_, err = NewTemplates.New(i).Parse(string(data))
		if err != nil {
			logger.Critical("template parse error , %s", err)
		}
		sa.units[i] = "active" // TODO put unit info in here
	}
	sa.templates = NewTemplates
	logger.Debug("%s", sa.units)
}

func (wh *WebHandler) RocketHandler() {
	// bind the test file system
	// TODO bind this to the primary config and include fs
	fs := &Diskfs{"./rocket"}
	//fs := &IPfsfs{"QmSeHSnaGkgyKc5a5WiWLuqAAa7socn2DgoSpSmgJPZAy8"}
	RocketACI = fs

	// create the spawn api
	TheSpawn = NewSpawnAPI(fs)

	t, err := template.New("rocket").Parse(MetaDiscovery)
	tmpl = t
	if err != nil {
		logger.Error("template error", err)
	}
	//wh.router.GET("/rocket", wh.Discovery)
	wh.router.GET("/rocket/:name", wh.Discovery)
	wh.router.GET("/images/:source/:rocket/:imageName", wh.AciImage)
	// access for spawn
	wh.router.GET("/spawn/list", wh.UnitList)
	wh.router.GET("/spawn/unit/:name", wh.GetUnit)

}

// BIG TODO
// handle unit file parsing , indexing and serving
// for spawn monster.

func (wh *WebHandler) UnitList(c *gin.Context) {
	// TODO hand list out of current available unit files
	c.IndentedJSON(200, TheSpawn.units)
}

func (wh *WebHandler) GetUnit(c *gin.Context) {
	// TODO hand list out of current available unit files
	UnitName := c.Params.ByName("name")
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

// perform action template
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

var MetaDiscovery = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
      <meta name="ac-discovery" content="{{ .BaseIP }}/rocket/{{ .AciName }} http://{{ .BaseIP }}/images/{name}-{version}-{os}-{arch}.{ext}">
  </head>
<html>
`

//<meta name="ac-discovery-pubkeys" content="example.com/{{ .AciName }} https://example.com/pubkeys.gpg">
