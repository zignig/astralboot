// TFTP server
package main

import (
	"errors"
	"io"
	"os"

	tftp "github.com/zignig/go-tftp/server"
)

var localConf *Config

// HandleWrite : writing is disabled in this service
func HandleWrite(filename string) (w io.Writer, err error) {
	err = errors.New("Server is read only")
	return
}

// HandlewritendleRead : read a ROfs file and send over tftp
func HandleRead(filename string) (r io.Reader, err error) {
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
	s := tftp.NewServer("", HandleRead, HandleWrite, logger)
	e := s.Serve(":69")
	if e != nil {
		logger.Error("tftp error, %s", e)
		os.Exit(1)
	}
}
