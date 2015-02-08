package main

// file system abstraction

import (
	"github.com/spf13/afero"
)

var AppFs afero.Fs = &afero.MemMapFs{}
