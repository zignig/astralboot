package main

// interface for ROfs in ipfs
import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"net/http"
	"net/url"
)

const (
	api      = "/api/v0/"
	ipfsHost = "localhost:5001"
	Max      = 600 // investigate byte limit
)

type IPfsfs struct {
	base string
}

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
	logger.Debug(u.String())
	resp, err = http.Get(u.String())
	if resp.StatusCode != 200 {
		return resp, errors.New(resp.Status)
	}
	if err != nil {
		return resp, err
	}
	return resp, err
}

func (fs *IPfsfs) Ls(name string) (data []byte, err error) {
	htr, err := fs.Req("ls", fs.base+"/"+name)
	if err != nil {
		return data, err
	}
	data, err = ioutil.ReadAll(htr.Body)
	return data, err
}

func (fs *IPfsfs) Get(s string) (f io.ReadCloser, err error) {
	data, err := fs.Req("cat", fs.base+"/"+s)
	return data.Body, err
}

// ipfs listing is in json , this is the marsahlling interface

type Item struct {
	Name string
	Hash string
	Size int64
}

type List struct {
	Hash  string
	Links []Item
}

type Listing struct {
	Objects []List
}

func (fs *IPfsfs) List(path string) (items []string, err error) {
	resp, err := fs.Ls(path)
	li := &Listing{}
	if err != nil {
		return items, err
	}
	fmt.Println("start unmarshall")
	merr := json.Unmarshal(resp, &li)
	if merr != nil {
		fmt.Println("Unmarshall error ", err)
		return items, merr
	}
	for _, it := range li.Objects[0].Links {
		//fmt.Println("listing ", i, it)
		items = append(items, it.Name)
	}
	return items, nil
}
