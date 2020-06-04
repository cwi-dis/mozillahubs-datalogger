# Mozilla Hubs Datalogger

This repository contains a simple HTTP server written in Go that logs activity
on Mozilla Hubs to a file. This activity information is submitted via a `POST`
request. The server focuses on simplicity and speed and because of this, it
only contains the bare essentials. To compile the server, run the following
command in the project's root directory (provided you have the Go compiler
installed):

    go build

After that, you can start the server by calling launching the generated
executable:

    ./mozillahubs-datalogger saveDir

Alternatively, you can also run the server directly without compiling it:

    go run server.go saveDir

Where `saveDir` is the name of the directory where the logged data should be
stored. By default, the server listens on port 5000. This can be changed by
passing in the desired port with the flag `-p`:

    ./mozillahubs-datalogger -p 5001 saveDir
