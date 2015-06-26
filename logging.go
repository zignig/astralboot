// logging creator
package main

import (
	"flag"
	"github.com/op/go-logging"
	"os"
)

var logger = logging.MustGetLogger("example")

var format = logging.MustStringFormatter(
	"%{color}%{time:15:04:05.000} %{shortfunc} > %{level:.7s} %{id:03x}%{color:reset} %{message}",
)

//LogSetup : set up the logging for information output
func LogSetup() {

	logFlag := flag.Int("v", 0, "Set Logging level")
	flag.Parse()
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend1Formatter := logging.NewBackendFormatter(backend1, format)
	backend1Leveled := logging.AddModuleLevel(backend1Formatter)
	switch *logFlag {
	case 0:
		backend1Leveled.SetLevel(logging.CRITICAL, "example")
	case 1:
		backend1Leveled.SetLevel(logging.ERROR, "example")
	case 2:
		backend1Leveled.SetLevel(logging.INFO, "example")
	case 3:
		backend1Leveled.SetLevel(logging.DEBUG, "example")
	}
	logging.SetBackend(backend1Leveled)
}
