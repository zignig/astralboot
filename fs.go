package main

// file system abstraction

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

var AppFs afero.Fs = &afero.MemMapFs{}

type ipfsFS struct{}

func (ipfsFS) Name() string { return "ipfsFS" }
func (ipfsFS) Create(name string) (afero.File, error) {
	return os.Create(name)
}
func (ipfsFS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}
func (ipfsFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
func (ipfsFS) Open(name string) (afero.File, error) {
	return os.Open(name)
}
func (ipfsFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return os.OpenFile(name, flag, perm)
}
func (ipfsFS) Remove(name string) error {
	return os.Remove(name)
}
func (ipfsFS) RemoveAll(path string) error {
	return os.RemoveAll(path)
}
func (ipfsFS) Rename(oldname, newname string) error {
	return os.Rename(oldname, newname)
}
func (ipfsFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (ipfsFS) Chmod(name string, mode os.FileMode) error {
	return os.Chmod(name, mode)
}
func (ipfsFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return os.Chtimes(name, atime, mtime)
}
