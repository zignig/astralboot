package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"text/template"
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
	wh.router.Static("/boot", "./data/boot/coreos")
	//router.Static("/boot", "./data/boot/debian")

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

func (w *WebHandler) Starter(c *gin.Context) {
}

func (w *WebHandler) Lister(c *gin.Context) {
	fmt.Println("lister")
	fmt.Println(w.config.OSList)
	var j []OS
	j = append(j, OS{"debian", "Debian"})
	j = append(j, OS{"coreos", "CoreOS"})
	for i, v := range w.config.OSList {
		fmt.Println(i, v)
	}
	err := w.templates.ExecuteTemplate(c.Writer, "list", j)
	if err != nil {
		fmt.Println("template error ", err)
	}
}

// basic selector for installing OS
var OsSelector = `#!ipxe

:top
menu Choose and operating sytem {{ range .}}
item {{ .Name }} {{ .Description }}{{ end }}
choose os
goto $os{{ range .}}
:{{ .Name }}
chain http://192.168.2.1/start/{{ .Name }}/${net0/mac}
goto top
{{ end }}

`

var defaultText = `#!ipxe

kernel http://192.168.2.1/boot/linux priority=critical auto=true url=http://192.168.2.1/boot/preseed
initrd http://192.168.2.1/boot/initrd.gz
boot

`
var coreText = `#!ipxe

kernel http://192.168.2.1/boot/coreos_production_pxe.vmlinuz console=tty0 coreos.autologin=tty0 root=/dev/sda1 cloud-config-url=http://192.168.2.1/cloud
initrd http://192.168.2.1/boot/coreos_production_pxe_image.cpio.gz
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
