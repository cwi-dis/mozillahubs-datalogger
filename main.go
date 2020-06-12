package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cwi-dis/mozillahubs-datalogger/server"
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

	server.StartServer(flag.Arg(0), *port)
}
