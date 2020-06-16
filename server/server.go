package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"sync"
	"time"

	"github.com/cwi-dis/mozillahubs-datalogger/util"
)

// SuccessResponse is the structure of the data that is returned as JSON upon
// a successfully completed request.
type SuccessResponse struct {
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

// ErrorResponse is the structure of the data that is returned as JSON when the
// request encountered an error.
type ErrorResponse struct {
	Status string `json:"status"`
}

// CreateHandlerWithPath returns a HTTP handler function which reacts to POST
// requests, takes the body data, parses it as JSON and writes the resulting
// data to a file at the path which is given by a parameter to this function.
func CreateHandlerWithPath(saveDir string) func(http.ResponseWriter, *http.Request) {
	// Initialise mutex to guaranteee exclusive access to output file
	var mutex = &sync.Mutex{}

	// Return HTTP handler function
	return func(writer http.ResponseWriter, req *http.Request) {
		// Make sure request is a POST request
		if req.Method == http.MethodPost {
			log.Println("Processing request")
			start := time.Now()

			// Attempt to decode the request body as JSON
			bodyData, err := util.ParseRequestBody(req)

			if err != nil {
				log.Println("Could not decode body data:", err)

				// Send error message with code 400 to client
				msg, _ := json.Marshal(&ErrorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			// Generate filename and save path
			fileName := time.Now().Format("datalog-2006-01-02.csv.gz")
			savePath := path.Join(saveDir, fileName)

			// Lock the mutex for file access
			mutex.Lock()

			// Attempt to write the data to a file
			if err := util.WriteToFile(savePath, bodyData); err != nil {
				// Unlock mutex and handle error
				mutex.Unlock()
				log.Println("Could not save data to file:", err)

				// Send error message with code 400 to client
				msg, _ := json.Marshal(&ErrorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			// Unlock mutex
			mutex.Unlock()

			// Generate response JSON with current timestamp
			msg, _ := json.Marshal(&SuccessResponse{
				Status: "ok",
				Time:   util.GetTimestamp(),
			})

			log.Printf("createHandlerWithPath %.3f", util.ToMSec(time.Now().Sub(start)))
			// Write response to client
			fmt.Fprintf(writer, string(msg))
		} else {
			// Send 404 to client if request is not POST
			http.NotFound(writer, req)
		}
	}
}

// StartServer starts a new HTTP server at the given port which saves data
// received from clients in the given directory under an automatically
// generated name. The server mounts a handler for the route POST /mozillahubs
func StartServer(saveDir string, port int) {
	portSpec := fmt.Sprintf(":%d", port)

	log.Printf("Server listening on port %d", port)

	http.HandleFunc("/mozillahubs", CreateHandlerWithPath(saveDir))
	err := http.ListenAndServe(portSpec, nil)

	if err != nil {
		log.Println(err)
	}
}
