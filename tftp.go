package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	tftp "github.com/pin/tftp"
	"github.com/zignig/cohort/assets"
)

var m map[string][]byte

func HandleWrite(filename string, r *io.PipeReader) {
	_, exists := m[filename]
	if exists {
		r.CloseWithError(fmt.Errorf("File already exists: %s", filename))
		return
	}
	buffer := &bytes.Buffer{}
	c, e := buffer.ReadFrom(r)
	if e != nil {
		fmt.Fprintf(os.Stderr, "Can't receive %s: %v\n", filename, e)
	} else {
		fmt.Fprintf(os.Stderr, "Received %s (%d bytes)\n", filename, c)
		m[filename] = buffer.Bytes()
	}
}

func HandleRead(filename string, w *io.PipeWriter) {
	b, exists := m[filename]
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
	log := log.New(os.Stderr, "", log.Ldate|log.Ltime)
	s := tftp.Server{addr, HandleWrite, HandleRead, log}
	e = s.Serve()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
}
