// TFTP server
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"

	tftp "github.com/pin/tftp"
)

var localConf *Config

// HandleWrite : writing is disabled in this service
func HandleWrite(filename string, r *io.PipeReader) {
	r.CloseWithError(fmt.Errorf("server is read only"))
}

// HandleRead : read a ROfs file and send over tftp
func HandleRead(filename string, w *io.PipeWriter) {
	fmt.Printf("Filename : %v \n", []byte(filename))
	var exists bool
	d, err := localConf.fs.Get("tftp/" + filename[0:len(filename)-1])
	defer d.Close()
	fmt.Println(d, err)
	if err == nil {
		exists = true
	}
	if exists {
		// copy all the data into a buffer
		data, err := ioutil.ReadAll(d)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Copy Error : %v\n", err)
		}
		buf := bytes.NewBuffer(data)
		c, e := io.Copy(w, buf)
		d.Close()
		if e != nil {
			fmt.Fprintf(os.Stderr, "Can't send %s: %v\n", filename, e)
		} else {
			fmt.Fprintf(os.Stderr, "Sent %s (%d bytes)\n", filename, c)
		}
		w.Close()
	} else {
		w.CloseWithError(fmt.Errorf("file does not exists: %s", filename))
	}
}

// tftp server
// TODO fix logging
func tftpServer(conf *Config) {
	localConf = conf
	addr, e := net.ResolveUDPAddr("udp", ":69")
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
