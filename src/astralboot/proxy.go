// interface for ROfs in ipfs
package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

	"net/http"
	"net/url"
)

const (
	api      = "/api/v0/"
	ipfsHost = "localhost:5001"
)

//IPfsfs : file system ROfs interface
type IPfsfs struct {
	base string
}

//Req : base request for ipfs access
func (fs *IPfsfs) Req(path string, arg string) (resp *http.Response, size int64, err error) {
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
		return nil, 0, err
	}
	size = resp.ContentLength
	if resp.StatusCode != 200 {
		return resp, 0, errors.New(resp.Status)
	}
	if err != nil {
		return resp, 0, err
	}
	return resp, size, err
}

//Stat : Check if the file system exist
func (fs *IPfsfs) Stat() (stat bool) {
	_, _, err := fs.Req("id", "")
	if err != nil {
		logger.Critical("IPFS stat , %s", err)
		return false
	}
	return true
}

//Ls :  Get the file listing ( json blob )
func (fs *IPfsfs) Ls(name string) (data []byte, err error) {
	logger.Debug("get listing for %s", name)
	htr, _, err := fs.Req("ls", "/ipfs/"+fs.base+"/"+name+"/")
	if err != nil {
		return data, err
	}
	data, err = ioutil.ReadAll(htr.Body)
	return data, err
}

//Get : get a file out of ipfs ( ROfs interface )
func (fs *IPfsfs) Get(s string) (f io.ReadCloser, size int64, err error) {
	data, size, err := fs.Req("cat", "/ipfs/"+fs.base+"/"+s)
	return data.Body, size, err
}

// ipfs listing is in json , this is the marsahlling interface

//Item : individual item
type Item struct {
	Name string
	Hash string
	Size int64
}

//List : list of items
type List struct {
	Hash  string
	Links []Item
}

//Listing : wrapper for json struct
type Listing struct {
	Objects []List
}

//List : get directory listing ( ROfs interface )
func (fs *IPfsfs) List(path string) (items []string, err error) {
	resp, err := fs.Ls(path)
	li := &Listing{}
	if err != nil {
		return items, err
	}
	merr := json.Unmarshal(resp, &li)
	if merr != nil {
		logger.Critical("Unmarshall error ", err)
		return items, merr
	}
	for _, it := range li.Objects[0].Links {
		items = append(items, it.Name)
	}
	return items, nil
}
