package main

import (
	"github.com/gin-gonic/gin"
	"html/template"
)

var tmpl *template.Template

func (wh *WebHandler) RocketHandler() {
	t, err := template.New("rocket").Parse(MetaDiscovery)
	tmpl = t
	if err != nil {
		logger.Error("template error", err)
	}
	wh.router.GET("/rocket", wh.Discovery)
	wh.router.GET("/rocket/:name", wh.Discovery)
}

// perform action template
func (wh *WebHandler) Discovery(c *gin.Context) {
	logger.Debug(c.Request.RequestURI)
	queryMap := c.Request.URL.Query()
	val, ok := queryMap["ac-discovery"]
	logger.Debug("%s -> %s", ok, val)
	if ok {
		err := tmpl.ExecuteTemplate(c.Writer, "rocket", wh.config)
		if err != nil {
			logger.Error("template error ", err)
		}
		return
	}
	// return actual image here
}

func (wh *WebHandler) ACI(c *gin.Context) {
}

var MetaDiscovery = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
      <meta name="ac-discovery" content="{{ .BaseIP }}/rocket/hello http://{{ .BaseIP }}/rocket/images/{name}-{version}-{os}-{arch}.{ext}">
      <meta name="ac-discovery-pubkeys" content="example.com/hello https://example.com/pubkeys.gpg">
  </head>
<html>
`
