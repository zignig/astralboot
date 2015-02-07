package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	tftp "github.com/pin/tftp"
	"github.com/zignig/cohort/assets"
)

var m map[string][]byte

func HandleWrite(filename string, r *io.PipeReader) {
	r.CloseWithError(fmt.Errorf("Server is Read Only"))
}

func HandleRead(filename string, w *io.PipeWriter) {
	fmt.Printf("Filename : %v \n", []byte(filename))
	for i := range m {
		fmt.Println([]byte(i))
	}
	b, exists := m[filename]
	fmt.Println("exists ", exists)
	if exists {
		buffer := bytes.NewBuffer(b)
		c, e := buffer.WriteTo(w)
		if e != nil {
			fmt.Fprintf(os.Stderr, "Can't send %s: %v\n", filename, e)
		} else {
			fmt.Fprintf(os.Stderr, "Sent %s (%d bytes)\n", filename, c)
		}
		w.Close()
	} else {
		w.CloseWithError(fmt.Errorf("File not exists: %s", filename))
	}
}

func tftpServer(conf *Config, cache *assets.Cache) {
	m = make(map[string][]byte)
	fmt.Println("start tftp")
	addrStr := flag.String("l", ":69", "Address to listen")
	flag.Parse()
	addr, e := net.ResolveUDPAddr("udp", *addrStr)
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		return
	}
	fmt.Printf("load undi")
	d, e := ioutil.ReadFile("data/tftp/undionly.kpxe")
	fmt.Println(e)
	m["undionly.kpxe\x00"] = d
	for i := range m {
		fmt.Println(i)
	}
	log := log.New(os.Stderr, "", log.Ldate|log.Ltime)
	s := tftp.Server{addr, HandleWrite, HandleRead, log}
	e = s.Serve()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
}
