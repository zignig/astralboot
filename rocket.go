package main

import (
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"net"
)

var tmpl *template.Template

// run the rocket aci out of a local fs for now
var RocketACI ROfs

// rocket template construct
type rktTmpl struct {
	BaseIP  net.IP
	AciName string
}

func (wh *WebHandler) RocketHandler() {
	t, err := template.New("rocket").Parse(MetaDiscovery)
	tmpl = t
	if err != nil {
		logger.Error("template error", err)
	}
	//wh.router.GET("/rocket", wh.Discovery)
	wh.router.GET("/rocket/:name", wh.Discovery)
	wh.router.GET("/images/:source/:rocket/:imageName", wh.AciImage)

	// bind the test file system
	// TODO bind this to the primary config and include fs
	fs := &Diskfs{"./rocket"}
	RocketACI = fs
}

func (wh *WebHandler) AciImage(c *gin.Context) {
	logger.Debug(c.Request.RequestURI)
	AciName := c.Params.ByName("imageName")
	logger.Debug(AciName)
	fh, err := RocketACI.Get(AciName)
	if err != nil {
		logger.Debug("Rocker file error : %s", err)
		c.AbortWithStatus(404)
	}
	io.Copy(c.Writer, fh)
}

// perform action template
func (wh *WebHandler) Discovery(c *gin.Context) {
	logger.Debug(c.Request.RequestURI)
	queryMap := c.Request.URL.Query()
	val, ok := queryMap["ac-discovery"]
	logger.Debug("%s -> %s", ok, val)
	if ok {
		t := rktTmpl{}
		t.BaseIP = wh.config.BaseIP
		t.AciName = c.Params.ByName("name")
		err := tmpl.ExecuteTemplate(c.Writer, "rocket", t)
		if err != nil {
			logger.Error("template error ", err)
		}
		return
	}
}

func (wh *WebHandler) ACI(c *gin.Context) {
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
