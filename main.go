package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cwi-dis/mozillahubs-datalogger/server"
)

// main parses the command line arguments which contain the output path as
// positional argument and optionally a port for the HTTP server specified by
// the flag -p
func main() {
	// Ignore SIGHUP so the process isn't killed when the terminal session ends
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Signal(syscall.SIGHUP))

		for {
			<-ch
		}
	}()

	port := flag.Int("p", 6000, "Port to listen on")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("USAGE:", os.Args[0], "[-p port] saveDir")
		os.Exit(1)
	}

	server.StartServer(flag.Arg(0), *port)
}
