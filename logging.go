package main

import (
	"github.com/op/go-logging"
	"os"
)

var logger = logging.MustGetLogger("example")

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} > %{level:.7s} %{id:03x}%{color:reset} %{message}",
)

func LogSetup() {
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.DEBUG, "")
	logging.SetBackend(backend1Formatter)
}
