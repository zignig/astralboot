// File system abstraction interface
package main

import (
	"io"
	"io/ioutil"
	"os"
)

type localFS struct {
	Base string
}

var sl = string(os.PathSeparator)

//ROfs :  basic interface for readonly file system
type ROfs interface {
	// get a string list of the directory
	List(name string) (names []string, err error)
	// get an io reader of the data
	Get(name string) (f io.ReadCloser, size int64, err error)
	// check if the file system exists
	Stat() (stat bool)
}

//Diskfs : basic disk base file system
type Diskfs struct {
	base string
}

//List : get a directory listing
func (fs *Diskfs) List(name string) (names []string, err error) {
	fi, err := ioutil.ReadDir(fs.base + sl + name)
	names = make([]string, len(fi))
	for i, f := range fi {
		names[i] = f.Name()
	}
	return names, err
}

//Get : gets a file as an io.ReadCloser ( don't forget to close )
func (fs *Diskfs) Get(name string) (f io.ReadCloser, size int64, err error) {
	logger.Debug("FS Get Path : %s", name)
	f, err = os.Open(fs.base + sl + name)
	fi, err := os.Stat(fs.base + sl + name)
	if fi != nil {
		size = fi.Size()
	}
	return f, size, err
}

//Stat : checks to see if file system exists
func (fs *Diskfs) Stat() (stat bool) {
	_, err := os.Stat(fs.base)
	if err != nil {
		logger.Error("FS stat %s", err)
		return false
	}
	return true
}
