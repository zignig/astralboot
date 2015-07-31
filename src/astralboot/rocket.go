// Serve rocket files and systemd units
package main

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"net"
	"strconv"
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

// RocketHandler : add the rocket and spawn parts to the wh router
func (wh *WebHandler) RocketHandler() {
	rocketRef := wh.config.Refs.Rocket
	var fs ROfs
	if wh.config.IPFS == true {
		if rocketRef == "" {
			logger.Debug("Using rkt ref from base boot")
			fs = &IPfsfs{wh.config.Refs.Boot + "/rkt"}
		} else {
			fs = &IPfsfs{rocketRef}
		}
	} else {
		fs = &Diskfs{"./data/rkt"}
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
}

// AciImage : sends athe requested rocket aci file
func (wh *WebHandler) AciImage(c *gin.Context) {
	logger.Debug(c.Request.RequestURI)
	AciName := c.Params.ByName("imageName")
	logger.Debug(AciName)
	fh, size, err := RocketACI.Get(AciName)
	if err != nil {
		logger.Error("Rocket file error : %s", err)
		c.AbortWithStatus(404)
	}
	logger.Notice("Serving ACI : %s", AciName)
	if size > 0 {
		c.Writer.Header().Set("Content-Length", strconv.FormatInt(size, 10))
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
		logger.Notice("Rocket file : %s", t.AciName)
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
