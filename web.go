package main

import (
	"fmt"
	"net"
	"text/template"

	"github.com/gin-gonic/gin"
)

// construct for web services
type WebHandler struct {
	router    *gin.Engine
	config    *Config
	templates *template.Template
	store     *Store
}

func NewWebServer(c *Config, l *Store) *WebHandler {
	wh := &WebHandler{}
	// create the router
	wh.router = gin.Default()
	// bind the store
	wh.store = l
	// bind the config
	wh.config = c

	wh.router.GET("/ipxe/start", func(c *gin.Context) {
		c.String(200, coreText)
	})
	wh.router.GET("/cloud", func(c *gin.Context) {
		c.String(200, cloudConfig)
	})
	wh.router.GET("/boot/*path", wh.Images)

	// templates
	t, err := template.New("list").Parse(OsSelector)
	if err != nil {
		fmt.Println("template error")
		return nil
	}
	wh.templates = t
	wh.router.GET("/choose", wh.Lister)
	wh.router.GET("/start/:dist/:mac", wh.Starter)
	return wh
}

func (wh *WebHandler) Run() {
	wh.router.Run(":80")
}

func (wh *WebHandler) Images(c *gin.Context) {
	// hand out base iamges
	path := c.Params.ByName("path")
	fmt.Println(path)
}

func (w *WebHandler) Starter(c *gin.Context) {
	name := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	fmt.Println("starter call")
	fmt.Println(name, mac)
	c.String(200, defaultText)
	macString, err := net.ParseMAC(mac)
	if err != nil {
		fmt.Println("mac update error ", err)
		return
	}
	w.store.UpdateActive(macString)

}

func (w *WebHandler) Lister(c *gin.Context) {
	err := w.templates.ExecuteTemplate(c.Writer, "list", w.config)
	if err != nil {
		fmt.Println("template error ", err)
	}
}

// basic selector for installing OS
var OsSelector = `#!ipxe

:top{{ $serverIP := .BaseIP }}
menu Choose an operating sytem {{ range .OSList}}
item {{ .Name }} {{ .Description }}{{ end }}
choose os && goto ${os}
{{ range .OSList}}
:{{ .Name }}
chain http://{{ $serverIP }}/start/{{ .Name }}/${net0/mac}
goto top
{{ end }}

`

var defaultText = `#!ipxe

kernel http://192.168.2.1/boot/linux priority=critical auto=true url=http://192.168.2.1/boot/preseed
initrd http://192.168.2.1/boot/initrd.gz
boot

`
var coreText = `#!ipxe

kernel http://192.168.1.1/boot/coreos_production_pxe.vmlinuz console=tty0 coreos.autologin=tty0 root=/dev/sda1 cloud-config-url=http://192.168.1.1/cloud
initrd http://192.168.1.1/boot/coreos_production_pxe_image.cpio.gz
boot

`

var cloudConfig = `#cloud-config
coreos:
  units:
    - name: etcd.service
      command: start
    - name: fleet.service
      command: start
`
