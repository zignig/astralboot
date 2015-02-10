package main

// file system abstraction

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

type localFS struct {
	Base string
}

var sl string = string(os.PathSeparator)

func (l localFS) Name() string { return "localFS" }

func (l localFS) Create(name string) (afero.File, error) {
	return os.Create(l.Base + sl + name)
}
func (l localFS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(l.Base+sl+name, perm)
}
func (l localFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(l.Base+sl+path, perm)
}
func (l localFS) Open(name string) (afero.File, error) {
	return os.Open(l.Base + sl + name)
}
func (l localFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return os.OpenFile(l.Base+sl+name, flag, perm)
}
func (l localFS) Remove(name string) error {
	return os.Remove(l.Base + sl + name)
}
func (l localFS) RemoveAll(path string) error {
	return os.RemoveAll(l.Base + sl + path)
}
func (l localFS) Rename(oldname, newname string) error {
	return os.Rename(l.Base+sl+oldname, l.Base+sl+newname)
}
func (l localFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(l.Base + sl + name)
}
func (l localFS) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(l.Base+sl+name, mode)
}
func (l localFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(l.Base+sl+name, atime, mtime)
}
