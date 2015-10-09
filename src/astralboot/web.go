// Web services for boot and ipxe menus
package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"text/template"

	"github.com/gin-gonic/gin"
)

// WebHandler : construct for web services
type WebHandler struct {
	router    *gin.Engine
	config    *Config
	store     *Store
	templates *template.Template
	fs        ROfs
}

// NewWebServer : create and configure a new web server
func NewWebServer(c *Config, l *Store, level int) *WebHandler {
	wh := &WebHandler{}
	// create the router
	if level == 0 {
		gin.SetMode(gin.ReleaseMode)
		wh.router = gin.New()
	}
	if level == 1 {
		gin.SetMode(gin.ReleaseMode)
		wh.router = gin.Default()
	}
	if level > 1 {
		wh.router = gin.Default()
	}
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
	// rocket handlers
	wh.RocketHandler()
	// chose and operating system
	wh.router.GET("/choose", wh.Lister)
	wh.router.GET("/choose/:dist/:mac", wh.Chooser)
	wh.router.GET("/class/:dist/:mac", wh.ClassChooser)
	wh.router.GET("/setclass/:dist/:class/:mac", wh.ClassSet)
	// get the boot line for your operating system
	wh.router.GET("/boot/:dist/:mac", wh.Starter)
	// load the kernel and file system
	wh.router.GET("/image/:dist/*path", wh.Images)
	// serve the bin folder
	wh.router.GET("/bin/*path", wh.Binaries)
	// actions for each distro
	wh.router.GET("/action/:dist/:action", wh.Action)
	// configs for each distro
	wh.router.GET("/config/:dist/:action", wh.Config)
	if wh.config.Spawn {
		wh.SpawnHandler()
	}
	return wh
}

// Run : run the web server
func (wh *WebHandler) Run() {
	logger.Error("Web Server error %v", wh.router.Run(wh.config.BaseIP.String()+":80"))
}

// TemplateData : template construct
type TemplateData struct {
	Name    string
	IP      net.IP
	BaseIP  net.IP                // the IP of this server
	Cluster map[string]*LeaseList // used for coreos etcd cluster for now
	Config  *Config
	Lease   *Lease
}

// GenTemplateData : generate template data from a mac address
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
		td.Name = fmt.Sprintf("node%d", lease.ID)
	}
	td.IP = lease.GetIP()
	td.BaseIP = wh.config.BaseIP
	td.Cluster = wh.store.DistLease(dist)
	return td
}

// GetIP : just get the client ip
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

// Config : perform config template
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
	logger.Notice("Perform %s from %s on %s", action, dist, td.Lease.Name)
	if td.Lease.Class != "" {
		logger.Info("Class %s", td.Lease.Class)
		action = c.Params.ByName("action") + "-" + td.Lease.Class
	}
	err = wh.config.OSList[dist].templates.ExecuteTemplate(c.Writer, action, td)
	if err != nil {
		logger.Critical("action fail %s", err)
	}
}

// Action : perform template for distro template files
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

//Binaries : hands back binaries in the data bin folder
func (wh *WebHandler) Binaries(c *gin.Context) {
	path := c.Params.ByName("path")
	logger.Info("Get bin %s ", path)
	fh, size, err := wh.fs.Get("bin/" + path)
	defer fh.Close()
	if err != nil {
		logger.Error("web error ", err)
		return
	}
	c.Writer.WriteHeader(200)
	if size > 0 {
		c.Writer.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	}
	io.Copy(c.Writer, fh)
	c.Writer.Flush()
}

// Images : hands back boot images and kernels
func (wh *WebHandler) Images(c *gin.Context) {
	dist := c.Params.ByName("dist")
	path := c.Params.ByName("path")
	logger.Info("Get %s at %s ", dist, path)
	fh, size, err := wh.fs.Get("boot/" + dist + "/images/" + path)
	defer fh.Close()
	if err != nil {
		logger.Error("web error ", err)
		return
	}
	c.Writer.WriteHeader(200)
	if size > 0 {
		c.Writer.Header().Set("Content-Length", strconv.FormatInt(size, 10))
	}
	io.Copy(c.Writer, fh)
	c.Writer.Flush()
}

// Chooser : generates selection menu for os choice
func (wh *WebHandler) Chooser(c *gin.Context) {
	dist := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	logger.Info("Choosing os for %s on %s", dist, mac)
	macString, err := net.ParseMAC(mac)
	if err != nil {
		logger.Error("mac update error %s", err)
		return
	}
	wh.store.UpdateActive(macString, dist)
	logger.Critical("%v", wh.config.OSList[dist])
	err = wh.config.OSList[dist].templates.ExecuteTemplate(c.Writer, "start", wh.config)
}

// ClassChooser : menu for class choice
// first choose a class for the given os
// need to select the sub class from a menu
func (wh *WebHandler) ClassChooser(c *gin.Context) {
	dist := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	logger.Info("Choosing os for %s on %s", dist, mac)
	// add the distro and class to the w.config passed to template
	m := make(map[string]interface{})
	m["config"] = wh.config
	m["dist"] = dist
	m["mac"] = mac
	m["classes"] = wh.config.OSList[dist].Classes
	err := wh.templates.ExecuteTemplate(c.Writer, "class", m)
	if err != nil {
		logger.Error("class template error ", err)
	}
}

// ClassSet : update the lease to have the selected class
func (wh *WebHandler) ClassSet(c *gin.Context) {
	dist := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	class := c.Params.ByName("class")
	// set the class of the lease
	macString, err := net.ParseMAC(mac)
	if err != nil {
		logger.Error("mac update error ", err)
		return
	}
	wh.store.UpdateClass(macString, dist, class)
	wh.config.OSList[dist].templates.ExecuteTemplate(c.Writer, "start", wh.config)
}

// Starter : boot into selected os with template start
func (wh *WebHandler) Starter(c *gin.Context) {
	dist := c.Params.ByName("dist")
	mac := c.Params.ByName("mac")
	logger.Info("Starting os for %s on %s", dist, mac)
	wh.config.OSList[dist].templates.ExecuteTemplate(c.Writer, "start", wh.config)
}

// Lister : select from the os list
func (wh *WebHandler) Lister(c *gin.Context) {
	logger.Critical("Select Machine type")
	err := wh.templates.ExecuteTemplate(c.Writer, "list", wh.config)
	if err != nil {
		logger.Error("template error ", err)
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
