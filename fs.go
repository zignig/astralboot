package main

// file system abstraction

import (
	"github.com/zignig/cohort/assets"
	"io"
	"io/ioutil"
	"os"
)

type localFS struct {
	Base string
}

var sl string = string(os.PathSeparator)

// basic interface for readonly file system
type ROfs interface {
	// get a string list of the directory
	List(name string) (names []string, err error)
	// get an io reader of the data
	Get(name string) (f io.ReadCloser, err error)
}

// basic disk FS
type Diskfs struct {
	base string
}

func (fs *Diskfs) List(name string) (names []string, err error) {
	fi, err := ioutil.ReadDir(fs.base + sl + name)
	names = make([]string, len(fi))
	for i, f := range fi {
		names[i] = f.Name()
	}
	return names, err
}
func (fs *Diskfs) Get(name string) (f io.ReadCloser, err error) {
	f, err = os.Open(fs.base + sl + name)
	return
}

// basic ipfs FS
type IPfsfs struct {
	// base ipfs reference
	base string
	// cache object for ipfs data
	cache *assets.Cache
}

func (fs *IPfsfs) List(name string) (names []string, err error) {
	names, err = fs.cache.Listing(fs.base + "/" + name)
	return
}

func (fs *IPfsfs) Get(name string) (f io.ReadCloser, err error) {
	return
}
