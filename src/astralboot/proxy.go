// interface for ROfs in ipfs
package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

	"net/http"
	"net/url"
	"sync"
)

const (
	api      = "/api/v0/"
	ipfsHost = "localhost:5001"
)

//IPfsfs : file system ROfs interface
type IPfsfs struct {
	base     string
	sizes    map[string]int
	sizeLock sync.Mutex
}

func init() {
	http.DefaultClient.Transport = &http.Transport{DisableKeepAlives: true}
}

func NewIPfsfs(base string) (fs *IPfsfs) {
	fs = &IPfsfs{base: base}
	fs.sizes = make(map[string]int)
	return fs
}

//Req : base request for ipfs access
func (fs *IPfsfs) Req(path string, arg string) (resp *http.Response, err error) {
	u := url.URL{}
	u.Scheme = "http"
	u.Host = ipfsHost
	u.Path = api + path
	if arg != "" {
		val := url.Values{}
		val.Set("arg", arg)
		val.Set("encoding", "json")
		u.RawQuery = val.Encode()
	}
	logger.Debug("URL : %s", u.String())
	resp, err = http.Get(u.String())
	if resp == nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return resp, errors.New(resp.Status)
	}
	if err != nil {
		return resp, err
	}
	return resp, err
}

//Stat : Check if the file system exist
func (fs *IPfsfs) Stat() (stat bool) {
	_, err := fs.Req("id", "")
	if err != nil {
		logger.Critical("IPFS stat , %s", err)
		return false
	}
	return true
}

//Ls :  Get the file listing ( json blob )
func (fs *IPfsfs) Ls(name string) (data []byte, err error) {
	logger.Debug("get listing for %s", name)
	htr, err := fs.Req("file/ls", "/ipfs/"+fs.base+"/"+name)
	if err != nil {
		return data, err
	}
	data, err = ioutil.ReadAll(htr.Body)
	return data, err
}

//Get : get a file out of ipfs ( ROfs interface )
func (fs *IPfsfs) Get(s string) (f io.ReadCloser, size int64, err error) {
	path := "/ipfs/" + fs.base + "/" + s
	data, err := fs.Req("cat", path)
	// calc size
	size, err = fs.Size(s)
	return data.Body, size, err
}

// ipfs listing is in json , this is the marsahlling interface

//Item : individual item
type Item struct {
	Name string
	Hash string
	Size int64
	Type string
}

//List : list of items
type Object struct {
	Hash  string
	Size  int64
	Type  string
	Links []Item
}

//Listing : wrapper for json struct
type Listing struct {
	Arguments map[string]string
	Objects   map[string]Object
}

func (fs *IPfsfs) fullListing(path string) (l *Listing, err error) {
	resp, err := fs.Ls(path)
	l = &Listing{}
	if err != nil {
		return l, err
	}
	merr := json.Unmarshal(resp, &l)
	if merr != nil {
		logger.Critical("Unmarshall error ", err)
		return l, merr
	}
	return
}

//Size: get file
func (fs *IPfsfs) Size(path string) (size int64, err error) {
	li, err := fs.fullListing(path)
	ref := li.Arguments["/ipfs/"+fs.base+"/"+path]
	size = li.Objects[ref].Size
	return size, err
}

//List : get directory listing ( ROfs interface )
func (fs *IPfsfs) List(path string) (items []string, err error) {
	li, err := fs.fullListing(path)
	ref := li.Arguments["/ipfs/"+fs.base+"/"+path]
	for _, it := range li.Objects[ref].Links {
		items = append(items, it.Name)
	}
	return items, nil
}
