package main

import (
	"fmt"
	"github.com/zignig/cohort/assets"
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

	f, err := fs.Get("tftp/undionly.kpxe")
	fmt.Println(f, err)
	fmt.Println("---- IPFS system ------")
	var remote_fs ROfs = &IPfsfs{config.Ref, cache}

	names, err = remote_fs.List("boot")
	fmt.Println(names, err)
}
