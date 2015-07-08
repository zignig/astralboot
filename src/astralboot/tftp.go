// TFTP server
package main

import (
	"bytes"
	"errors"
	tftp "github.com/zignig/go-tftp/server"
	"io"
	"io/ioutil"
	"sync"
)

var localConf *Config

// store the tftp files in ram
var storage map[string][]byte
var storageLock sync.Mutex

// HandleWrite : writing is disabled in this service
func HandleWrite(filename string) (w io.Writer, err error) {
	err = errors.New("Server is read only")
	return
}

func getFile(filename string) (r io.Reader, err error) {
	logger.Debug("tftp get file %s", filename)
	_, ok := storage[filename]
	data := []byte{}
	if !ok {
		storageLock.Lock()
		logger.Notice("tftp cache loading %s", filename)
		r, _, err = localConf.fs.Get("/tftp/" + filename)
		data, err := ioutil.ReadAll(r)
		storage[filename] = data
		if err != nil {
			err = errors.New("Fail")
			return nil, err
		}
		storageLock.Unlock()
		return bytes.NewBuffer(data), err
	}
	data = storage[filename]
	return bytes.NewBuffer(data), err
}

// HandlewritendleRead : read a ROfs file and send over tftp
func HandleRead(filename string) (r io.Reader, err error) {
	r, err = getFile(filename)
	return
}

// tftp server
func tftpServer(conf *Config) {
	localConf = conf
	storage = make(map[string][]byte)
	_, err := getFile("/undionly.kpxe")
	if err != nil {
		logger.Fatal("TFTP preload error : %s", err)
	}
	s := tftp.NewServer("", HandleRead, HandleWrite, logger)
	e := s.Serve(conf.BaseIP.String() + ":69")
	if e != nil {
		logger.Error("tftp error, %s", e)
	}
}
