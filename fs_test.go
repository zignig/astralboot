package main

import (
	"fmt"
	"github.com/zignig/cohort/assets"
	"io/ioutil"
	"testing"
)

func TestTiles(t *testing.T) {
	cache := assets.NewCache()
	config := GetConfig("config.toml", cache)
	config.PrintConfig()
	fmt.Println("---- LOCAL TESTS ------")
	fmt.Println("---- File system ------")
	var fs ROfs = &Diskfs{"./data"}

	names, err := fs.List("boot")
	fmt.Println(names, err)

	f, err := fs.Get("templates/ipxe/menu.txt")
	fmt.Print(f, err)
	data, err := ioutil.ReadAll(f)
	fmt.Println(string(data), err)

	fmt.Println("---- IPFS system ------")
	var remote_fs ROfs = &IPfsfs{config.Ref}

	names, err = remote_fs.List("boot")
	fmt.Println(names, err)
	f, err = remote_fs.Get("templates/ipxe/menu.txt")
	fmt.Print(f, err)
	data, err = ioutil.ReadAll(f)
	fmt.Println(string(data), err)
}
