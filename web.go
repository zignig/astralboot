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
	store     *Store
	templates *template.Template
	fs        ROfs
}

func NewWebServer(c *Config, l *Store) *WebHandler {
	wh := &WebHandler{}
	// create the router
	//gin.SetMode(gin.ReleaseMode)
	wh.router = gin.Default()
	// bind the lease db
	wh.store = l
	// bind the config
	wh.config = c
	// bind the config to the file store
	wh.fs = c.fs

	// templates
	// base os selector
	t, err := template.New("list").Parse(OsSelector)
	if err != nil {
		logger.Critical("template error : %v", err)
		return nil
	}
	// class selector
	_, err = t.New("class").Parse(ClassSelector)
	if err != nil {
		logger.Critical("template error : %v", err)
		return nil
	}

	wh.templates = t
	// chose and operating system
	wh.router.GET("/choose", wh.Lister)
	wh.router.GET("/choose/:dist/:mac", wh.Chooser)
	wh.router.GET("/class/:dist/:mac", wh.ClassChooser)
	wh.router.GET("/setclass/:dist/:class/:mac", wh.ClassSet)
	// get the boot line for your operating system
	wh.router.GET("/boot/:dist/:mac", wh.Starter)
	// load the kernel and file system
	wh.router.GET("/image/:dist/*path", wh.Images)
	// actions for each distro
	wh.router.GET("/action/:dist/:action", wh.Action)
	// configs for each distro
	wh.router.GET("/config/:dist/:action", wh.Config)
	// rocket handlers
	wh.RocketHandler()
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

// Data Construct for templating
// includes config , lease
// Adds  IP at top level
type TemplateData struct {
	Name    string
	IP      net.IP
	BaseIP  net.IP                // the IP of this server
	Cluster map[string]*LeaseList // used for coreos etcd cluster for now
	Config  *Config
	Lease   *Lease
}

// generate template data from a mac address
// TODO and keep a cache so it is not regenerated every request
func (wh *WebHandler) GenTemplateData(ip net.IP, dist string) *TemplateData {
	td := &TemplateData{}
	td.Config = wh.config
	lease, err := wh.store.GetFromIP(ip)
	if err != nil {
		logger.Error("Get lease error , %s", err)
	}
	td.Lease = lease
	if lease.Name != "" {
		td.Name = lease.Name
	} else {
		td.Name = fmt.Sprintf("node%d", lease.Id)
	}
	td.IP = lease.GetIP()
	td.BaseIP = wh.config.BaseIP
	td.Cluster = wh.store.DistLease(dist)
	return td
}

// just get the client ip
func GetIP(c *gin.Context) (ip net.IP, err error) {
	tmp := c.ClientIP()
	ipStr, _, err := net.SplitHostPort(tmp)
	ip = net.ParseIP(ipStr)
	if err != nil {
		logger.Error("Client IP fail , %s", err)
		return nil, err
	}
	return ip, nil
}

// perform config template
// config requests  name and appends the device class
// so you can have a template per server class

func (wh *WebHandler) Config(c *gin.Context) {
	dist := c.Params.ByName("dist")
	action := c.Params.ByName("action")
	client, err := GetIP(c)
	if err != nil {
		return
	}
	td := wh.GenTemplateData(client, dist)
	//logger.Info("Template Data : %v", td)
	//logger.Info("Client ip is %s", client)
	logger.Info("Perform %s from %s ", action, dist)
	//	logger.Info("Lease Info ", td.Lease)
	if td.Lease.Class != "" {
		logger.Info("Class %s", td.Lease.Class)
		action = c.Params.ByName("action") + "-" + td.Lease.Class
	}
	err = wh.config.OSList[dist].templates.ExecuteTemplate(c.Writer, action, td)
	if err != nil {
		logger.Critical("action fail %s", err)
	}
}

// perform action template
func (wh *WebHandler) Action(c *gin.Context) {
	dist := c.Params.ByName("dist")
	action := c.Params.ByName("action")
	client, err := GetIP(c)
	if err != nil {
		return
	}
	td := wh.GenTemplateData(client, dist)
	logger.Info("Template Data : %v", td)
	logger.Info("Client ip is %s", client)
	logger.Info("Perform %s from %s ", action, dist)
	err = wh.config.OSList[dist].templates.ExecuteTemplate(c.Writer, action, td)
	if err != nil {
		logger.Critical("action fail %s", err)
	}
}

// hands back boot images and kernels
func (wh *WebHandler) Images(c *gin.Context) {
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

// first boot to choose os
// generates selection menu for os choice
func (w *WebHandler) Chooser(c *gin.Context) {
	dist := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	logger.Info("Choosing os for %s on %s", dist, mac)
	macString, err := net.ParseMAC(mac)
	if err != nil {
		fmt.Println("mac update error ", err)
		return
	}
	w.store.UpdateActive(macString, dist)
	logger.Critical("%v", w.config.OSList[dist])
	err = w.config.OSList[dist].templates.ExecuteTemplate(c.Writer, "start", w.config)
}

// first choose a class for the given os
// need to select the sub class from a menu
func (w *WebHandler) ClassChooser(c *gin.Context) {
	dist := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	logger.Info("Choosing os for %s on %s", dist, mac)
	// add the distro and class to the w.config passed to template
	m := make(map[string]interface{})
	m["config"] = w.config
	m["dist"] = dist
	m["mac"] = mac
	m["classes"] = w.config.OSList[dist].Classes
	err := w.templates.ExecuteTemplate(c.Writer, "class", m)
	if err != nil {
		fmt.Println("class template error ", err)
	}
}

// update the lease to have the selected class
func (w *WebHandler) ClassSet(c *gin.Context) {
	dist := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	class := c.Params.ByName("class")
	// set the class of the lease
	macString, err := net.ParseMAC(mac)
	if err != nil {
		fmt.Println("mac update error ", err)
		return
	}
	w.store.UpdateClass(macString, dist, class)
	w.config.OSList[dist].templates.ExecuteTemplate(c.Writer, "start", w.config)
}

// boot into selected os
func (w *WebHandler) Starter(c *gin.Context) {
	dist := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	logger.Info("Starting os for %s on %s", dist, mac)
	w.config.OSList[dist].templates.ExecuteTemplate(c.Writer, "start", w.config)
}

// select from the os list
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
item {{ .Name }} {{ .Description }} {{ if .HasClasses }}>{{ end }}{{ end }}
choose os && goto ${os}
{{ range .OSList}}
:{{ .Name }}
chain http://{{ $serverIP}}/{{ if .HasClasses }}class{{ else }}choose{{ end }}/{{.Name}}/${net0/mac}
goto top
{{ end }}

`

// class selector for image class
var ClassSelector = `#!ipxe

:top{{ $serverIP := .config.BaseIP }}{{ $dist := .dist }}
menu Choose a systems class from {{ .dist }}{{ range .classes }}
item {{ .}} {{ . }}{{ end }}
choose os && goto ${os}
{{ range .classes }}
:{{ . }}
chain http://{{ $serverIP }}/setclass/{{ $dist }}/{{ . }}/${net0/mac}
goto top
{{ end }}

`
