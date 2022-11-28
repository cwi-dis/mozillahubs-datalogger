package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cwi-dis/mozillahubs-datalogger/server"
	"github.com/cwi-dis/mozillahubs-datalogger/util"
)

// main parses the command line arguments which contain the output path as
// positional argument and optionally a port for the HTTP server specified by
// the flag -p
func main() {
	port := flag.Int("p", 6000, "Port to listen on")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("USAGE:", os.Args[0], "[-p port] saveDir")
		os.Exit(1)
	}

	// Get target directory from command line args
	saveDir := flag.Arg(0)

	// Check if target directory exists and create it if not
	if err := util.CheckAndCreateFolder(saveDir); err != nil {
		fmt.Println("Could not create output directory", saveDir)
		os.Exit(1)
	}

	// Ignore SIGHUP so the process isn't killed when the terminal session ends
	util.IgnoreSighup()

	// Start server
	server.StartServer(saveDir, *port)
}
