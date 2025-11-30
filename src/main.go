package main

import (
	"flag"

	APTGlobal "github.com/Nivesh00/automatic-pod-terminator/src/global"
)

func init() {
	
	// Get flags
	logLevel := flag.String("log-level", "warn", "Log level for the controller. Valid values are [debug, info, warn, error]. Default is warn.")
	flag.Parse()

	// Create logger
	APTGlobal.CreateLogger(logLevel)

	APTGlobal.Logger.Debug("finished initializing program")
}

func main() {

}