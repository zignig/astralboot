package main

import (
	"fmt"
	"io"
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
	fs        ROfs
}

func NewWebServer(c *Config, l *Store) *WebHandler {
	wh := &WebHandler{}
	// create the router
	wh.router = gin.Default()
	// bind the store
	wh.store = l
	// bind the config
	wh.config = c
	// bind the config to the file store
	wh.fs = c.fs

	wh.router.GET("/ipxe/start", func(c *gin.Context) {
		c.String(200, coreText)
	})

	// templates
	t, err := template.New("list").Parse(OsSelector)
	if err != nil {
		fmt.Println("template error")
		return nil
	}
	wh.templates = t
	// chose and operating system
	wh.router.GET("/choose", wh.Lister)
	// get the boot line for your operating system
	wh.router.GET("/boot/:dist/:mac", wh.Starter)
	// load the kernel and file system
	wh.router.GET("/image/:dist/*path", wh.Images)
	// actions for each distro
	wh.router.GET("/action/:dist/:action", wh.Action)
	// TODO
	// preseed / config
	// post install
	// finalise
	// close
	return wh
}

func (wh *WebHandler) Run() {
	wh.router.Run(":80")
}

// perform action template
func (wh *WebHandler) Action(c *gin.Context) {
	dist := c.Params.ByName("dist")
	action := c.Params.ByName("action")
	logger.Info("Perform %s from %s ", action, dist)
	err := wh.config.OSList[dist].templates.ExecuteTemplate(c.Writer, action, wh.config)
	if err != nil {
		logger.Critical("action fail %s", err)
	}
}

func (wh *WebHandler) Images(c *gin.Context) {
	// hand out base images
	dist := c.Params.ByName("dist")
	path := c.Params.ByName("path")
	logger.Info("Get %s at %s ", dist, path)
	fh, err := wh.fs.Get("boot/" + dist + "/images/" + path)
	defer fh.Close()
	if err != nil {
		fmt.Println("web error ", err)
		return
	}
	c.Writer.WriteHeader(200)
	io.Copy(c.Writer, fh)
	c.Writer.Flush()

}

func (w *WebHandler) Starter(c *gin.Context) {
	dist := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	logger.Info("Starting os for %s on %s", dist, mac)
	macString, err := net.ParseMAC(mac)
	if err != nil {
		fmt.Println("mac update error ", err)
		return
	}
	w.store.UpdateActive(macString, dist)
	logger.Critical("%v", w.config.OSList[dist])
	err = w.config.OSList[dist].templates.ExecuteTemplate(c.Writer, "start", w.config)
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
chain http://{{ $serverIP }}/boot/{{ .Name }}/${net0/mac}
goto top
{{ end }}

`

// testing tempates
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
