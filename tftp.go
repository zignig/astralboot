// TFTP server
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	tftp "github.com/whyrusleeping/go-tftp/server"
)

var localConf *Config

// HandleWrite : writing is disabled in this service
func HandleWrite(filename string) (w io.Writer, err error) {
	err = errors.New("Server is write only")
	return
}

// HandlewritendleRead : read a ROfs file and send over tftp
func HandleRead(filename string) (r io.Reader, err error) {
	fmt.Printf("Filename : %v \n", filename)
	r, _, err = localConf.fs.Get("/tftp/" + filename)
	if err != nil {
		err = errors.New("Fail")
	}
	return
}

// tftp server
// TODO fix logging
func tftpServer(conf *Config) {
	localConf = conf
	s := tftp.NewServer("", HandleRead, HandleWrite)
	e := s.Serve(":69")
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
}
