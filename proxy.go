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
	//TODO need to parse and return http status
	//logger.Debug(u.String())
	resp, err = http.Get(u.String())
	if resp.StatusCode != 200 {
		return resp, errors.New(resp.Status)
	}
	if err != nil {
		return resp, err
	}
	return resp, err
}

//Ls :  Get the file listing ( json blob )
func (fs *IPfsfs) Ls(name string) (data []byte, err error) {
	htr, err := fs.Req("ls", fs.base+"/"+name)
	if err != nil {
		return data, err
	}
	data, err = ioutil.ReadAll(htr.Body)
	return data, err
}

//Get : get a file out of ipfs ( ROfs interface )
func (fs *IPfsfs) Get(s string) (f io.ReadCloser, size int64, err error) {
	data, err := fs.Req("cat", fs.base+"/"+s)
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
